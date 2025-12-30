package models

// UserRoleLink represents the up_users_role_lnk table
type UserRoleLink struct {
	ID      uint    `json:"id" gorm:"primarykey"`
	UserID  uint    `json:"user_id" gorm:"column:user_id"`
	RoleID  uint    `json:"role_id" gorm:"column:role_id"`
	UserOrd float64 `json:"user_ord" gorm:"column:user_ord"`
}

// TableName sets the insert table name for UserRoleLink
func (UserRoleLink) TableName() string {
	return "up_users_role_lnk"
}
