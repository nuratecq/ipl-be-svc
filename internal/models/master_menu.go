package models

import (
	"time"
)

// MasterMenu represents the master_menus table
type MasterMenu struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	DocumentID  string     `json:"document_id" gorm:"column:document_id"`
	NamaMenu    string     `json:"nama_menu" gorm:"column:nama_menu"`
	KodeMenu    string     `json:"kode_menu" gorm:"column:kode_menu"`
	UrutanMenu  *int       `json:"urutan_menu" gorm:"column:urutan_menu"`
	IsActive    *bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at"`
	Locale      *string    `json:"locale"`
}

// TableName sets the insert table name for MasterMenu
func (MasterMenu) TableName() string {
	return "master_menus"
}
