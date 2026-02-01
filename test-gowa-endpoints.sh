#!/bin/bash

# Test different GOWA endpoint paths with Basic Auth
# Run this on VPS

GOWA_URL="https://gow.nicola.id"
DEVICE_ID="default"
PHONE="6281389592985"
MESSAGE="Test"
AUTH="admin:@Nandha20"  # Update with your actual credentials

echo "🧪 Testing GOWA Endpoints with Basic Auth"
echo "=========================================="
echo ""

# Test different endpoint paths
endpoints=(
  "/send/text"
  "/api/send/text"
  "/send/message"
  "/api/send/message"
  "/message/send"
  "/api/message/send"
)

for endpoint in "${endpoints[@]}"; do
  echo "Testing: POST $endpoint?device_id=$DEVICE_ID"
  response=$(curl -s -w "\n%{http_code}" -X POST "$GOWA_URL$endpoint?device_id=$DEVICE_ID" \
    -u "$AUTH" \
    -H "Content-Type: application/json" \
    -d "{\"phone\":\"$PHONE\",\"message\":\"$MESSAGE\"}")
  
  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | head -n-1)
  
  if [ "$http_code" = "200" ]; then
    echo "✅ SUCCESS! Status: $http_code"
    echo "Response: $body"
    echo ""
    echo "🎉 Found working endpoint: $endpoint"
    exit 0
  else
    echo "❌ Failed. Status: $http_code"
    echo "Response: $body"
  fi
  echo ""
done

echo "❌ No working endpoint found!"
echo ""
echo "Try checking GOWA dashboard or documentation for correct endpoint."
