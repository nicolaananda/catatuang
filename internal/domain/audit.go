package domain

import (
	"encoding/json"
	"time"
)

// Audit actions
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
	ActionUndo   = "UNDO"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          int64           `json:"id"`
	UserID      *int64          `json:"user_id,omitempty"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type,omitempty"`
	EntityID    *int64          `json:"entity_id,omitempty"`
	OldValue    json.RawMessage `json:"old_value,omitempty"`
	NewValue    json.RawMessage `json:"new_value,omitempty"`
	PerformedBy string          `json:"performed_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// AdminAction represents an admin operation
type AdminAction struct {
	ID           int64           `json:"id"`
	AdminMSISDN  string          `json:"admin_msisdn"`
	Action       string          `json:"action"`
	TargetMSISDN string          `json:"target_msisdn,omitempty"`
	Details      json.RawMessage `json:"details,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}
