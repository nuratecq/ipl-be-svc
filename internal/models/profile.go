package models

import (
	"time"
)

// Profile represents the profiles table
type Profile struct {
	ID           uint       `json:"id" gorm:"primarykey"`
	DocumentID   string     `json:"document_id" gorm:"column:document_id"`
	NamaPenghuni string     `json:"nama_penghuni" gorm:"column:nama_penghuni"`
	NoHP         string     `json:"no_hp" gorm:"column:no_hp"`
	NoTelp       string     `json:"no_telp" gorm:"column:no_telp"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedByID  *int       `json:"created_by_id"`
	UpdatedByID  *int       `json:"updated_by_id"`
	Locale       *string    `json:"locale"`
}

// TableName sets the insert table name for Profile
func (Profile) TableName() string {
	return "profiles"
}
