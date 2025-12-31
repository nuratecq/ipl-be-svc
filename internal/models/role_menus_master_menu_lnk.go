package models

// RoleMenuMasterMenuLink represents the role_menus_master_menu_lnk table
type RoleMenuMasterMenuLink struct {
	ID           uint     `json:"id" gorm:"primarykey"`
	RoleMenuID   uint     `json:"role_menu_id" gorm:"column:role_menu_id"`
	MasterMenuID uint     `json:"master_menu_id" gorm:"column:master_menu_id"`
	RoleMenuOrd  *float64 `json:"role_menu_ord" gorm:"column:role_menu_ord"`
}

// TableName sets the insert table name for RoleMenuMasterMenuLink
func (RoleMenuMasterMenuLink) TableName() string {
	return "role_menus_master_menu_lnk"
}
