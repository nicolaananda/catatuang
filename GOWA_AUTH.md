# GOWA Basic Auth Configuration

## Problem
GOWA API menggunakan Basic Authentication (username:password), bukan Bearer token.

## Solution

Update `.env` di VPS dengan format:

```bash
# GOWA API Token format: username:password
GOWA_API_TOKEN=admin:your-password-here
```

Atau jika username dan password sama:
```bash
GOWA_API_TOKEN=admin
```

## Test

```bash
# Test dengan Basic Auth
curl -X POST "https://gow.nicola.id/send/text?device_id=default" \
  -u "admin:your-password" \
  -H "Content-Type: application/json" \
  -d '{"phone":"6281389592985","message":"Test"}'
```

## Deploy

```bash
cd /var/www/catatuang

# Update .env dengan username:password yang benar
nano .env
# Set: GOWA_API_TOKEN=username:password

# Pull latest code
git pull origin main

# Rebuild
docker compose down
docker compose up -d --build

# Test
sleep 10
docker compose logs app | tail -20
```
