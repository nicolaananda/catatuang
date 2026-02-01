package domain

import "time"

// ParsedTransaction represents AI-parsed transaction data
type ParsedTransaction struct {
	Type        string    `json:"type"` // INCOME or EXPENSE
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Confidence  float64   `json:"confidence"`
}

// NeedsConfirmation checks if confidence is in the medium range
func (pt *ParsedTransaction) NeedsConfirmation() bool {
	return pt.Confidence >= 0.4 && pt.Confidence < 0.7
}

// ShouldReject checks if confidence is too low
func (pt *ParsedTransaction) ShouldReject() bool {
	return pt.Confidence < 0.4
}

// ShouldAutoSave checks if confidence is high enough
func (pt *ParsedTransaction) ShouldAutoSave() bool {
	return pt.Confidence >= 0.7
}
