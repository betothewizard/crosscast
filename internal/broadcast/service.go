package broadcast

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/betothewizard/crosscast/internal/config"
	"github.com/betothewizard/crosscast/internal/platform/discord"
	"github.com/betothewizard/crosscast/internal/platform/telegram"
)

// DeliveryResult records delivery metadata for a specific platform.
type DeliveryResult struct {
	Platform  string `json:"platform"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id,omitempty"`
}

// BroadcastResult represents the result of a broadcast operation.
type BroadcastResult struct {
	Success    bool                      `json:"success"`
	Target     string                    `json:"target"`
	Deliveries map[string]DeliveryResult `json:"deliveries"`
	Timestamp  time.Time                 `json:"timestamp"`
}

// Service orchestrates multi-platform broadcasting.
type Service struct {
	cfg      *config.Config
	discord  *discord.Bot
	telegram *telegram.Bot
	resolver *Resolver
}

// NewService creates a new broadcasting Service.
func NewService(cfg *config.Config, discordBot *discord.Bot, telegramBot *telegram.Bot) *Service {
	return &Service{
		cfg:      cfg,
		discord:  discordBot,
		telegram: telegramBot,
		resolver: NewResolver(cfg),
	}
}

// Broadcast resolves targets, prepares messages, and dispatches to platforms.
func (s *Service) Broadcast(ctx context.Context, payload *BroadcastPayload) (*BroadcastResult, error) {
	res, err := s.resolver.Resolve(payload.Target, payload.Platform)
	if err != nil {
		return nil, err
	}

	deliveries := make(map[string]DeliveryResult)
	var errs []string

	// 1. Deliver to Discord
	if res.DiscordChannelID != "" {
		if s.discord != nil && s.discord.IsReady() {
			discordMsg := ToDiscordMessage(payload, res.DefaultColor)
			sentMsg, err := s.discord.SendMessage(res.DiscordChannelID, discordMsg)
			if err != nil {
				errs = append(errs, fmt.Sprintf("discord: %v", err))
			} else {
				deliveries["discord"] = DeliveryResult{
					Platform:  "discord",
					ChannelID: res.DiscordChannelID,
					MessageID: sentMsg.ID,
				}
			}
		} else {
			errs = append(errs, "discord: bot session not active")
		}
	}

	// 2. Deliver to Telegram
	if res.TelegramChatID != "" {
		if s.telegram != nil && s.telegram.IsReady() {
			tgMsg := ToTelegramMessage(payload)
			sentMsg, err := s.telegram.SendMessage(res.TelegramChatID, tgMsg)
			if err != nil {
				errs = append(errs, fmt.Sprintf("telegram: %v", err))
			} else {
				deliveries["telegram"] = DeliveryResult{
					Platform:  "telegram",
					ChannelID: res.TelegramChatID,
					MessageID: strconv.Itoa(sentMsg.MessageID),
				}
			}
		} else {
			errs = append(errs, "telegram: bot session not active")
		}
	}

	if len(deliveries) == 0 {
		errMsg := "failed to deliver to any platform"
		if len(errs) > 0 {
			errMsg = strings.Join(errs, "; ")
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	return &BroadcastResult{
		Success:    true,
		Target:     payload.Target,
		Deliveries: deliveries,
		Timestamp:  time.Now().UTC(),
	}, nil
}

// DiscordStatus returns the telemetry status of the Discord bot.
func (s *Service) DiscordStatus() any {
	if s.discord == nil {
		return nil
	}
	return s.discord.GetStatus()
}

// TelegramStatus returns the telemetry status of the Telegram bot.
func (s *Service) TelegramStatus() any {
	if s.telegram == nil {
		return nil
	}
	return s.telegram.GetStatus()
}
