package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// ChannelTarget defines destination IDs for a target alias across supported platforms.
type ChannelTarget struct {
	Discord  string `json:"discord,omitempty"`
	Telegram string `json:"telegram,omitempty"`
}

// Config holds runtime configuration for Crosscast.
type Config struct {
	Port             int
	APIKey           string
	DiscordToken     string
	DiscordAppID     string
	DiscordGuildID   string                   // Optional: restrict slash commands to a specific guild for instant updates
	DefaultChannelID string                   // Fallback Discord channel ID when target is "default" or omitted
	TelegramToken    string
	TelegramChatID   string                   // Fallback Telegram chat ID when target is "default" or omitted
	Channels         map[string]ChannelTarget // Target name -> Platform destinations
	StatusMessage    string                   // Presence custom status message
}

// Load loads configuration from environment variables and an optional .env file.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[config] notice: %v", err)
		}
	}

	port := getEnvAsInt("PORT", 6403)
	apiKey := strings.TrimSpace(os.Getenv("CROSSCAST_API_KEY"))
	discordToken := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	discordAppID := strings.TrimSpace(os.Getenv("DISCORD_APP_ID"))
	discordGuildID := strings.TrimSpace(os.Getenv("DISCORD_GUILD_ID"))
	defaultChannelID := strings.TrimSpace(os.Getenv("DISCORD_DEFAULT_CHANNEL_ID"))

	telegramToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	telegramChatID := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))

	statusMsg := strings.TrimSpace(os.Getenv("CROSSCAST_STATUS_MESSAGE"))
	if statusMsg == "" {
		statusMsg = "Broadcasting with Crosscast"
	}

	channels := make(map[string]ChannelTarget)

	// Parse CROSSCAST_CHANNELS strictly from JSON object string
	if rawChannels := strings.TrimSpace(os.Getenv("CROSSCAST_CHANNELS")); rawChannels != "" {
		var rawMap map[string]json.RawMessage
		if err := json.Unmarshal([]byte(rawChannels), &rawMap); err == nil {
			for k, rawVal := range rawMap {
				kClean := strings.ToLower(strings.TrimSpace(k))
				trimmed := strings.TrimSpace(string(rawVal))

				// If value is a nested JSON object: {"discord": "...", "telegram": "..."}
				if strings.HasPrefix(trimmed, "{") {
					var ct ChannelTarget
					if err := json.Unmarshal(rawVal, &ct); err == nil {
						channels[kClean] = ct
					}
				} else {
					// If value is a simple string: "-100...", "@channel", or "123456..."
					var s string
					if err := json.Unmarshal(rawVal, &s); err == nil {
						s = strings.TrimSpace(s)
						ct := ChannelTarget{}
						if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "@") || strings.HasPrefix(s, "tg:") || strings.HasPrefix(s, "telegram:") {
							ct.Telegram = strings.TrimPrefix(strings.TrimPrefix(s, "tg:"), "telegram:")
						} else {
							ct.Discord = strings.TrimPrefix(s, "discord:")
						}
						channels[kClean] = ct
					}
				}
			}
		} else {
			log.Printf("[config] warning: failed to parse CROSSCAST_CHANNELS as JSON: %v", err)
		}
	}

	return &Config{
		Port:             port,
		APIKey:           apiKey,
		DiscordToken:     discordToken,
		DiscordAppID:     discordAppID,
		DiscordGuildID:   discordGuildID,
		DefaultChannelID: defaultChannelID,
		TelegramToken:    telegramToken,
		TelegramChatID:   telegramChatID,
		Channels:         channels,
		StatusMessage:    statusMsg,
	}
}

func getEnvAsInt(key string, fallback int) int {
	valStr := strings.TrimSpace(os.Getenv(key))
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}
