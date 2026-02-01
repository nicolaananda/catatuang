#!/bin/bash

# Test GOWA API endpoints
# Run this on VPS to test different API formats

GOWA_URL="https://gow.nicola.id"
DEVICE_ID="default"
PHONE="6281389592985"
MESSAGE="Test from script"

echo "🧪 Testing GOWA API Endpoints"
echo "================================"
echo ""

# Test 1: /send/text with device_id query param
echo "Test 1: POST /send/text?device_id=$DEVICE_ID"
curl -X POST "$GOWA_URL/send/text?device_id=$DEVICE_ID" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE\",\"message\":\"$MESSAGE\"}" \
  -v
echo ""
echo ""

# Test 2: /send/text without device_id
echo "Test 2: POST /send/text (no device_id)"
curl -X POST "$GOWA_URL/send/text" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE\",\"message\":\"$MESSAGE\"}" \
  -v
echo ""
echo ""

# Test 3: Check if /api prefix needed
echo "Test 3: POST /api/send/text?device_id=$DEVICE_ID"
curl -X POST "$GOWA_URL/api/send/text?device_id=$DEVICE_ID" \
  -H "Content-Type: application/json" \
  -d "{\"phone\":\"$PHONE\",\"message\":\"$MESSAGE\"}" \
  -v
echo ""
echo ""

# Test 4: GET /app/devices to see available devices
echo "Test 4: GET /app/devices"
curl -X GET "$GOWA_URL/app/devices" -v
echo ""
echo ""

echo "✅ Tests complete!"
echo ""
echo "Check which endpoint returns 200 OK"
