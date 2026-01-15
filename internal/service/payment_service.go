package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/repository"
	"ipl-be-svc/pkg/logger"
)

// MayarConfig holds Mayar API configuration
type MayarConfig struct {
	AuthKey string
	BaseURL string
}

// MayarItem represents an item in the invoice
type MayarItem struct {
	Quantity    int    `json:"quantity"`
	Rate        int64  `json:"rate"`
	Description string `json:"description"`
}

// MayarCreateInvoiceRequest represents the Mayar invoice creation request
type MayarCreateInvoiceRequest struct {
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Mobile      string      `json:"mobile"`
	RedirectURL string      `json:"redirectUrl"`
	Description string      `json:"description"`
	ExpiredAt   string      `json:"expiredAt"`
	Items       []MayarItem `json:"items"`
}

// MayarCreateInvoiceResponse represents the Mayar API response
type MayarCreateInvoiceResponse struct {
	StatusCode int    `json:"statusCode"`
	Messages   string `json:"messages"`
	Data       struct {
		ID            string `json:"id"`
		TransactionID string `json:"transactionId"`
		Link          string `json:"link"`
		ExpiredAt     int64  `json:"expiredAt"`
	} `json:"data"`
}

// MayarService defines the interface for Mayar payment operations
type MayarService interface {
	CreatePaymentLink(billings []*models.Billing, billingIDsStr string, documentIDs string) (string, error)
	CreateInvoice(req *MayarCreateInvoiceRequest) (*MayarCreateInvoiceResponse, error)
}

// PaymentService defines the interface for payment operations
type PaymentService interface {
	CreatePaymentLink(billingID uint) (*PaymentLinkResponse, error)
	CreatePaymentLinkMultiple(billingIDs []uint) (*PaymentLinkResponse, error)
}

