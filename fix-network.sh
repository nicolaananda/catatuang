#!/bin/bash

# Fix Docker Network Issue on VPS
# Run this on your VPS server

echo "🔧 Fixing Docker Network Issue..."
echo ""

# Step 1: Stop containers
echo "1️⃣ Stopping containers..."
docker-compose down

# Step 2: Remove old network
echo "2️⃣ Cleaning up networks..."
docker network prune -f

# Step 3: Restart Docker daemon (optional but helps)
echo "3️⃣ Restarting Docker daemon..."
sudo systemctl restart docker
sleep 3

# Step 4: Start containers with fresh network
echo "4️⃣ Starting containers with fresh network..."
docker-compose up -d --force-recreate

# Step 5: Wait for postgres to be ready
echo "5️⃣ Waiting for PostgreSQL..."
sleep 10

# Step 6: Check status
echo "6️⃣ Checking status..."
docker-compose ps

# Step 7: Check logs
echo ""
echo "📊 App logs:"
docker-compose logs app | tail -10

echo ""
echo "📊 Postgres logs:"
docker-compose logs postgres | tail -5

# Step 8: Test connection
echo ""
echo "🧪 Testing health endpoint..."
sleep 3
curl -s http://localhost:1101/health || echo "Health check failed"

echo ""
echo "✅ Done! Check logs above for status."
