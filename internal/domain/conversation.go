package domain

import (
	"encoding/json"
	"time"
)

// Conversation states
const (
	StateNewUser              = "NEW_USER"
	StateOnboardingSelectPlan = "ONBOARDING_SELECT_PLAN"
	StateActive               = "ACTIVE"
	StateAwaitingConfirm      = "AWAITING_CONFIRM_RECORD"
	StateEditingTransaction   = "EDITING_TRANSACTION"
	StateError                = "ERROR_STATE"
)

// ConversationState represents user's current conversation state
type ConversationState struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	State     string          `json:"state"`
	Context   json.RawMessage `json:"context,omitempty"`
	ExpiresAt time.Time       `json:"expires_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// IsExpired checks if state has expired
func (cs *ConversationState) IsExpired() bool {
	return time.Now().After(cs.ExpiresAt)
}

// ConfirmContext for AWAITING_CONFIRM_RECORD state
type ConfirmContext struct {
	ParsedTransaction *ParsedTransaction `json:"parsed_transaction"`
	OriginalMessage   string             `json:"original_message"`
}

// EditContext for EDITING_TRANSACTION state
type EditContext struct {
	TransactionID int64  `json:"transaction_id"`
	Field         string `json:"field"`
}
