package domain

import "time"

type FaceVerificationRequest struct {
	CustomerID        string  `json:"customer_id" validate:"required"`
	SourceImageBase64 string  `json:"source_image_base64" validate:"required"`
	TargetImageBase64 string  `json:"target_image_base64" validate:"required"`
	Threshold         float64 `json:"threshold"`
}

type FaceVerificationResponse struct {
	Match           bool      `json:"match"`
	ConfidenceScore float64   `json:"confidence_score"`
	LivenessPassed  bool      `json:"liveness_passed"`
	ProcessingTimeMs int64    `json:"processing_time_ms"`
	Timestamp       time.Time `json:"timestamp"`
}

type FingerprintVerificationRequest struct {
	CustomerID      string `json:"customer_id" validate:"required"`
	TemplateISO     string `json:"template_iso" validate:"required"`
	CapturedTemplate string `json:"captured_template" validate:"required"`
}

type FingerprintVerificationResponse struct {
	Matched          bool      `json:"matched"`
	NFIQQualityScore int       `json:"nfiq_quality_score"`
	Score            float64   `json:"score"`
	Timestamp        time.Time `json:"timestamp"`
}

type BiometricAuditLog struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	CustomerID string    `json:"customer_id" gorm:"index"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Score      float64   `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
}

type BiometricRepository interface {
	SaveAuditLog(log *BiometricAuditLog) error
	GetAuditLog(customerID string) ([]BiometricAuditLog, error)
}

type BiometricUsecase interface {
	VerifyFace(req *FaceVerificationRequest) (*FaceVerificationResponse, error)
	VerifyFingerprint(req *FingerprintVerificationRequest) (*FingerprintVerificationResponse, error)
}
