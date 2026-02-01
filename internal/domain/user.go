package domain

import "time"

// User plan types
const (
	PlanFree           = "FREE"
	PlanPremium        = "PREMIUM"
	PlanPendingPremium = "PENDING_PREMIUM"
)

// User represents a WhatsApp user
type User struct {
	ID           int64      `json:"id"`
	MSISDN       string     `json:"msisdn"`
	Plan         string     `json:"plan"`
	FreeTxCount  int        `json:"free_tx_count"`
	PremiumUntil *time.Time `json:"premium_until,omitempty"`
	IsBlocked    bool       `json:"is_blocked"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CanRecord checks if user can record a new transaction
func (u *User) CanRecord(freeLimit int) bool {
	if u.IsBlocked {
		return false
	}
	
	if u.Plan == PlanPremium {
		// Check if premium is still valid
		if u.PremiumUntil != nil && u.PremiumUntil.After(time.Now()) {
			return true
		}
		// Premium expired, treat as free
	}
	
	// Free user or expired premium
	return u.FreeTxCount < freeLimit
}

// IsPremium checks if user has active premium
func (u *User) IsPremium() bool {
	return u.Plan == PlanPremium && u.PremiumUntil != nil && u.PremiumUntil.After(time.Now())
}
