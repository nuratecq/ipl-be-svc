package models

import (
	"time"
)

// LogSchedullers represents the log_schedullers table
type LogSchedullers struct {
	ID               uint       `json:"id" gorm:"primarykey"`
	DocumentID       *string    `json:"document_id" gorm:"column:document_id"`
	SchedullerCode   *string    `json:"scheduller_code" gorm:"column:scheduller_code"`
	Message          *string    `json:"message" gorm:"column:message"`
	StatusScheduller *string    `json:"status_scheduller" gorm:"column:status_scheduller"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	PublishedAt      *time.Time `json:"published_at"`
	CreatedByID      *int       `json:"created_by_id"`
	UpdatedByID      *int       `json:"updated_by_id"`
	Locale           *string    `json:"locale"`
}

// TableName sets the insert table name for LogSchedullers
func (LogSchedullers) TableName() string {
	return "log_schedullers"
}