// PaymentLinkResponse represents the response for payment link creation
type PaymentLinkResponse struct {
	BillingID     uint   `json:"billing_id,omitempty"`
	BillingIDs    []uint `json:"billing_ids,omitempty"`
	Amount        int64  `json:"amount"`
	PaymentURL    string `json:"payment_url"`
	Description   string `json:"description"`
	InvoiceID     string `json:"invoice_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	ExpiredAt     int64  `json:"expired_at,omitempty"`
	DocumentID    string `json:"document_id,omitempty"`
}

// paymentService implements PaymentService
type paymentService struct {
	billingRepo  repository.BillingRepository
	mayarService MayarService
	logger       *logger.Logger
}

// NewPaymentService creates a new instance of PaymentService
func NewPaymentService(billingRepo repository.BillingRepository, mayarService MayarService, logger *logger.Logger) PaymentService {
	return &paymentService{
		billingRepo:  billingRepo,
		mayarService: mayarService,
		logger:       logger,
	}
}

// CreatePaymentLink creates a Mayar payment link for a billing record
func (s *paymentService) CreatePaymentLink(billingID uint) (*PaymentLinkResponse, error) {
	// Get billing record
	billing, err := s.billingRepo.GetBillingByID(billingID)
	if err != nil {
		s.logger.WithError(err).WithField("billing_id", billingID).Error("Failed to get billing record")
		return nil, fmt.Errorf("billing record not found: %w", err)
	}
	s.logger.WithField("billing", billing).Info("Retrieved billing record")

	// Validate nominal exists
	if billing.Nominal == nil || *billing.Nominal <= 0 {
		s.logger.WithField("billing_id", billingID).Error("Invalid billing nominal")
		return nil, fmt.Errorf("invalid billing nominal")
	}

	// Create billing IDs string for webhook parsing
	billingIDsStr := fmt.Sprintf("%d", billingID)

	// Get document ID
	documentID := ""
	if billing.DocumentID != nil {
		documentID = *billing.DocumentID
	}

	// Create billings slice for payment link
	billings := []*models.Billing{billing}

	// Create Mayar payment link
	paymentURL, err := s.mayarService.CreatePaymentLink(billings, billingIDsStr, documentID)
	if err != nil {
		s.logger.WithError(err).WithField("billing_id", billingID).Error("Failed to create Mayar payment link")
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	return &PaymentLinkResponse{
		BillingID:   billingID,
		Amount:      *billing.Nominal,
		PaymentURL:  paymentURL,
		DocumentID:  documentID,
		Description: fmt.Sprintf("Payment for %d billings", len(billings)),
	}, nil
}

// CreatePaymentLinkMultiple creates a Mayar payment link for multiple billing records
func (s *paymentService) CreatePaymentLinkMultiple(billingIDs []uint) (*PaymentLinkResponse, error) {
	if len(billingIDs) == 0 {
		return nil, fmt.Errorf("billing IDs cannot be empty")
	}

	// Get all billings using WHERE IN (optimized query)
	billings, err := s.billingRepo.GetBillingsByIDs(billingIDs)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get billing records")
		return nil, fmt.Errorf("failed to get billing records: %w", err)
	}

	if len(billings) == 0 {
		return nil, fmt.Errorf("no billing records found")
	}

	// Validate and collect document IDs
	var totalAmount int64 = 0
	var listDocumentIDs []string

	for _, billing := range billings {
		// Validate nominal exists
		if billing.Nominal == nil || *billing.Nominal <= 0 {
			s.logger.WithField("billing_id", billing.ID).Error("Invalid billing nominal")
			return nil, fmt.Errorf("invalid billing nominal for ID %d", billing.ID)
		}

		totalAmount += *billing.Nominal

		if billing.DocumentID != nil {
			listDocumentIDs = append(listDocumentIDs, *billing.DocumentID)
		}
	}

	// Create billing IDs string for webhook parsing (comma-separated)
	billingIDsStr := strings.Join(func() []string {
		parts := make([]string, len(billingIDs))
		for i, id := range billingIDs {
			parts[i] = fmt.Sprintf("%d", id)
		}
		return parts
	}(), ",")

	// Create document IDs string
	documentIDsStr := strings.Join(listDocumentIDs, ", ")

	// Create Mayar payment link
	paymentURL, err := s.mayarService.CreatePaymentLink(billings, billingIDsStr, documentIDsStr)
	if err != nil {
		s.logger.WithError(err).WithField("billing_ids", billingIDs).Error("Failed to create Mayar payment link")
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	return &PaymentLinkResponse{
		BillingIDs:  billingIDs,
		Amount:      totalAmount,
		PaymentURL:  paymentURL,
		Description: fmt.Sprintf("Payment for %d billings", len(billingIDs)),
	}, nil
}

// mayarService implements MayarService
type mayarService struct {
	paymentConfigRepo repository.PaymentConfigRepository
	logger            *logger.Logger
	config            MayarConfig
}

// NewMayarService creates a new instance of MayarService
func NewMayarService(paymentConfigRepo repository.PaymentConfigRepository, logger *logger.Logger) MayarService {
	config := MayarConfig{
		AuthKey: os.Getenv("MAYAR_AUTH_KEY"),
		BaseURL: os.Getenv("MAYAR_BASE_URL"),
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.mayar.id/hl/v1"
	}

	return &mayarService{
		paymentConfigRepo: paymentConfigRepo,
		logger:            logger,
		config:            config,
	}
}

// CreateInvoice creates an invoice using Mayar API
func (m *mayarService) CreateInvoice(req *MayarCreateInvoiceRequest) (*MayarCreateInvoiceResponse, error) {
	url := fmt.Sprintf("%s/invoice/create", m.config.BaseURL)

	// Marshal request to JSON
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.AuthKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// Log request
	m.logger.WithFields(map[string]interface{}{
		"url":  url,
		"body": string(bodyJSON),
	}).Info("📡 Sending request to Mayar API...")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response
	m.logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"response":    string(body),
	}).Info("Mayar API Response")

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Mayar API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var mayarResp MayarCreateInvoiceResponse
	if err := json.Unmarshal(body, &mayarResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Debug log the parsed response
	m.logger.WithFields(map[string]interface{}{
		"status":         mayarResp.StatusCode,
		"message":        mayarResp.Messages,
		"invoice_id":     mayarResp.Data.ID,
		"transaction_id": mayarResp.Data.TransactionID,
		"payment_link":   mayarResp.Data.Link,
	}).Info("Successfully parsed Mayar response")

	return &mayarResp, nil
}

// CreatePaymentLink creates a payment link using Mayar service
// billingIDsStr: comma-separated billing IDs (e.g., "1372,67" or "1372")
// documentIDs: comma-separated document IDs for reference
func (m *mayarService) CreatePaymentLink(billings []*models.Billing, billingIDsStr string, documentIDs string) (string, error) {
	// Get payment config from database
	paymentConfig, err := m.paymentConfigRepo.GetActivePaymentConfig()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get payment config")
		return "", fmt.Errorf("failed to get payment config: %w", err)
	}

	// Use customer information from payment config
	userName := "Admin"
	if paymentConfig.AdminName != nil {
		userName = *paymentConfig.AdminName
	}
	userEmail := "admin@mail.com"
	if paymentConfig.AdminEmail != nil {
		userEmail = *paymentConfig.AdminEmail
	}
	userPhone := "-"
	if paymentConfig.AdminPhone != nil {
		userPhone = *paymentConfig.AdminPhone
	}

	// Calculate admin fee based on unique months
	monthsMap := make(map[string]bool)
	for _, billing := range billings {
		if billing.Bulan != nil && billing.Tahun != nil {
			monthKey := fmt.Sprintf("%d-%d", *billing.Tahun, *billing.Bulan)
			monthsMap[monthKey] = true
		}
	}

	totalMonths := len(monthsMap)
	basePaymentFee := int64(0)
	if paymentConfig.PaymentFee != nil {
		basePaymentFee = *paymentConfig.PaymentFee
	}

	// Calculate admin fee with optimized logic
	var adminFee int64
	isFixedFee := paymentConfig.IsFixedFee != nil && *paymentConfig.IsFixedFee

	if isFixedFee {
		// Fixed Fee mode: use base fee directly
		adminFee = basePaymentFee
	} else {
		// Non-fixed Fee mode: calculate based on months
		// Get discount threshold (default 6 months)
		minMonths := 6
		if paymentConfig.MinMonthDiscount != nil {
			minMonths = *paymentConfig.MinMonthDiscount
		}

		if totalMonths < minMonths {
			// If less than 6 months, multiply months by payment fee
			adminFee = int64(totalMonths) * basePaymentFee
		} else {
			// If paying for 6+ months, use fixed discount amount (default 20000)
			if paymentConfig.MaxFee != nil {
				adminFee = *paymentConfig.MaxFee
			} else {
				adminFee = int64(totalMonths) * basePaymentFee
			}
		}
	}

	m.logger.WithFields(map[string]interface{}{
		"total_months":     totalMonths,
		"base_payment_fee": basePaymentFee,
		"is_fixed_fee":     isFixedFee,
		"calculated_fee":   adminFee,
	}).Info("Calculated admin fee based on payment config")

	// Validate auth key
	if m.config.AuthKey == "" {
		return "", fmt.Errorf("Mayar auth key not configured")
	}

	// Set expiration to 30 days from now
	expiredAt := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)

	// Format product description for webhook parsing
	productDescription := billingIDsStr
	if documentIDs != "" {
		productDescription = fmt.Sprintf("%s (DocumentID: %s)", billingIDsStr, documentIDs)
	} else {
		productDescription = fmt.Sprintf("%s (DocumentID: N/A)", billingIDsStr)
	}

	// Create detailed items for each billing
	items := make([]MayarItem, 0, len(billings)+1)
	for _, billing := range billings {
		if billing.Nominal == nil {
			continue
		}

		// Create description for billing item
		desc := fmt.Sprintf("Billing ID %d", billing.DocumentID)
		if billing.NamaBilling != nil {
			desc = *billing.NamaBilling
		}
		if billing.Bulan != nil && billing.Tahun != nil {
			desc = fmt.Sprintf("%s - %d/%d", desc, *billing.Bulan, *billing.Tahun)
		}

		items = append(items, MayarItem{
			Quantity:    1,
			Rate:        *billing.Nominal,
			Description: desc,
		})
	}

	// Add admin fee as separate item
	items = append(items, MayarItem{
		Quantity:    1,
		Rate:        adminFee,
		Description: "Admin Fee",
	})

	// Create invoice request
	invoiceReq := &MayarCreateInvoiceRequest{
		Name:        userName,
		Email:       userEmail,
		Mobile:      userPhone,
		RedirectURL: "https://web.mayar.id",
		Description: productDescription,
		ExpiredAt:   expiredAt,
		Items:       items,
	}

	m.logger.WithFields(map[string]interface{}{
		"billing_ids":         billingIDsStr,
		"document_ids":        documentIDs,
		"product_description": productDescription,
		"total_items":         len(items),
		"admin_fee":           adminFee,
	}).Info("Creating Mayar payment link")

	// Create invoice
	result, err := m.CreateInvoice(invoiceReq)
	if err != nil {
		m.logger.WithError(err).Error("Failed to create Mayar invoice")
		return "", err
	}

	// Extract payment link from response
	if result.Data.Link == "" {
		m.logger.Error("Payment link not found in response")
		return "", fmt.Errorf("payment link not found in response")
	}

	m.logger.WithFields(map[string]interface{}{
		"billing_ids":         billingIDsStr,
		"document_ids":        documentIDs,
		"product_description": productDescription,
		"payment_link":        result.Data.Link,
		"invoice_id":          result.Data.ID,
		"transaction_id":      result.Data.TransactionID,
	}).Info("Mayar payment link created successfully")

	return result.Data.Link, nil
}
