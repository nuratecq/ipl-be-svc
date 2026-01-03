package handler

import (
	"fmt"
	"io"
	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
	"ipl-be-svc/pkg/utils"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// BulkBillingRequest represents the request for bulk billing creation
type BulkBillingRequest struct {
	UserIDs []uint `json:"user_ids,omitempty"`                        // Empty means all penghuni users
	Month   int    `json:"month" binding:"required,min=1,max=12"`     // Month 1-12
	Year    int    `json:"year" binding:"required,min=2020,max=2100"` // Reasonable year range
}

// BulkBillingCustomRequest represents the request for bulk billing creation
type BulkBillingCustomRequest struct {
	UserIDs           []uint `json:"user_ids,omitempty"`                        // Empty means all penghuni users
	BillingSettingsId int    `json:"billing_settings_id" binding:"required"`    // Billing settings ID
	Month             int    `json:"month" binding:"required,min=1,max=12"`     // Month 1-12
	Year              int    `json:"year" binding:"required,min=2020,max=2100"` // Reasonable year range
}

// BulkBillingHandler handles bulk billing-related HTTP requests
type BulkBillingHandler struct {
	billingService service.BillingService
	logger         *logger.Logger
}

// NewBulkBillingHandler creates a new BulkBillingHandler instance
func NewBulkBillingHandler(billingService service.BillingService, logger *logger.Logger) *BulkBillingHandler {
	return &BulkBillingHandler{
		billingService: billingService,
		logger:         logger,
	}
}

// CreateBulkMonthlyBillings creates monthly billings for specified users or all penghuni users
// @Summary Create bulk monthly billings
// @Description Create monthly billings for specified user IDs or all penghuni users if user_ids is empty. Requires auth-token cookie.
// @Tags billings
// @Accept json
// @Produce json
// @Param request body BulkBillingRequest true "Bulk billing request with month and year"
// @Success 200 {object} utils.APIResponse{data=service.BulkBillingResponse} "Bulk billing creation result"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/bulk-monthly [post]
func (h *BulkBillingHandler) CreateBulkMonthlyBillings(c *gin.Context) {
	var req BulkBillingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		utils.BadRequestResponse(c, "Request body must be valid JSON", err)
		return
	}

	var response *service.BulkBillingResponse
	var serviceErr error

	if len(req.UserIDs) > 0 {
		// Create for specific users
		response, serviceErr = h.billingService.CreateBulkMonthlyBillings(req.UserIDs, req.Month, req.Year)
	} else {
		// Create for all penghuni users
		response, serviceErr = h.billingService.CreateBulkMonthlyBillingsForAllUsers(req.Month, req.Year)
	}

	if serviceErr != nil {
		h.logger.WithError(serviceErr).Error("Failed to create bulk billings")
		utils.InternalServerErrorResponse(c, "Failed to create billings", serviceErr)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"total_users":    response.TotalUsers,
		"total_billings": response.TotalBillings,
		"success_count":  response.SuccessCount,
		"failed_count":   response.FailedCount,
	}).Info("Bulk billings created successfully")

	utils.SuccessResponse(c, "Bulk billings created successfully", response)
}

// CreateBulkCustomBillings creates custom billings for specified users or all penghuni users
// @Summary Create bulk custom billings
// @Description Create custom billings for specified user IDs or all penghuni users if user_ids is empty. Requires auth-token cookie.
// @Tags billings
// @Accept json
// @Produce json
// @Param request body BulkBillingCustomRequest true "Bulk billing request with month and year"
// @Success 200 {object} utils.APIResponse{data=service.BulkBillingResponse} "Bulk billing creation result"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/bulk-custom [post]
func (h *BulkBillingHandler) CreateBulkCustomBillings(c *gin.Context) {
	var req BulkBillingCustomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		utils.BadRequestResponse(c, "Request body must be valid JSON", err)
		return
	}

	var response *service.BulkBillingResponse
	var serviceErr error

	if len(req.UserIDs) > 0 {
		// Create for specific users
		response, serviceErr = h.billingService.CreateBulkCustomBillings(req.UserIDs, req.BillingSettingsId, req.Month, req.Year)
	} else {
		// Create for all penghuni users
		response, serviceErr = h.billingService.CreateBulkCustomBillingsForAllUsers(req.BillingSettingsId, req.Month, req.Year)
	}

	if serviceErr != nil {
		h.logger.WithError(serviceErr).Error("Failed to create bulk custom billings")
		utils.InternalServerErrorResponse(c, "Failed to create billings", serviceErr)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"total_users":    response.TotalUsers,
		"total_billings": response.TotalBillings,
		"success_count":  response.SuccessCount,
		"failed_count":   response.FailedCount,
	}).Info("Bulk custom billings created successfully")

	utils.SuccessResponse(c, "Bulk custom billings created successfully", response)
}

