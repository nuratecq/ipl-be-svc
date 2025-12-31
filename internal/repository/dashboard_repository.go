package repository

import (
	"ipl-be-svc/internal/models/response"

	"gorm.io/gorm"
)

// DashboardRepository defines the interface for dashboard data operations
type DashboardRepository interface {
	GetDashboardStatistics(rt int, bulan, tahun *int) (*response.DashboardStatisticsResponse, error)
	GetBillingList(rt, bulan, tahun *int, page, limit int) ([]*response.BillingListItem, int64, error)
}

// dashboardRepository implements DashboardRepository
type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new instance of DashboardRepository
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}

// GetDashboardStatistics retrieves dashboard statistics by RT with optional bulan and tahun filters
func (r *dashboardRepository) GetDashboardStatistics(rt int, bulan, tahun *int) (*response.DashboardStatisticsResponse, error) {
	var result response.DashboardStatisticsResponse

	query := `
		SELECT
			COUNT(*) FILTER (WHERE bsbl.master_general_status_id = 2) AS belum_bayar,
			COUNT(*) AS total
		FROM billings_profile_id_lnk bpil
		JOIN billings b 
			ON b.id = bpil.t_billing_id
		   AND b.published_at IS NOT NULL
		JOIN up_users_profile_lnk uupl 
			ON uupl.user_id = bpil.user_id
		JOIN profiles p 
			ON p.id = uupl.profile_id
		   AND p.published_at IS NOT NULL
		   AND p.rt = ?
		JOIN billings_status_bill_lnk bsbl 
			ON bsbl.t_billing_id = b.id
	`

	var args []interface{}
	args = append(args, rt)

	// Add bulan filter if provided
	if bulan != nil {
		query += " AND b.bulan = ?"
		args = append(args, *bulan)
	}

	// Add tahun filter if provided
	if tahun != nil {
		query += " AND b.tahun = ?"
		args = append(args, *tahun)
	}

	err := r.db.Raw(query, args...).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBillingList retrieves billing list with optional RT, bulan, tahun filters and pagination
func (r *dashboardRepository) GetBillingList(rt, bulan, tahun *int, page, limit int) ([]*response.BillingListItem, int64, error) {
	var billings []*response.BillingListItem
	var total int64

	// Base query for counting
	countQuery := `
		SELECT COUNT(*)
		FROM billings_profile_id_lnk bpil
		JOIN billings b 
			ON b.id = bpil.t_billing_id
		   AND b.published_at IS NOT NULL
		JOIN up_users_profile_lnk uupl 
			ON uupl.user_id = bpil.user_id
		JOIN profiles p 
			ON p.id = uupl.profile_id
		   AND p.published_at IS NOT NULL
	`

	// Base query for data
	dataQuery := `
		SELECT
			b.nominal, b.bulan, b.tahun, mgs.status_name, p.rt, p.nama_penghuni
		FROM billings_profile_id_lnk bpil
		JOIN billings b 
			ON b.id = bpil.t_billing_id
		   AND b.published_at IS NOT NULL
		JOIN up_users_profile_lnk uupl 
			ON uupl.user_id = bpil.user_id
		JOIN profiles p 
			ON p.id = uupl.profile_id
		   AND p.published_at IS NOT NULL
	`

	// Build args slice for dynamic parameters
	var countArgs []interface{}
	var dataArgs []interface{}

	// Add RT filter if provided
	if rt != nil {
		countQuery += " AND p.rt = ?"
		dataQuery += " AND p.rt = ?"
		countArgs = append(countArgs, *rt)
		dataArgs = append(dataArgs, *rt)
	}

	// Add joins for status
	countQuery += `
		JOIN billings_status_bill_lnk bsbl 
			ON bsbl.t_billing_id = b.id
		JOIN master_general_statuses mgs 
			ON bsbl.master_general_status_id = mgs.id
	`

	dataQuery += `
		JOIN billings_status_bill_lnk bsbl 
			ON bsbl.t_billing_id = b.id
		JOIN master_general_statuses mgs 
			ON bsbl.master_general_status_id = mgs.id
	`

	// Add bulan filter if provided
	if bulan != nil {
		countQuery += " AND b.bulan = ?"
		dataQuery += " AND b.bulan = ?"
		countArgs = append(countArgs, *bulan)
		dataArgs = append(dataArgs, *bulan)
	}

	// Add tahun filter if provided
	if tahun != nil {
		countQuery += " AND b.tahun = ?"
		dataQuery += " AND b.tahun = ?"
		countArgs = append(countArgs, *tahun)
		dataArgs = append(dataArgs, *tahun)
	}

	// Add ORDER BY and pagination to data query
	dataQuery += `
		ORDER BY b.tahun DESC, b.bulan DESC
		LIMIT ? OFFSET ?
	`

	// Calculate offset
	offset := (page - 1) * limit

	// Add pagination params to dataArgs
	dataArgs = append(dataArgs, limit, offset)

	// Execute count query
	err := r.db.Raw(countQuery, countArgs...).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Execute data query
	err = r.db.Raw(dataQuery, dataArgs...).Scan(&billings).Error
	if err != nil {
		return nil, 0, err
	}

	return billings, total, nil
}
