package models

// BillingPenghuniResponse represents the response structure for billing penghuni list
type BillingPenghuniResponse struct {
	BillingID     uint   `json:"billing_id" example:"10"`                // Billing ID
	DocumentID    string `json:"document_id" example:"abc123def456"`     // User document ID
	Email         string `json:"email" example:"john.doe@example.com"`   // User email address
	ID            uint   `json:"id" example:"1"`                         // User ID
	NamaPenghuni  string `json:"nama_penghuni" example:"John Doe"`       // Resident name
	NoHP          string `json:"no_hp" example:"+6281234567890"`         // Phone number
	NoTelp        string `json:"no_telp" example:"021-12345678"`         // Telephone number
	RoleID        uint   `json:"role_id" example:"5"`                    // Role ID
	RoleName      string `json:"role_name" example:"Penghuni"`           // Role name
	RoleType      string `json:"role_type" example:"penghuni"`           // Role type
	Username      string `json:"username" example:"john_doe"`            // Username
	Nominal       int64  `json:"nominal" example:"500000"`               // Total nominal amount (summed per billing period)
	StatusBilling string `json:"status_billing" example:"Belum Dibayar"` // Billing status
	Bulan         string `json:"bulan" example:"November"`               // Month name
	Tahun         int    `json:"tahun" example:"2025"`                   // Year
	BillingIDs    []uint `json:"billings_id,omitempty" example:"10,11" // Related billing IDs for the user/period`
}
