package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	apiURL   string
	apiToken string
	client   *http.Client
}

func NewClient(apiURL, apiToken string) *Client {
	return &Client{
		apiURL:   apiURL,
		apiToken: apiToken,
		client:   &http.Client{},
	}
}

type SendMessageRequest struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

func (c *Client) SendMessage(to, message string) error {
	// GOWA expects phone number without @ suffix
	// Remove @s.whatsapp.net if present
	phone := to
	if len(phone) > 15 {
		// Extract just the number part (before @)
		if idx := bytes.IndexByte([]byte(phone), '@'); idx > 0 {
			phone = phone[:idx]
		}
	}

	req := SendMessageRequest{
		Phone:   phone,
		Message: message,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// GOWA endpoint is /send/text
	httpReq, err := http.NewRequest("POST", c.apiURL+"/send/text", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// GOWA doesn't use Authorization header - it's open API

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) DownloadMedia(mediaURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// GOWA doesn't require auth for media download

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download media: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media: %w", err)
	}

	return data, nil
}
