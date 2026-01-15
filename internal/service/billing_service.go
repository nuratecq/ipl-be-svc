package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/models/response"
	"ipl-be-svc/internal/repository"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// BillingService defines the interface for billing business operations
type BillingService interface {
	CreateBulkMonthlyBillings(userIDs []uint, month int, year int) (*BulkBillingResponse, error)
	CreateBulkCustomBillings(userIDs []uint, billingSettingsId int, month int, year int) (*BulkBillingResponse, error)
	CreateBulkMonthlyBillingsForAllUsers(month int, year int) (*BulkBillingResponse, error)
	CreateBulkCustomBillingsForAllUsers(month int, billingSettingsId int, year int) (*BulkBillingResponse, error)
	GetBillingPenghuni(search string, page int, limit int) ([]*models.BillingPenghuniResponse, int64, error)
	ConfirmPayment(listIds []uint) error
	GetBillingPenghuniAll() ([]*models.BillingPenghuniResponse, error)
	GetProfileBillingWithFilters(search string, bulan *int, tahun *int, rt *int, statusID *int, page int, limit int) ([]*response.ProfileBillingResponse, int64, error)
	GetBillingByProfileID(profileID uint, bulan *int, tahun *int, statusID *int, rt *int, page int, limit int) ([]*response.BillingByProfileResponse, int64, error)
	GetBillingStatistics(search string, bulan *int, tahun *int, rt *int, statusIDs []int) (*response.BillingStatisticsResponse, error)
	ExportBillingToExcel(bulan *int, tahun *int, statusID *int, rt *int) ([]byte, string, error)
	// Attachments
	UploadBillingAttachment(billingID uint, filename string, content []byte) (*models.BillingAttachment, error)
	GetBillingAttachments(billingID uint) ([]*models.BillingAttachment, error)
	GetBillingAttachmentByID(id uint) (*models.BillingAttachment, error)
}

// BulkBillingResponse represents the response for bulk billing creation
type BulkBillingResponse struct {
	TotalUsers    int      `json:"total_users"`
	TotalBillings int      `json:"total_billings"`
	SuccessCount  int      `json:"success_count"`
	FailedCount   int      `json:"failed_count"`
	Errors        []string `json:"errors,omitempty"`
}

// billingService implements BillingService
type billingService struct {
	billingRepo repository.BillingRepository
	db          *gorm.DB
}

// NewBillingService creates a new instance of BillingService
func NewBillingService(billingRepo repository.BillingRepository, db *gorm.DB) BillingService {
	return &billingService{
		billingRepo: billingRepo,
		db:          db,
	}
}

