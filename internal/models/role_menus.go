package models

import (
	"time"
)

// RoleMenu represents the role_menus table
type RoleMenu struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	DocumentID  *string    `json:"document_id" gorm:"column:document_id"`
	RoleMenuOrd *float64   `json:"role_menu_ord" gorm:"column:role_menu_ord"`
	IsActive    *bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedByID *int       `json:"created_by_id"`
	UpdatedByID *int       `json:"updated_by_id"`
	// Relationships
	MasterMenus []MasterMenu `json:"master_menus,omitempty" gorm:"many2many:role_menus_master_menu_lnk;joinForeignKey:role_menu_id;joinReferences:master_menu_id"`
	Roles       []Role       `json:"roles,omitempty" gorm:"many2many:role_menus_role_lnk;joinForeignKey:role_menu_id;joinReferences:role_id"`
}

// TableName sets the insert table name for RoleMenu
func (RoleMenu) TableName() string {
	return "role_menus"
}
