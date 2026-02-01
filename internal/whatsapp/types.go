package whatsapp

import (
	"strings"
)

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
	// Ensure phone number has @s.whatsapp.net suffix for sending messages
	phone := m.SenderID
	if !strings.Contains(phone, "@") {
		return phone + "@s.whatsapp.net"
	}
	return phone
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
