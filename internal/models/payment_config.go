package models

import (
	"time"
)

// PaymentConfig represents the log_schedullers table
type PaymentConfig struct {
	ID               uint       `json:"id" gorm:"primarykey"`
	PaymentFee       *int64     `json:"payment_fee" gorm:"column:payment_fee"`
	AdminName        *string    `json:"admin_name" gorm:"column:admin_name"`
	AdminEmail       *string    `json:"admin_email" gorm:"column:admin_email"`
	AdminPhone       *string    `json:"admin_phone" gorm:"column:admin_phone"`
	MinMonthDiscount *int       `json:"min_month_discount" gorm:"column:min_month_discount"`
	MaxFee           *int64     `json:"max_fee" gorm:"column:max_fee"`
	IsFixedFee       *bool      `json:"is_fixed_fee" gorm:"column:is_fixed_fee"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	PublishedAt      *time.Time `json:"published_at"`
	CreatedByID      *int       `json:"created_by_id"`
	UpdatedByID      *int       `json:"updated_by_id"`
	Locale           *string    `json:"locale"`
}

// TableName sets the insert table name for PaymentConfig
func (PaymentConfig) TableName() string {
	return "payment_configs"
}
