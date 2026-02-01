package domain

import (
	"fmt"
	"time"
)

// Transaction types
const (
	TypeIncome  = "INCOME"
	TypeExpense = "EXPENSE"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID              int64     `json:"id"`
	TxID            string    `json:"tx_id"`
	UserID          int64     `json:"user_id"`
	Type            string    `json:"type"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category,omitempty"`
	Description     string    `json:"description,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	WAMessageID     string    `json:"wa_message_id,omitempty"`
	AIConfidence    float64   `json:"ai_confidence,omitempty"`
	AIVersion       string    `json:"ai_version,omitempty"`
	IsDeleted       bool      `json:"is_deleted"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GenerateTxID generates a unique transaction ID using WA message ID
// This prevents race conditions since WA message ID is unique before DB insert
func GenerateTxID(waMessageID string) string {
	// Use last 8 chars of WA message ID + timestamp for uniqueness
	timestamp := time.Now().Unix()

	// Extract last 8 chars of WA message ID for readability
	suffix := waMessageID
	if len(waMessageID) > 8 {
		suffix = waMessageID[len(waMessageID)-8:]
	}

	return fmt.Sprintf("TX#%s-%d", suffix, timestamp)
}
