package models

// RoleMenuRoleLink represents the role_menus_role_lnk table
type RoleMenuRoleLink struct {
	ID          uint     `json:"id" gorm:"primarykey"`
	RoleMenuID  uint     `json:"role_menu_id" gorm:"column:role_menu_id"`
	RoleID      uint     `json:"role_id" gorm:"column:role_id"`
	RoleMenuOrd *float64 `json:"role_menu_ord" gorm:"column:role_menu_ord"`
}

// TableName sets the insert table name for RoleMenuRoleLink
func (RoleMenuRoleLink) TableName() string {
	return "role_menus_role_lnk"
}
