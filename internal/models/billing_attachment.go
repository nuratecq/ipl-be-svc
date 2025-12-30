package models

import "time"

// BillingAttachment stores metadata for files uploaded against a billing
type BillingAttachment struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	BillingID uint       `json:"billing_id" gorm:"column:t_billing_id"`
	FileName  string     `json:"file_name"`
	FilePath  string     `json:"file_path"`
	CreatedAt *time.Time `json:"created_at"`
}

func (BillingAttachment) TableName() string {
	return "billing_attachments"
}
