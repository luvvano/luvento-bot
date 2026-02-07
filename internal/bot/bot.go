package bot

import (
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/luvvano/luvento-bot/internal/storage"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	storage *storage.Storage
	stop    chan struct{}
}

func New(token string, store *storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}

	slog.Info("authorized on telegram", "username", api.Self.UserName)

	return &Bot{
		api:     api,
		storage: store,
		stop:    make(chan struct{}),
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-b.stop:
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				b.handleCommand(update.Message)
			}
		}
	}
}

func (b *Bot) Stop() {
	close(b.stop)
	b.api.StopReceivingUpdates()
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.cmdStart(msg)
	case "addgroup":
		b.cmdAddGroup(msg)
	case "removegroup":
		b.cmdRemoveGroup(msg)
	case "status":
		b.cmdStatus(msg)
	case "help":
		b.cmdHelp(msg)
	}
}

func (b *Bot) cmdStart(msg *tgbotapi.Message) {
	text := `🤖 *Luvento Notification Bot*

Я отправляю уведомления о важных событиях:
• Новые регистрации пользователей
• Сообщения в поддержку
• Ошибки сервера

*Команды:*
/addgroup — добавить эту группу в рассылку
/removegroup — убрать группу из рассылки
/status — показать статус
/help — помощь`

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "Markdown"
	b.api.Send(reply)
}

func (b *Bot) cmdHelp(msg *tgbotapi.Message) {
	b.cmdStart(msg)
}

func (b *Bot) cmdAddGroup(msg *tgbotapi.Message) {
	// Check if it's a group chat
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Эта команда работает только в группах")
		b.api.Send(reply)
		return
	}

	// Check if user is admin
	isAdmin, err := b.isUserAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		slog.Error("failed to check admin status", "error", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка проверки прав")
		b.api.Send(reply)
		return
	}

	if !isAdmin {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Только администраторы могут добавлять группу")
		b.api.Send(reply)
		return
	}

	// Add group
	err = b.storage.AddGroup(msg.Chat.ID, msg.Chat.Title, msg.From.ID)
	if err != nil {
		slog.Error("failed to add group", "error", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка добавления группы")
		b.api.Send(reply)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "✅ Группа добавлена в рассылку уведомлений")
	b.api.Send(reply)
	slog.Info("group added", "chat_id", msg.Chat.ID, "title", msg.Chat.Title, "by", msg.From.ID)
}

func (b *Bot) cmdRemoveGroup(msg *tgbotapi.Message) {
	if msg.Chat.Type != "group" && msg.Chat.Type != "supergroup" {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Эта команда работает только в группах")
		b.api.Send(reply)
		return
	}

	isAdmin, err := b.isUserAdmin(msg.Chat.ID, msg.From.ID)
	if err != nil {
		slog.Error("failed to check admin status", "error", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка проверки прав")
		b.api.Send(reply)
		return
	}

	if !isAdmin {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Только администраторы могут удалять группу")
		b.api.Send(reply)
		return
	}

	err = b.storage.RemoveGroup(msg.Chat.ID)
	if err != nil {
		slog.Error("failed to remove group", "error", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка удаления группы")
		b.api.Send(reply)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "✅ Группа удалена из рассылки")
	b.api.Send(reply)
	slog.Info("group removed", "chat_id", msg.Chat.ID, "by", msg.From.ID)
}

func (b *Bot) cmdStatus(msg *tgbotapi.Message) {
	groups, err := b.storage.GetAllGroups()
	if err != nil {
		slog.Error("failed to get groups", "error", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка получения статуса")
		b.api.Send(reply)
		return
	}

	var text string
	if len(groups) == 0 {
		text = "📊 *Статус*\n\nНет подписанных групп"
	} else {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📊 *Статус*\n\nПодписанных групп: %d\n\n", len(groups)))
		for _, g := range groups {
			sb.WriteString(fmt.Sprintf("• %s\n", g.Title))
		}
		text = sb.String()
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "Markdown"
	b.api.Send(reply)
}

func (b *Bot) isUserAdmin(chatID int64, userID int64) (bool, error) {
	admins, err := b.api.GetChatAdministrators(tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
	})
	if err != nil {
		return false, err
	}

	for _, admin := range admins {
		if admin.User.ID == userID {
			return true, nil
		}
	}

	return false, nil
}

// SendToAllGroups sends a message to all registered groups
func (b *Bot) SendToAllGroups(text string) error {
	groups, err := b.storage.GetAllGroups()
	if err != nil {
		return fmt.Errorf("get groups: %w", err)
	}

	for _, g := range groups {
		msg := tgbotapi.NewMessage(g.ChatID, text)
		msg.ParseMode = "Markdown"
		_, err := b.api.Send(msg)
		if err != nil {
			slog.Error("failed to send to group", "chat_id", g.ChatID, "error", err)
		}
	}

	return nil
}
