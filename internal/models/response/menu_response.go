package response

// MenuResponse represents the menu response structure
type MenuResponse struct {
	ID          uint    `json:"id" example:"1"`
	DocumentID  string  `json:"document_id" example:"mo5qqs8ezbruui07t91p6da8"`
	NamaMenu    string  `json:"nama_menu" example:"Master Data"`
	KodeMenu    string  `json:"kode_menu" example:"master-data"`
	UrutanMenu  *int    `json:"urutan_menu" example:"1"`
	IsActive    *bool   `json:"is_active" example:"true"`
	PublishedAt *string `json:"published_at,omitempty" example:"2025-10-23T15:16:28.206Z"`
}
