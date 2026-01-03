package response

// DashboardStatisticsResponse represents dashboard statistics response
type DashboardStatisticsResponse struct {
	BelumBayar int `json:"belum_bayar" example:"5"`
	SudahBayar int `json:"sudah_bayar" example:"10"`
	Total      int `json:"total" example:"20"`
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

// ProfileBillingResponse represents profile billing data
type ProfileBillingResponse struct {
	ID           uint   `json:"id" example:"1"`
	NamaPenghuni string `json:"nama_penghuni" example:"John Doe"`
	NamaPemilik  string `json:"nama_pemilik" example:"Jane Doe"`
	Blok         string `json:"blok" example:"A"`
	RT           int    `json:"rt" example:"9"`
}

// BillingByProfileResponse represents billing data by profile ID
type BillingByProfileResponse struct {
	ID          uint   `json:"id" example:"1"`
	ProfileID   uint   `json:"profile_id" example:"654"`
	NamaBilling string `json:"nama_billing" example:"Iuran Bulanan"`
	Bulan       int    `json:"bulan" example:"1"`
	Tahun       int    `json:"tahun" example:"2026"`
	Nominal     int    `json:"nominal" example:"100000"`
	StatusID    uint   `json:"status_id" example:"2"`
	StatusName  string `json:"status_name" example:"Belum Dibayar"`
	Keterangan  string `json:"keterangan" example:"Iuran wajib bulanan"`
}

// BillingStatisticsResponse represents billing statistics data
type BillingStatisticsResponse struct {
	TotalBilling      int64 `json:"total_billing" example:"10"`
	TotalSudahDibayar int64 `json:"total_sudah_dibayar" example:"7"`
	TotalBelumDibayar int64 `json:"total_belum_dibayar" example:"3"`
	TotalNominal      int64 `json:"total_nominal" example:"1000000"`
}
