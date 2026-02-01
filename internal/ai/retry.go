package ai

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig defines retry behavior for AI calls
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
}

// WithRetry wraps an AI call with retry logic
func WithRetry[T any](ctx context.Context, cfg RetryConfig, fn func(context.Context) (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(cfg.Delay * time.Duration(attempt)):
			}
		}

		result, lastErr = fn(ctx)
		if lastErr == nil {
			return result, nil
		}

		// Log retry attempt
		fmt.Printf("AI call failed (attempt %d/%d): %v\n", attempt+1, cfg.MaxRetries+1, lastErr)
	}

	return result, fmt.Errorf("AI call failed after %d retries: %w", cfg.MaxRetries, lastErr)
}
