package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramToken string
	WebhookAPIKey string
	DatabasePath  string
	Port          string
	OwnerID       int64 // Telegram user ID of the bot owner
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

	var ownerID int64
	if ownerIDStr := os.Getenv("OWNER_ID"); ownerIDStr != "" {
		var err error
		ownerID, err = strconv.ParseInt(ownerIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid OWNER_ID: %w", err)
		}
	}

	return &Config{
		TelegramToken: token,
		WebhookAPIKey: apiKey,
		DatabasePath:  dbPath,
		Port:          port,
		OwnerID:       ownerID,
	}, nil
}
