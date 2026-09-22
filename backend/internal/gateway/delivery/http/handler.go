package http

import (
	"fmt"
	"time"

	"github.com/cbe/kyc-biometric-gateway/internal/gateway/usecase"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/config"
	"github.com/cbe/kyc-biometric-gateway/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type GatewayHandler struct {
	cfg     *config.Config
	usecase usecase.GatewayUsecase
}

func NewGatewayHandler(cfg *config.Config, uc usecase.GatewayUsecase) *GatewayHandler {
	return &GatewayHandler{
		cfg:     cfg,
		usecase: uc,
	}
}

func (h *GatewayHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.ClientID == "" || req.ClientSecret == "" {
		return response.Error(c, fiber.StatusBadRequest, "client_id and client_secret required", nil)
	}

	claims := jwt.MapClaims{
		"sub":  req.ClientID,
		"role": "bank_partner",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Token generation failed", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Authentication successful", fiber.Map{
		"access_token": t,
		"token_type":   "Bearer",
		"expires_in":   86400,
	})
}

func (h *GatewayHandler) ProcessFullOnboarding(c *fiber.Ctx) error {
	var req usecase.FullOnboardingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid onboarding request", err.Error())
	}

	res, err := h.usecase.ProcessFullOnboarding(c.Context(), &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Full onboarding orchestration failed", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Full onboarding completed", res)
}

func (h *GatewayHandler) ProxyToKYC(c *fiber.Ctx) error {
	targetPath := c.Params("*")
	targetURL := fmt.Sprintf("%s/api/v1/kyc/%s", h.cfg.KYCServiceURL, targetPath)

	bytes, status, err := h.usecase.ProxyRequest(c.Context(), targetURL, c.Method(), c.Body())
	if err != nil {
		return response.Error(c, status, "KYC microservice unavailable", err.Error())
	}

	c.Set("Content-Type", "application/json")
	return c.Status(status).Send(bytes)
}

func (h *GatewayHandler) ProxyToBiometric(c *fiber.Ctx) error {
	targetPath := c.Params("*")
	targetURL := fmt.Sprintf("%s/api/v1/biometric/%s", h.cfg.BiometricServiceURL, targetPath)

	bytes, status, err := h.usecase.ProxyRequest(c.Context(), targetURL, c.Method(), c.Body())
	if err != nil {
		return response.Error(c, status, "Biometric microservice unavailable", err.Error())
	}

	c.Set("Content-Type", "application/json")
	return c.Status(status).Send(bytes)
}

func (h *GatewayHandler) HealthCheck(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, "API Gateway is healthy", fiber.Map{
		"service": "api-gateway",
		"status":  "UP",
		"time":    time.Now(),
	})
}
