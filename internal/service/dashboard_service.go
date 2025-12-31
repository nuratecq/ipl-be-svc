package service

import (
	"fmt"
	"ipl-be-svc/internal/models/response"
	"ipl-be-svc/internal/repository"
	"ipl-be-svc/pkg/logger"
)

// DashboardService interface defines dashboard service methods
type DashboardService interface {
	GetDashboardStatistics(rt int, bulan, tahun *int) (*response.DashboardStatisticsResponse, error)
	GetBillingList(rt, bulan, tahun *int, page, limit int) ([]*response.BillingListItem, int64, error)
}

// dashboardService implements DashboardService interface
type dashboardService struct {
	dashboardRepo repository.DashboardRepository
	logger        *logger.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(dashboardRepo repository.DashboardRepository, logger *logger.Logger) DashboardService {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
		logger:        logger,
	}
}

// GetDashboardStatistics gets dashboard statistics by RT with optional bulan and tahun filters
func (s *dashboardService) GetDashboardStatistics(rt int, bulan, tahun *int) (*response.DashboardStatisticsResponse, error) {
	if rt <= 0 {
		s.logger.WithField("rt", rt).Error("Invalid RT parameter")
		return nil, fmt.Errorf("invalid RT parameter")
	}

	statistics, err := s.dashboardRepo.GetDashboardStatistics(rt, bulan, tahun)
	if err != nil {
		s.logger.WithError(err).WithField("rt", rt).Error("Failed to get dashboard statistics")
		return nil, err
	}

	logFields := map[string]interface{}{
		"rt":          rt,
		"belum_bayar": statistics.BelumBayar,
		"total":       statistics.Total,
	}
	if bulan != nil {
		logFields["bulan"] = *bulan
	}
	if tahun != nil {
		logFields["tahun"] = *tahun
	}
	s.logger.WithFields(logFields).Info("Dashboard statistics retrieved successfully")

	return statistics, nil
}

// GetBillingList gets billing list with optional RT, bulan, tahun filters and pagination
func (s *dashboardService) GetBillingList(rt, bulan, tahun *int, page, limit int) ([]*response.BillingListItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Validate RT if provided
	if rt != nil && *rt <= 0 {
		s.logger.WithField("rt", *rt).Error("Invalid RT parameter")
		return nil, 0, fmt.Errorf("invalid RT parameter")
	}

	// Validate bulan if provided
	if bulan != nil && (*bulan < 1 || *bulan > 12) {
		s.logger.WithField("bulan", *bulan).Error("Invalid bulan parameter")
		return nil, 0, fmt.Errorf("invalid bulan parameter, must be between 1-12")
	}

	billings, total, err := s.dashboardRepo.GetBillingList(rt, bulan, tahun, page, limit)
	if err != nil {
		s.logger.WithError(err).WithFields(map[string]interface{}{
			"rt":    rt,
			"bulan": bulan,
			"tahun": tahun,
			"page":  page,
			"limit": limit,
		}).Error("Failed to get billing list")
		return nil, 0, err
	}

	logFields := map[string]interface{}{
		"page":  page,
		"limit": limit,
		"total": total,
		"count": len(billings),
	}
	if rt != nil {
		logFields["rt"] = *rt
	}
	if bulan != nil {
		logFields["bulan"] = *bulan
	}
	if tahun != nil {
		logFields["tahun"] = *tahun
	}
	s.logger.WithFields(logFields).Info("Billing list retrieved successfully")

	return billings, total, nil
}
