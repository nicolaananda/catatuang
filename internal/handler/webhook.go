package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nicolaananda/catatuang/internal/ai"
	"github.com/nicolaananda/catatuang/internal/config"
	"github.com/nicolaananda/catatuang/internal/domain"
	"github.com/nicolaananda/catatuang/internal/repository"
	"github.com/nicolaananda/catatuang/internal/service"
	"github.com/nicolaananda/catatuang/internal/statemachine"
	"github.com/nicolaananda/catatuang/internal/whatsapp"
)

type WebhookHandler struct {
	cfg           *config.Config
	db            *sql.DB
	waClient      *whatsapp.Client
	textParser    *ai.TextParser
	visionParser  *ai.VisionParser
	userService   *service.UserService
	txService     *service.TransactionService
	reportService *service.ReportService
	stateMachine  *statemachine.StateMachine
	dedupRepo     *repository.DedupRepository
	auditRepo     *repository.AuditRepository
}

func NewWebhookHandler(
	cfg *config.Config,
	db *sql.DB,
	waClient *whatsapp.Client,
	textParser *ai.TextParser,
	visionParser *ai.VisionParser,
	userService *service.UserService,
	txService *service.TransactionService,
	reportService *service.ReportService,
	stateMachine *statemachine.StateMachine,
	dedupRepo *repository.DedupRepository,
	auditRepo *repository.AuditRepository,
) *WebhookHandler {
	return &WebhookHandler{
		cfg:           cfg,
		db:            db,
		waClient:      waClient,
		textParser:    textParser,
		visionParser:  visionParser,
		userService:   userService,
		txService:     txService,
		reportService: reportService,
		stateMachine:  stateMachine,
		dedupRepo:     dedupRepo,
		auditRepo:     auditRepo,
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body first (needed for signature verification)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	// GOWA uses X-Hub-Signature-256 header with HMAC SHA256
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature != "" {
		// Verify HMAC signature
		if !verifyHMACSignature(body, signature, h.cfg.GowaWebhookSecret) {
			log.Printf("Webhook HMAC signature verification failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	} else {
		// Fallback: check for simple secret headers
		secret := r.Header.Get("X-Webhook-Secret")
		if secret == "" {
			secret = r.Header.Get("X-Api-Key")
		}
		if secret == "" {
			secret = r.Header.Get("Authorization")
			if strings.HasPrefix(secret, "Bearer ") {
				secret = strings.TrimPrefix(secret, "Bearer ")
			}
		}

		if secret != h.cfg.GowaWebhookSecret {
			log.Printf("Webhook auth failed. Expected: %s, Got: %s", h.cfg.GowaWebhookSecret, secret)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Debug: log the payload to see what GOWA actually sends
	log.Printf("Webhook payload: %s", string(body))

	var msg whatsapp.IncomingMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Failed to unmarshal webhook payload: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process message asynchronously
	go h.processMessage(context.Background(), &msg)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *WebhookHandler) processMessage(ctx context.Context, msg *whatsapp.IncomingMessage) {
	// Check deduplication
	processed, err := h.dedupRepo.IsProcessed(ctx, msg.GetMessageID())
	if err != nil {
		log.Printf("Failed to check dedup: %v", err)
		return
	}
	if processed {
		log.Printf("Message already processed: %s", msg.GetMessageID())
		return
	}

	// Mark as processed
	if err := h.dedupRepo.MarkProcessed(ctx, msg.GetMessageID()); err != nil {
		log.Printf("Failed to mark processed: %v", err)
		return
	}

	// Get or create user
	user, isNew, err := h.userService.GetOrCreateUser(ctx, msg.GetFrom())
	if err != nil {
		log.Printf("Failed to get/create user: %v", err)
		h.sendMessage(msg.GetFrom(), "Maaf, terjadi kesalahan sistem 😔")
		return
	}

	// Check if user is blocked
	if user.IsBlocked {
		h.sendMessage(msg.GetFrom(), "Akun Anda diblokir. Hubungi admin untuk informasi lebih lanjut.")
		return
	}

	// Handle admin commands
	if msg.GetFrom() == h.cfg.AdminMSISDN {
		if h.handleAdminCommand(ctx, msg) {
			return
		}
	}

	// Handle new user onboarding
	if isNew {
		h.handleOnboarding(ctx, user, msg)
		return
	}

	// Get conversation state
	state, err := h.stateMachine.GetState(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to get state: %v", err)
		state = &domain.ConversationState{State: domain.StateActive}
	}

	// Route based on state
	switch state.State {
	case domain.StateOnboardingSelectPlan:
		h.handlePlanSelection(ctx, user, msg)
	case domain.StateActive:
		h.handleActiveState(ctx, user, msg)
	default:
		h.handleActiveState(ctx, user, msg)
	}
}

func (h *WebhookHandler) handleOnboarding(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	// Set state to onboarding
	h.stateMachine.SetState(ctx, user.ID, domain.StateOnboardingSelectPlan, nil, h.cfg.StateExpiryMinutes)

	onboardingMsg := `Halo! Aku bot pencatat keuangan 📒

Pilih paket:
1️⃣ Free (10 transaksi)
2️⃣ Premium Rp10rb/bulan (hubungi admin 081389592985)

Ketik *1* atau *2* untuk memilih.`

	h.sendMessage(msg.GetFrom(), onboardingMsg)
}

func (h *WebhookHandler) handlePlanSelection(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	text := strings.TrimSpace(msg.GetText())

	if text == "1" {
		user.Plan = domain.PlanFree
		h.userService.GetOrCreateUser(ctx, user.MSISDN) // This will update
		h.stateMachine.ClearState(ctx, user.ID)
		h.sendMessage(msg.GetFrom(), "✅ Paket Free aktif! Kamu bisa mencatat hingga 10 transaksi.\n\nContoh penggunaan:\n• catat pemasukan 100000 gaji\n• beli bensin 50rb\n• atau kirim foto struk!")
	} else if text == "2" {
		user.Plan = domain.PlanPendingPremium
		h.userService.GetOrCreateUser(ctx, user.MSISDN)
		h.stateMachine.ClearState(ctx, user.ID)
		h.sendMessage(msg.GetFrom(), "📞 Silakan hubungi admin di 081389592985 untuk upgrade ke Premium.\n\nSementara itu, kamu bisa pakai paket Free (10 transaksi).")
	} else {
		h.sendMessage(msg.GetFrom(), "Pilihan tidak valid. Ketik *1* untuk Free atau *2* untuk Premium.")
	}
}

func (h *WebhookHandler) handleActiveState(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	text := strings.ToLower(strings.TrimSpace(msg.GetText()))

	// Check for report requests
	if strings.Contains(text, "rekap") || strings.Contains(text, "laporan") {
		h.handleReportRequest(ctx, user, msg)
		return
	}

	// Check for undo
	if text == "undo" || text == "batal" {
		h.handleUndo(ctx, user, msg)
		return
	}

	// Handle image
	if msg.IsImage() {
		h.handleImageTransaction(ctx, user, msg)
		return
	}

	// Handle text transaction
	if msg.IsText() && ai.ShouldTriggerParsing(text) {
		h.handleTextTransaction(ctx, user, msg)
		return
	}

	// Default help message
	h.sendMessage(msg.GetFrom(), `Aku bisa bantu kamu:
• Catat transaksi: "catat pemasukan 100rb gaji"
• Kirim foto struk
• Lihat rekap: "rekap hari ini", "rekap bulan ini"
• Undo transaksi terakhir: "undo"`)
}

func (h *WebhookHandler) handleTextTransaction(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	// Parse with AI
	parsed, err := ai.WithRetry(ctx, ai.RetryConfig{
		MaxRetries: h.cfg.AIMaxRetries,
		Delay:      2 * time.Second,
	}, func(ctx context.Context) (*domain.ParsedTransaction, error) {
		return h.textParser.Parse(ctx, msg.GetText())
	})

	if err != nil {
		log.Printf("AI parsing failed: %v", err)
		h.sendMessage(msg.GetFrom(), "Maaf, aku belum bisa memahami pesan ini 😅\n\nContoh: catat pemasukan 100000 gaji")
		return
	}

	// Check confidence
	if parsed.ShouldReject() {
		h.sendMessage(msg.GetFrom(), "Aku kurang yakin dengan transaksi ini 🤔\n\nCoba tulis lebih jelas, contoh:\n• catat pemasukan 100000 gaji\n• beli bensin 50rb")
		return
	}

	if parsed.NeedsConfirmation() {
		// TODO: Implement confirmation flow
		h.sendMessage(msg.GetFrom(), fmt.Sprintf("Konfirmasi transaksi:\n%s Rp%.0f - %s\n\nKetik *ya* untuk simpan atau *tidak* untuk batal.",
			parsed.Type, parsed.Amount, parsed.Description))
		return
	}

	// Auto-save (high confidence)
	tx, err := h.txService.RecordTransaction(ctx, user, parsed, msg.GetMessageID(), h.cfg.OpenAIModel, h.cfg.FreeTransactionLimit)
	if err != nil {
		if strings.Contains(err.Error(), "free limit") {
			h.sendMessage(msg.GetFrom(), "❌ Limit free sudah habis (10 transaksi).\n\nUpgrade ke Premium? Hubungi admin 081389592985")
		} else {
			log.Printf("Failed to record transaction: %v", err)
			h.sendMessage(msg.GetFrom(), "Maaf, gagal menyimpan transaksi 😔")
		}
		return
	}

	emoji := "💰"
	if parsed.Type == domain.TypeExpense {
		emoji = "💸"
	}

	h.sendMessage(msg.GetFrom(), fmt.Sprintf("✅ Transaksi tersimpan!\n\n%s %s\nRp %.0f - %s\n\nID: %s\nKetik *undo* dalam 60 detik untuk membatalkan.",
		emoji, parsed.Type, parsed.Amount, parsed.Description, tx.TxID))
}

func (h *WebhookHandler) handleImageTransaction(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	// TODO: Implement image handling for GOWA format
	// GOWA sends media in different format than expected
	h.sendMessage(msg.GetFrom(), "Maaf, fitur gambar sedang dalam pengembangan 🚧")

	// Parse with vision AI
	parsed, err := ai.WithRetry(ctx, ai.RetryConfig{
		MaxRetries: h.cfg.AIMaxRetries,
		Delay:      2 * time.Second,
	}, func(ctx context.Context) (*domain.ParsedTransaction, error) {
		// Assuming imageData is available from msg.GetImage() or similar
		// This part of the code was incomplete/incorrect in the original
		// For now, we'll assume imageData is a placeholder.
		// In a real scenario, you'd extract the image data from the incoming message.
		var imageData []byte // Placeholder for actual image data extraction
		return h.visionParser.ParseImage(ctx, imageData)
	})

	if err != nil || parsed.ShouldReject() {
		h.sendMessage(msg.GetFrom(), "Aku belum bisa membaca gambar ini 😅\n\nBisa kirim ulang atau ketik manual?")
		return
	}

	// Record transaction
	tx, err := h.txService.RecordTransaction(ctx, user, parsed, msg.GetMessageID(), h.cfg.OpenAIModel, h.cfg.FreeTransactionLimit)
	if err != nil {
		if strings.Contains(err.Error(), "free limit") {
			h.sendMessage(msg.GetFrom(), "❌ Limit free sudah habis (10 transaksi).")
		} else {
			h.sendMessage(msg.GetFrom(), "Gagal menyimpan transaksi 😔")
		}
		return
	}

	h.sendMessage(msg.GetFrom(), fmt.Sprintf("✅ Transaksi dari gambar tersimpan!\n\nRp %.0f - %s\nID: %s",
		parsed.Amount, parsed.Description, tx.TxID))
}

func (h *WebhookHandler) handleUndo(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	err := h.txService.UndoTransaction(ctx, user.ID, h.cfg.UndoWindowSeconds)
	if err != nil {
		if strings.Contains(err.Error(), "no transaction") {
			h.sendMessage(msg.GetFrom(), "Tidak ada transaksi untuk dibatalkan.")
		} else if strings.Contains(err.Error(), "window expired") {
			h.sendMessage(msg.GetFrom(), "Waktu undo sudah habis (60 detik).")
		} else {
			h.sendMessage(msg.GetFrom(), "Gagal membatalkan transaksi 😔")
		}
		return
	}

	h.sendMessage(msg.GetFrom(), "✅ Transaksi terakhir dibatalkan!")
}

func (h *WebhookHandler) handleReportRequest(ctx context.Context, user *domain.User, msg *whatsapp.IncomingMessage) {
	text := strings.ToLower(msg.GetText())
	loc, _ := h.cfg.GetLocation()

	var report string
	var err error

	if strings.Contains(text, "hari ini") || strings.Contains(text, "harian") {
		report, err = h.reportService.GetDailyReport(ctx, user.ID, loc)
	} else if strings.Contains(text, "minggu") {
		report, err = h.reportService.GetWeeklyReport(ctx, user.ID, loc)
	} else if strings.Contains(text, "bulan") {
		report, err = h.reportService.GetMonthlyReport(ctx, user.ID, loc)
	} else {
		report, err = h.reportService.GetDailyReport(ctx, user.ID, loc)
	}

	if err != nil {
		log.Printf("Failed to generate report: %v", err)
		h.sendMessage(msg.GetFrom(), "Gagal membuat rekap 😔")
		return
	}

	h.sendMessage(msg.GetFrom(), report)
}

func (h *WebhookHandler) handleAdminCommand(ctx context.Context, msg *whatsapp.IncomingMessage) bool {
	text := strings.TrimSpace(msg.GetText())

	// upgrade <msisdn> monthly <dd/mm>
	if strings.HasPrefix(text, "upgrade ") {
		parts := strings.Fields(text)
		if len(parts) >= 4 {
			msisdn := parts[1]
			startDate := parts[3]

			// Parse date (dd/mm)
			dateParts := strings.Split(startDate, "/")
			if len(dateParts) == 2 {
				day, _ := strconv.Atoi(dateParts[0])
				month, _ := strconv.Atoi(dateParts[1])
				now := time.Now()
				start := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)

				err := h.userService.UpgradeToPremium(ctx, msisdn, start, 1)
				if err != nil {
					h.sendMessage(msg.GetFrom(), fmt.Sprintf("Failed: %v", err))
				} else {
					h.auditRepo.LogAdminAction(ctx, msg.GetFrom(), "upgrade", msisdn, map[string]interface{}{
						"start_date": start,
						"months":     1,
					})
					h.sendMessage(msg.GetFrom(), fmt.Sprintf("✅ %s upgraded to Premium", msisdn))
					h.sendMessage(msisdn, "🎉 Akun kamu sudah di-upgrade ke Premium! Unlimited transaksi.")
				}
				return true
			}
		}
	}

	// status <msisdn>
	if strings.HasPrefix(text, "status ") {
		parts := strings.Fields(text)
		if len(parts) == 2 {
			msisdn := parts[1]
			user, err := h.userService.GetUserStatus(ctx, msisdn)
			if err != nil {
				h.sendMessage(msg.GetFrom(), fmt.Sprintf("Error: %v", err))
			} else {
				status := fmt.Sprintf("User: %s\nPlan: %s\nTx Count: %d\nBlocked: %v",
					user.MSISDN, user.Plan, user.FreeTxCount, user.IsBlocked)
				if user.PremiumUntil != nil {
					status += fmt.Sprintf("\nPremium Until: %s", user.PremiumUntil.Format("2006-01-02"))
				}
				h.sendMessage(msg.GetFrom(), status)
			}
			return true
		}
	}

	return false
}

func (h *WebhookHandler) sendMessage(to, message string) {
	if err := h.waClient.SendMessage(to, message); err != nil {
		log.Printf("Failed to send message to %s: %v", to, err)
	}
}

// verifyHMACSignature verifies the X-Hub-Signature-256 header from GOWA webhook
func verifyHMACSignature(payload []byte, signatureHeader, secret string) bool {
	// signatureHeader format: "sha256=<hex_signature>"
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}

	expectedSignature := strings.TrimPrefix(signatureHeader, "sha256=")

	// Compute HMAC SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures (constant time to prevent timing attacks)
	return hmac.Equal([]byte(computedSignature), []byte(expectedSignature))
}
