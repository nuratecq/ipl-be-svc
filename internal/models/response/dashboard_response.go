package response

// DashboardStatisticsResponse represents dashboard statistics response
type DashboardStatisticsResponse struct {
	BelumBayar int `json:"belum_bayar" example:"5"`
	SudahBayar int `json:"sudah_bayar" example:"15"`
}

// BillingListItem represents a single billing item in the list
type BillingListItem struct {
	Nominal      float64 `json:"nominal" example:"100000"`
	Bulan        int     `json:"bulan" example:"12"`
	Tahun        int     `json:"tahun" example:"2025"`
	StatusName   string  `json:"status_name" example:"Lunas"`
	RT           int     `json:"rt" example:"8"`
	NamaPenghuni string  `json:"nama_penghuni" example:"John Doe"`
}
