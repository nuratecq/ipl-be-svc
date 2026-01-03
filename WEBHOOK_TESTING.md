# Mayar Webhook Testing Guide

## Test Scenarios

### Scenario 1: Single Billing Payment
**Product Description:** `1372 (DocumentID: monthly-6763f269-d01d-401f-842c-903db75936d3)`

**Expected Result:** Billing ID `1372` confirmed

**Test Webhook Payload:**
```bash
curl -X POST http://localhost:8080/api/v1/billings/confirm-payment \
  -H "Content-Type: application/json" \
  -d '{
    "event": "payment.received",
    "data": {
      "id": "test-invoice-001",
      "transactionId": "test-transaction-001",
      "status": "SUCCESS",
      "transactionStatus": "created",
      "createdAt": "2026-01-03T07:13:36.565Z",
      "updatedAt": "2026-01-03T07:14:00.582Z",
      "customerName": "Penghuni IPL",
      "customerEmail": "billing@ipl.com",
      "customerMobile": "08123456789",
      "amount": 165000,
      "productDescription": "1372 (DocumentID: monthly-6763f269-d01d-401f-842c-903db75936d3)",
      "paymentMethod": "QRIS"
    }
  }'
```

### Scenario 2: Multiple Billings Payment
**Product Description:** `1372,67 (DocumentID: monthly-6763f269-d01d-401f-842c-903db75936d3, monthly-527dbcbd-e923-44ef-9b6f-ecaa495d6e84)`

**Expected Result:** Billing IDs `1372` and `67` confirmed

**Test Webhook Payload:**
```bash
curl -X POST http://localhost:8080/api/v1/billings/confirm-payment \
  -H "Content-Type: application/json" \
  -d '{
    "event": "payment.received",
    "data": {
      "id": "test-invoice-002",
      "transactionId": "test-transaction-002",
      "status": "SUCCESS",
      "transactionStatus": "created",
      "createdAt": "2026-01-03T07:13:36.565Z",
      "updatedAt": "2026-01-03T07:14:00.582Z",
      "customerName": "Penghuni IPL",
      "customerEmail": "billing@ipl.com",
      "customerMobile": "08123456789",
      "amount": 330000,
      "productDescription": "1372,67 (DocumentID: monthly-6763f269-d01d-401f-842c-903db75936d3, monthly-527dbcbd-e923-44ef-9b6f-ecaa495d6e84)",
      "paymentMethod": "QRIS"
    }
  }'
```

### Scenario 3: Three Billings Payment
**Product Description:** `100,200,300 (DocumentID: doc-1, doc-2, doc-3)`

**Expected Result:** Billing IDs `100`, `200`, and `300` confirmed

**Test Webhook Payload:**
```bash
curl -X POST http://localhost:8080/api/v1/billings/confirm-payment \
  -H "Content-Type: application/json" \
  -d '{
    "event": "payment.received",
    "data": {
      "id": "test-invoice-003",
      "transactionId": "test-transaction-003",
      "status": "SUCCESS",
      "transactionStatus": "created",
      "createdAt": "2026-01-03T07:13:36.565Z",
      "updatedAt": "2026-01-03T07:14:00.582Z",
      "customerName": "Penghuni IPL",
      "customerEmail": "billing@ipl.com",
      "customerMobile": "08123456789",
      "amount": 500000,
      "productDescription": "100,200,300 (DocumentID: doc-1, doc-2, doc-3)",
      "paymentMethod": "BANK_TRANSFER"
    }
  }'
```

## Parsing Logic Verification

### Test Case 1: Extract IDs from "1372,67 (DocumentID: ...)"
```
Input: "1372,67 (DocumentID: monthly-6763f269-d01d-401f-842c-903db75936d3, monthly-527dbcbd-e923-44ef-9b6f-ecaa495d6e84)"

Step 1: Split by space
Result: ["1372,67", "(DocumentID:", "...)"]

Step 2: Take first part
Result: "1372,67"

Step 3: Split by comma
Result: ["1372", "67"]

Step 4: Convert to uint
Result: [1372, 67]
```

