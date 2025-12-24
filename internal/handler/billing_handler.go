package handler

import (
	"fmt"
	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"
	"ipl-be-svc/pkg/utils"
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

// GetBillingPenghuni retrieves all billing data for penghuni users
// @Summary Get billing penghuni list with summed nominals
// @Description Get all billing data for penghuni users with complete information including profile, role, and billing status. Nominal amounts are summed per user per billing period (month/year).
// @Tags billings
// @Accept json
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]models.BillingPenghuniResponse} "Billing penghuni retrieved successfully"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Router /api/v1/billings/penghuni [get]
func (h *BulkBillingHandler) GetBillingPenghuni(c *gin.Context) {
	results, err := h.billingService.GetBillingPenghuni()
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
