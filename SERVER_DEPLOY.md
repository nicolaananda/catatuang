#!/bin/bash

# Server Deployment Commands for catat.nicola.id
# Run these commands on your production server

echo "🚀 Deploying WhatsApp Finance Bot to catat.nicola.id"
echo "====================================================="
echo ""

# Step 1: Clone repository
echo "📦 Step 1: Clone repository"
echo "cd /var/www"
echo "git clone git@github.com:nicolaananda/catatuang.git"
echo "cd catatuang"
echo ""

# Step 2: Setup environment
echo "⚙️  Step 2: Setup environment"
echo "cp .env.example .env"
echo "nano .env  # Edit with production values"
echo ""
echo "Required values:"
echo "  - DATABASE_URL (strong password!)"
echo "  - OPENAI_API_KEY"
echo "  - GOWA_WEBHOOK_SECRET"
echo "  - GOWA_API_TOKEN"
echo ""

# Step 3: Start Docker services
echo "🐳 Step 3: Start Docker services"
echo "docker-compose up -d --build"
echo ""

# Step 4: Run migrations
echo "📊 Step 4: Run database migrations"
echo "sleep 10  # Wait for postgres"
echo "docker-compose exec -T app ./migrate -direction=up"
echo ""

# Step 5: Test service
echo "🧪 Step 5: Test service"
echo "curl http://localhost:1101/health"
echo "# Should return: OK"
echo ""

# Step 6: Configure Nginx
echo "🌐 Step 6: Configure Nginx"
echo "sudo cp nginx/catat.nicola.id.conf /etc/nginx/sites-available/catat.nicola.id"
echo "sudo ln -s /etc/nginx/sites-available/catat.nicola.id /etc/nginx/sites-enabled/"
echo "sudo nginx -t"
echo "sudo systemctl reload nginx"
echo ""

# Step 7: Install SSL certificate
echo "🔒 Step 7: Install SSL certificate"
echo "sudo certbot --nginx -d catat.nicola.id"
echo ""

# Step 8: Update GOWA webhook
echo "🔧 Step 8: Update GOWA webhook"
echo "Login to: https://gow.nicola.id"
echo "Set webhook URL: https://catat.nicola.id/webhook"
echo "Set header: X-Webhook-Secret: [your-secret]"
echo ""

# Step 9: Test production
echo "✅ Step 9: Test production"
echo "curl https://catat.nicola.id/health"
echo "open https://catat.nicola.id  # Admin panel"
echo ""

# Monitoring
echo "📊 Monitoring commands:"
echo "docker-compose logs -f app"
echo "docker-compose ps"
echo "sudo tail -f /var/log/nginx/catat.nicola.id.access.log"
echo ""

echo "🎉 Deployment complete!"