// CreateBulkMonthlyBillings creates monthly billings for specified user IDs
func (s *billingService) CreateBulkMonthlyBillings(userIDs []uint, month int, year int) (*BulkBillingResponse, error) {
	// Always use admin user (ID 1) as the creator
	adminID := 1
	createdByInt := &adminID

	// Get default status ("Belum Dibayar")
	var defaultStatus models.MasterGeneralStatus
	if err := s.db.Table("master_general_statuses").Where("status_name = ? AND published_at IS NOT NULL", "Belum Dibayar").First(&defaultStatus).Error; err != nil {
		// If no default status found, get first available status
		if err := s.db.Table("master_general_statuses").Where("published_at IS NOT NULL").First(&defaultStatus).Error; err != nil {
			return nil, fmt.Errorf("failed to get default status: %w", err)
		}
	}

	// Get setting billings
	settings, err := s.billingRepo.GetActiveMonthlySettingBillings()
	if err != nil {
		return nil, fmt.Errorf("failed to get setting billings: %w", err)
	}

	if len(settings) == 0 {
		return nil, fmt.Errorf("no active monthly setting billings found")
	}

	// Get users with profiles
	var users []*models.User
	if len(userIDs) > 0 {
		// Filter specific users
		for _, userID := range userIDs {
			user, err := s.getUserWithProfile(userID)
			if err != nil {
				continue // Skip if user not found or no profile
			}
			users = append(users, user)
		}
	} else {
		// Get all penghuni users
		users, err = s.billingRepo.GetUsersWithPenghuniRoleWithoutBilling(month, year)
		if err != nil {
			return nil, fmt.Errorf("failed to get penghuni users: %w", err)
		}
	}

	if len(users) == 0 {
		return &BulkBillingResponse{
			TotalUsers:    0,
			TotalBillings: 0,
			SuccessCount:  0,
			FailedCount:   0,
		}, nil
	}

	// Prepare billings and links
	var billings []*models.Billing
	var links []*models.BillingProfileLink
	var statusLinks []*models.BillingStatusBillLink
	var kategoriLinks []*models.BillingKategoriTransaksiLink
	now := time.Now()

	for _, user := range users {
		for _, setting := range settings {
			// Skip settings that are not published
			if setting.PublishedAt == nil {
				continue
			}

			// Generate document ID
			docID := "monthly-" + uuid.New().String()

			// Convert nominal from float64 to int64
			nominal := int64(setting.Nominal)

			// Use provided month and year
			billingMonth := month
			billingYear := year

			// Set PublishedAt based on setting's PublishedAt
			var billingPublishedAt *time.Time
			if setting.PublishedAt != nil {
				billingPublishedAt = &now
			} else {
				billingPublishedAt = nil
			}

			// Create billing
			billing := &models.Billing{
				DocumentID:  &docID,
				NamaBilling: &setting.NamaBilling,
				Keterangan:  &setting.Keterangan,
				Bulan:       &billingMonth,
				Tahun:       &billingYear,
				Nominal:     &nominal,
				CreatedAt:   &now,
				UpdatedAt:   &now,
				PublishedAt: billingPublishedAt,
				CreatedByID: createdByInt,
				UpdatedByID: createdByInt,
			}
			billings = append(billings, billing)

			// Create link
			link := &models.BillingProfileLink{
				BillingID: billing.ID, // Will be set after insert
				ProfileID: user.ID,    // Use user ID directly
			}
			links = append(links, link)

			// Create status link
			statusLink := &models.BillingStatusBillLink{
				BillingID:             billing.ID, // Will be set after insert
				MasterGeneralStatusID: defaultStatus.ID,
			}
			statusLinks = append(statusLinks, statusLink)

			// Create kategori transaksi link
			kategoriLink := &models.BillingKategoriTransaksiLink{
				BillingID:                 billing.ID, // Will be set after insert
				MasterKategoriTransaksiID: 1,
			}
			kategoriLinks = append(kategoriLinks, kategoriLink)
		}
	}

	// Execute in transaction
	response := &BulkBillingResponse{
		TotalUsers:    len(users),
		TotalBillings: len(billings),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create billings
		if err := tx.CreateInBatches(billings, 100).Error; err != nil {
			return fmt.Errorf("failed to create billings: %w", err)
		}

		// Update links with billing IDs
		for i, billing := range billings {
			if i < len(links) {
				links[i].BillingID = billing.ID
			}
			if i < len(statusLinks) {
				statusLinks[i].BillingID = billing.ID
			}
			if i < len(kategoriLinks) {
				kategoriLinks[i].BillingID = billing.ID
			}
		}

		// Create profile links
		if err := tx.CreateInBatches(links, 100).Error; err != nil {
			return fmt.Errorf("failed to create billing profile links: %w", err)
		}

		// Create status bill links
		if err := tx.CreateInBatches(statusLinks, 100).Error; err != nil {
			return fmt.Errorf("failed to create billing status bill links: %w", err)
		}

		// Create kategori transaksi links
		if err := tx.CreateInBatches(kategoriLinks, 100).Error; err != nil {
			return fmt.Errorf("failed to create billing kategori transaksi links: %w", err)
		}

		response.SuccessCount = len(billings)
		return nil
	})

	if err != nil {
		response.FailedCount = len(billings)
		response.Errors = []string{err.Error()}
	}

	return response, nil
}

