package models

import (
	"time"
)

// User represents the up_users table
type User struct {
	ID                 uint       `json:"id" gorm:"primarykey"`
	DocumentID         string     `json:"document_id" gorm:"column:document_id"`
	Username           string     `json:"username" gorm:"column:username"`
	Email              string     `json:"email" gorm:"column:email"`
	Provider           string     `json:"provider" gorm:"column:provider"`
	Password           string     `json:"password" gorm:"column:password"`
	ResetPasswordToken *string    `json:"reset_password_token" gorm:"column:reset_password_token"`
	ConfirmationToken  *string    `json:"confirmation_token" gorm:"column:confirmation_token"`
	Confirmed          *bool      `json:"confirmed" gorm:"column:confirmed"`
	Blocked            *bool      `json:"blocked" gorm:"column:blocked"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	PublishedAt        *time.Time `json:"published_at"`
	CreatedByID        *int       `json:"created_by_id"`
	UpdatedByID        *int       `json:"updated_by_id"`
	Locale             *string    `json:"locale"`
}

// TableName sets the insert table name for User
func (User) TableName() string {
	return "up_users"
}
