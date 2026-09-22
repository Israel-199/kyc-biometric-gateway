package domain

import "time"

type NationalIDVerificationRequest struct {
	NationalID  string `json:"national_id" validate:"required"`
	FullName    string `json:"full_name" validate:"required"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
	Gender      string `json:"gender"`
}

type NationalIDVerificationResponse struct {
	Verified      bool      `json:"verified"`
	NationalID    string    `json:"national_id"`
	FullName      string    `json:"full_name"`
	IssueDate     string    `json:"issue_date"`
	ExpiryDate    string    `json:"expiry_date"`
	Status        string    `json:"status"`
	Tier          string    `json:"tier"`
	AMLPassed     bool      `json:"aml_passed"`
	VerifiedAt    time.Time `json:"verified_at"`
}

type DocumentOCRRequest struct {
	DocumentType string `json:"document_type" validate:"required"`
	ImageBase64  string `json:"image_base64" validate:"required"`
}

type DocumentOCRResponse struct {
	DocumentType string             `json:"document_type"`
	DocumentNo   string             `json:"document_no"`
	ExtractedData map[string]string `json:"extracted_data"`
	Confidence   float64            `json:"confidence"`
	IsTampered   bool               `json:"is_tampered"`
	Timestamp    time.Time          `json:"timestamp"`
}

type KYCAuditRecord struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	NationalID  string    `json:"national_id" gorm:"index"`
	FullName    string    `json:"full_name"`
	DocType     string    `json:"doc_type"`
	VerificationStatus string `json:"verification_status"`
	Tier        string    `json:"tier"`
	AMLStatus   string    `json:"aml_status"`
	CreatedAt   time.Time `json:"created_at"`
}

type KYCRepository interface {
	SaveAuditRecord(record *KYCAuditRecord) error
	GetAuditRecord(nationalID string) (*KYCAuditRecord, error)
}

type KYCUsecase interface {
	VerifyNationalID(req *NationalIDVerificationRequest) (*NationalIDVerificationResponse, error)
	ProcessDocumentOCR(req *DocumentOCRRequest) (*DocumentOCRResponse, error)
}