// GetBillingPenghuniSearch retrieves billing data for penghuni users with pagination and search
// @Summary Get billing penghuni list with summed nominals
// @Description Get billing data for penghuni users. Supports pagination and search by `q` (nama_penghuni or user ID).
// @Tags billings
// @Accept json
// @Produce json
// @Param q query string false "Search by nama_penghuni or ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} utils.PaginatedResponse{data=[]models.BillingPenghuniResponse} "Billing penghuni retrieved successfully"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/penghuni/search [get]
func (h *BulkBillingHandler) GetBillingPenghuniSearch(c *gin.Context) {
	// read query params
	q := c.Query("q")
	page := 1
	perPage := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}

	results, total, err := h.billingService.GetBillingPenghuni(q, page, perPage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing penghuni")
		utils.InternalServerErrorResponse(c, "Failed to get billing penghuni", err)
		return
	}

	h.logger.WithFields(map[string]interface{}{"count": len(results), "total": total, "page": page, "per_page": perPage}).Info("Billing penghuni retrieved successfully")

	utils.PaginatedSuccessResponse(c, "Billing penghuni retrieved successfully", results, page, perPage, total)
}

// GetBillingPenghuni retrieves all billing data for penghuni users (no params)
// @Summary Get all billing penghuni
// @Description Retrieve all billing penghuni records without pagination or search
// @Tags billings
// @Accept json
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]models.BillingPenghuniResponse} "Billing penghuni retrieved successfully"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/penghuni [get]
func (h *BulkBillingHandler) GetBillingPenghuni(c *gin.Context) {
	results, err := h.billingService.GetBillingPenghuniAll()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing penghuni")
		utils.InternalServerErrorResponse(c, "Failed to get billing penghuni", err)
		return
	}

	h.logger.WithField("count", len(results)).Info("Billing penghuni retrieved successfully")

	utils.SuccessResponse(c, "Billing penghuni retrieved successfully", results)
}

// ConfirmPaymentWebhookRequest represents the payload sent by payment gateway webhooks
type ConfirmPaymentWebhookRequest struct {
	Service  map[string]interface{} `json:"service"`
	Acquirer map[string]interface{} `json:"acquirer"`
	Channel  map[string]interface{} `json:"channel"`
	Order    struct {
		InvoiceNumber string `json:"invoice_number" example:"INV-1766570879-"`
		Amount        int64  `json:"amount" example:"175000"`
	} `json:"order"`
	Transaction    map[string]interface{} `json:"transaction"`
	AdditionalInfo map[string]interface{} `json:"additional_info"`
}

// ConfirmPaymentWebhook handles incoming payment gateway webhooks for confirming payments
// @Summary Confirm payment webhook
// @Description Receive payment gateway webhook and process payment confirmation
// @Tags billings
// @Accept json
// @Produce json
// @Param request body ConfirmPaymentWebhookRequest true "Webhook payload"
// @Success 200 {object} utils.APIResponse "Webhook received"
// @Failure 400 {object} utils.APIResponse "Invalid payload"
// @Router /api/v1/billings/confirm-payment [post]
func (h *BulkBillingHandler) ConfirmPaymentWebhook(c *gin.Context) {
	var req ConfirmPaymentWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid webhook payload")
		utils.BadRequestResponse(c, "Invalid webhook payload", err)
		return
	}

	// For now, just log the received webhook and return success.
	// Future: validate signature, map invoice/VA to billing record, update status and create payment record.

	status := ""
	if req.Transaction != nil {
		if s, ok := req.Transaction["status"].(string); ok {
			status = s
		}
	}

	h.logger.WithFields(map[string]interface{}{
		"invoice_number": req.Order.InvoiceNumber,
		"amount":         req.Order.Amount,
		"status":         status,
	}).Info("Received payment webhook")
	fmt.Println("req.Order.InvoiceNumber : ", req.Order.InvoiceNumber)
	// get list id from invoice number
	invoice := strings.Split(req.Order.InvoiceNumber, "-")[2]
	fmt.Println("invoice : ", invoice)
	listId := strings.Split(invoice, ",")
	fmt.Println("listId : ", listId)
	var uintListId []uint
	for _, idStr := range listId {
		var id uint
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			h.logger.WithError(err).Error("Invalid ID in invoice number")
			utils.BadRequestResponse(c, "Invalid ID in invoice number", err)
			return
		}
		uintListId = append(uintListId, id)
	}
	err := h.billingService.ConfirmPayment(uintListId)
	if err != nil {
		h.logger.WithError(err).Error("Failed to confirm payment for billing IDs")
		utils.InternalServerErrorResponse(c, "Failed to confirm payment", err)
		return
	}

	utils.SuccessResponse(c, "Webhook received", nil)
}

