package models

// UserDetail represents the user detail response
type UserDetail struct {
	ID           uint   `json:"id" gorm:"column:id"`
	NamaPenghuni string `json:"nama_penghuni" gorm:"column:nama_penghuni"`
	NoHP         string `json:"no_hp" gorm:"column:no_hp"`
	NoTelp       string `json:"no_telp" gorm:"column:no_telp"`
	DocumentID   string `json:"document_id" gorm:"column:document_id"`
	Email        string `json:"email" gorm:"column:email"`
	UserID       uint   `json:"user_id" gorm:"column:user_id"`
	Username     string `json:"username" gorm:"column:username"`
	RoleName     string `json:"role_name" gorm:"column:name"`
	RoleID       uint   `json:"role_id" gorm:"column:role_id"`
	RoleType     string `json:"role_type" gorm:"column:role_type"`
}

// TableName sets the insert table name for UserDetail
func (UserDetail) TableName() string {
	return "profiles"
}
