package repository

import (
	"ipl-be-svc/internal/models"

	"gorm.io/gorm"
)

// BillingRepository defines the interface for billing data operations
type BillingRepository interface {
	GetBillingByID(id uint) (*models.Billing, error)
	GetUsersWithPenghuniRole() ([]*models.User, error)
	GetActiveMonthlySettingBillings() ([]*models.SettingBilling, error)
	CreateBulkBillings(billings []*models.Billing) error
	CreateBulkBillingProfileLinks(links []*models.BillingProfileLink) error
	GetBillingPenghuni() ([]*models.BillingPenghuniResponse, error)
}

// billingRepository implements BillingRepository
type billingRepository struct {
	db *gorm.DB
}

// NewBillingRepository creates a new instance of BillingRepository
func NewBillingRepository(db *gorm.DB) BillingRepository {
	return &billingRepository{
		db: db,
	}
}

// GetBillingByID retrieves a billing record by ID
func (r *billingRepository) GetBillingByID(id uint) (*models.Billing, error) {
	var billing models.Billing

	err := r.db.Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, err
	}

	return &billing, nil
}

// GetUsersWithPenghuniRole retrieves all users with role type "penghuni"
func (r *billingRepository) GetUsersWithPenghuniRole() ([]*models.User, error) {
	var users []*models.User

	err := r.db.Table("up_users").
		Joins("JOIN up_users_role_lnk url ON up_users.id = url.user_id").
		Joins("JOIN up_roles r ON url.role_id = r.id").
		Where("r.type = ?", "penghuni").
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetActiveMonthlySettingBillings retrieves all active monthly setting billings
func (r *billingRepository) GetActiveMonthlySettingBillings() ([]*models.SettingBilling, error) {
	var settings []*models.SettingBilling

	err := r.db.Where("jenis_billing = ? AND is_active = ? AND published_at IS NOT NULL", "bulanan", true).Find(&settings).Error
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// CreateBulkBillings creates multiple billing records in a transaction
func (r *billingRepository) CreateBulkBillings(billings []*models.Billing) error {
	return r.db.CreateInBatches(billings, 100).Error
}

// CreateBulkBillingProfileLinks creates multiple billing-profile links in a transaction
func (r *billingRepository) CreateBulkBillingProfileLinks(links []*models.BillingProfileLink) error {
	return r.db.CreateInBatches(links, 100).Error
}

// GetBillingPenghuni retrieves all billing data for penghuni users with complete information
func (r *billingRepository) GetBillingPenghuni() ([]*models.BillingPenghuniResponse, error) {
	var results []*models.BillingPenghuniResponse

	monthNames := map[int]string{
		1: "January", 2: "February", 3: "March", 4: "April",
		5: "May", 6: "June", 7: "July", 8: "August",
		9: "September", 10: "October", 11: "November", 12: "December",
	}

	query := `
		SELECT 
			u.document_id,
			u.email,
			u.id,
			p.nama_penghuni,
			COALESCE(p.no_hp, '') as no_hp,
			COALESCE(p.no_telp, '') as no_telp,
			r.id as role_id,
			r.name as role_name,
			r.type as role_type,
			u.username,
			SUM(COALESCE(b.nominal, 0)) as nominal,
			COALESCE(MAX(mgs.status_name), 'Belum Dibayar') as status_billing,
			COALESCE(b.bulan, 0) as bulan,
			COALESCE(b.tahun, 0) as tahun
		FROM up_users u
		INNER JOIN up_users_role_lnk url ON u.id = url.user_id
		INNER JOIN up_roles r ON url.role_id = r.id
		INNER JOIN up_users_profile_lnk pul ON u.id = pul.user_id
		INNER JOIN profiles p ON pul.profile_id = p.id
		LEFT JOIN billings_profile_id_lnk bpl ON u.id = bpl.user_id
		LEFT JOIN billings b ON bpl.t_billing_id = b.id
		LEFT JOIN billings_status_bill_lnk bsbl ON b.id = bsbl.t_billing_id
		LEFT JOIN master_general_statuses mgs ON bsbl.master_general_status_id = mgs.id
		WHERE r.type = 'penghuni'
		AND b.published_at IS NOT NULL
		AND p.published_at IS NOT NULL
		GROUP BY u.document_id, u.email, u.id, p.nama_penghuni, p.no_hp, p.no_telp, r.id, r.name, r.type, u.username, b.bulan, b.tahun
		ORDER BY u.id, b.tahun DESC, b.bulan DESC
	`

	rows, err := r.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var result models.BillingPenghuniResponse
		var bulan int

		err := rows.Scan(
			&result.DocumentID,
			&result.Email,
			&result.ID,
			&result.NamaPenghuni,
			&result.NoHP,
			&result.NoTelp,
			&result.RoleID,
			&result.RoleName,
			&result.RoleType,
			&result.Username,
			&result.Nominal,
			&result.StatusBilling,
			&bulan,
			&result.Tahun,
		)
		if err != nil {
			return nil, err
		}

		// Convert month number to month name
		if monthName, ok := monthNames[bulan]; ok {
			result.Bulan = monthName
		} else {
			result.Bulan = ""
		}

		results = append(results, &result)
	}

	return results, nil
}
