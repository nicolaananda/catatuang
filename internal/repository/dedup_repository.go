package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type DedupRepository struct {
	db *sql.DB
}

func NewDedupRepository(db *sql.DB) *DedupRepository {
	return &DedupRepository{db: db}
}

func (r *DedupRepository) IsProcessed(ctx context.Context, waMessageID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM message_dedup WHERE wa_message_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, waMessageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check dedup: %w", err)
	}

	return exists, nil
}

func (r *DedupRepository) MarkProcessed(ctx context.Context, waMessageID string) error {
	query := `INSERT INTO message_dedup (wa_message_id) VALUES ($1) ON CONFLICT DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, waMessageID)
	if err != nil {
		return fmt.Errorf("failed to mark processed: %w", err)
	}

	return nil
}

func (r *DedupRepository) CleanupOld(ctx context.Context, olderThanHours int) error {
	query := `DELETE FROM message_dedup WHERE processed_at < NOW() - INTERVAL '1 hour' * $1`

	_, err := r.db.ExecContext(ctx, query, olderThanHours)
	if err != nil {
		return fmt.Errorf("failed to cleanup dedup: %w", err)
	}

	return nil
}
