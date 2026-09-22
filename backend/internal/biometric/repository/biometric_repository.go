package repository

import (
	"github.com/cbe/kyc-biometric-gateway/internal/biometric/domain"
	"gorm.io/gorm"
)

type biometricRepository struct {
	db *gorm.DB
}

func NewBiometricRepository(db *gorm.DB) domain.BiometricRepository {
	if db != nil {
		_ = db.AutoMigrate(&domain.BiometricAuditLog{})
	}
	return &biometricRepository{db: db}
}

func (r *biometricRepository) SaveAuditLog(log *domain.BiometricAuditLog) error {
	if r.db == nil {
		return nil
	}
	return r.db.Create(log).Error
}

func (r *biometricRepository) GetAuditLog(customerID string) ([]domain.BiometricAuditLog, error) {
	var logs []domain.BiometricAuditLog
	if r.db == nil {
		return logs, nil
	}
	err := r.db.Where("customer_id = ?", customerID).Order("created_at desc").Find(&logs).Error
	return logs, err
}
