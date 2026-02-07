package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramToken string
	WebhookAPIKey string
	DatabasePath  string
	Port          string
}

func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	apiKey := os.Getenv("WEBHOOK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEBHOOK_API_KEY is required")
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "/data/bot.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		TelegramToken: token,
		WebhookAPIKey: apiKey,
		DatabasePath:  dbPath,
		Port:          port,
	}, nil
}
