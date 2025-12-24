package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ipl-be-svc/internal/repository"
	"ipl-be-svc/pkg/logger"

	"github.com/google/uuid"
)

// DokuConfig holds DOKU API configuration
type DokuConfig struct {
	ClientID  string
	SecretKey string
	BaseURL   string
}

// DokuOrder represents order details for DOKU checkout
type DokuOrder struct {
	Amount        int64          `json:"amount"`
	InvoiceNumber string         `json:"invoice_number"`
	Currency      string         `json:"currency"`
	SessionID     string         `json:"session_id"`
	CallbackURL   string         `json:"callback_url"`
	LineItems     []DokuLineItem `json:"line_items"`
}

// DokuLineItem represents a line item in the order
type DokuLineItem struct {
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
}

// DokuLineItemResponse represents a line item in the response (with string price)
type DokuLineItemResponse struct {
	Name     string `json:"name"`
	Price    string `json:"price"`
	Quantity int    `json:"quantity"`
}

// DokuPayment represents payment configuration
type DokuPayment struct {
	PaymentDueDate int `json:"payment_due_date"`
}

// DokuCustomer represents customer information
type DokuCustomer struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Country string `json:"country"`
}

// DokuCheckoutRequest represents the complete DOKU checkout request
type DokuCheckoutRequest struct {
	Order    DokuOrder    `json:"order"`
	Payment  DokuPayment  `json:"payment"`
	Customer DokuCustomer `json:"customer"`
}

// DokuCheckoutResponse represents the actual DOKU API response structure
type DokuCheckoutResponse struct {
	Message  []string `json:"message"`
	Response struct {
		Order struct {
			Amount        string                 `json:"amount"`
			InvoiceNumber string                 `json:"invoice_number"`
			Currency      string                 `json:"currency"`
			SessionID     string                 `json:"session_id"`
			CallbackURL   string                 `json:"callback_url"`
			LineItems     []DokuLineItemResponse `json:"line_items"`
		} `json:"order"`
		Payment struct {
			PaymentMethodTypes []string `json:"payment_method_types"`
			PaymentDueDate     int      `json:"payment_due_date"`
			TokenID            string   `json:"token_id"`
			URL                string   `json:"url"`
			ExpiredDate        string   `json:"expired_date"`
			ExpiredDatetime    string   `json:"expired_datetime"`
		} `json:"payment"`
		Customer struct {
			Email   string `json:"email"`
			Phone   string `json:"phone"`
			Name    string `json:"name"`
			Address string `json:"address"`
			Country string `json:"country"`
		} `json:"customer"`
		AdditionalInfo struct {
			Origin struct {
				Product   string `json:"product"`
				System    string `json:"system"`
				APIFormat string `json:"apiFormat"`
				Source    string `json:"source"`
			} `json:"origin"`
			LineItems []DokuLineItemResponse `json:"line_items"`
		} `json:"additional_info"`
		UUID    interface{} `json:"uuid"` // Can be int64 or float64 depending on size
		Headers struct {
			RequestID string `json:"request_id"`
			Signature string `json:"signature"`
			Date      string `json:"date"`
			ClientID  string `json:"client_id"`
		} `json:"headers"`
	} `json:"response"`
}

// DokuService defines the interface for DOKU payment operations
type DokuService interface {
	CreatePaymentLink(amount int64, description string) (string, error)
	InitiateDokuCheckout(clientID, secretKey string, amount int64, description string) (*DokuCheckoutResponse, error)
}

// PaymentService defines the interface for payment operations
type PaymentService interface {
	CreatePaymentLink(billingID uint) (*PaymentLinkResponse, error)
	CreatePaymentLinkMultiple(billingIDs []uint) (*PaymentLinkResponse, error)
}

// PaymentLinkResponse represents the response for payment link creation
type PaymentLinkResponse struct {
	BillingID   uint   `json:"billing_id,omitempty"`
	BillingIDs  []uint `json:"billing_ids,omitempty"`
	Amount      int64  `json:"amount"`
	PaymentURL  string `json:"payment_url"`
	Description string `json:"description"`
}

// paymentService implements PaymentService
type paymentService struct {
	billingRepo repository.BillingRepository
	dokuService DokuService
	logger      *logger.Logger
}

