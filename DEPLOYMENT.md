# Production Deployment Guide

## 📦 Push to GitHub

### 1. Initialize Git (if not already done)

```bash
cd /Users/nicolaanandadwiervantoro/SE/catatuang

# Initialize git
git init

# Add remote
git remote add origin git@github.com:nicolaananda/catatuang.git

# Add all files
git add .

# Commit
git commit -m "Initial commit: WhatsApp Finance Bot with Docker"

# Push to main branch
git push -u origin main
```

### 2. Verify .gitignore

Make sure `.env` is ignored (already in .gitignore):
```
.env
bin/
*.log
```

## 🚀 Deploy to Server

### Prerequisites on Server

- Docker & Docker Compose installed
- Git installed
- Domain/subdomain pointing to server
- SSL certificate (optional but recommended)

### Step-by-Step Deployment

#### 1. SSH to Your Server

```bash
ssh user@your-server.com
```

#### 2. Clone Repository

```bash
cd /var/www  # or your preferred directory
git clone git@github.com:nicolaananda/catatuang.git
cd catatuang
```

#### 3. Create Production .env

```bash
cp .env.example .env
nano .env
```

**Fill in production values:**
```bash
# Database (use strong password!)
DATABASE_URL=postgresql://catatuang:STRONG_PASSWORD_HERE@postgres:5432/catatuang?sslmode=disable

# OpenAI
OPENAI_API_KEY=sk-proj-your-real-key-here
OPENAI_MODEL=gpt-4o-mini

# GOWA WhatsApp
GOWA_WEBHOOK_SECRET=your-secure-secret-here
GOWA_API_URL=https://gow.nicola.id
GOWA_API_TOKEN=your-real-gowa-token

# Server
PORT=8080
ADMIN_PANEL_PORT=8081

# Admin
ADMIN_MSISDN=081389592985

# App
TIMEZONE=Asia/Jakarta
AI_TIMEOUT_SECONDS=12
AI_MAX_RETRIES=2
STATE_EXPIRY_MINUTES=30
UNDO_WINDOW_SECONDS=60
FREE_TRANSACTION_LIMIT=10
```

#### 4. Update docker-compose.yml for Production

Edit `docker-compose.yml` to use production database password:

```bash
nano docker-compose.yml
```

Update postgres password to match your .env:
```yaml
postgres:
  environment:
    - POSTGRES_PASSWORD=STRONG_PASSWORD_HERE  # Same as in DATABASE_URL
```

#### 5. Start Services

```bash
# Build and start containers
docker-compose up -d --build

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

#### 6. Run Database Migration

```bash
# Wait for postgres to be ready (5-10 seconds)
sleep 10

# Run migration
docker-compose exec -T app ./migrate -direction=up

# Verify
docker-compose exec postgres psql -U catatuang -d catatuang -c "\dt"
```

#### 7. Test Server

```bash
# Health check
curl http://localhost:8080/health

# Should return: OK
```

### 🌐 Configure Nginx Reverse Proxy

#### 1. Create Nginx Config

```bash
sudo nano /etc/nginx/sites-available/catatuang
```

**Basic HTTP config:**
```nginx
server {
    listen 80;
    server_name catatuang.nicola.id;  # Your domain

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Webhook endpoint
    location /webhook {
        proxy_pass http://localhost:8080/webhook;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Increase timeout for AI processing
        proxy_read_timeout 30s;
        proxy_connect_timeout 30s;
    }
}
```

#### 2. Enable Site

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/catatuang /etc/nginx/sites-enabled/

# Test config
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

#### 3. Add SSL with Certbot (Recommended)

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d catatuang.nicola.id

# Auto-renewal is configured automatically
```

### 🔧 Configure GOWA Webhook

Update GOWA webhook to production URL:

**Login to GOWA**: https://gow.nicola.id
- Username: `admin`
- Password: `@Nandha20`

**Set Webhook:**
- **URL**: `https://catatuang.nicola.id/webhook`
- **Method**: `POST`
- **Header**: `X-Webhook-Secret: your-secure-secret-here`

### 🧪 Test Production

#### 1. Test Health Endpoint

```bash
curl https://catatuang.nicola.id/health
```

#### 2. Test Admin Panel

Open browser: `https://catatuang.nicola.id`

#### 3. Test WhatsApp Flow

Send message to bot:
1. `Hi` → Onboarding
2. `1` → Free plan
3. `catat pemasukan 100rb gaji` → Transaction

### 📊 Monitor Production

#### View Logs

```bash
# All logs
docker-compose logs -f

# Just app
docker-compose logs -f app

# Last 100 lines
docker-compose logs --tail=100 app
```

#### Check Container Status

```bash
docker-compose ps
```

#### Check Database

```bash
docker-compose exec postgres psql -U catatuang -d catatuang

# In psql:
\dt                                    # List tables
SELECT COUNT(*) FROM users;            # Count users
SELECT * FROM transactions LIMIT 10;   # View transactions
\q                                     # Quit
```

### 🔄 Update Deployment

When you push updates to GitHub:

```bash
# On server
cd /var/www/catatuang

# Pull latest
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose up -d --build

# Run migrations if needed
docker-compose exec -T app ./migrate -direction=up
```

### 🛡️ Security Checklist

- [ ] Use strong database password
- [ ] Use secure webhook secret
- [ ] Enable SSL/HTTPS
- [ ] Restrict admin panel access (add auth)
- [ ] Set up firewall (UFW)
- [ ] Regular backups of database
- [ ] Monitor logs for suspicious activity

### 💾 Database Backup

```bash
# Backup
docker-compose exec postgres pg_dump -U catatuang catatuang > backup_$(date +%Y%m%d).sql

# Restore
docker-compose exec -T postgres psql -U catatuang catatuang < backup_20260202.sql
```

### 🔥 Firewall Setup (UFW)

```bash
# Allow SSH
sudo ufw allow 22

# Allow HTTP/HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Enable firewall
sudo ufw enable
```

### 📈 Production Monitoring

Consider adding:
- **Uptime monitoring**: UptimeRobot, Pingdom
- **Error tracking**: Sentry
- **Logging**: ELK stack, Loki
- **Metrics**: Prometheus + Grafana

### 🚨 Troubleshooting

#### Container won't start
```bash
docker-compose logs app
docker-compose logs postgres
```

#### Database connection failed
```bash
# Check postgres is ready
docker-compose exec postgres pg_isready

# Check connection string in .env
```

#### Webhook not receiving
```bash
# Check nginx logs
sudo tail -f /var/log/nginx/error.log

# Check app logs
docker-compose logs -f app
```

#### Out of memory
```bash
# Check resources
docker stats

# Restart containers
docker-compose restart
```

## 📝 Quick Commands Reference

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# Restart
docker-compose restart

# Rebuild
docker-compose up -d --build

# Logs
docker-compose logs -f app

# Shell into container
docker-compose exec app sh

# Database shell
docker-compose exec postgres psql -U catatuang -d catatuang
```

## ✅ Deployment Complete!

Your WhatsApp Finance Bot is now running in production! 🎉
