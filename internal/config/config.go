package config

import (
	"encoding/json"
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
	Channels         map[string]string // Target name -> Channel ID (parsed from JSON in CROSSCAST_CHANNELS)
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

	// Build target channels map from CROSSCAST_CHANNELS (strictly JSON object string)
	channels := make(map[string]string)
	if rawChannels := strings.TrimSpace(os.Getenv("CROSSCAST_CHANNELS")); rawChannels != "" {
		var jsonMap map[string]string
		if err := json.Unmarshal([]byte(rawChannels), &jsonMap); err == nil {
			for k, v := range jsonMap {
				kClean := strings.ToLower(strings.TrimSpace(k))
				vClean := strings.TrimSpace(v)
				if kClean != "" && vClean != "" {
					channels[kClean] = vClean
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
