#!/bin/bash

# Fix Docker Network Issue on VPS (Safe - Only affects catatuang)
# Run this on your VPS server

echo "🔧 Fixing Docker Network Issue for catatuang..."
echo ""
echo "⚠️  This will only restart catatuang containers"
echo "    Other Docker apps (sigantara, catsflix, etc) will NOT be affected"
echo ""

# Step 1: Stop catatuang containers
echo "1️⃣ Stopping catatuang containers..."
docker-compose down

# Step 2: Remove ONLY catatuang network (safe)
echo "2️⃣ Removing catatuang network..."
docker network rm catatuang_default 2>/dev/null || echo "Network already removed"

# Step 3: Start containers with fresh network
echo "3️⃣ Starting containers with fresh network..."
docker compose up -d --force-recreate

# Step 4: Wait for postgres to be ready
echo "4️⃣ Waiting for PostgreSQL..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U catatuang &>/dev/null; then
        echo "✅ PostgreSQL is ready"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 2
done

# Step 5: Check status
echo ""
echo "5️⃣ Checking status..."
docker-compose ps

# Step 6: Check logs
echo ""
echo "📊 App logs:"
docker-compose logs app | tail -10

# Step 7: Test connection
echo ""
echo "🧪 Testing health endpoint..."
sleep 3
if curl -s http://localhost:1101/health | grep -q "OK"; then
    echo "✅ Health check PASSED"
else
    echo "❌ Health check FAILED"
    echo ""
    echo "Run this to see full logs:"
    echo "  docker-compose logs app"
fi

echo ""
echo "✅ Done!"