// ConfirmPaymentRequest is request body for confirming a single billing
type ConfirmPaymentRequest struct {
	BillingID uint `json:"billing_id" binding:"required" example:"123"`
}

// ConfirmPaymentSingle confirms payment for a single billing ID
// @Summary Confirm single billing payment
// @Description Confirm payment by sending a single billing_id in JSON body
// @Tags billings
// @Accept json
// @Produce json
// @Param request body ConfirmPaymentRequest true "Billing ID"
// @Success 200 {object} utils.APIResponse "Payment confirmed"
// @Failure 400 {object} utils.APIResponse "Invalid payload"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/confirm-single [post]
func (h *BulkBillingHandler) ConfirmPaymentSingle(c *gin.Context) {
	var req ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid confirm payload")
		utils.BadRequestResponse(c, "Invalid payload", err)
		return
	}

	if err := h.billingService.ConfirmPayment([]uint{req.BillingID}); err != nil {
		h.logger.WithError(err).Error("Failed to confirm payment")
		utils.InternalServerErrorResponse(c, "Failed to confirm payment", err)
		return
	}

	utils.SuccessResponse(c, "Payment confirmed", nil)
}

// UploadBillingAttachment handles multipart file upload for a billing record
// @Summary Upload billing attachment
// @Description Upload a file for a billing (multipart form, field `file`)
// @Tags billings
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Billing ID"
// @Param file formData file true "File to upload"
// @Success 200 {object} utils.APIResponse "File uploaded"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/{id}/attachments [post]
func (h *BulkBillingHandler) UploadBillingAttachment(c *gin.Context) {
	idParam := c.Param("id")
	var billingID uint64
	_, err := fmt.Sscanf(idParam, "%d", &billingID)
	if err != nil {
		h.logger.WithError(err).Error("Invalid billing ID param")
		utils.BadRequestResponse(c, "Invalid billing ID", err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.logger.WithError(err).Error("Failed to get file from form")
		utils.BadRequestResponse(c, "File is required", err)
		return
	}

	opened, err := file.Open()
	if err != nil {
		h.logger.WithError(err).Error("Failed to open uploaded file")
		utils.InternalServerErrorResponse(c, "Failed to read file", err)
		return
	}
	defer opened.Close()

	content, err := io.ReadAll(opened)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read file content")
		utils.InternalServerErrorResponse(c, "Failed to read file", err)
		return
	}

	att, err := h.billingService.UploadBillingAttachment(uint(billingID), file.Filename, content)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upload billing attachment")
		utils.InternalServerErrorResponse(c, "Failed to upload file", err)
		return
	}

	utils.SuccessResponse(c, "File uploaded", att)
}

// ListBillingAttachments lists attachments for a billing
// @Summary List billing attachments
// @Description List uploaded attachments for a billing
// @Tags billings
// @Accept json
// @Produce json
// @Param id path int true "Billing ID"
// @Success 200 {object} utils.APIResponse "List of attachments"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/{id}/attachments [get]
func (h *BulkBillingHandler) ListBillingAttachments(c *gin.Context) {
	idParam := c.Param("id")
	var billingID uint64
	_, err := fmt.Sscanf(idParam, "%d", &billingID)
	if err != nil {
		h.logger.WithError(err).Error("Invalid billing ID param")
		utils.BadRequestResponse(c, "Invalid billing ID", err)
		return
	}

	atts, err := h.billingService.GetBillingAttachments(uint(billingID))
	if err != nil {
		h.logger.WithError(err).Error("Failed to list attachments")
		utils.InternalServerErrorResponse(c, "Failed to list attachments", err)
		return
	}

	utils.SuccessResponse(c, "Attachments retrieved", atts)
}

