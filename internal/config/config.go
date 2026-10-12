package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration for Crosscast.
type Config struct {
	Port             int
	APIKey           string
	DiscordToken     string
	DiscordAppID     string
	DiscordGuildID   string            // Optional: restrict slash commands to a specific guild for instant updates
	DefaultChannelID string            // Fallback channel ID when target is "default" or omitted
	Channels         map[string]string // Target name -> Channel ID (loaded from CROSSCAST_CHANNELS or CROSSCAST_CHANNEL_*)
	StatusMessage    string            // Discord presence custom status message
	TelegramToken    string
	TelegramChatID   string
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

	statusMsg := strings.TrimSpace(os.Getenv("CROSSCAST_STATUS_MESSAGE"))
	if statusMsg == "" {
		statusMsg = "Broadcasting with Crosscast 📡"
	}

	telegramToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	telegramChatID := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))

	// Build target channels map
	channels := make(map[string]string)

	// 1. Parse comma-separated CROSSCAST_CHANNELS e.g. "alerts:123456,announcements:789012"
	if rawChannels := os.Getenv("CROSSCAST_CHANNELS"); rawChannels != "" {
		pairs := strings.Split(rawChannels, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(strings.TrimSpace(pair), ":", 2)
			if len(kv) == 2 {
				k := strings.ToLower(strings.TrimSpace(kv[0]))
				v := strings.TrimSpace(kv[1])
				if k != "" && v != "" {
					channels[k] = v
				}
			}
		}
	}

	// 2. Parse individual CROSSCAST_CHANNEL_<TARGET> variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], "CROSSCAST_CHANNEL_") {
			target := strings.ToLower(strings.TrimPrefix(parts[0], "CROSSCAST_CHANNEL_"))
			val := strings.TrimSpace(parts[1])
			if target != "" && val != "" {
				channels[target] = val
			}
		}
	}

	return &Config{
		Port:             port,
		APIKey:           apiKey,
		DiscordToken:     discordToken,
		DiscordAppID:     discordAppID,
		DiscordGuildID:   discordGuildID,
		DefaultChannelID: defaultChannelID,
		Channels:         channels,
		StatusMessage:    statusMsg,
		TelegramToken:    telegramToken,
		TelegramChatID:   telegramChatID,
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
