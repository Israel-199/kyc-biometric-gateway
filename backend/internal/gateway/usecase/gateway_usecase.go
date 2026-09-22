package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cbe/kyc-biometric-gateway/internal/platform/circuitbreaker"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/config"
)

type FullOnboardingRequest struct {
	CustomerID        string `json:"customer_id"`
	NationalID        string `json:"national_id"`
	FullName          string `json:"full_name"`
	DateOfBirth       string `json:"date_of_birth"`
	SourceImageBase64 string `json:"source_image_base64"`
	TargetImageBase64 string `json:"target_image_base64"`
}

type FullOnboardingResponse struct {
	OnboardingStatus string      `json:"onboarding_status"`
	KYCResult        interface{} `json:"kyc_result"`
	BiometricResult  interface{} `json:"biometric_result"`
	TotalDurationMs  int64       `json:"total_duration_ms"`
}

type GatewayUsecase interface {
	ProcessFullOnboarding(ctx context.Context, req *FullOnboardingRequest) (*FullOnboardingResponse, error)
	ProxyRequest(ctx context.Context, targetURL string, method string, body []byte) ([]byte, int, error)
}

type gatewayUsecase struct {
	cfg            *config.Config
	httpClient     *http.Client
	circuitBreaker *circuitbreaker.CircuitBreaker
}

func NewGatewayUsecase(cfg *config.Config, cb *circuitbreaker.CircuitBreaker) GatewayUsecase {
	tr := &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	return &gatewayUsecase{
		cfg:            cfg,
		httpClient:     client,
		circuitBreaker: cb,
	}
}

func (u *gatewayUsecase) ProxyRequest(ctx context.Context, targetURL string, method string, body []byte) ([]byte, int, error) {
	respVal, err := u.circuitBreaker.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := u.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("downstream error status %d", resp.StatusCode)
		}

		return struct {
			Bytes      []byte
			StatusCode int
		}{Bytes: respBytes, StatusCode: resp.StatusCode}, nil
	})

	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}

	res := respVal.(struct {
		Bytes      []byte
		StatusCode int
	})

	return res.Bytes, res.StatusCode, nil
}

func (u *gatewayUsecase) ProcessFullOnboarding(ctx context.Context, req *FullOnboardingRequest) (*FullOnboardingResponse, error) {
	start := time.Now()

	kycPayload, _ := json.Marshal(map[string]string{
		"national_id":   req.NationalID,
		"full_name":     req.FullName,
		"date_of_birth": req.DateOfBirth,
	})

	bioPayload, _ := json.Marshal(map[string]interface{}{
		"customer_id":         req.CustomerID,
		"source_image_base64": req.SourceImageBase64,
		"target_image_base64": req.TargetImageBase64,
	})

	type asyncResult struct {
		data interface{}
		err  error
	}

	kycChan := make(chan asyncResult, 1)
	bioChan := make(chan asyncResult, 1)

	go func() {
		targetURL := fmt.Sprintf("%s/api/v1/kyc/verify-national-id", u.cfg.KYCServiceURL)
		b, _, err := u.ProxyRequest(ctx, targetURL, "POST", kycPayload)
		if err != nil {
			kycChan <- asyncResult{err: err}
			return
		}
		var parsed map[string]interface{}
		_ = json.Unmarshal(b, &parsed)
		kycChan <- asyncResult{data: parsed}
	}()

	go func() {
		targetURL := fmt.Sprintf("%s/api/v1/biometric/verify-face", u.cfg.BiometricServiceURL)
		b, _, err := u.ProxyRequest(ctx, targetURL, "POST", bioPayload)
		if err != nil {
			bioChan <- asyncResult{err: err}
			return
		}
		var parsed map[string]interface{}
		_ = json.Unmarshal(b, &parsed)
		bioChan <- asyncResult{data: parsed}
	}()

	kycRes := <-kycChan
	bioRes := <-bioChan

	onboardingStatus := "APPROVED"
	if kycRes.err != nil || bioRes.err != nil {
		onboardingStatus = "FAILED"
	}

	return &FullOnboardingResponse{
		OnboardingStatus: onboardingStatus,
		KYCResult:        kycRes.data,
		BiometricResult:  bioRes.data,
		TotalDurationMs:  time.Since(start).Milliseconds(),
	}, nil
}