// NewPaymentService creates a new instance of PaymentService
func NewPaymentService(billingRepo repository.BillingRepository, dokuService DokuService, logger *logger.Logger) PaymentService {
	return &paymentService{
		billingRepo: billingRepo,
		dokuService: dokuService,
		logger:      logger,
	}
}

// CreatePaymentLink creates a DOKU payment link for a billing record
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

	// Create description
	description := fmt.Sprintf("Payment for Billing ID %d", billingID)
	if billing.Bulan != nil && billing.Tahun != nil {
		description = fmt.Sprintf("Payment for %d/%d - Billing ID %d", *billing.Bulan, *billing.Tahun, billingID)
	}

	// Create DOKU payment link
	paymentURL, err := s.dokuService.CreatePaymentLink(*billing.Nominal, description)
	if err != nil {
		s.logger.WithError(err).WithField("billing_id", billingID).Error("Failed to create DOKU payment link")
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	return &PaymentLinkResponse{
		BillingID:   billingID,
		Amount:      *billing.Nominal,
		PaymentURL:  paymentURL,
		Description: description,
	}, nil
}

// CreatePaymentLinkMultiple creates a DOKU payment link for multiple billing records
func (s *paymentService) CreatePaymentLinkMultiple(billingIDs []uint) (*PaymentLinkResponse, error) {
	if len(billingIDs) == 0 {
		return nil, fmt.Errorf("billing IDs cannot be empty")
	}

	var totalAmount int64 = 0
	var listBillingIDs []uint

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

		// Create description part
		listBillingIDs = append(listBillingIDs, billingID)

	}

	// Create combined description
	description := strings.Join(func() []string {
		parts := make([]string, len(billingIDs))
		for i, id := range billingIDs {
			parts[i] = fmt.Sprintf("%d", id)
		}
		return parts
	}(), ",")

	fmt.Println("totalAmount : ", totalAmount)
	fmt.Println("Desc : ", description)

	// Create DOKU payment link
	paymentURL, err := s.dokuService.CreatePaymentLink(totalAmount, description)
	if err != nil {
		s.logger.WithError(err).WithField("billing_ids", billingIDs).Error("Failed to create DOKU payment link")
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	return &PaymentLinkResponse{
		BillingIDs:  billingIDs,
		Amount:      totalAmount,
		PaymentURL:  paymentURL,
		Description: description,
	}, nil
}

// dokuService implements DokuService
type dokuService struct {
	logger *logger.Logger
	config DokuConfig
}

// NewDokuService creates a new instance of DokuService
func NewDokuService(logger *logger.Logger) DokuService {
	config := DokuConfig{
		ClientID:  os.Getenv("DOKU_CLIENT_ID"),
		SecretKey: os.Getenv("DOKU_SECRET_KEY"),
		BaseURL:   "https://api-sandbox.doku.com",
	}

	if config.ClientID == "" {
		config.ClientID = "BRN-0241-1762176502792" // Default dari Python code
	}
	if config.SecretKey == "" {
		config.SecretKey = "SK-PaILsZudZTytTSTNCmUV" // Default dari Python code
	}

	return &dokuService{
		logger: logger,
		config: config,
	}
}

// generateSignature creates HMACSHA256 signature exactly like Python code
func (d *dokuService) generateSignature(clientID, secretKey, requestID, requestTimestamp, requestTarget, body string) string {
	// Step 1: Digest body menggunakan SHA256 lalu encode base64
	bodyHash := sha256.Sum256([]byte(body))
	digestBase64 := base64.StdEncoding.EncodeToString(bodyHash[:])

	// Step 2: Gabungkan semua komponen signature
	signatureComponents := fmt.Sprintf("Client-Id:%s\nRequest-Id:%s\nRequest-Timestamp:%s\nRequest-Target:%s\nDigest:%s",
		clientID, requestID, requestTimestamp, requestTarget, digestBase64)

	// Step 3: Buat HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(signatureComponents))
	signatureHMAC := h.Sum(nil)
	signatureBase64 := base64.StdEncoding.EncodeToString(signatureHMAC)

	return fmt.Sprintf("HMACSHA256=%s", signatureBase64)
}

