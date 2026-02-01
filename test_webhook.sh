#!/bin/bash

# Test script for WhatsApp Finance Bot webhook

echo "🔍 Getting ngrok URL..."

# Wait for ngrok to be ready
sleep 2

# Get ngrok URL from API
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels 2>/dev/null | grep -o '"public_url":"https://[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$NGROK_URL" ]; then
    echo "❌ Could not get ngrok URL. Make sure ngrok is running:"
    echo "   ngrok http 8080"
    echo ""
    echo "Or check manually at: http://localhost:4040"
    exit 1
fi

echo "✅ Ngrok URL: $NGROK_URL"
echo ""
echo "📋 GOWA Webhook Configuration:"
echo "   URL: $NGROK_URL/webhook"
echo "   Method: POST"
echo "   Header: X-Webhook-Secret: apiku"
echo ""
echo "🧪 Testing webhook endpoint..."

# Test webhook with sample message
curl -X POST "$NGROK_URL/webhook" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: apiku" \
  -H "ngrok-skip-browser-warning: true" \
  -d '{
    "message_id": "test123",
    "from": "6281389592985",
    "text": "Hi",
    "timestamp": 1738431548
  }'

echo ""
echo ""
echo "✅ Test complete! Check docker logs:"
echo "   docker-compose logs -f app"
