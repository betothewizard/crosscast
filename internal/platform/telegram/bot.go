package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/betothewizard/crosscast/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotStatus represents real-time operational metrics of the Telegram session.
type BotStatus struct {
	Enabled   bool   `json:"enabled"`
	Online    bool   `json:"online"`
	Username  string `json:"username,omitempty"`
	Uptime    string `json:"uptime"`
	UptimeSec int64  `json:"uptime_seconds"`
}

// Bot manages the 24/7 Telegram bot session, long-polling, and message broadcasting.
type Bot struct {
	cfg       *config.Config
	api       *tgbotapi.BotAPI
	startTime time.Time
	ready     bool
	cancel    context.CancelFunc
}

// New creates a new Telegram Bot instance.
func New(cfg *config.Config) *Bot {
	return &Bot{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// Start initializes the Telegram bot API and launches the update polling loop.
func (b *Bot) Start(ctx context.Context) error {
	if b.cfg.TelegramToken == "" {
		log.Printf("[telegram] notice: TELEGRAM_BOT_TOKEN is empty; Telegram gateway is disabled")
		return nil
	}

	api, err := tgbotapi.NewBotAPI(b.cfg.TelegramToken)
	if err != nil {
		return fmt.Errorf("failed to initialize telegram bot: %w", err)
	}
	b.api = api
	b.ready = true

	log.Printf("[telegram] authenticated as @%s (Bot ID: %d)", api.Self.UserName, api.Self.ID)

	pollCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	// Launch background command poller
	go b.pollUpdates(pollCtx)

	return nil
}

// Close gracefully stops the Telegram polling loop.
func (b *Bot) Close() error {
	if b.cancel != nil {
		b.cancel()
	}
	b.ready = false
	return nil
}

// IsReady returns whether the Telegram bot is authenticated and active.
func (b *Bot) IsReady() bool {
	return b.ready && b.api != nil
}

// Message represents an outbound Telegram message payload.
type Message struct {
	Text       string
	ImageURL   string
	ButtonURL  string
	ButtonText string
}

// SendMessage delivers a formatted message or card to a designated Telegram chat ID or channel username.
func (b *Bot) SendMessage(target string, msg *Message) (*tgbotapi.Message, error) {
	if !b.IsReady() {
		return nil, fmt.Errorf("telegram bot is not configured or offline (check TELEGRAM_BOT_TOKEN)")
	}

	target = strings.TrimSpace(target)
	var chatID int64
	var channelUsername string

	if strings.HasPrefix(target, "@") {
		channelUsername = target
	} else {
		id, err := strconv.ParseInt(target, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid telegram chat id %q: must be numeric or @channel", target)
		}
		chatID = id
	}

	if msg == nil || (msg.Text == "" && msg.ImageURL == "") {
		return nil, fmt.Errorf("message text or image cannot be empty")
	}

	// Prepare inline button if URL is provided
	var keyboard *tgbotapi.InlineKeyboardMarkup
	if msg.ButtonURL != "" {
		btnLabel := msg.ButtonText
		if btnLabel == "" {
			btnLabel = "View Details"
		}
		btn := tgbotapi.NewInlineKeyboardButtonURL(btnLabel, msg.ButtonURL)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
		keyboard = &kb
	}

	// Case 1: Send Photo with HTML Caption
	if msg.ImageURL != "" {
		var photo tgbotapi.PhotoConfig
		if channelUsername != "" {
			photo = tgbotapi.NewPhotoToChannel(channelUsername, tgbotapi.FileURL(msg.ImageURL))
		} else {
			photo = tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(msg.ImageURL))
		}
		photo.Caption = msg.Text
		photo.ParseMode = tgbotapi.ModeHTML
		if keyboard != nil {
			photo.ReplyMarkup = *keyboard
		}

		sentMsg, err := b.api.Send(photo)
		if err != nil {
			return nil, fmt.Errorf("failed to send telegram photo: %w", err)
		}
		return &sentMsg, nil
	}

	// Case 2: Send Text Message
	var msgCfg tgbotapi.MessageConfig
	if channelUsername != "" {
		msgCfg = tgbotapi.NewMessageToChannel(channelUsername, msg.Text)
	} else {
		msgCfg = tgbotapi.NewMessage(chatID, msg.Text)
	}
	msgCfg.ParseMode = tgbotapi.ModeHTML
	msgCfg.DisableWebPagePreview = false
	if keyboard != nil {
		msgCfg.ReplyMarkup = *keyboard
	}

	sentMsg, err := b.api.Send(msgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to send telegram message: %w", err)
	}
	return &sentMsg, nil
}

// GetStatus returns the operational status and telemetry of the Telegram bot.
func (b *Bot) GetStatus() BotStatus {
	status := BotStatus{
		Enabled:   b.cfg.TelegramToken != "",
		Online:    b.ready && b.api != nil,
		Uptime:    time.Since(b.startTime).Round(time.Second).String(),
		UptimeSec: int64(time.Since(b.startTime).Seconds()),
	}
	if b.api != nil {
		status.Username = b.api.Self.UserName
	}
	return status
}

func (b *Bot) pollUpdates(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message != nil && update.Message.IsCommand() {
				b.handleCommand(update.Message)
			}
		}
	}
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "ping":
		reply := tgbotapi.NewMessage(msg.Chat.ID, "<b>Pong!</b> Telegram Gateway is operational.")
		reply.ParseMode = tgbotapi.ModeHTML
		b.api.Send(reply)

	case "status":
		status := b.GetStatus()
		text := fmt.Sprintf("<b>Crosscast Gateway Status</b>\n\n<b>Telegram:</b> Online (@%s)\n<b>Uptime:</b> <code>%s</code>",
			status.Username, status.Uptime)
		reply := tgbotapi.NewMessage(msg.Chat.ID, text)
		reply.ParseMode = tgbotapi.ModeHTML
		b.api.Send(reply)

	case "start":
		text := "Hello! I am <b>Crosscast</b>, an ultra-lightweight multi-tenant social & community broadcasting gateway.\n\nCommands: /ping, /status"
		reply := tgbotapi.NewMessage(msg.Chat.ID, text)
		reply.ParseMode = tgbotapi.ModeHTML
		b.api.Send(reply)
	}
}
