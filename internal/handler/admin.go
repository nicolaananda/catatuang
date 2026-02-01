package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nicolaananda/catatuang/internal/service"
)

type AdminHandler struct {
	userService *service.UserService
	txService   *service.TransactionService
}

func NewAdminHandler(userService *service.UserService, txService *service.TransactionService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		txService:   txService,
	}
}

func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) UpgradeUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN    string `json:"msisdn"`
		StartDate string `json:"start_date"` // DD/MM format
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse date
	parts := strings.Split(req.StartDate, "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	day, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	now := time.Now()
	startDate := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)

	if err := h.userService.UpgradeToPremium(context.Background(), req.MSISDN, startDate, 1); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN string `json:"msisdn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.BlockUser(context.Background(), req.MSISDN); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN string `json:"msisdn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.UnblockUser(context.Background(), req.MSISDN); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN string `json:"msisdn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.DeleteUser(context.Background(), req.MSISDN); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *AdminHandler) DowngradePremium(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN string `json:"msisdn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.userService.DowngradePremium(context.Background(), req.MSISDN); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *AdminHandler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	msisdn := r.URL.Query().Get("msisdn")
	if msisdn == "" {
		http.Error(w, "msisdn required", http.StatusBadRequest)
		return
	}

	// Get user
	user, err := h.userService.GetUserByMSISDN(context.Background(), msisdn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get transactions for last 30 days
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	transactions, err := h.txService.GetTransactionsByDateRange(context.Background(), user.ID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
