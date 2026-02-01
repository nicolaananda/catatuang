# Testing Guide - WhatsApp Finance Bot

## ✅ Docker Setup Complete

Your bot is now running locally with Docker!

- **API Server**: http://localhost:8080
- **Admin Panel**: http://localhost:8080
- **Database**: PostgreSQL on localhost:5432
- **Webhook Endpoint**: http://localhost:8080/webhook

## 🧪 Testing Steps

### 1. Test Admin Panel

Open your browser:
```bash
open http://localhost:8080
```

You should see the admin dashboard with user statistics.

### 2. Configure GOWA Webhook

Login to GOWA at https://gow.nicola.id:
- Username: `admin`
- Password: `@Nandha20`

Configure webhook:
- **URL**: `http://your-public-url/webhook` (use ngrok for local testing)
- **Method**: POST
- **Header**: `X-Webhook-Secret: apiku`

### 3. Expose Local Server (for WhatsApp testing)

Use ngrok to expose your local server:

```bash
# Install ngrok if not already installed
brew install ngrok

# Expose port 8080
ngrok http 8080
```

Copy the ngrok URL (e.g., `https://abc123.ngrok.io`) and set it as the webhook URL in GOWA:
```
https://abc123.ngrok.io/webhook
```

### 4. Test WhatsApp Flow

Send messages from your WhatsApp to the bot number:

**New User Onboarding:**
```
Hi
```
Expected: Bot asks you to choose plan (1 for Free, 2 for Premium)

**Choose Free Plan:**
```
1
```
Expected: Confirmation message with usage examples

**Record Transaction (Text):**
```
catat pemasukan 100000 gaji
```
Expected: Transaction saved with TX ID

**Record Transaction (Indonesian Slang):**
```
beli bensin 50rb
```
Expected: Transaction saved (50,000 rupiah)

**Send Receipt Image:**
Send a photo of a receipt or bank transfer
Expected: Bot extracts amount and saves transaction

**View Daily Report:**
```
rekap hari ini
```
Expected: Summary of today's transactions

**Undo Last Transaction:**
```
undo
```
Expected: Last transaction cancelled (within 60 seconds)

### 5. Test Admin Commands

From admin WhatsApp (081389592985):

**Check User Status:**
```
status 6281234567890
```

**Upgrade User to Premium:**
```
upgrade 6281234567890 monthly 01/02
```
Expected: User upgraded, receives confirmation message

### 6. Monitor Logs

Watch real-time logs:
```bash
docker-compose logs -f app
```

### 7. Test API Endpoints

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Get All Users (Admin API):**
```bash
curl http://localhost:8080/api/admin/users
```

**Upgrade User (Admin API):**
```bash
curl -X POST http://localhost:8080/api/admin/upgrade \
  -H "Content-Type: application/json" \
  -d '{"msisdn":"6281234567890","start_date":"01/02"}'
```

## 🔍 Troubleshooting

### Container not starting
```bash
docker-compose logs app
docker-compose logs postgres
```

### Database connection issues
```bash
# Check if postgres is ready
docker-compose exec postgres pg_isready

# Connect to database
docker-compose exec postgres psql -U catatuang -d catatuang
```

### Reset everything
```bash
docker-compose down -v
docker-compose up -d --build
docker-compose exec -T app ./migrate -direction=up
```

## 📊 Verify Database

Connect to PostgreSQL:
```bash
docker-compose exec postgres psql -U catatuang -d catatuang
```

Check tables:
```sql
\dt
SELECT * FROM users;
SELECT * FROM transactions;
SELECT * FROM conversation_states;
```

## 🎯 Expected Behavior

1. **New User**: Gets onboarding message
2. **Free User**: Can record up to 10 transactions
3. **Premium User**: Unlimited transactions
4. **AI Parsing**: Confidence-based auto-save/confirm/reject
5. **Undo**: Works within 60 seconds
6. **Reports**: Daily/weekly/monthly summaries
7. **Admin**: Can upgrade users, check status

## 🐛 Common Issues

**Issue**: Webhook not receiving messages
**Solution**: Make sure ngrok is running and GOWA webhook URL is correct

**Issue**: AI parsing fails
**Solution**: Check OpenAI API key in `.env`

**Issue**: Database connection refused
**Solution**: Wait a few seconds for postgres to be ready, or restart containers

**Issue**: Free limit not enforcing
**Solution**: Check `free_tx_count` in database

## 📝 Next Steps

1. ✅ Test complete user journey
2. ✅ Verify free limit enforcement
3. ✅ Test admin upgrade flow
4. ✅ Check audit logs
5. Deploy to production VPS
6. Add monitoring/alerts

## 🛑 Stop Containers

```bash
docker-compose down
```

## 🔄 Restart Containers

```bash
docker-compose restart
```

## 📦 View Container Status

```bash
docker-compose ps
docker-compose logs -f
```
