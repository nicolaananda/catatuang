package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/nicolaananda/catatuang/internal/ai"
	"github.com/nicolaananda/catatuang/internal/config"
	"github.com/nicolaananda/catatuang/internal/handler"
	"github.com/nicolaananda/catatuang/internal/repository"
	"github.com/nicolaananda/catatuang/internal/service"
	"github.com/nicolaananda/catatuang/internal/statemachine"
	"github.com/nicolaananda/catatuang/internal/whatsapp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Get timezone
	loc, err := cfg.GetLocation()
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	dedupRepo := repository.NewDedupRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	txService := service.NewTransactionService(txRepo, userRepo, auditRepo, db)
	reportService := service.NewReportService(txRepo)

	// Initialize AI parsers
	textParser := ai.NewTextParser(cfg.OpenAIAPIKey, cfg.OpenAIModel, loc)
	visionParser := ai.NewVisionParser(cfg.OpenAIAPIKey, cfg.OpenAIModel, loc)

	// Initialize WhatsApp client
	waClient := whatsapp.NewClient(cfg.GowaAPIURL, cfg.GowaAPIToken)

	// Initialize state machine
	stateMachine := statemachine.NewStateMachine(db)

	// Initialize webhook handler
	webhookHandler := handler.NewWebhookHandler(
		cfg,
		db,
		waClient,
		textParser,
		visionParser,
		userService,
		txService,
		reportService,
		stateMachine,
		dedupRepo,
		auditRepo,
	)

	// Initialize admin handler
	adminHandler := handler.NewAdminHandler(userService)

	// Setup HTTP server
	http.Handle("/webhook", webhookHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Admin API endpoints
	http.HandleFunc("/api/admin/users", adminHandler.GetUsers)
	http.HandleFunc("/api/admin/upgrade", adminHandler.UpgradeUser)
	http.HandleFunc("/api/admin/block", adminHandler.BlockUser)
	http.HandleFunc("/api/admin/unblock", adminHandler.UnblockUser)

	// Serve admin panel
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📊 Admin panel: http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