// InitiateDokuCheckout initiates DOKU checkout payment exactly like Python code
func (d *dokuService) InitiateDokuCheckout(clientID, secretKey string, amount int64, description string) (*DokuCheckoutResponse, error) {
	// --- Konfigurasi dasar ---
	url := fmt.Sprintf("%s/checkout/v1/payment", d.config.BaseURL)
	requestTarget := "/checkout/v1/payment"

	// --- Generate ID & Timestamp ---
	requestID := uuid.New().String()
	requestTimestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// --- Payload body ---
	payload := DokuCheckoutRequest{
		Order: DokuOrder{
			Amount:        amount + 5000, // Tambah service fee
			InvoiceNumber: fmt.Sprintf("INV-%d-%s", time.Now().Unix(), description),
			Currency:      "IDR",
			SessionID:     "SU5WFDferd561dfasfasdfae123c",
			CallbackURL:   "https://doku.com/",
			LineItems: []DokuLineItem{
				{Name: "Biaya IPL", Price: amount, Quantity: 1},
				{Name: "Biaya Layanan", Price: 5000, Quantity: 1},
			},
		},
		Payment: DokuPayment{PaymentDueDate: 60},
		Customer: DokuCustomer{
			Name:    "",
			Email:   "",
			Phone:   "+6285694566147",
			Address: "Plaza Asia Office Park Unit 3",
			Country: "ID",
		},
	}

	// Konversi ke string JSON (compact, separators=(',', ':'))
	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// --- Generate Signature ---
	signature := d.generateSignature(clientID, secretKey, requestID, requestTimestamp, requestTarget, string(bodyJSON))

	// --- Header HTTP ---
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Request-Id", requestID)
	req.Header.Set("Request-Timestamp", requestTimestamp)
	req.Header.Set("Signature", signature)

	// --- Eksekusi request ---
	d.logger.Info("📡 Sending request to DOKU Sandbox...")
	d.logger.WithFields(map[string]interface{}{
		"url":               url,
		"request_id":        requestID,
		"request_timestamp": requestTimestamp,
		"signature":         signature,
	}).Info("Request Info")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// --- Baca response ---
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging
	d.logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"response":    string(body),
	}).Info("DOKU API Response")

	// Parse response
	var dokuResp DokuCheckoutResponse
	if err := json.Unmarshal(body, &dokuResp); err != nil {
		// If we can't parse as expected structure, try generic
		var genericResp map[string]interface{}
		if err := json.Unmarshal(body, &genericResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		d.logger.WithField("response", genericResp).Error("Unexpected response structure")
		return nil, fmt.Errorf("unexpected response structure: %v", genericResp)
	}

	// Debug log the parsed response
	d.logger.WithFields(map[string]interface{}{
		"message":     dokuResp.Message,
		"payment_url": dokuResp.Response.Payment.URL,
		"token_id":    dokuResp.Response.Payment.TokenID,
	}).Info("Successfully parsed DOKU response")

	return &dokuResp, nil
}

// CreatePaymentLink creates a payment link using DOKU service
func (d *dokuService) CreatePaymentLink(amount int64, description string) (string, error) {

	d.logger.WithFields(map[string]interface{}{
		"amount":      amount,
		"description": description,
	}).Info("Creating DOKU payment link")

	// Use configured credentials
	clientID := d.config.ClientID
	secretKey := d.config.SecretKey

	if clientID == "" || secretKey == "" {
		return "", fmt.Errorf("DOKU credentials not configured")
	}

	// Initiate DOKU checkout
	result, err := d.InitiateDokuCheckout(clientID, secretKey, amount, description)
	if err != nil {
		d.logger.WithError(err).Error("Failed to initiate DOKU checkout")
		return "", err
	}

	// Extract payment URL from response
	if result.Response.Payment.URL == "" {
		d.logger.Error("Payment URL not found in response")
		return "", fmt.Errorf("payment URL not found in response")
	}

	d.logger.WithFields(map[string]interface{}{
		"amount":      amount,
		"description": description,
		"payment_url": result.Response.Payment.URL,
	}).Info("DOKU payment link created successfully")

	return result.Response.Payment.URL, nil
}
