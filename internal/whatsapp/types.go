package whatsapp

// IncomingMessage represents a webhook message from GOWA
type IncomingMessage struct {
	MessageID string `json:"message_id"`
	From      string `json:"from"`
	Text      string `json:"text,omitempty"`
	MediaURL  string `json:"media_url,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// IsImage checks if message contains an image
func (m *IncomingMessage) IsImage() bool {
	return m.MediaType == "image" && m.MediaURL != ""
}

// IsText checks if message is text
func (m *IncomingMessage) IsText() bool {
	return m.Text != "" && m.MediaURL == ""
}
