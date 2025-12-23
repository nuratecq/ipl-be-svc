package models

// ProfileUserLink represents the profiles_user_lnk table
type ProfileUserLink struct {
	ID        uint `json:"id" gorm:"primarykey"`
	ProfileID uint `json:"profile_id" gorm:"column:profile_id"`
	UserID    uint `json:"user_id" gorm:"column:user_id"`
}

// TableName sets the insert table name for ProfileUserLink
func (ProfileUserLink) TableName() string {
	return "up_users_profile_lnk"
}
