package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
	openai "github.com/sashabaranov/go-openai"
)

type TextParser struct {
	client   *openai.Client
	model    string
	timezone *time.Location
}

func NewTextParser(apiKey, model string, timezone *time.Location) *TextParser {
	return &TextParser{
		client:   openai.NewClient(apiKey),
		model:    model,
		timezone: timezone,
	}
}

func (p *TextParser) Parse(ctx context.Context, message string) (*domain.ParsedTransaction, error) {
	// Normalize Indonesian slang
	normalized := p.normalizeAmount(message)

	// Get current date in user's timezone
	today := time.Now().In(p.timezone).Format("2006-01-02")

	systemPrompt := fmt.Sprintf(`You are a financial transaction parser for Indonesian users.
Parse the message into a structured transaction.

IMPORTANT: Today's date is %s. Use this as the default date if no date is mentioned.

Rules:
1. Determine if it's INCOME or EXPENSE
2. Extract the amount (handle "rb" = ribu/1000, "jt" = juta/1000000)
3. Identify category (e.g., "gaji", "makan", "transport", "belanja")
4. Extract description
5. Parse date if mentioned, otherwise use TODAY (%s)
6. Provide confidence score (0.0-1.0)

Return ONLY valid JSON in this exact format:
{
  "type": "INCOME" or "EXPENSE",
  "amount": number,
  "category": "string",
  "description": "string",
  "date": "YYYY-MM-DD",
  "confidence": 0.0-1.0
}

Examples:
- "catat pemasukan 10000 gaji" → INCOME, 10000, "gaji", "gaji", %s, 0.95
- "beli bensin 50rb" → EXPENSE, 50000, "transport", "beli bensin", %s, 0.9
- "dapat uang dari jual motor 20 juta" → INCOME, 20000000, "penjualan", "jual motor", %s, 0.85`, today, today, today, today, today)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: normalized},
		},
		Temperature: 0.3,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
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

// normalizeAmount converts Indonesian slang to numbers
func (p *TextParser) normalizeAmount(text string) string {
	// Convert "50rb" → "50000", "20jt" → "20000000"

	// Match patterns like "50rb", "50 rb", "50ribu"
	reRibu := regexp.MustCompile(`(\d+)\s*(rb|ribu)`)
	text = reRibu.ReplaceAllStringFunc(text, func(match string) string {
		parts := reRibu.FindStringSubmatch(match)
		if len(parts) >= 2 {
			num, _ := strconv.Atoi(parts[1])
			return strconv.Itoa(num * 1000)
		}
		return match
	})

	// Match patterns like "20jt", "20 jt", "20juta"
	reJuta := regexp.MustCompile(`(\d+)\s*(jt|juta)`)
	text = reJuta.ReplaceAllStringFunc(text, func(match string) string {
		parts := reJuta.FindStringSubmatch(match)
		if len(parts) >= 2 {
			num, _ := strconv.Atoi(parts[1])
			return strconv.Itoa(num * 1000000)
		}
		return match
	})

	return text
}

// ShouldTriggerParsing checks if message should trigger transaction parsing
func ShouldTriggerParsing(message string) bool {
	message = strings.ToLower(message)

	// Explicit keywords
	keywords := []string{"catat", "pencatatan", "simpan", "record"}
	for _, kw := range keywords {
		if strings.Contains(message, kw) {
			return true
		}
	}

	// Check for amount patterns (numbers with rb/jt or just numbers with context)
	hasAmount := regexp.MustCompile(`\d+\s*(rb|ribu|jt|juta|rupiah|rp)`).MatchString(message)
	if hasAmount {
		return true
	}

	// Check for transaction keywords
	txKeywords := []string{"beli", "bayar", "dapat", "terima", "gaji", "pemasukan", "pengeluaran", "belanja"}
	for _, kw := range txKeywords {
		if strings.Contains(message, kw) && regexp.MustCompile(`\d+`).MatchString(message) {
			return true
		}
	}

	return false
}
