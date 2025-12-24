package handler

import (
	"log"
	"net/http"
	"strconv"

	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService service.PaymentService
	logger         *logger.Logger
}

// CreatePaymentLinkMultipleRequest represents the request body for creating payment link for multiple billings
type CreatePaymentLinkMultipleRequest struct {
	BillingIDs []uint `json:"billing_ids" binding:"required" example:"6,2"` // Array of billing IDs to create payment link for
}

// NewPaymentHandler creates a new PaymentHandler instance
func NewPaymentHandler(paymentService service.PaymentService, logger *logger.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// CreatePaymentLink creates a payment link for a billing record
// @Summary Create payment link
// @Description Create a DOKU payment link for a billing record by ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Billing ID"
// @Success 200 {object} service.PaymentLinkResponse "Payment link created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid billing ID"
// @Failure 404 {object} map[string]interface{} "Billing not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/payments/billing/{id}/link [post]
func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {
	// Get billing ID from path parameter
	idParam := c.Param("id")
	log.Println("Id Param:", idParam)
	billingID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.WithError(err).WithField("id_param", idParam).Error("Invalid billing ID parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid billing ID",
			"message": "Billing ID must be a valid number",
		})
		return
	}

	// Create payment link
	response, err := h.paymentService.CreatePaymentLink(uint(billingID))
	if err != nil {
		h.logger.WithError(err).WithField("billing_id", billingID).Error("Failed to create payment link")

		// Check if it's a not found error
		if err.Error() == "billing record not found" || err.Error() == "invalid billing nominal" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Billing not found",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create payment link",
			"message": "Internal server error",
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"billing_id":  billingID,
		"amount":      response.Amount,
		"payment_url": response.PaymentURL,
	}).Info("Payment link created successfully")

	c.JSON(http.StatusOK, response)
}

// CreatePaymentLinkMultiple creates a payment link for multiple billing records
// @Summary Create payment link for multiple billings
// @Description Create a DOKU payment link for multiple billing records by IDs
// @Tags payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentLinkMultipleRequest true "Billing IDs"
// @Success 200 {object} service.PaymentLinkResponse "Payment link created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid billing IDs"
// @Failure 404 {object} map[string]interface{} "Billing not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/payments/billing/link [post]
func (h *PaymentHandler) CreatePaymentLinkMultiple(c *gin.Context) {
	var request CreatePaymentLinkMultipleRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "billing_ids is required and must be an array of numbers",
		})
		return
	}

	if len(request.BillingIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "billing_ids cannot be empty",
		})
		return
	}

	// Create payment link
	response, err := h.paymentService.CreatePaymentLinkMultiple(request.BillingIDs)
	if err != nil {
		h.logger.WithError(err).WithField("billing_ids", request.BillingIDs).Error("Failed to create payment link")

		// Check if it's a not found error
		if err.Error() == "billing record not found" || err.Error() == "invalid billing nominal" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Billing not found",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create payment link",
			"message": "Internal server error",
		})
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"billing_ids": request.BillingIDs,
		"amount":      response.Amount,
		"payment_url": response.PaymentURL,
	}).Info("Payment link created successfully for multiple billings")

	c.JSON(http.StatusOK, response)
}
