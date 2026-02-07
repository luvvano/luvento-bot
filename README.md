# Luvento Notification Bot

Telegram bot for Luvento technical notifications.

## Features

- 👤 New user registration notifications
- 💬 Support message notifications  
- 🚨 Server error notifications
- 📢 Multi-group support

## Setup

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Yes | Telegram bot token from @BotFather |
| `WEBHOOK_API_KEY` | Yes | API key for webhook authentication |
| `DATABASE_PATH` | No | SQLite database path (default: `/data/bot.db`) |
| `PORT` | No | HTTP server port (default: `8080`) |

### Bot Commands

- `/start` - Bot info
- `/addgroup` - Add this group to notifications (admin only)
- `/removegroup` - Remove group from notifications (admin only)
- `/status` - Show status and registered groups

## API Endpoints

All endpoints require `X-API-Key` header.

### POST /webhook/user-registered

```json
{
  "email": "user@example.com",
  "createdAt": "2026-02-07T10:00:00Z",
  "metadata": {
    "country": "Cyprus",
    "city": "Nicosia",
    "browser": "Chrome",
    "os": "macOS",
    "referrer": "google.com",
    "ip": "1.2.3.4"
  }
}
```

### POST /webhook/support-message

```json
{
  "userEmail": "user@example.com",
  "message": "Help needed...",
  "createdAt": "2026-02-07T10:00:00Z"
}
```

### POST /webhook/server-error

```json
{
  "service": "luvento-back",
  "error": "Connection refused",
  "stack": "...",
  "createdAt": "2026-02-07T10:00:00Z"
}
```

## Deployment

Deploy via Coolify with docker-compose.

### Traefik Route

Webhook endpoint: `https://cal.luvano.pro/bot-webhook/webhook/*`

Example: `https://cal.luvano.pro/bot-webhook/webhook/user-registered`

## Integration with luvento-back

Add HTTP call on user registration:

```go
func (s *Service) notifyNewUser(user *models.User) {
    payload := map[string]any{
        "email":     user.Email,
        "createdAt": user.CreatedAt,
        "metadata": map[string]string{
            "country": user.Country,
            // ...
        },
    }
    
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", 
        os.Getenv("NOTIFICATION_BOT_URL")+"/webhook/user-registered", 
        bytes.NewReader(body))
    req.Header.Set("X-API-Key", os.Getenv("NOTIFICATION_BOT_API_KEY"))
    req.Header.Set("Content-Type", "application/json")
    
    http.DefaultClient.Do(req)
}
```

## License

MIT
