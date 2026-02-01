# Quick Test Guide

## ✅ Webhook Configured

Your GOWA webhook is now configured:
- **URL**: `https://unpinioned-nonceremoniously-lezlie.ngrok-free.dev/webhook`
- **Secret**: `apiku`

## 🧪 Test Now

### 1. Send a WhatsApp Message

From any WhatsApp number, send a message to your GOWA bot number:

```
Hi
```

### 2. Watch the Logs

In your terminal, watch for incoming messages:

```bash
docker-compose logs -f app
```

You should see:
- Message received
- User created/fetched
- Onboarding message sent

### 3. Expected Flow

**First Message: "Hi"**
```
You should receive:
┌─────────────────────────────────┐
│ Halo! Aku bot pencatat keuangan│
│                                 │
│ Pilih paket:                    │
│ 1️⃣ Free (10 transaksi)         │
│ 2️⃣ Premium Rp10rb/bulan        │
│                                 │
│ Ketik *1* atau *2* untuk memilih│
└─────────────────────────────────┘
```

**Reply: "1"**
```
✅ Paket Free aktif! 
Kamu bisa mencatat hingga 10 transaksi.

Contoh penggunaan:
• catat pemasukan 100000 gaji
• beli bensin 50rb
• atau kirim foto struk!
```

**Record Transaction: "catat pemasukan 100rb gaji"**
```
✅ Transaksi tersimpan!

💰 INCOME
Rp 100000 - gaji

ID: [TX_ID]
Ketik *undo* dalam 60 detik untuk membatalkan.
```

## ⚠️ If Messages Don't Send

If you see "Unauthorized" error in logs, the GOWA API token is incorrect.

**Fix:**
1. Get correct token from GOWA dashboard
2. Update `.env`:
   ```
   GOWA_API_TOKEN=your-real-token
   ```
3. Restart:
   ```bash
   docker-compose restart app
   ```

## 🔍 Debugging

### Check if webhook is receiving messages:
```bash
# Watch logs in real-time
docker-compose logs -f app

# Check last 100 lines
docker-compose logs --tail=100 app
```

### Check ngrok requests:
Open http://localhost:4040/inspect/http

### Test webhook manually:
```bash
./test_webhook.sh
```

### Check database:
```bash
docker-compose exec postgres psql -U catatuang -d catatuang -c "SELECT * FROM users;"
```

## 📊 Admin Panel

View users in real-time: http://localhost:8080

## 🎯 Test Checklist

- [ ] Send "Hi" → Get onboarding
- [ ] Reply "1" → Get confirmation
- [ ] Send "catat pemasukan 100rb gaji" → Transaction saved
- [ ] Send "rekap hari ini" → Get report
- [ ] Send "undo" → Transaction cancelled
- [ ] Send receipt image → Transaction extracted
- [ ] Check admin panel → See user listed

## 🚀 Ready!

Everything is configured. Just send a WhatsApp message to test!
