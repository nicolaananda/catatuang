# 🔧 Port 5432 Conflict - Quick Fix

## Problem
```
failed to bind host port 0.0.0.0:5432/tcp: address already in use
```

Port 5432 sudah dipakai oleh PostgreSQL lain di VPS.

## ✅ Solution (Already Fixed in Latest Code)

Port 5432 sudah di-remove dari `docker-compose.yml`. 

**Kenapa aman?**
- App container connect ke postgres via **Docker network** (internal)
- Tidak perlu expose port 5432 ke host
- Tidak akan conflict dengan PostgreSQL lain

## 🚀 Di VPS, Jalankan:

```bash
cd /var/www/catatuang

# Pull latest fix
git pull origin main

# Restart dengan config baru
docker compose down
docker compose up -d

# Wait for postgres
sleep 10

# Check logs
docker compose logs app | tail -20
```

## ✅ Expected Result

Seharusnya sekarang berhasil:
```
✅ Connected to database
🚀 Server starting on port 8080
```

## 🧪 Test

```bash
curl http://localhost:1101/health
# Should return: OK
```

## 📝 Technical Details

**Before:**
```yaml
postgres:
  ports:
    - "5432:5432"  # ❌ Conflict dengan PostgreSQL lain
```

**After:**
```yaml
postgres:
  # Port tidak di-expose
  # App connect via Docker network: postgres:5432
```

**Connection:**
- External tools: ❌ Tidak bisa connect dari luar
- App container: ✅ Bisa connect via `postgres:5432`

## 🔍 Jika Perlu Akses Database dari Luar

Gunakan port lain, misalnya 5433:

```yaml
postgres:
  ports:
    - "5433:5432"  # Expose di port 5433
```

Tapi untuk production, **tidak recommended** expose database port.
