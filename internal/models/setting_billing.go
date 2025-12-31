package models

import (
	"time"
)

// SettingBilling represents the setting_billings table
type SettingBilling struct {
	ID           uint       `json:"id" gorm:"primarykey"`
	DocumentID   string     `json:"document_id" gorm:"column:document_id"`
	NamaBilling  string     `json:"nama_billing" gorm:"column:nama_billing"`
	Nominal      float64    `json:"nominal" gorm:"column:nominal"`
	Keterangan   string     `json:"keterangan" gorm:"column:keterangan"`
	JenisBilling string     `json:"jenis_billing" gorm:"column:jenis_billing"`
	IsActive     *bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedByID  *int       `json:"created_by_id"`
	UpdatedByID  *int       `json:"updated_by_id"`
	Locale       *string    `json:"locale"`
}

// TableName sets the insert table name for SettingBilling
func (SettingBilling) TableName() string {
	return "setting_billings"
}
