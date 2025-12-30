package models

import (
	"time"
)

// Role represents the up_roles table
type Role struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	DocumentID  string     `json:"document_id" gorm:"column:document_id"`
	Name        string     `json:"name" gorm:"column:name"`
	Description string     `json:"description" gorm:"column:description"`
	Type        string     `json:"type" gorm:"column:type"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedByID *int       `json:"created_by_id"`
	UpdatedByID *int       `json:"updated_by_id"`
	Locale      *string    `json:"locale"`
}

// TableName sets the insert table name for Role
func (Role) TableName() string {
	return "up_roles"
}
