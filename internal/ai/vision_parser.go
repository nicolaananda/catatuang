package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
	openai "github.com/sashabaranov/go-openai"
)

type VisionParser struct {
	client   *openai.Client
	model    string
	timezone *time.Location
}

func NewVisionParser(apiKey, model string, timezone *time.Location) *VisionParser {
	return &VisionParser{
		client:   openai.NewClient(apiKey),
		model:    model,
		timezone: timezone,
	}
}

func (p *VisionParser) ParseImage(ctx context.Context, imageData []byte) (*domain.ParsedTransaction, error) {
	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	systemPrompt := `You are a receipt/transaction image parser for Indonesian financial transactions.
Extract transaction information from the image (receipt, bank transfer screenshot, etc).

Return ONLY valid JSON in this exact format:
{
  "type": "INCOME" or "EXPENSE",
  "amount": number,
  "category": "string",
  "description": "string (merchant name or transfer description)",
  "date": "YYYY-MM-DD",
  "confidence": 0.0-1.0
}

Rules:
1. For receipts → EXPENSE, extract total amount and merchant name
2. For transfer screenshots → check if incoming (INCOME) or outgoing (EXPENSE)
3. Extract date from image, use today if not visible
4. Provide high confidence (>0.8) only if amount and type are clear
5. If image is unclear or not a transaction, return confidence < 0.4`

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: systemPrompt,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
				},
			},
		},
		Temperature: 0.3,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("openai vision API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	content := resp.Choices[0].Message.Content

	// Parse JSON response
	var result struct {
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
		Confidence  float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Parse date
	txDate, err := time.ParseInLocation("2006-01-02", result.Date, p.timezone)
	if err != nil {
		// Default to today if parsing fails
		txDate = time.Now().In(p.timezone)
	}

	return &domain.ParsedTransaction{
		Type:        result.Type,
		Amount:      result.Amount,
		Category:    result.Category,
		Description: result.Description,
		Date:        txDate,
		Confidence:  result.Confidence,
	}, nil
}
