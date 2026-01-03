# Migration from Doku to Mayar Payment Gateway

## Overview
The IPL Backend Service has been migrated from Doku payment gateway to Mayar payment gateway. This document outlines the changes made and how to use the new Mayar integration.

## Changes Made

### 1. Configuration
- Added Mayar configuration in `internal/config/config.go`
- Environment variables required:
  ```bash
  MAYAR_AUTH_KEY=your-mayar-auth-key
  MAYAR_BASE_URL=https://api.mayar.id/hl/v1
  ```

### 2. Service Layer
- Replaced `DokuService` with `MayarService` in `internal/service/payment_service.go`
- New Mayar service implements:
  - `CreateInvoice()` - Creates invoice using Mayar API
  - `CreatePaymentLink()` - Creates payment link for billing

### 3. Payment Flow
The new Mayar payment flow:
1. Client requests payment link via API
2. Service fetches billing information
3. Service creates Mayar invoice with:
   - Customer name (default: "Penghuni IPL")
   - Customer email (default: "billing@ipl.com")
   - Customer phone (default: "08123456789")
   - Amount from billing
   - Description with billing ID and month/year
   - Expiration date (30 days from creation)
4. Mayar API returns payment link
5. Client receives payment link in response

## API Endpoints

### Create Payment Link for Single Billing
```bash
POST /api/v1/payments/billing/{id}/link
```

**Example Request:**
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/payments/billing/1/link \
  --header 'Cookie: auth-token=your-jwt-token'
```

**Example Response:**
```json
{
  "billing_id": 1,
  "amount": 100000,
  "payment_url": "https://nasri-adzlani.myr.id/invoices/xxxxx",
  "description": "Payment for 12/2025 - Billing ID 1"
}
```

### Create Payment Link for Multiple Billings
```bash
POST /api/v1/payments/billing/link
```

**Example Request:**
```bash
curl --request POST \
  --url http://localhost:8080/api/v1/payments/billing/link \
  --header 'Content-Type: application/json' \
  --header 'Cookie: auth-token=your-jwt-token' \
  --data '{
    "billing_ids": [1, 2, 3]
  }'
```

**Example Response:**
```json
{
  "billing_ids": [1, 2, 3],
  "amount": 300000,
  "payment_url": "https://nasri-adzlani.myr.id/invoices/xxxxx",
  "description": "1,2,3"
}
```

## Mayar API Integration

### Creating Invoice Request Format
The service automatically formats the invoice description to enable webhook parsing:
- Product description format: `<billing_ids> (DocumentID: <document_ids>)`
- Example: `"1372,67 (DocumentID: monthly-xxx, monthly-yyy)"`
- This format allows the webhook handler to extract billing IDs

```bash
curl --request POST \
  --url https://api.mayar.id/hl/v1/invoice/create \
  --header 'Authorization: Bearer YOUR_MAYAR_AUTH_KEY' \
  --header 'Content-Type: application/json' \
  --data '{
    "name": "Customer Name",
    "email": "customer@email.com",
    "mobile": "08123456789",
    "redirectUrl": "https://web.mayar.id",
    "description": "1372,67 (DocumentID: monthly-xxx, monthly-yyy)",
    "expiredAt": "2026-12-01T09:41:09.401Z",
    "items": [{
      "quantity": 1,
      "rate": 100000,
      "description": "Payment for 2 billings"
    }]
  }'
```

### Response Format
```json
{
  "statusCode": 200,
  "messages": "success",
  "data": {
    "id": "invoice-id",
    "transactionId": "transaction-id",
    "link": "https://nasri-adzlani.myr.id/invoices/xxxxx",
    "expiredAt": 1796118069401
  }
}
```

### Webhook Endpoint
The webhook endpoint `/api/v1/billings/confirm-payment` receives payment confirmations from Mayar.

**Webhook Request Format:**
```json
{
  "event": "payment.received",
  "data": {
    "id": "b31fce13-1d1c-4bab-8bae-63bc2142a14e",
    "transactionId": "b31fce13-1d1c-4bab-8bae-63bc2142a14e",
    "status": "SUCCESS",
    "transactionStatus": "created",
    "amount": 2000,
    "productDescription": "1372,67 (DocumentID: monthly-xxx, monthly-yyy)",
    "paymentMethod": "QRIS",
    "customerName": "Penghuni IPL",
    "customerEmail": "billing@ipl.com",
    "customerMobile": "08123456789"
  }
}
```

**Billing ID Parsing:**
The webhook handler extracts billing IDs from `productDescription`:
- Format: `<billing_ids> (DocumentID: ...)`
- Examples:
  - `"1372,67 (DocumentID: ...)"`  → Billing IDs: `[1372, 67]`
  - `"1372 (DocumentID: ...)"`     → Billing IDs: `[1372]`

The handler:
1. Splits `productDescription` by space and takes the first part
2. Splits that part by comma to get individual billing IDs
3. Confirms payment for all extracted billing IDs

## Environment Setup

### .env File
```bash
# Mayar Payment Configuration
MAYAR_AUTH_KEY=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
MAYAR_BASE_URL=https://api.mayar.id/hl/v1
```

## Testing

### Manual Testing
1. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

2. Create a payment link:
   ```bash
   curl --request POST \
     --url http://localhost:8080/api/v1/payments/billing/1/link \
     --header 'Cookie: auth-token=your-jwt-token'
   ```

3. Check the response for payment URL

4. Visit the payment URL to verify it works

## Migration Notes

### Customer Information
Currently using default customer information:
- Name: "Penghuni IPL"
- Email: "billing@ipl.com"
- Phone: "08123456789"

Future enhancement: Fetch actual customer information from user profile.

### Expiration
Invoice expires after 30 days from creation date.

### Backward Compatibility
The old Doku configuration is still present in the config but is marked as deprecated.

## Troubleshooting

### Error: "Mayar auth key not configured"
- Check that `MAYAR_AUTH_KEY` is set in your `.env` file
- Verify the auth key is valid and not expired

### Error: "Payment link not found in response"
- Check Mayar API credentials
- Verify the API endpoint is correct
- Check API logs for detailed error messages

### Error: "billing record not found"
- Verify the billing ID exists in the database
- Check that the billing has a valid nominal value

## Additional Resources

- Mayar API Documentation: https://mayar.id/docs
- Mayar Dashboard: https://web.mayar.id
