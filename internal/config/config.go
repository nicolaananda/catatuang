package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseURL string

	// OpenAI
	OpenAIAPIKey string
	OpenAIModel  string

	// GOWA WhatsApp
	GowaWebhookSecret string
	GowaAPIURL        string
	GowaAPIToken      string

	// Server
	Port           string
	AdminPanelPort string

	// Admin
	AdminMSISDN string

	// App Settings
	Timezone              string
	AITimeoutSeconds      int
	AIMaxRetries          int
	StateExpiryMinutes    int
	UndoWindowSeconds     int
	FreeTransactionLimit  int
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:       getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		GowaWebhookSecret: getEnv("GOWA_WEBHOOK_SECRET", ""),
		GowaAPIURL:        getEnv("GOWA_API_URL", ""),
		GowaAPIToken:      getEnv("GOWA_API_TOKEN", ""),
		Port:              getEnv("PORT", "8080"),
		AdminPanelPort:    getEnv("ADMIN_PANEL_PORT", "8081"),
		AdminMSISDN:       getEnv("ADMIN_MSISDN", "081389592985"),
		Timezone:          getEnv("TIMEZONE", "Asia/Jakarta"),
		AITimeoutSeconds:  getEnvInt("AI_TIMEOUT_SECONDS", 12),
		AIMaxRetries:      getEnvInt("AI_MAX_RETRIES", 2),
		StateExpiryMinutes: getEnvInt("STATE_EXPIRY_MINUTES", 30),
		UndoWindowSeconds: getEnvInt("UNDO_WINDOW_SECONDS", 60),
		FreeTransactionLimit: getEnvInt("FREE_TRANSACTION_LIMIT", 10),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	if c.GowaWebhookSecret == "" {
		return fmt.Errorf("GOWA_WEBHOOK_SECRET is required")
	}
	return nil
}

func (c *Config) GetLocation() (*time.Location, error) {
	return time.LoadLocation(c.Timezone)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
