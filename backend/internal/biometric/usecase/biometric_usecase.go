package usecase

import (
	"context"
	"math"
	"time"

	"github.com/cbe/kyc-biometric-gateway/internal/biometric/domain"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/metrics"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/workerpool"
	"github.com/cbe/kyc-biometric-gateway/pkg/crypto"
	"github.com/google/uuid"
)

type biometricUsecase struct {
	repo       domain.BiometricRepository
	workerPool *workerpool.WorkerPool
}

func NewBiometricUsecase(repo domain.BiometricRepository, wp *workerpool.WorkerPool) domain.BiometricUsecase {
	return &biometricUsecase{
		repo:       repo,
		workerPool: wp,
	}
}

func (u *biometricUsecase) VerifyFace(req *domain.FaceVerificationRequest) (*domain.FaceVerificationResponse, error) {
	startTime := time.Now()

	hashSource := crypto.HashSHA256([]byte(req.SourceImageBase64))
	hashTarget := crypto.HashSHA256([]byte(req.TargetImageBase64))

	var confidence float64
	if hashSource == hashTarget {
		confidence = 0.998
	} else {
		similarity := calculateSimulatedCosineSimilarity(hashSource, hashTarget)
		confidence = similarity
	}

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.85
	}

	match := confidence >= threshold
	liveness := confidence > 0.70

	processingTime := time.Since(startTime).Milliseconds()

	status := "MATCH"
	if !match {
		status = "MISMATCH"
	}

	metrics.BiometricVerificationsTotal.WithLabelValues("face", status).Inc()

	logEntry := &domain.BiometricAuditLog{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		Type:       "FACE_MATCH",
		Status:     status,
		Score:      confidence,
		CreatedAt:  time.Now(),
	}

	if u.workerPool != nil {
		_ = u.workerPool.Submit(func(ctx context.Context) error {
			return u.repo.SaveAuditLog(logEntry)
		})
	} else {
		_ = u.repo.SaveAuditLog(logEntry)
	}

	return &domain.FaceVerificationResponse{
		Match:            match,
		ConfidenceScore:  math.Round(confidence*10000) / 10000,
		LivenessPassed:   liveness,
		ProcessingTimeMs: processingTime,
		Timestamp:        time.Now(),
	}, nil
}

func (u *biometricUsecase) VerifyFingerprint(req *domain.FingerprintVerificationRequest) (*domain.FingerprintVerificationResponse, error) {
	hashTemplate := crypto.HashSHA256([]byte(req.TemplateISO))
	hashCaptured := crypto.HashSHA256([]byte(req.CapturedTemplate))

	matched := hashTemplate == hashCaptured || calculateSimulatedCosineSimilarity(hashTemplate, hashCaptured) >= 0.88
	score := 0.95
	if !matched {
		score = 0.42
	}

	status := "MATCH"
	if !matched {
		status = "MISMATCH"
	}

	metrics.BiometricVerificationsTotal.WithLabelValues("fingerprint", status).Inc()

	logEntry := &domain.BiometricAuditLog{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		Type:       "FINGERPRINT",
		Status:     status,
		Score:      score,
		CreatedAt:  time.Now(),
	}

	if u.workerPool != nil {
		_ = u.workerPool.Submit(func(ctx context.Context) error {
			return u.repo.SaveAuditLog(logEntry)
		})
	} else {
		_ = u.repo.SaveAuditLog(logEntry)
	}

	return &domain.FingerprintVerificationResponse{
		Matched:          matched,
		NFIQQualityScore: 1,
		Score:            score,
		Timestamp:        time.Now(),
	}, nil
}

func calculateSimulatedCosineSimilarity(h1, h2 string) float64 {
	var diff int
	minLen := len(h1)
	if len(h2) < minLen {
		minLen = len(h2)
	}
	for i := 0; i < minLen; i++ {
		if h1[i] == h2[i] {
			diff++
		}
	}
	return float64(diff)/float64(minLen)*0.4 + 0.6
}
