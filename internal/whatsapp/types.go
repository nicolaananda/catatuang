package whatsapp

import "strings"

// IncomingMessage represents a webhook message from GOWA
type IncomingMessage struct {
	ChatID    string      `json:"chat_id"`
	From      string      `json:"from"`
	Message   MessageData `json:"message"`
	Pushname  string      `json:"pushname"`
	SenderID  string      `json:"sender_id"`
	Timestamp string      `json:"timestamp"`
}

// MessageData represents the nested message object
type MessageData struct {
	Text          string `json:"text"`
	ID            string `json:"id"`
	RepliedID     string `json:"replied_id,omitempty"`
	QuotedMessage string `json:"quoted_message,omitempty"`
}

// GetMessageID returns the message ID
func (m *IncomingMessage) GetMessageID() string {
	return m.Message.ID
}

// GetFrom returns the sender ID (phone number with WhatsApp suffix)
func (m *IncomingMessage) GetFrom() string {
	// For LID users, GOWA sends:
	// - sender_id: "88270922903758" (LID)
	// - from: "6281617985577@s.whatsapp.net" (actual phone)
	// We need to use the 'from' field to get the real phone number

	// First, try to extract phone from 'from' field if it contains @s.whatsapp.net
	if strings.Contains(m.From, "@s.whatsapp.net") {
		// Extract phone number before @s.whatsapp.net
		parts := strings.Split(m.From, "@")
		if len(parts) > 0 {
			phone := parts[0]
			// Return with @s.whatsapp.net suffix
			return phone + "@s.whatsapp.net"
		}
	}

	// Fallback to sender_id
	phone := m.SenderID

	// Skip if empty
	if phone == "" {
		return ""
	}

	// If already has @ suffix
	if strings.Contains(phone, "@") {
		// For LID format (e.g., "88270922903758@lid"), convert to standard format
		if strings.HasSuffix(phone, "@lid") {
			phone = strings.TrimSuffix(phone, "@lid")
			return phone + "@s.whatsapp.net"
		}
		return phone
	}

	// Add standard WhatsApp suffix
	return phone + "@s.whatsapp.net"
}

// GetText returns the message text
func (m *IncomingMessage) GetText() string {
	return m.Message.Text
}

// IsImage checks if message contains an image
func (m *IncomingMessage) IsImage() bool {
	// GOWA sends media in different format, will handle later
	return false
}

// IsText checks if message is text
func (m *IncomingMessage) IsText() bool {
	return m.Message.Text != ""
}
