package repository

import (
	"ipl-be-svc/internal/models"

	"gorm.io/gorm"
)

// PaymentConfigRepository defines the interface for payment config data operations
type PaymentConfigRepository interface {
	GetActivePaymentConfig() (*models.PaymentConfig, error)
}

// paymentConfigRepository implements PaymentConfigRepository
type paymentConfigRepository struct {
	db *gorm.DB
}

// NewPaymentConfigRepository creates a new instance of PaymentConfigRepository
func NewPaymentConfigRepository(db *gorm.DB) PaymentConfigRepository {
	return &paymentConfigRepository{
		db: db,
	}
}

// GetActivePaymentConfig retrieves the active payment configuration
func (r *paymentConfigRepository) GetActivePaymentConfig() (*models.PaymentConfig, error) {
	var config models.PaymentConfig

	err := r.db.Where("published_at IS NOT NULL").Order("id DESC").First(&config).Error
	if err != nil {
		return nil, err
	}

	return &config, nil
}