// CreateBulkCustomBillings creates custom billings for specified user IDs
func (s *billingService) CreateBulkCustomBillings(userIDs []uint, billingSettingsId int, month int, year int) (*BulkBillingResponse, error) {
	// Always use admin user (ID 1) as the creator
	adminID := 1
	createdByInt := &adminID

	// Get default status ("Belum Dibayar")
	var defaultStatus models.MasterGeneralStatus
	if err := s.db.Table("master_general_statuses").Where("status_name = ? AND published_at IS NOT NULL", "Belum Dibayar").First(&defaultStatus).Error; err != nil {
		// If no default status found, get first available status
		if err := s.db.Table("master_general_statuses").Where("published_at IS NOT NULL").First(&defaultStatus).Error; err != nil {
			return nil, fmt.Errorf("failed to get default status: %w", err)
		}
	}

	// Get setting billings
	setting, err := s.billingRepo.GetBillingSettingsByID(uint(billingSettingsId))
	if err != nil {
		return nil, fmt.Errorf("failed to get setting billings: %w", err)
	}

	// Get users with profiles
	var users []*models.User
	if len(userIDs) > 0 {
		// Filter specific users
		for _, userID := range userIDs {
			user, err := s.getUserWithProfile(userID)
			if err != nil {
				continue // Skip if user not found or no profile
			}
			users = append(users, user)
		}
	} else {
		// Get all penghuni users
		users, err = s.billingRepo.GetUsersWithPenghuniRoleWithoutBilling(month, year)
		if err != nil {
			return nil, fmt.Errorf("failed to get penghuni users: %w", err)
		}
	}

	if len(users) == 0 {
		return &BulkBillingResponse{
			TotalUsers:    0,
			TotalBillings: 0,
			SuccessCount:  0,
			FailedCount:   0,
		}, nil
	}

	// Prepare billings and links
	var billings []*models.Billing
	var links []*models.BillingProfileLink
	var statusLinks []*models.BillingStatusBillLink
	var kategoriLinks []*models.BillingKategoriTransaksiLink
	now := time.Now()

	for _, user := range users {
		// Skip settings that are not published
		if setting.PublishedAt == nil {
			continue
		}
		// Generate document ID
		docID := "custom-" + uuid.New().String()

		// Convert nominal from float64 to int64
		nominal := setting.Nominal

		// Use provided month and year
		billingMonth := month
		billingYear := year

		// Set PublishedAt based on setting's PublishedAt
		var billingPublishedAt *time.Time
		if setting.PublishedAt != nil {
			billingPublishedAt = &now
		} else {
			billingPublishedAt = nil
		}

		// Create billing
		nominalPtr := int64(nominal)
		billing := &models.Billing{
			DocumentID:  &docID,
			NamaBilling: &setting.NamaBilling,
			Keterangan:  &setting.Keterangan,
			Bulan:       &billingMonth,
			Tahun:       &billingYear,
			Nominal:     &nominalPtr,
			CreatedAt:   &now,
			UpdatedAt:   &now,
			PublishedAt: billingPublishedAt,
			CreatedByID: createdByInt,
			UpdatedByID: createdByInt,
		}
		billings = append(billings, billing)

		// Create link
		link := &models.BillingProfileLink{
			BillingID: billing.ID, // Will be set after insert
			ProfileID: user.ID,    // Use user ID directly
		}
		links = append(links, link)

		// Create status link
		statusLink := &models.BillingStatusBillLink{
			BillingID:             billing.ID, // Will be set after insert
			MasterGeneralStatusID: defaultStatus.ID,
		}
		statusLinks = append(statusLinks, statusLink)

		// Create kategori transaksi link
		kategoriLink := &models.BillingKategoriTransaksiLink{
			BillingID:                 billing.ID, // Will be set after insert
			MasterKategoriTransaksiID: 1,
		}
		kategoriLinks = append(kategoriLinks, kategoriLink)
	}

	// Execute in transaction
	response := &BulkBillingResponse{
		TotalUsers:    len(users),
		TotalBillings: len(billings),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create billings
		if err := tx.CreateInBatches(billings, 100).Error; err != nil {
			return fmt.Errorf("failed to create billings: %w", err)
		}

		// Update links with billing IDs
		for i, billing := range billings {
			if i < len(links) {
				links[i].BillingID = billing.ID
			}
			if i < len(statusLinks) {
				statusLinks[i].BillingID = billing.ID
			}
			if i < len(kategoriLinks) {
				kategoriLinks[i].BillingID = billing.ID
			}
		}

		// Create profile links
		if err := tx.CreateInBatches(links, 100).Error; err != nil {
			return fmt.Errorf("failed to create billing profile links: %w", err)
		}

		// Create status bill links
		if err := tx.CreateInBatches(statusLinks, 100).Error; err != nil {
			return fmt.Errorf("failed to create billing status bill links: %w", err)
		}

		// Create kategori transaksi links
		if err := tx.CreateInBatches(kategoriLinks, 100).Error; err != nil {
			return fmt.Errorf("failed to create billing kategori transaksi links: %w", err)
		}

		response.SuccessCount = len(billings)
		return nil
	})

	if err != nil {
		response.FailedCount = len(billings)
		response.Errors = []string{err.Error()}
	}

	return response, nil
}

