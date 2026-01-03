package handler

import (
	"strconv"

	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
	"ipl-be-svc/pkg/utils"

	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	dashboardService service.DashboardService
	logger           *logger.Logger
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(dashboardService service.DashboardService, logger *logger.Logger) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		logger:           logger,
	}
}

// GetDashboardStatistics handles GET /api/v1/dashboard/statistics
// @Summary Get dashboard statistics
// @Description Get dashboard statistics with optional RT, bulan, and tahun filters. If rt=0 or not provided, no RT filter will be applied.
// @Tags dashboard
// @Accept json
// @Produce json
// @Param rt query int false "Filter by RT (optional, if 0 or not provided, no RT filter applied)"
// @Param bulan query int false "Filter by month (1-12)"
// @Param tahun query int false "Filter by year"
// @Success 200 {object} utils.APIResponse "Successfully retrieved dashboard statistics"
// @Failure 400 {object} utils.APIResponse "Bad request - invalid parameter"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/dashboard/statistics [get]
func (h *DashboardHandler) GetDashboardStatistics(c *gin.Context) {
	// Get optional rt parameter
	var rt *int
	rtStr := c.Query("rt")
	if rtStr != "" {
		rtValue, err := strconv.Atoi(rtStr)
		if err != nil {
			h.logger.WithError(err).WithField("rt", rtStr).Error("Invalid RT parameter format")
			utils.BadRequestResponse(c, "Invalid RT parameter format", err)
			return
		}
		rt = &rtValue
	}

	// Get optional bulan parameter
	var bulan *int
	bulanStr := c.Query("bulan")
	if bulanStr != "" {
		bulanValue, err := strconv.Atoi(bulanStr)
		if err != nil {
			h.logger.WithError(err).WithField("bulan", bulanStr).Error("Invalid bulan parameter format")
			utils.BadRequestResponse(c, "Invalid bulan parameter format", err)
			return
		}
		bulan = &bulanValue
	}

	// Get optional tahun parameter
	var tahun *int
	tahunStr := c.Query("tahun")
	if tahunStr != "" {
		tahunValue, err := strconv.Atoi(tahunStr)
		if err != nil {
			h.logger.WithError(err).WithField("tahun", tahunStr).Error("Invalid tahun parameter format")
			utils.BadRequestResponse(c, "Invalid tahun parameter format", err)
			return
		}
		tahun = &tahunValue
	}

	statistics, err := h.dashboardService.GetDashboardStatistics(rt, bulan, tahun)
	if err != nil {
		h.logger.WithError(err).WithField("rt", rt).Error("Failed to get dashboard statistics")
		utils.InternalServerErrorResponse(c, "Failed to retrieve dashboard statistics", err)
		return
	}

	utils.SuccessResponse(c, "Dashboard statistics retrieved successfully", statistics)
}

// GetBillingList handles GET /api/v1/dashboard/billings
// @Summary Get billing list with pagination
// @Description Get list of billings with optional RT, bulan, tahun filters and pagination
// @Tags dashboard
// @Accept json
// @Produce json
// @Param rt query int false "RT (Rukun Tetangga) number - optional, if not provided will return all"
// @Param bulan query int false "Month (1-12) - optional"
// @Param tahun query int false "Year - optional"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} utils.PaginatedResponse "Successfully retrieved billing list"
// @Failure 400 {object} utils.APIResponse "Bad request - invalid parameters"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/dashboard/billings [get]
func (h *DashboardHandler) GetBillingList(c *gin.Context) {
	// Get pagination parameters
	page, limit := utils.GetPaginationParams(c)

	// Get RT parameter (optional)
	var rt *int
	rtStr := c.Query("rt")
	if rtStr != "" {
		rtValue, err := strconv.Atoi(rtStr)
		if err != nil {
			h.logger.WithError(err).WithField("rt", rtStr).Error("Invalid RT parameter format")
			utils.BadRequestResponse(c, "Invalid RT parameter format", err)
			return
		}
		rt = &rtValue
	}

	// Get bulan parameter (optional)
	var bulan *int
	bulanStr := c.Query("bulan")
	if bulanStr != "" {
		bulanValue, err := strconv.Atoi(bulanStr)
		if err != nil {
			h.logger.WithError(err).WithField("bulan", bulanStr).Error("Invalid bulan parameter format")
			utils.BadRequestResponse(c, "Invalid bulan parameter format", err)
			return
		}
		bulan = &bulanValue
	}

	// Get tahun parameter (optional)
	var tahun *int
	tahunStr := c.Query("tahun")
	if tahunStr != "" {
		tahunValue, err := strconv.Atoi(tahunStr)
		if err != nil {
			h.logger.WithError(err).WithField("tahun", tahunStr).Error("Invalid tahun parameter format")
			utils.BadRequestResponse(c, "Invalid tahun parameter format", err)
			return
		}
		tahun = &tahunValue
	}

	// Get billing list
	billings, total, err := h.dashboardService.GetBillingList(rt, bulan, tahun, page, limit)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"rt":    rt,
			"bulan": bulan,
			"tahun": tahun,
			"page":  page,
			"limit": limit,
		}).Error("Failed to get billing list")
		utils.InternalServerErrorResponse(c, "Failed to retrieve billing list", err)
		return
	}

	utils.PaginatedSuccessResponse(c, "Billing list retrieved successfully", billings, page, limit, total)
}
