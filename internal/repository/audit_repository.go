package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nicolaananda/catatuang/internal/domain"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Log(ctx context.Context, userID int64, action, entityType string, entityID int64, oldValue, newValue json.RawMessage, performedBy string) error {
	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_value, new_value, performed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query, userID, action, entityType, entityID, oldValue, newValue, performedBy)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (r *AuditRepository) GetByUser(ctx context.Context, userID int64, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_value, new_value, performed_by, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.EntityType, &log.EntityID,
			&log.OldValue, &log.NewValue, &log.PerformedBy, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *AuditRepository) LogAdminAction(ctx context.Context, adminMSISDN, action, targetMSISDN string, details interface{}) error {
	detailsJSON, _ := json.Marshal(details)

	query := `
		INSERT INTO admin_actions (admin_msisdn, action, target_msisdn, details)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, adminMSISDN, action, targetMSISDN, detailsJSON)
	if err != nil {
		return fmt.Errorf("failed to log admin action: %w", err)
	}

	return nil
}
