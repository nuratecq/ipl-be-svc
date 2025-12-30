package models

import (
	"time"
)

// Billing represents the billings table
type Billing struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	DocumentID  *string    `json:"document_id" gorm:"column:document_id"`
	Bulan       *int       `json:"bulan" gorm:"column:bulan"`
	Tahun       *int       `json:"tahun" gorm:"column:tahun"`
	Nominal     *int64     `json:"nominal" gorm:"column:nominal"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedByID *int       `json:"created_by_id"`
	UpdatedByID *int       `json:"updated_by_id"`
	Locale      *string    `json:"locale"`
}

// TableName sets the insert table name for Billing
func (Billing) TableName() string {
	return "billings"
}