// CreateBulkMonthlyBillingsForAllUsers creates monthly billings for all penghuni users
func (s *billingService) CreateBulkMonthlyBillingsForAllUsers(month int, year int) (*BulkBillingResponse, error) {
	return s.CreateBulkMonthlyBillings([]uint{}, month, year)
}

// CreateBulkCustomBillingsForAllUsers creates custom billings for all penghuni users
func (s *billingService) CreateBulkCustomBillingsForAllUsers(month int, billingSettingsId int, year int) (*BulkBillingResponse, error) {
	return s.CreateBulkCustomBillings([]uint{}, billingSettingsId, month, year)
}

// getUserWithProfile gets user with profile information
func (s *billingService) getUserWithProfile(userID uint) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	// Check if user has profile via join
	var count int64
	err = s.db.Table("profiles").
		Joins("JOIN up_users_profile_lnk pul ON profiles.id = pul.profile_id").
		Where("pul.user_id = ?", userID).
		Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("user has no profile")
	}

	return &user, nil
}

// GetBillingPenghuni retrieves all billing data for penghuni users
func (s *billingService) GetBillingPenghuni(search string, page int, limit int) ([]*models.BillingPenghuniResponse, int64, error) {
	return s.billingRepo.GetBillingPenghuni(search, page, limit)
}

func (s *billingService) GetBillingPenghuniAll() ([]*models.BillingPenghuniResponse, error) {
	return s.billingRepo.GetBillingPenghuniAll()
}

