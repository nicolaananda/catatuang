# 🔧 VPS Troubleshooting - Docker Network Issue

## Problem

App container tidak bisa connect ke postgres:
```
Failed to ping database: dial tcp: lookup postgres on 127.0.0.11:53: no such host
```

## Quick Fix (di VPS)

### Option 1: Run Fix Script

```bash
cd /var/www/catatuang
./fix-network.sh
```

### Option 2: Manual Fix

```bash
# Stop containers
docker-compose down

# Clean networks
docker network prune -f

# Restart Docker
sudo systemctl restart docker

# Start with fresh network
docker-compose up -d --force-recreate

# Wait for postgres
sleep 10

# Check logs
docker-compose logs app | tail -20
```

## Verify Fix

```bash
# Check containers
docker-compose ps

# Check app logs (should see "Connected to database")
docker-compose logs app | grep -i "connected"

# Test health
curl http://localhost:1101/health
```

## If Still Not Working

### Check Docker Network

```bash
# List networks
docker network ls

# Inspect catatuang network
docker network inspect catatuang_default

# Check if both containers are in same network
docker inspect catatuang-app-1 | grep NetworkMode
docker inspect catatuang-postgres-1 | grep NetworkMode
```

### Check .env File

```bash
# Make sure DATABASE_URL uses "postgres" as hostname
cat .env | grep DATABASE_URL

# Should be:
# DATABASE_URL=postgresql://catatuang:catatuang123@postgres:5432/catatuang?sslmode=disable
```

### Rebuild Everything

```bash
# Nuclear option - rebuild from scratch
docker-compose down -v  # WARNING: This deletes data!
docker-compose up -d --build
sleep 10
docker-compose exec -T app ./migrate -direction=up
```

## Common Causes

1. **Old Docker version** - Update Docker
2. **Firewall blocking** - Check UFW rules
3. **DNS issues** - Restart Docker daemon
4. **Network conflicts** - Prune unused networks

## Check Docker Version

```bash
docker --version
docker-compose --version

# Should be:
# Docker version 20.10+ 
# Docker Compose version 2.0+
```

## Update Docker (if needed)

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Or use Docker's official script
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
```

## Success Indicators

When fixed, you should see:
```
✅ Connected to database
🚀 Server starting on port 8080
📊 Admin panel: http://localhost:8080
```

## Test After Fix

```bash
# Health check
curl http://localhost:1101/health
# Should return: OK

# Admin panel
curl http://localhost:1101
# Should return HTML

# Check database
docker-compose exec postgres psql -U catatuang -d catatuang -c "SELECT 1;"
# Should return: 1
```

## Still Having Issues?

Run diagnostics:

```bash
# Full diagnostic
echo "=== Docker Info ==="
docker info

echo "=== Networks ==="
docker network ls

echo "=== Containers ==="
docker-compose ps

echo "=== App Logs ==="
docker-compose logs app | tail -50

echo "=== Postgres Logs ==="
docker-compose logs postgres | tail -20

echo "=== Network Inspect ==="
docker network inspect catatuang_default
```

Send output to debug further.