### Test Case 2: Extract ID from "1372 (DocumentID: ...)"
```
Input: "1372 (DocumentID: monthly-6763f269-d01d-401f-842c-903db75936d3)"

Step 1: Split by space
Result: ["1372", "(DocumentID:", "...)"]

Step 2: Take first part
Result: "1372"

Step 3: Split by comma
Result: ["1372"]

Step 4: Convert to uint
Result: [1372]
```

## Expected Response Format

### Success Response
```json
{
  "status": "success",
  "message": "Webhook received and payment confirmed",
  "data": {
    "billing_ids": [1372, 67],
    "event": "payment.received",
    "status": "SUCCESS"
  }
}
```

### Error Response - Invalid Format
```json
{
  "status": "error",
  "message": "Invalid billing ID format: abc",
  "error": "..."
}
```

### Error Response - Empty Description
```json
{
  "status": "error",
  "message": "Empty product description",
  "error": null
}
```

## Verification Steps

1. **Check Server Logs**
   - Look for: "Received Mayar payment webhook"
   - Verify: billing IDs are correctly parsed
   - Confirm: "Parsed billing IDs from webhook"

2. **Check Database**
   ```sql
   SELECT id, nominal, bulan, tahun 
   FROM billings 
   WHERE id IN (1372, 67);
   ```
   Verify the billing status has been updated

3. **Check Payment Confirmation**
   - Verify `ConfirmPayment()` service method was called
   - Check for any errors in the logs

## Full Workflow Test

### Step 1: Create Payment Link
```bash
curl -X POST http://localhost:8080/api/v1/payments/billing/link \
  -H "Content-Type: application/json" \
  -H "Cookie: auth-token=your-jwt-token" \
  -d '{"billing_ids": [1372, 67]}'
```

**Expected Response:**
```json
{
  "billing_ids": [1372, 67],
  "amount": 330000,
  "payment_url": "https://nasri-adzlani.myr.id/invoices/xxxxx",
  "description": "Payment for 2 billings"
}
```

### Step 2: Note the Product Description Format
The invoice created will have description: `1372,67 (DocumentID: monthly-xxx, monthly-yyy)`

### Step 3: Simulate Webhook (or wait for actual payment)
Use the webhook test from Scenario 2 above

### Step 4: Verify Payment Confirmed
Check that billing records 1372 and 67 are marked as paid

## Debugging Tips

### Enable Debug Logging
Check these log entries:
- "Billing IDs string:" - Should show comma-separated IDs
- "ID strings:" - Should show array of ID strings
- "Parsed billing IDs from webhook:" - Final parsed IDs

### Common Issues

**Issue:** "No valid billing IDs found"
- **Cause:** Product description format is incorrect
- **Fix:** Ensure format is `<ids> (DocumentID: ...)`

**Issue:** "Invalid billing ID format"
- **Cause:** Non-numeric characters in billing IDs
- **Fix:** Verify IDs are numbers only, separated by commas

**Issue:** "Empty product description"
- **Cause:** Webhook missing productDescription field
- **Fix:** Check webhook payload structure

## Mayar Webhook Configuration

To configure Mayar to send webhooks to your endpoint:

1. Log in to Mayar Dashboard: https://web.mayar.id
2. Go to Settings > Webhooks
3. Add webhook URL: `https://your-domain.com/api/v1/billings/confirm-payment`
4. Select event: `payment.received`
5. Save configuration

## Production Checklist

- [ ] Webhook URL configured in Mayar dashboard
- [ ] HTTPS enabled for production webhook endpoint
- [ ] Webhook signature validation implemented (future enhancement)
- [ ] Error monitoring configured
- [ ] Database backup before testing
- [ ] Rollback plan ready
- [ ] Customer notification system tested
