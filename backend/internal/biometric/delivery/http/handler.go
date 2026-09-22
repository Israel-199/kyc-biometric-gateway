package http

import (
	"github.com/cbe/kyc-biometric-gateway/internal/biometric/domain"
	"github.com/cbe/kyc-biometric-gateway/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type BiometricHandler struct {
	usecase domain.BiometricUsecase
}

func NewBiometricHandler(app *fiber.App, usecase domain.BiometricUsecase) {
	h := &BiometricHandler{usecase: usecase}

	api := app.Group("/api/v1/biometric")
	api.Post("/verify-face", h.VerifyFace)
	api.Post("/verify-fingerprint", h.VerifyFingerprint)
	api.Get("/health", h.HealthCheck)
}

func (h *BiometricHandler) VerifyFace(c *fiber.Ctx) error {
	var req domain.FaceVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.CustomerID == "" || req.SourceImageBase64 == "" || req.TargetImageBase64 == "" {
		return response.Error(c, fiber.StatusBadRequest, "customer_id, source_image_base64 and target_image_base64 are required", nil)
	}

	res, err := h.usecase.VerifyFace(&req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Face verification processing failed", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Face verification completed successfully", res)
}

func (h *BiometricHandler) VerifyFingerprint(c *fiber.Ctx) error {
	var req domain.FingerprintVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.CustomerID == "" || req.TemplateISO == "" || req.CapturedTemplate == "" {
		return response.Error(c, fiber.StatusBadRequest, "customer_id, template_iso and captured_template are required", nil)
	}

	res, err := h.usecase.VerifyFingerprint(&req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Fingerprint verification failed", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Fingerprint verification completed successfully", res)
}

func (h *BiometricHandler) HealthCheck(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, "Biometric microservice is healthy", fiber.Map{
		"service": "biometric-service",
		"status":  "UP",
	})
}
