package repository

import (
	"ipl-be-svc/internal/models"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// BillingRepository defines the interface for billing data operations
type BillingRepository interface {
	GetBillingByID(id uint) (*models.Billing, error)
	GetBillingSettingsByID(id uint) (*models.SettingBilling, error)
	GetUsersWithPenghuniRole() ([]*models.User, error)
	GetActiveMonthlySettingBillings() ([]*models.SettingBilling, error)
	CreateBulkBillings(billings []*models.Billing) error
	CreateBulkBillingProfileLinks(links []*models.BillingProfileLink) error
	GetBillingPenghuni(search string, page int, limit int) ([]*models.BillingPenghuniResponse, int64, error)
	GetBillingPenghuniAll() ([]*models.BillingPenghuniResponse, error)
	// Note: attachment file operations are handled on disk (not persisted to DB)
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

// GetBillingSettingsByID retrieves a billing setting record by ID
func (r *billingRepository) GetBillingSettingsByID(id uint) (*models.SettingBilling, error) {
	var setting models.SettingBilling

	err := r.db.Where("id = ?", id).First(&setting).Error
	if err != nil {
		return nil, err
	}

	return &setting, nil
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

// (no DB-backed attachment methods; file attachments are stored on disk)

// (removed old GetBillingPenghuni - use the paginated version with search)

// GetBillingPenghuni retrieves billing data for penghuni users with pagination and optional search (by nama_penghuni or user id)
func (r *billingRepository) GetBillingPenghuni(search string, page int, limit int) ([]*models.BillingPenghuniResponse, int64, error) {
	var results []*models.BillingPenghuniResponse

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Base query (same as previous implementation)
	baseQuery := `
		SELECT 
			string_agg(DISTINCT b.id::text, ',') as billings_ids,
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
	`

	args := []interface{}{}

	// Add search filter if present
	if strings.TrimSpace(search) != "" {
		// if search is numeric, search by id OR name; otherwise search by name
		if _, err := strconv.Atoi(search); err == nil {
			// numeric search: user id match OR name ILIKE
			baseQuery += " AND (u.id = ? OR p.nama_penghuni ILIKE ?)"
			args = append(args, search, "%"+search+"%")
		} else {
			baseQuery += " AND p.nama_penghuni ILIKE ?"
			args = append(args, "%"+search+"%")
		}
	}

	// GROUP BY and ORDER, then LIMIT/OFFSET
	dataQuery := baseQuery + `
		GROUP BY u.document_id, u.email, u.id, p.nama_penghuni, p.no_hp, p.no_telp, r.id, r.name, r.type, u.username, b.bulan, b.tahun
		ORDER BY u.id, b.tahun DESC, b.bulan DESC
		LIMIT ? OFFSET ?
	`

	// Count total distinct groups (user + month + year) using a lightweight subquery
	countBase := `
		SELECT CONCAT(u.id, '-', COALESCE(b.bulan::text, ''), '-', COALESCE(b.tahun::text, '')) as grp
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
	`

	if strings.TrimSpace(search) != "" {
		if _, err := strconv.Atoi(search); err == nil {
			countBase += " AND (u.id = ? OR p.nama_penghuni ILIKE ?)"
		} else {
			countBase += " AND p.nama_penghuni ILIKE ?"
		}
	}

	countQuery := "SELECT COUNT(*) FROM (" + countBase + ` GROUP BY u.id, b.bulan, b.tahun) as sub`

	var total int64
	countArgs := append([]interface{}{}, args...)
	if err := r.db.Raw(countQuery, countArgs...).Row().Scan(&total); err != nil {
		return nil, 0, err
	}

	// run data query with limit/offset
	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.Raw(dataQuery, queryArgs...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	monthNames := map[int]string{
		1: "January", 2: "February", 3: "March", 4: "April",
		5: "May", 6: "June", 7: "July", 8: "August",
		9: "September", 10: "October", 11: "November", 12: "December",
	}

	for rows.Next() {
		var result models.BillingPenghuniResponse
		var billingsIDsStr *string
		var bulan int

		err := rows.Scan(
			&result.BillingID,
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
			return nil, 0, err
		}

		if monthName, ok := monthNames[bulan]; ok {
			result.Bulan = monthName
		} else {
			result.Bulan = ""
		}

		// parse billings_ids string into slice of uint
		result.BillingIDs = []uint{}
		if billingsIDsStr != nil && *billingsIDsStr != "" {
			parts := strings.Split(*billingsIDsStr, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				if id64, err := strconv.ParseUint(p, 10, 64); err == nil {
					result.BillingIDs = append(result.BillingIDs, uint(id64))
				}
			}
		}

		results = append(results, &result)
	}

	return results, total, nil
}

// GetBillingPenghuniAll retrieves billing data for penghuni users without pagination/search
func (r *billingRepository) GetBillingPenghuniAll() ([]*models.BillingPenghuniResponse, error) {
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
			&result.BillingID,
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

		if monthName, ok := monthNames[bulan]; ok {
			result.Bulan = monthName
		} else {
			result.Bulan = ""
		}

		// billing ids not currently selected for all-list; if need, user can use /penghuni/search
		results = append(results, &result)
	}

	return results, nil
}
