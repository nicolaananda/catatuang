package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (tx_id, user_id, type, amount, category, description, transaction_date, wa_message_id, ai_confidence, ai_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		tx.TxID,
		tx.UserID,
		tx.Type,
		tx.Amount,
		tx.Category,
		tx.Description,
		tx.TransactionDate,
		tx.WAMessageID,
		tx.AIConfidence,
		tx.AIVersion,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	query := `
		SELECT id, tx_id, user_id, type, amount, category, description, transaction_date, 
		       wa_message_id, ai_confidence, ai_version, is_deleted, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	tx := &domain.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Category, &tx.Description,
		&tx.TransactionDate, &tx.WAMessageID, &tx.AIConfidence, &tx.AIVersion,
		&tx.IsDeleted, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

func (r *TransactionRepository) GetByTxID(ctx context.Context, txID string) (*domain.Transaction, error) {
	query := `
		SELECT id, tx_id, user_id, type, amount, category, description, transaction_date, 
		       wa_message_id, ai_confidence, ai_version, is_deleted, created_at, updated_at
		FROM transactions
		WHERE tx_id = $1
	`

	tx := &domain.Transaction{}
	err := r.db.QueryRowContext(ctx, query, txID).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Category, &tx.Description,
		&tx.TransactionDate, &tx.WAMessageID, &tx.AIConfidence, &tx.AIVersion,
		&tx.IsDeleted, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

func (r *TransactionRepository) GetLastByUser(ctx context.Context, userID int64) (*domain.Transaction, error) {
	query := `
		SELECT id, tx_id, user_id, type, amount, category, description, transaction_date, 
		       wa_message_id, ai_confidence, ai_version, is_deleted, created_at, updated_at
		FROM transactions
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT 1
	`

	tx := &domain.Transaction{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&tx.ID, &tx.TxID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Category, &tx.Description,
		&tx.TransactionDate, &tx.WAMessageID, &tx.AIConfidence, &tx.AIVersion,
		&tx.IsDeleted, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last transaction: %w", err)
	}

	return tx, nil
}

func (r *TransactionRepository) GetByUserAndDateRange(ctx context.Context, userID int64, start, end time.Time) ([]*domain.Transaction, error) {
	query := `
		SELECT id, tx_id, user_id, type, amount, category, description, transaction_date, 
		       wa_message_id, ai_confidence, ai_version, is_deleted, created_at, updated_at
		FROM transactions
		WHERE user_id = $1 AND is_deleted = false 
		  AND transaction_date >= $2 AND transaction_date < $3
		ORDER BY transaction_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.TxID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Category, &tx.Description,
			&tx.TransactionDate, &tx.WAMessageID, &tx.AIConfidence, &tx.AIVersion,
			&tx.IsDeleted, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	query := `
		UPDATE transactions
		SET type = $1, amount = $2, category = $3, description = $4, transaction_date = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(ctx, query,
		tx.Type, tx.Amount, tx.Category, tx.Description, tx.TransactionDate, tx.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE transactions SET is_deleted = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) HardDelete(ctx context.Context, id int64) error {
	query := `DELETE FROM transactions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete transaction: %w", err)
	}
	return nil
}
