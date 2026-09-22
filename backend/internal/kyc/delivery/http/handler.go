package http

import (
	"github.com/cbe/kyc-biometric-gateway/internal/kyc/domain"
	"github.com/cbe/kyc-biometric-gateway/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type KYCHandler struct {
	usecase domain.KYCUsecase
}

func NewKYCHandler(app *fiber.App, usecase domain.KYCUsecase) {
	h := &KYCHandler{usecase: usecase}

	api := app.Group("/api/v1/kyc")
	api.Post("/verify-national-id", h.VerifyNationalID)
	api.Post("/document-ocr", h.ProcessDocumentOCR)
	api.Get("/health", h.HealthCheck)
}

func (h *KYCHandler) VerifyNationalID(c *fiber.Ctx) error {
	var req domain.NationalIDVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.NationalID == "" || req.FullName == "" {
		return response.Error(c, fiber.StatusBadRequest, "national_id and full_name are required", nil)
	}

	res, err := h.usecase.VerifyNationalID(&req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "National ID verification failed", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "National ID verified successfully", res)
}

func (h *KYCHandler) ProcessDocumentOCR(c *fiber.Ctx) error {
	var req domain.DocumentOCRRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.DocumentType == "" || req.ImageBase64 == "" {
		return response.Error(c, fiber.StatusBadRequest, "document_type and image_base64 are required", nil)
	}

	res, err := h.usecase.ProcessDocumentOCR(&req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Document OCR extraction failed", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Document OCR processed successfully", res)
}

func (h *KYCHandler) HealthCheck(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, "KYC microservice is healthy", fiber.Map{
		"service": "kyc-service",
		"status":  "UP",
	})
}