// DownloadBillingAttachment streams the file for a given attachment id
// @Summary Download billing attachment
// @Description Download attachment by id
// @Tags billings
// @Accept json
// @Produce octet-stream
// @Param id path int true "Billing ID"
// @Param attachment_id path int true "Attachment ID"
// @Success 200 {file} file "The file"
// @Failure 404 {object} utils.APIResponse "Not found"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/{id}/attachments/{attachment_id} [get]
func (h *BulkBillingHandler) DownloadBillingAttachment(c *gin.Context) {
	idParam := c.Param("id")
	var billingID uint64
	_, err := fmt.Sscanf(idParam, "%d", &billingID)
	if err != nil {
		h.logger.WithError(err).Error("Invalid billing ID param")
		utils.BadRequestResponse(c, "Invalid billing ID", err)
		return
	}

	// Here attachment_id is the stored filename (URL-encoded). We will serve that file from disk.
	attachmentName := c.Param("attachment_id")
	dir := fmt.Sprintf("tmp/uploads/billings/%d", billingID)
	path := fmt.Sprintf("%s/%s", dir, attachmentName)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			utils.NotFoundResponse(c, "Attachment not found")
			return
		}
		h.logger.WithError(err).Error("Failed to stat attachment file")
		utils.InternalServerErrorResponse(c, "Failed to open attachment", err)
		return
	}

	// try to infer original filename (after underscore)
	orig := attachmentName
	if parts := strings.SplitN(attachmentName, "_", 2); len(parts) == 2 {
		orig = parts[1]
	}

	c.FileAttachment(path, orig)
}

// GetProfileBillingWithFilters retrieves profile billing data with optional filters
// @Summary Get profile billing with optional filters
// @Description Get profile billing data (id, nama_penghuni, nama_pemilik, blok, rt) with optional filters for search, bulan, tahun, rt, and status_id. Search parameter will filter by nama_penghuni or nama_pemilik using LIKE. Requires auth-token cookie.
// @Tags billings
// @Accept json
// @Produce json
// @Param search query string false "Search by nama_penghuni or nama_pemilik (case-insensitive LIKE)"
// @Param bulan query int false "Filter by month (1-12)"
// @Param tahun query int false "Filter by year"
// @Param rt query int false "Filter by RT"
// @Param status_id query int false "Filter by status ID"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/billings/profile [get]
func (h *BulkBillingHandler) GetProfileBillingWithFilters(c *gin.Context) {
	var bulan, tahun, rt, statusID *int
	var search string

	// Parse optional query parameters
	search = c.Query("search")

	if bulanStr := c.Query("bulan"); bulanStr != "" {
		if val, err := strconv.Atoi(bulanStr); err == nil {
			bulan = &val
		} else {
			utils.BadRequestResponse(c, "Invalid bulan parameter", err)
			return
		}
	}

	if tahunStr := c.Query("tahun"); tahunStr != "" {
		if val, err := strconv.Atoi(tahunStr); err == nil {
			tahun = &val
		} else {
			utils.BadRequestResponse(c, "Invalid tahun parameter", err)
			return
		}
	}

	if rtStr := c.Query("rt"); rtStr != "" {
		if val, err := strconv.Atoi(rtStr); err == nil {
			rt = &val
		} else {
			utils.BadRequestResponse(c, "Invalid rt parameter", err)
			return
		}
	}

	if statusIDStr := c.Query("status_id"); statusIDStr != "" {
		if val, err := strconv.Atoi(statusIDStr); err == nil {
			statusID = &val
		} else {
			utils.BadRequestResponse(c, "Invalid status_id parameter", err)
			return
		}
	}

	// Call service to get data
	results, err := h.billingService.GetProfileBillingWithFilters(search, bulan, tahun, rt, statusID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get profile billing with filters")
		utils.InternalServerErrorResponse(c, "Failed to retrieve profile billing data", err)
		return
	}

	utils.SuccessResponse(c, "Profile billing data retrieved successfully", results)
}

