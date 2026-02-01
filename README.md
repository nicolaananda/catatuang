# WhatsApp Finance Bot (Catatuang)

Bot WhatsApp untuk mencatat keuangan pribadi menggunakan AI (OpenAI 4o-mini).

## Features

- ✅ Pencatatan transaksi via teks natural language
- ✅ Pencatatan via gambar (struk, transfer)
- ✅ Rekap harian, mingguan, bulanan
- ✅ Edit, delete, undo transaksi
- ✅ Free plan (10 transaksi) & Premium (unlimited)
- ✅ Admin panel web
- ✅ Audit trail lengkap
- ✅ Message deduplication (idempotency)

## Tech Stack

- **Backend**: Go
- **Database**: PostgreSQL
- **AI**: OpenAI 4o-mini (text & vision)
- **WhatsApp**: GOWA Engine
- **Deployment**: Docker

## Setup

### 1. Prerequisites

- Go 1.21+
- PostgreSQL 15+
- OpenAI API key
- GOWA WhatsApp instance

### 2. Environment Variables

Copy `.env.example` to `.env` and fill in:

```bash
cp .env.example .env
```

Required variables:
- `DATABASE_URL` - PostgreSQL connection string
- `OPENAI_API_KEY` - OpenAI API key
- `GOWA_WEBHOOK_SECRET` - Webhook secret for GOWA
- `GOWA_API_URL` - GOWA API URL
- `GOWA_API_TOKEN` - GOWA API token

### 3. Database Migration

```bash
go run cmd/migrate/main.go -direction=up
```

### 4. Run Locally

```bash
go run cmd/api/main.go
```

Server will start on `http://localhost:8080`

### 5. Docker Deployment

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop
docker-compose down
```

## Usage

### User Commands

**Pencatatan Transaksi:**
- `catat pemasukan 100000 gaji`
- `beli bensin 50rb`
- `dapat uang dari jual motor 20 juta`
- Kirim foto struk/transfer

**Rekap:**
- `rekap hari ini`
- `rekap minggu ini`
- `rekap bulan ini`

**Undo:**
- `undo` (dalam 60 detik setelah transaksi)

### Admin Commands (WhatsApp)

Hanya untuk nomor admin (081389592985):

- `upgrade <msisdn> monthly <dd/mm>` - Upgrade user ke premium
- `status <msisdn>` - Cek status user
- `block <msisdn>` - Block user
- `unblock <msisdn>` - Unblock user

### Admin Panel (Web)

Akses: `http://localhost:8080`

Features:
- View all users
- Upgrade/downgrade users
- Block/unblock users
- View statistics

## API Endpoints

- `POST /webhook` - GOWA webhook (requires X-Webhook-Secret header)
- `GET /health` - Health check
- `GET /api/admin/users` - Get all users
- `POST /api/admin/upgrade` - Upgrade user
- `POST /api/admin/block` - Block user
- `POST /api/admin/unblock` - Unblock user

## Project Structure

```
catatuang/
├── cmd/
│   ├── api/           # Main API server
│   └── migrate/       # DB migration tool
├── internal/
│   ├── config/        # Configuration
│   ├── domain/        # Domain models
│   ├── repository/    # Data access
│   ├── service/       # Business logic
│   ├── handler/       # HTTP handlers
│   ├── ai/            # OpenAI integration
│   ├── whatsapp/      # GOWA integration
│   └── statemachine/  # Conversation state
├── web/               # Admin panel
├── migrations/        # SQL migrations
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o bin/api cmd/api/main.go
go build -o bin/migrate cmd/migrate/main.go
```

## Production Deployment

1. Set up PostgreSQL database
2. Configure environment variables
3. Run migrations
4. Deploy with Docker:

```bash
docker build -t catatuang:latest .
docker run -d \
  --name catatuang \
  -p 8080:8080 \
  --env-file .env \
  catatuang:latest
```

5. Configure GOWA webhook to point to `https://your-domain.com/webhook`

## License

MIT
