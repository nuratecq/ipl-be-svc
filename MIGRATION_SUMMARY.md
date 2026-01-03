# Summary of Mayar Payment Gateway Migration

## Overview
Successfully migrated the IPL Backend Service from Doku to Mayar payment gateway, including webhook integration for payment confirmations.

## Files Modified

### 1. Configuration Files
- **internal/config/config.go**
  - Added `MayarConfig` struct with `AuthKey` and `BaseURL` fields
  - Updated `Config` struct to include `Mayar` field
  - Marked `DokuConfig` as deprecated
  - Added environment variable loading for Mayar settings

### 2. Service Layer
- **internal/service/payment_service.go**
  - Replaced `DokuService` interface with `MayarService` interface
  - Removed all Doku-related types (DokuOrder, DokuLineItem, DokuCheckoutRequest, etc.)
  - Added Mayar-specific types:
    - `MayarConfig`
    - `MayarItem`
    - `MayarCreateInvoiceRequest`
    - `MayarCreateInvoiceResponse`
    - `MayarService` interface
  - Implemented `mayarService` struct with methods:
    - `CreateInvoice()` - Creates invoice via Mayar API
    - `CreatePaymentLink()` - Formats billing IDs and creates payment link
  - Updated `CreatePaymentLink()` to format product description as: `<billing_ids> (DocumentID: <document_ids>)`
  - Updated `CreatePaymentLinkMultiple()` to handle comma-separated billing IDs
  - Set invoice expiration to 30 days (instead of 1 day)

### 3. Handler Layer
- **internal/handler/billing_handler.go**
  - Added new `MayarWebhookRequest` struct to handle Mayar webhook payload
  - Kept old `ConfirmPaymentWebhookRequest` for backward compatibility (marked as deprecated)
  - Updated `ConfirmPaymentWebhook()` handler to:
    - Accept Mayar webhook format with `event` and `data` fields
    - Parse billing IDs from `data.productDescription` field
    - Support both single and comma-separated billing IDs
    - Extract IDs by splitting on space (to remove document ID part) then comma
  - Enhanced logging for webhook processing

- **internal/handler/payment_handler.go**
  - Updated documentation to mention Mayar instead of Doku

### 4. Main Application
- **cmd/server/main.go**
  - Changed initialization from `NewDokuService()` to `NewMayarService()`
  - Updated service dependency injection

## Key Features

### Product Description Format
The system uses a specific format for the invoice description that enables webhook parsing:
```
<billing_ids> (DocumentID: <document_ids>)
```

Examples:
- Single billing: `"1372 (DocumentID: monthly-xxx)"`
- Multiple billings: `"1372,67 (DocumentID: monthly-xxx, monthly-yyy)"`

### Webhook Parsing Logic
1. Receives Mayar webhook with `productDescription` field
2. Splits by space to get first part (billing IDs)
3. Splits by comma to get individual IDs
4. Converts strings to uint and confirms payment for all IDs

### Environment Variables
Required environment variables in `.env`:
```bash
MAYAR_AUTH_KEY=<your-bearer-token>
MAYAR_BASE_URL=https://api.mayar.id/hl/v1
```

## API Endpoints

### Create Payment Link (Single Billing)
- **Endpoint:** `POST /api/v1/payments/billing/:id/link`
- **Response includes:** billing_id, amount, payment_url, description

### Create Payment Link (Multiple Billings)
- **Endpoint:** `POST /api/v1/payments/billing/link`
- **Body:** `{"billing_ids": [1, 2, 3]}`
- **Response includes:** billing_ids, amount, payment_url, description

### Webhook Endpoint
- **Endpoint:** `POST /api/v1/billings/confirm-payment`
- **Accepts:** Mayar webhook payload with event "payment.received"
- **Processes:** Extracts billing IDs and confirms payment

## Mayar API Integration

### Invoice Creation
- **URL:** `https://api.mayar.id/hl/v1/invoice/create`
- **Method:** POST
- **Headers:**
  - `Authorization: Bearer <MAYAR_AUTH_KEY>`
  - `Content-Type: application/json`
- **Body:**
  - `name`: Customer name
  - `email`: Customer email
  - `mobile`: Customer phone
  - `description`: Billing IDs with document IDs
  - `expiredAt`: Invoice expiration date (30 days)
  - `items`: Array with rate and description

### Webhook Format
Mayar sends webhooks with:
- `event`: "payment.received"
- `data.productDescription`: Contains billing IDs
- `data.status`: Payment status
- `data.transactionId`: Transaction ID

## Testing

### Build Command
```bash
go build -o bin/server ./cmd/server/
```

### Run Server
```bash
go run cmd/server/main.go
```

### Test Payment Link Creation
```bash
# Single billing
curl -X POST http://localhost:8080/api/v1/payments/billing/1/link \
  -H "Cookie: auth-token=your-jwt-token"

# Multiple billings
curl -X POST http://localhost:8080/api/v1/payments/billing/link \
  -H "Content-Type: application/json" \
  -H "Cookie: auth-token=your-jwt-token" \
  -d '{"billing_ids": [1, 2, 3]}'
```

### Test Webhook
```bash
curl -X POST http://localhost:8080/api/v1/billings/confirm-payment \
  -H "Content-Type: application/json" \
  -d '{
    "event": "payment.received",
    "data": {
      "productDescription": "1372,67 (DocumentID: monthly-xxx)",
      "status": "SUCCESS",
      "transactionId": "test-123"
    }
  }'
```

## Migration Notes

### Breaking Changes
- Doku service completely replaced with Mayar
- Webhook payload format changed
- Product description format is now critical for webhook parsing

### Backward Compatibility
- Old Doku configuration still present but marked as deprecated
- Old webhook struct kept for reference but not used

### Default Customer Information
Currently using hardcoded defaults:
- Name: "Penghuni IPL"
- Email: "billing@ipl.com"
- Phone: "08123456789"

**Future Enhancement:** Fetch actual customer data from user profiles

## Documentation
- **MAYAR_MIGRATION.md**: Detailed migration guide and API documentation
- Swagger documentation updated with new Mayar endpoints

## Success Metrics
✅ Build successful without errors
✅ Server starts and connects to database
✅ Payment links created successfully via Mayar API
✅ Webhook endpoint ready to receive Mayar callbacks
✅ Billing ID parsing logic implemented and tested

## Next Steps
1. Test webhook with actual Mayar payment callbacks
2. Consider fetching real customer information from profiles
3. Add webhook signature validation for security
4. Monitor Mayar API logs for any issues
5. Update Swagger documentation if needed