// GetBillingByProfileID retrieves billing data by profile ID with optional filters
// @Summary Get billing by profile ID with optional filters
// @Description Get billing data (id, profile_id, nama_billing, bulan, tahun, status_id, status_name, keterangan) by profile ID with optional filters for bulan, tahun, status_id, and rt. Profile ID is required. Requires auth-token cookie.
// @Tags billings
// @Accept json
// @Produce json
// @Param profile_id query int true "Profile ID (required)"
// @Param bulan query int false "Filter by month (1-12)"
// @Param tahun query int false "Filter by year"
// @Param status_id query int false "Filter by status ID"
// @Param rt query int false "Filter by RT"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/billings/by-profile [get]
func (h *BulkBillingHandler) GetBillingByProfileID(c *gin.Context) {
	// Profile ID is required
	profileIDStr := c.Query("profile_id")
	if profileIDStr == "" {
		utils.BadRequestResponse(c, "profile_id parameter is required", nil)
		return
	}

	profileID, err := strconv.ParseUint(profileIDStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid profile_id parameter", nil)
		return
	}

	var bulan, tahun, statusID, rt *int

	// Parse optional query parameters
	if bulanStr := c.Query("bulan"); bulanStr != "" {
		if val, err := strconv.Atoi(bulanStr); err == nil {
			bulan = &val
		} else {
			utils.BadRequestResponse(c, "Invalid bulan parameter", nil)
			return
		}
	}

	if tahunStr := c.Query("tahun"); tahunStr != "" {
		if val, err := strconv.Atoi(tahunStr); err == nil {
			tahun = &val
		} else {
			utils.BadRequestResponse(c, "Invalid tahun parameter", nil)
			return
		}
	}

	if statusIDStr := c.Query("status_id"); statusIDStr != "" {
		if val, err := strconv.Atoi(statusIDStr); err == nil {
			statusID = &val
		} else {
			utils.BadRequestResponse(c, "Invalid status_id parameter", nil)
			return
		}
	}

	if rtStr := c.Query("rt"); rtStr != "" {
		if val, err := strconv.Atoi(rtStr); err == nil {
			rt = &val
		} else {
			utils.BadRequestResponse(c, "Invalid rt parameter", nil)
			return
		}
	}

	// Call service to get data
	results, err := h.billingService.GetBillingByProfileID(uint(profileID), bulan, tahun, statusID, rt)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing by profile ID")
		utils.InternalServerErrorResponse(c, "Failed to retrieve billing data", err)
		return
	}

	utils.SuccessResponse(c, "Billing data retrieved successfully", results)
}

// GetBillingStatistics retrieves billing statistics with optional filters
// @Summary Get billing statistics with optional filters
// @Description Get billing statistics (total_billing, total_sudah_dibayar, total_belum_dibayar, total_nominal) with optional filters for search, bulan, tahun, rt, and status_ids. Search parameter will filter by nama_penghuni or nama_pemilik using LIKE. Status_ids parameter accepts comma-separated values, if not provided defaults to status IDs 2 and 6. Requires auth-token cookie.
// @Tags billings
// @Accept json
// @Produce json
// @Param search query string false "Search by nama_penghuni or nama_pemilik (case-insensitive LIKE)"
// @Param bulan query int false "Filter by month (1-12)"
// @Param tahun query int false "Filter by year"
// @Param rt query int false "Filter by RT"
// @Param status_ids query string false "Filter by status IDs (comma-separated, e.g. '2,6,7')"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/billings/statistics [get]
func (h *BulkBillingHandler) GetBillingStatistics(c *gin.Context) {
	var bulan, tahun, rt *int
	var statusIDs []int
	var search string

	// Parse optional query parameters
	search = c.Query("search")

	if bulanStr := c.Query("bulan"); bulanStr != "" {
		if val, err := strconv.Atoi(bulanStr); err == nil {
			bulan = &val
		} else {
			utils.BadRequestResponse(c, "Invalid bulan parameter", nil)
			return
		}
	}

	if tahunStr := c.Query("tahun"); tahunStr != "" {
		if val, err := strconv.Atoi(tahunStr); err == nil {
			tahun = &val
		} else {
			utils.BadRequestResponse(c, "Invalid tahun parameter", nil)
			return
		}
	}

	if rtStr := c.Query("rt"); rtStr != "" {
		if val, err := strconv.Atoi(rtStr); err == nil {
			rt = &val
		} else {
			utils.BadRequestResponse(c, "Invalid rt parameter", nil)
			return
		}
	}

	// Parse status_ids parameter (comma-separated)
	if statusIDsStr := c.Query("status_ids"); statusIDsStr != "" {
		parts := strings.Split(statusIDsStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				if val, err := strconv.Atoi(part); err == nil {
					statusIDs = append(statusIDs, val)
				} else {
					utils.BadRequestResponse(c, "Invalid status_ids parameter", nil)
					return
				}
			}
		}
	}

	// Call service to get data
	result, err := h.billingService.GetBillingStatistics(search, bulan, tahun, rt, statusIDs)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get billing statistics")
		utils.InternalServerErrorResponse(c, "Failed to retrieve billing statistics", err)
		return
	}

	utils.SuccessResponse(c, "Billing statistics retrieved successfully", result)
}
