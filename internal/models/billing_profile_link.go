package models

// BillingProfileLink represents the billings_profile_id_lnk table
type BillingProfileLink struct {
	ID        uint `json:"id" gorm:"primarykey"`
	BillingID uint `json:"t_billing_id" gorm:"column:t_billing_id"`
	ProfileID uint `json:"user_id" gorm:"column:user_id"`
}

// TableName sets the insert table name for BillingProfileLink
func (BillingProfileLink) TableName() string {
	return "billings_profile_id_lnk"
}
