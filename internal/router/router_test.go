package router_test

import (
	"testing"

	"github.com/betothewizard/crosscast/internal/config"
	"github.com/betothewizard/crosscast/internal/router"
)

func TestRouter_Resolve_MultiPlatform(t *testing.T) {
	cfg := &config.Config{
		DefaultChannelID: "000000000000000000",
		TelegramChatID:   "-1000000000000",
		Channels: map[string]config.ChannelTarget{
			"alerts": {
				Discord:  "111111111111111111",
				Telegram: "-1001111111111",
			},
			"discord_only": {
				Discord: "222222222222222222",
			},
			"telegram_only": {
				Telegram: "-1002222222222",
			},
		},
	}
	r := router.New(cfg)

	// Case 1: Both platforms for alerts
	res, err := r.Resolve("alerts", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscordChannelID != "111111111111111111" {
		t.Errorf("discord id = %q, want 111111111111111111", res.DiscordChannelID)
	}
	if res.TelegramChatID != "-1001111111111" {
		t.Errorf("telegram id = %q, want -1001111111111", res.TelegramChatID)
	}

	// Case 2: Filter platform to telegram only
	res, err = r.Resolve("alerts", "telegram")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscordChannelID != "" {
		t.Errorf("discord id should be empty, got %q", res.DiscordChannelID)
	}
	if res.TelegramChatID != "-1001111111111" {
		t.Errorf("telegram id = %q, want -1001111111111", res.TelegramChatID)
	}

	// Case 3: Direct Telegram chat ID
	res, err = r.Resolve("-100999888777", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TelegramChatID != "-100999888777" {
		t.Errorf("telegram id = %q, want -100999888777", res.TelegramChatID)
	}

	// Case 4: Direct Discord snowflake
	res, err = r.Resolve("999999999999999999", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DiscordChannelID != "999999999999999999" {
		t.Errorf("discord id = %q, want 999999999999999999", res.DiscordChannelID)
	}

	// Case 5: Unknown target
	_, err = r.Resolve("unknown", "all")
	if err == nil {
		t.Fatalf("expected error for unknown target, got nil")
	}
}
