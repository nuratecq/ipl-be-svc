package response

// PenghuniUserResponse represents penghuni user response data
type PenghuniUserResponse struct {
	ID           uint   `json:"id" example:"1"`
	Username     string `json:"username" example:"john_doe"`
	Email        string `json:"email" example:"john.doe@example.com"`
	NamaPenghuni string `json:"nama_penghuni" example:"John Doe"`
	NoHP         string `json:"no_hp" example:"+6281234567890"`
	NoTelp       string `json:"no_telp" example:"021-12345678"`
	DocumentID   string `json:"document_id" example:"abc123def456"`
	RoleName     string `json:"role_name" example:"Penghuni"`
	RoleID       uint   `json:"role_id" example:"5"`
	RoleType     string `json:"role_type" example:"penghuni"`
}