func (s *billingService) ConfirmPayment(listIds []uint) error {
	// Run updates in a transaction: mark billings as paid and update status links
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range listIds {
			fmt.Println("id to confirm payment: ", id)

			// Update billing status links: set master_general_status_id = 6 for matching t_billing_id
			if err := tx.Model(&models.BillingStatusBillLink{}).
				Where("t_billing_id = ?", id).
				Updates(map[string]interface{}{
					"master_general_status_id": 6,
				}).Error; err != nil {
				return fmt.Errorf("failed to update billing status links: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// UploadBillingAttachment stores the uploaded file on disk and returns metadata (no DB persistence)
func (s *billingService) UploadBillingAttachment(billingID uint, filename string, content []byte) (*models.BillingAttachment, error) {
	// ensure billing exists
	if _, err := s.billingRepo.GetBillingByID(billingID); err != nil {
		return nil, fmt.Errorf("billing not found: %w", err)
	}

	// storage dir
	dir := fmt.Sprintf("tmp/uploads/billings/%d", billingID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	// unique filename
	fname := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
	path := fmt.Sprintf("%s/%s", dir, fname)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	now := time.Now()
	att := &models.BillingAttachment{
		BillingID: billingID,
		FileName:  filename,
		FilePath:  path,
		CreatedAt: &now,
	}

	// Do NOT persist to DB per request — just return metadata
	return att, nil
}

// GetBillingAttachments lists files on disk for a billing
func (s *billingService) GetBillingAttachments(billingID uint) ([]*models.BillingAttachment, error) {
	dir := fmt.Sprintf("tmp/uploads/billings/%d", billingID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		// if dir doesn't exist, return empty slice
		if os.IsNotExist(err) {
			return []*models.BillingAttachment{}, nil
		}
		return nil, fmt.Errorf("failed to read attachments dir: %w", err)
	}

	var res []*models.BillingAttachment
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		// original filename is after the first underscore in stored name
		stored := e.Name()
		parts := strings.SplitN(stored, "_", 2)
		orig := stored
		if len(parts) == 2 {
			orig = parts[1]
		}
		path := fmt.Sprintf("%s/%s", dir, stored)
		t := info.ModTime()
		res = append(res, &models.BillingAttachment{
			BillingID: billingID,
			FileName:  orig,
			FilePath:  path,
			CreatedAt: &t,
		})
	}

	return res, nil
}

// GetBillingAttachmentByFilename returns metadata for a single file by filename (stored name may have timestamp prefix)
func (s *billingService) GetBillingAttachmentByID(id uint) (*models.BillingAttachment, error) {
	// The previous signature accepted single numeric id (DB id). Since we don't persist, we can't support numeric lookup.
	return nil, fmt.Errorf("numeric lookup not supported: attachments are stored on disk without DB ids")
}

// GetProfileBillingWithFilters retrieves profile billing data with optional filters and supports pagination
func (s *billingService) GetProfileBillingWithFilters(search string, bulan *int, tahun *int, rt *int, statusID *int, page int, limit int) ([]*response.ProfileBillingResponse, int64, error) {
	return s.billingRepo.GetProfileBillingWithFilters(search, bulan, tahun, rt, statusID, page, limit)
}

// GetBillingByProfileID retrieves billing data by profile ID with optional filters and supports pagination
func (s *billingService) GetBillingByProfileID(profileID uint, bulan *int, tahun *int, statusID *int, rt *int, page int, limit int) ([]*response.BillingByProfileResponse, int64, error) {
	return s.billingRepo.GetBillingByProfileID(profileID, bulan, tahun, statusID, rt, page, limit)
}

// GetBillingStatistics retrieves billing statistics with optional filters
func (s *billingService) GetBillingStatistics(search string, bulan *int, tahun *int, rt *int, statusIDs []int) (*response.BillingStatisticsResponse, error) {
	return s.billingRepo.GetBillingStatistics(search, bulan, tahun, rt, statusIDs)
}

// ExportBillingToExcel exports billing data to Excel file
func (s *billingService) ExportBillingToExcel(bulan *int, tahun *int, statusID *int, rt *int) ([]byte, string, error) {
	// Get billing data from repository
	billings, err := s.billingRepo.GetBillingForExport(bulan, tahun, statusID, rt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get billing data: %w", err)
	}

	// Create a new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	sheetName := "Billing Data"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create sheet: %w", err)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Define headers
	headers := []string{"No", "Blok", "RT", "Nama Penghuni", "Nama Pemilik", "Nama Billing", "Bulan", "Tahun", "Nominal", "Status", "Keterangan"}

	// Write headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style for headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err == nil {
		f.SetCellStyle(sheetName, "A1", "K1", headerStyle)
	}

	// Month names mapping
	monthNames := map[int]string{
		1: "Januari", 2: "Februari", 3: "Maret", 4: "April",
		5: "Mei", 6: "Juni", 7: "Juli", 8: "Agustus",
		9: "September", 10: "Oktober", 11: "November", 12: "Desember",
	}

	// Write data
	for i, billing := range billings {
		row := i + 2

		// Get month name
		monthName := fmt.Sprintf("%d", billing.Bulan)
		if name, ok := monthNames[billing.Bulan]; ok {
			monthName = name
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), billing.Blok)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), billing.RT)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), billing.NamaPenghuni)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), billing.NamaPemilik)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), billing.NamaBilling)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), monthName)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), billing.Tahun)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), billing.Nominal)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), billing.StatusName)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), billing.Keterangan)
	}

	// Auto-fit columns
	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheetName, col, col, 15)
	}

	// Delete default Sheet1 if it exists
	if f.GetSheetName(0) == "Sheet1" && sheetName != "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("billing_export_%s.xlsx", timestamp)

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", fmt.Errorf("failed to write Excel file: %w", err)
	}

	return buffer.Bytes(), filename, nil
}
