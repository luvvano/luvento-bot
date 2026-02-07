package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luvvano/luvento-bot/internal/bot"
	"github.com/luvvano/luvento-bot/internal/config"
	"github.com/luvvano/luvento-bot/internal/storage"
	"github.com/luvvano/luvento-bot/internal/webhook"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize storage
	store, err := storage.New(cfg.DatabasePath)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize Telegram bot
	tgBot, err := bot.New(cfg.TelegramToken, store, cfg.OwnerID)
	if err != nil {
		slog.Error("failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}

	// Start bot in background
	go tgBot.Start()

	// Setup HTTP server for webhooks
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Webhook handlers
	webhookHandler := webhook.NewHandler(cfg.WebhookAPIKey, tgBot, store)
	r.Route("/webhook", func(r chi.Router) {
		r.Use(webhookHandler.APIKeyAuth)
		r.Post("/user-registered", webhookHandler.UserRegistered)
		r.Post("/support-message", webhookHandler.SupportMessage)
		r.Post("/server-error", webhookHandler.ServerError)
	})

	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("starting HTTP server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	tgBot.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
