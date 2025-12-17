#!/bin/bash
# Test script for Pismo API endpoints

BASE_URL="http://localhost:8080"

echo "═══════════════════════════════════════════════════"
echo "                  PISMO API TESTS                  "
echo "═══════════════════════════════════════════════════"

# 1. Health Check
echo ""
echo "🏥 GET /healthz"
echo "───────────────────────────────────────────────────"
curl -s "$BASE_URL/healthz" | jq . 2>/dev/null || curl -s "$BASE_URL/healthz"
echo ""

# 2. Create Account
echo ""
echo "👤 POST /accounts (Create Account)"
echo "───────────────────────────────────────────────────"
ACCOUNT_RESPONSE=$(curl -s -X POST "$BASE_URL/accounts" \
  -H "Content-Type: application/json" \
  -d '{"document_number": "12345678900"}')
echo "$ACCOUNT_RESPONSE" | jq . 2>/dev/null || echo "$ACCOUNT_RESPONSE"
ACCOUNT_ID=$(echo "$ACCOUNT_RESPONSE" | jq -r '.account_id // 1' 2>/dev/null || echo "1")
echo ""

# 3. Get Account
echo ""
echo "🔍 GET /accounts/{id} (Get Account)"
echo "───────────────────────────────────────────────────"
curl -s "$BASE_URL/accounts/$ACCOUNT_ID" | jq . 2>/dev/null || curl -s "$BASE_URL/accounts/$ACCOUNT_ID"
echo ""

# 4. Create Transaction - Purchase (Debit)
echo ""
echo "💳 POST /transactions (Purchase - Debit)"
echo "───────────────────────────────────────────────────"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -d "{\"account_id\": $ACCOUNT_ID, \"operation_type_id\": 1, \"amount\": 50.00}" | jq . 2>/dev/null || \
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -d "{\"account_id\": $ACCOUNT_ID, \"operation_type_id\": 1, \"amount\": 50.00}"
echo ""

# 5. Create Transaction - Payment (Credit)
echo ""
echo "💰 POST /transactions (Payment - Credit)"
echo "───────────────────────────────────────────────────"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -d "{\"account_id\": $ACCOUNT_ID, \"operation_type_id\": 4, \"amount\": 100.00}" | jq . 2>/dev/null || \
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -d "{\"account_id\": $ACCOUNT_ID, \"operation_type_id\": 4, \"amount\": 100.00}"
echo ""

# 6. Metrics (just check status)
echo ""
echo "📊 GET /metrics (Prometheus)"
echo "───────────────────────────────────────────────────"
METRICS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/metrics")
echo "Status: $METRICS_STATUS (OK if 200)"
echo ""

echo "═══════════════════════════════════════════════════"
echo "                  ✅ TESTS COMPLETE                "
echo "═══════════════════════════════════════════════════"
