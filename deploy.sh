#!/bin/bash

# Production Deployment Script for WhatsApp Finance Bot
# Run this script on your production server

set -e  # Exit on error

echo "🚀 WhatsApp Finance Bot - Production Deployment"
echo "================================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
    echo -e "${RED}❌ Please do not run as root${NC}"
    exit 1
fi

# Step 1: Clone repository
echo -e "${YELLOW}📦 Step 1: Cloning repository...${NC}"
if [ ! -d "catatuang" ]; then
    git clone git@github.com:nicolaananda/catatuang.git
    cd catatuang
else
    cd catatuang
    git pull origin main
fi

# Step 2: Check .env file
echo -e "${YELLOW}⚙️  Step 2: Checking environment configuration...${NC}"
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ .env file not found!${NC}"
    echo "Creating .env from example..."
    cp .env.example .env
    echo ""
    echo -e "${YELLOW}⚠️  IMPORTANT: Edit .env file with production values:${NC}"
    echo "   - DATABASE_URL (use strong password)"
    echo "   - OPENAI_API_KEY"
    echo "   - GOWA_WEBHOOK_SECRET (use secure secret)"
    echo "   - GOWA_API_TOKEN"
    echo ""
    echo "Run: nano .env"
    echo ""
    read -p "Press Enter after editing .env file..."
fi

# Step 3: Check Docker
echo -e "${YELLOW}🐳 Step 3: Checking Docker...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker not installed!${NC}"
    echo "Install Docker: https://docs.docker.com/engine/install/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose not installed!${NC}"
    echo "Install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✅ Docker and Docker Compose found${NC}"

# Step 4: Build and start containers
echo -e "${YELLOW}🔨 Step 4: Building and starting containers...${NC}"
docker-compose down 2>/dev/null || true
docker-compose up -d --build

# Step 5: Wait for PostgreSQL
echo -e "${YELLOW}⏳ Step 5: Waiting for PostgreSQL to be ready...${NC}"
sleep 10

# Check if postgres is ready
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U catatuang &>/dev/null; then
        echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
        break
    fi
    echo "Waiting for PostgreSQL... ($i/30)"
    sleep 2
done

# Step 6: Run migrations
echo -e "${YELLOW}📊 Step 6: Running database migrations...${NC}"
docker-compose exec -T app ./migrate -direction=up

# Step 7: Verify deployment
echo -e "${YELLOW}🧪 Step 7: Verifying deployment...${NC}"

# Check containers
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Containers are running${NC}"
else
    echo -e "${RED}❌ Containers failed to start${NC}"
    docker-compose logs
    exit 1
fi

# Check health endpoint
sleep 3
if curl -s http://localhost:8080/health | grep -q "OK"; then
    echo -e "${GREEN}✅ Health check passed${NC}"
else
    echo -e "${RED}❌ Health check failed${NC}"
    docker-compose logs app
    exit 1
fi

# Step 8: Display info
echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}🎉 Deployment Successful!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo "📊 Service Status:"
docker-compose ps
echo ""
echo "🌐 Endpoints:"
echo "   - Health: http://localhost:8080/health"
echo "   - Admin Panel: http://localhost:8080"
echo "   - Webhook: http://localhost:8080/webhook"
echo ""
echo "📝 Next Steps:"
echo "   1. Configure Nginx reverse proxy (see DEPLOYMENT.md)"
echo "   2. Set up SSL with Certbot"
echo "   3. Update GOWA webhook URL to your domain"
echo "   4. Test WhatsApp flow"
echo ""
echo "📊 View Logs:"
echo "   docker-compose logs -f app"
echo ""
echo "🛑 Stop Services:"
echo "   docker-compose down"
echo ""
echo -e "${GREEN}✅ Ready for production!${NC}"
