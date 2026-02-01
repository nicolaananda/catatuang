package statemachine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
)

type StateMachine struct {
	db *sql.DB
}

func NewStateMachine(db *sql.DB) *StateMachine {
	return &StateMachine{db: db}
}

func (sm *StateMachine) GetState(ctx context.Context, userID int64) (*domain.ConversationState, error) {
	query := `
		SELECT id, user_id, state, context, expires_at, created_at
		FROM conversation_states
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	state := &domain.ConversationState{}
	var contextJSON sql.NullString

	err := sm.db.QueryRowContext(ctx, query, userID).Scan(
		&state.ID,
		&state.UserID,
		&state.State,
		&contextJSON,
		&state.ExpiresAt,
		&state.CreatedAt,
	)

	if err == sql.ErrNoRows {
		// No active state, return ACTIVE as default
		return &domain.ConversationState{
			UserID: userID,
			State:  domain.StateActive,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Handle NULL context
	if contextJSON.Valid {
		state.Context = json.RawMessage(contextJSON.String)
	}

	return state, nil
}

func (sm *StateMachine) SetState(ctx context.Context, userID int64, state string, context interface{}, expiryMinutes int) error {
	// Serialize context to JSON
	var contextJSON []byte
	var err error
	if context != nil {
		contextJSON, err = json.Marshal(context)
		if err != nil {
			return fmt.Errorf("failed to marshal context: %w", err)
		}
	}

	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	query := `
		INSERT INTO conversation_states (user_id, state, context, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err = sm.db.ExecContext(ctx, query, userID, state, contextJSON, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	return nil
}

func (sm *StateMachine) ClearState(ctx context.Context, userID int64) error {
	query := `DELETE FROM conversation_states WHERE user_id = $1`
	_, err := sm.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear state: %w", err)
	}
	return nil
}

func (sm *StateMachine) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM conversation_states WHERE expires_at < NOW()`
	_, err := sm.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired states: %w", err)
	}
	return nil
}
