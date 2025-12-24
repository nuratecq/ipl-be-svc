package service

import (
	"fmt"
	"time"

	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillingService defines the interface for billing business operations
type BillingService interface {
	CreateBulkMonthlyBillings(userIDs []uint, month int, year int) (*BulkBillingResponse, error)
	CreateBulkCustomBillings(userIDs []uint, billingSettingsId int, month int, year int) (*BulkBillingResponse, error)
	CreateBulkMonthlyBillingsForAllUsers(month int, year int) (*BulkBillingResponse, error)
	CreateBulkCustomBillingsForAllUsers(month int, billingSettingsId int, year int) (*BulkBillingResponse, error)
	GetBillingPenghuni() ([]*models.BillingPenghuniResponse, error)
	ConfirmPayment(listIds []uint) error
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
		users, err = s.billingRepo.GetUsersWithPenghuniRole()
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
		users, err = s.billingRepo.GetUsersWithPenghuniRole()
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
func (s *billingService) GetBillingPenghuni() ([]*models.BillingPenghuniResponse, error) {
	return s.billingRepo.GetBillingPenghuni()
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
