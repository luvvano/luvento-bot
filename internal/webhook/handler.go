package webhook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/luvvano/luvento-bot/internal/bot"
	"github.com/luvvano/luvento-bot/internal/storage"
)

type Handler struct {
	apiKey  string
	bot     *bot.Bot
	storage *storage.Storage
}

func NewHandler(apiKey string, b *bot.Bot, s *storage.Storage) *Handler {
	return &Handler{
		apiKey:  apiKey,
		bot:     b,
		storage: s,
	}
}

// APIKeyAuth middleware to verify API key
func (h *Handler) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" || key != h.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserRegisteredPayload represents new user registration event
type UserRegisteredPayload struct {
	Email     string            `json:"email"`
	CreatedAt time.Time         `json:"createdAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// UserRegistered handles new user registration webhook
func (h *Handler) UserRegistered(w http.ResponseWriter, r *http.Request) {
	var payload UserRegisteredPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode payload", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Format message
	text := fmt.Sprintf("👤 *Новый пользователь*\n\n"+
		"📧 Email: `%s`\n"+
		"🕐 Время: %s",
		payload.Email,
		payload.CreatedAt.Format("02.01.2006 15:04:05"),
	)

	// Add metadata if present
	if len(payload.Metadata) > 0 {
		text += "\n\n📍 *Информация:*"
		if country := payload.Metadata["country"]; country != "" {
			text += fmt.Sprintf("\n• Страна: %s", country)
		}
		if city := payload.Metadata["city"]; city != "" {
			text += fmt.Sprintf("\n• Город: %s", city)
		}
		if browser := payload.Metadata["browser"]; browser != "" {
			text += fmt.Sprintf("\n• Браузер: %s", browser)
		}
		if os := payload.Metadata["os"]; os != "" {
			text += fmt.Sprintf("\n• ОС: %s", os)
		}
		if referrer := payload.Metadata["referrer"]; referrer != "" {
			text += fmt.Sprintf("\n• Источник: %s", referrer)
		}
		if ip := payload.Metadata["ip"]; ip != "" {
			text += fmt.Sprintf("\n• IP: `%s`", ip)
		}
	}

	if err := h.bot.SendToAllGroups(text); err != nil {
		slog.Error("failed to send notification", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("user registered notification sent", "email", payload.Email)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// SupportMessagePayload represents support chat message
type SupportMessagePayload struct {
	UserEmail string    `json:"userEmail"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// SupportMessage handles support message webhook
func (h *Handler) SupportMessage(w http.ResponseWriter, r *http.Request) {
	var payload SupportMessagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode payload", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	text := fmt.Sprintf("💬 *Сообщение в поддержку*\n\n"+
		"👤 От: `%s`\n"+
		"🕐 Время: %s\n\n"+
		"📝 Сообщение:\n%s",
		payload.UserEmail,
		payload.CreatedAt.Format("02.01.2006 15:04:05"),
		payload.Message,
	)

	if err := h.bot.SendToAllGroups(text); err != nil {
		slog.Error("failed to send notification", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("support message notification sent", "email", payload.UserEmail)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ServerErrorPayload represents server error event
type ServerErrorPayload struct {
	Service   string    `json:"service"`
	Error     string    `json:"error"`
	Stack     string    `json:"stack,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ServerError handles server error webhook
func (h *Handler) ServerError(w http.ResponseWriter, r *http.Request) {
	var payload ServerErrorPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode payload", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	text := fmt.Sprintf("🚨 *Ошибка на сервере*\n\n"+
		"🔧 Сервис: `%s`\n"+
		"🕐 Время: %s\n\n"+
		"❌ Ошибка:\n```\n%s\n```",
		payload.Service,
		payload.CreatedAt.Format("02.01.2006 15:04:05"),
		truncate(payload.Error, 500),
	)

	if payload.Stack != "" {
		text += fmt.Sprintf("\n\n📚 Stack trace:\n```\n%s\n```", truncate(payload.Stack, 500))
	}

	if err := h.bot.SendToAllGroups(text); err != nil {
		slog.Error("failed to send notification", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.Info("server error notification sent", "service", payload.Service)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
