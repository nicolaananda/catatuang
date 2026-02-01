package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
	"github.com/nicolaananda/catatuang/internal/repository"
)

type TransactionService struct {
	txRepo    *repository.TransactionRepository
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	db        *sql.DB
}

func NewTransactionService(
	txRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	db *sql.DB,
) *TransactionService {
	return &TransactionService{
		txRepo:    txRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		db:        db,
	}
}

func (s *TransactionService) RecordTransaction(ctx context.Context, user *domain.User, parsed *domain.ParsedTransaction, waMessageID, aiVersion string, freeLimit int) (*domain.Transaction, error) {
	// Check if user can record
	if !user.CanRecord(freeLimit) {
		return nil, fmt.Errorf("free limit exceeded")
	}

	// Start database transaction
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Create transaction
	tx := &domain.Transaction{
		UserID:          user.ID,
		Type:            parsed.Type,
		Amount:          parsed.Amount,
		Category:        parsed.Category,
		Description:     parsed.Description,
		TransactionDate: parsed.Date,
		WAMessageID:     waMessageID,
		AIConfidence:    parsed.Confidence,
		AIVersion:       aiVersion,
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Generate TX ID
	tx.TxID = domain.GenerateTxID(tx.ID)

	// Update TX ID
	if err := s.txRepo.Update(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to update tx_id: %w", err)
	}

	// Increment user's free transaction count if not premium
	if !user.IsPremium() {
		if err := s.userRepo.IncrementFreeTxCount(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("failed to increment count: %w", err)
		}
	}

	// Audit log
	newValue, _ := json.Marshal(tx)
	if err := s.auditRepo.Log(ctx, user.ID, domain.ActionCreate, "transaction", tx.ID, nil, newValue, "system"); err != nil {
		// Log but don't fail
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return tx, nil
}

func (s *TransactionService) UndoTransaction(ctx context.Context, userID int64, undoWindowSeconds int) error {
	// Get last transaction
	tx, err := s.txRepo.GetLastByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get last transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("no transaction to undo")
	}

	// Check if within undo window
	if time.Since(tx.CreatedAt) > time.Duration(undoWindowSeconds)*time.Second {
		return fmt.Errorf("undo window expired")
	}

	// Start database transaction
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer dbTx.Rollback()

	// Hard delete the transaction
	if err := s.txRepo.HardDelete(ctx, tx.ID); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Decrement user's free transaction count
	user, err := s.userRepo.GetByMSISDN(ctx, "")
	if err == nil && user != nil && !user.IsPremium() {
		user.FreeTxCount--
		if user.FreeTxCount < 0 {
			user.FreeTxCount = 0
		}
		_ = s.userRepo.Update(ctx, user)
	}

	// Audit log
	oldValue, _ := json.Marshal(tx)
	if err := s.auditRepo.Log(ctx, userID, domain.ActionUndo, "transaction", tx.ID, oldValue, nil, "user"); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Commit
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func (s *TransactionService) EditTransaction(ctx context.Context, txID string, updates map[string]interface{}) error {
	tx, err := s.txRepo.GetByTxID(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}

	oldValue, _ := json.Marshal(tx)

	// Apply updates
	if amount, ok := updates["amount"].(float64); ok {
		tx.Amount = amount
	}
	if category, ok := updates["category"].(string); ok {
		tx.Category = category
	}
	if description, ok := updates["description"].(string); ok {
		tx.Description = description
	}
	if txType, ok := updates["type"].(string); ok {
		tx.Type = txType
	}

	if err := s.txRepo.Update(ctx, tx); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Audit log
	newValue, _ := json.Marshal(tx)
	if err := s.auditRepo.Log(ctx, tx.UserID, domain.ActionUpdate, "transaction", tx.ID, oldValue, newValue, "user"); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, txID string) error {
	tx, err := s.txRepo.GetByTxID(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}

	oldValue, _ := json.Marshal(tx)

	if err := s.txRepo.SoftDelete(ctx, tx.ID); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Audit log
	if err := s.auditRepo.Log(ctx, tx.UserID, domain.ActionDelete, "transaction", tx.ID, oldValue, nil, "user"); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *TransactionService) GetTransactionsByDateRange(ctx context.Context, userID int64, start, end time.Time) ([]*domain.Transaction, error) {
	return s.txRepo.GetByUserAndDateRange(ctx, userID, start, end)
}
