package repository

import (
	"github.com/cbe/kyc-biometric-gateway/internal/kyc/domain"
	"gorm.io/gorm"
)

type kycRepository struct {
	db *gorm.DB
}

func NewKYCRepository(db *gorm.DB) domain.KYCRepository {
	if db != nil {
		_ = db.AutoMigrate(&domain.KYCAuditRecord{})
	}
	return &kycRepository{db: db}
}

func (r *kycRepository) SaveAuditRecord(record *domain.KYCAuditRecord) error {
	if r.db == nil {
		return nil
	}
	return r.db.Create(record).Error
}

func (r *kycRepository) GetAuditRecord(nationalID string) (*domain.KYCAuditRecord, error) {
	var record domain.KYCAuditRecord
	if r.db == nil {
		return &record, nil
	}
	err := r.db.Where("national_id = ?", nationalID).Order("created_at desc").First(&record).Error
	return &record, err
}
