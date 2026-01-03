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
	CreatePaymentLink(amount int64, billingIDsStr string, documentIDs string, humanDescription string, customerName string, customerEmail string, customerPhone string) (string, error)
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

	// Use default customer information
	userName := "Penghuni IPL"
	userEmail := "billing@ipl.com"
	userPhone := "08123456789"

	// Create billing IDs string for webhook parsing (just the ID)
	billingIDsStr := fmt.Sprintf("%d", billingID)

	// Get document ID
	documentID := ""
	if billing.DocumentID != nil {
		documentID = *billing.DocumentID
	}

	// Create human-readable description
	humanDescription := fmt.Sprintf("Payment for Billing ID %d", billingID)
	if billing.Bulan != nil && billing.Tahun != nil {
		humanDescription = fmt.Sprintf("Payment for %d/%d - Billing ID %d", *billing.Bulan, *billing.Tahun, billingID)
	}

	// Create Mayar payment link
	paymentURL, err := s.mayarService.CreatePaymentLink(*billing.Nominal, billingIDsStr, documentID, humanDescription, userName, userEmail, userPhone)
	if err != nil {
		s.logger.WithError(err).WithField("billing_id", billingID).Error("Failed to create Mayar payment link")
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	return &PaymentLinkResponse{
		BillingID:   billingID,
		Amount:      *billing.Nominal,
		PaymentURL:  paymentURL,
		DocumentID:  documentID,
		Description: humanDescription,
	}, nil
}

// CreatePaymentLinkMultiple creates a Mayar payment link for multiple billing records
func (s *paymentService) CreatePaymentLinkMultiple(billingIDs []uint) (*PaymentLinkResponse, error) {
	if len(billingIDs) == 0 {
		return nil, fmt.Errorf("billing IDs cannot be empty")
	}

	var totalAmount int64 = 0
	var listBillingIDs []uint
	var listDocumentIDs []string

	for _, billingID := range billingIDs {
		// Get billing record
		billing, err := s.billingRepo.GetBillingByID(billingID)
		if err != nil {
			s.logger.WithError(err).WithField("billing_id", billingID).Error("Failed to get billing record")
			return nil, fmt.Errorf("billing record not found for ID %d: %w", billingID, err)
		}

		// Validate nominal exists
		if billing.Nominal == nil || *billing.Nominal <= 0 {
			s.logger.WithField("billing_id", billingID).Error("Invalid billing nominal")
			return nil, fmt.Errorf("invalid billing nominal for ID %d", billingID)
		}

		totalAmount += *billing.Nominal
		listBillingIDs = append(listBillingIDs, billingID)
		if billing.DocumentID != nil {
			listDocumentIDs = append(listDocumentIDs, *billing.DocumentID)
		}
	}

	// Use default customer information
	userName := "Penghuni IPL"
	userEmail := "billing@ipl.com"
	userPhone := "08123456789"

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

	// Create human-readable description
	humanDescription := fmt.Sprintf("Payment for %d billings", len(billingIDs))

	// Create Mayar payment link
	paymentURL, err := s.mayarService.CreatePaymentLink(totalAmount, billingIDsStr, documentIDsStr, humanDescription, userName, userEmail, userPhone)
	if err != nil {
		s.logger.WithError(err).WithField("billing_ids", billingIDs).Error("Failed to create Mayar payment link")
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	return &PaymentLinkResponse{
		BillingIDs:  billingIDs,
		Amount:      totalAmount,
		PaymentURL:  paymentURL,
		Description: humanDescription,
	}, nil
}

// mayarService implements MayarService
type mayarService struct {
	logger *logger.Logger
	config MayarConfig
}

// NewMayarService creates a new instance of MayarService
func NewMayarService(logger *logger.Logger) MayarService {
	config := MayarConfig{
		AuthKey: os.Getenv("MAYAR_AUTH_KEY"),
		BaseURL: os.Getenv("MAYAR_BASE_URL"),
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.mayar.id/hl/v1"
	}

	return &mayarService{
		logger: logger,
		config: config,
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
// humanDescription: human-readable description for display
func (m *mayarService) CreatePaymentLink(amount int64, billingIDsStr string, documentIDs string, humanDescription string, customerName string, customerEmail string, customerPhone string) (string, error) {
	m.logger.WithFields(map[string]interface{}{
		"amount":            amount,
		"billing_ids":       billingIDsStr,
		"document_ids":      documentIDs,
		"human_description": humanDescription,
		"customer_name":     customerName,
		"customer_email":    customerEmail,
		"customer_phone":    customerPhone,
	}).Info("Creating Mayar payment link")

	// Validate auth key
	if m.config.AuthKey == "" {
		return "", fmt.Errorf("Mayar auth key not configured")
	}

	// Set expiration to 30 days from now
	expiredAt := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)

	// Format product description: "<billing_ids> (DocumentID: <document_ids>)"
	// This format is required for webhook parsing
	productDescription := billingIDsStr
	if documentIDs != "" {
		productDescription = fmt.Sprintf("%s (DocumentID: %s)", billingIDsStr, documentIDs)
	} else {
		productDescription = fmt.Sprintf("%s (DocumentID: N/A)", billingIDsStr)
	}

	// Create invoice request
	invoiceReq := &MayarCreateInvoiceRequest{
		Name:        customerName,
		Email:       customerEmail,
		Mobile:      customerPhone,
		RedirectURL: "https://web.mayar.id",
		Description: productDescription, // This will be in webhook's productDescription field
		ExpiredAt:   expiredAt,
		Items: []MayarItem{
			{
				Quantity:    1,
				Rate:        amount,
				Description: humanDescription,
			},
			{
				Quantity:    1,
				Rate:        5000, // Fixed admin fee
				Description: "Admin Fee",
			},
		},
	}

	m.logger.WithField("product_description", productDescription).Info("Product description for webhook parsing")

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
		"amount":              amount,
		"billing_ids":         billingIDsStr,
		"document_ids":        documentIDs,
		"product_description": productDescription,
		"payment_link":        result.Data.Link,
		"invoice_id":          result.Data.ID,
		"transaction_id":      result.Data.TransactionID,
	}).Info("Mayar payment link created successfully")

	return result.Data.Link, nil
}
