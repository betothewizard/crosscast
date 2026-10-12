package router

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/betothewizard/crosscast/internal/config"
	"github.com/bwmarrin/discordgo"
)

var (
	discordSnowflakeRegex = regexp.MustCompile(`^\d{17,20}$`)
	telegramChatIDRegex   = regexp.MustCompile(`^(-\d{5,20}|@\w{3,32})$`)
)

// Standard color palette (Hex)
const (
	ColorDefault = 0x5865F2 // Discord Blurple
	ColorSuccess = 0x10B981 // Emerald Green
	ColorWarning = 0xF59E0B // Amber Warning
	ColorDanger  = 0xEF4444 // Crimson Red
	ColorInfo    = 0x3B82F6 // Blue
)

// ResolvedTargets holds resolved destinations for Discord and Telegram.
type ResolvedTargets struct {
	DiscordChannelID string
	TelegramChatID   string
	DefaultColor     int
}

// Router manages routing rules between logical target names and destination platform channels.
type Router struct {
	cfg *config.Config
}

// New creates a new Router instance.
func New(cfg *config.Config) *Router {
	return &Router{cfg: cfg}
}

// Resolve maps a logical target and optional platform filter to resolved platform destinations.
func (r *Router) Resolve(target string, platform string) (*ResolvedTargets, error) {
	t := strings.ToLower(strings.TrimSpace(target))
	plat := strings.ToLower(strings.TrimSpace(platform))
	if plat == "" {
		plat = "all"
	}

	allowDiscord := plat == "all" || plat == "discord"
	allowTelegram := plat == "all" || plat == "telegram"

	res := &ResolvedTargets{
		DefaultColor: pickDefaultColor(t),
	}

	// 1. Handle default/empty target
	if t == "" || t == "default" {
		if allowDiscord && r.cfg.DefaultChannelID != "" {
			res.DiscordChannelID = r.cfg.DefaultChannelID
		}
		if allowTelegram && r.cfg.TelegramChatID != "" {
			res.TelegramChatID = r.cfg.TelegramChatID
		}
		if res.DiscordChannelID == "" && res.TelegramChatID == "" {
			return nil, fmt.Errorf("target is default/empty but no default channel is configured")
		}
		return res, nil
	}

	// 2. Check dynamic channel map from config
	if dest, exists := r.cfg.Channels[t]; exists {
		if allowDiscord && dest.Discord != "" {
			res.DiscordChannelID = dest.Discord
		}
		if allowTelegram && dest.Telegram != "" {
			res.TelegramChatID = dest.Telegram
		}
		if res.DiscordChannelID != "" || res.TelegramChatID != "" {
			return res, nil
		}
	}

	// 3. Direct Discord snowflake ID (17-20 digits)
	if allowDiscord && discordSnowflakeRegex.MatchString(t) {
		res.DiscordChannelID = t
		return res, nil
	}

	// 4. Direct Telegram chat ID (starts with '-' or '@')
	if allowTelegram && (telegramChatIDRegex.MatchString(t) || strings.HasPrefix(t, "-") || strings.HasPrefix(t, "@")) {
		res.TelegramChatID = t
		return res, nil
	}

	return nil, fmt.Errorf("target '%s' could not be resolved for platform '%s' (configure in CROSSCAST_CHANNELS)", target, plat)
}

// PrepareDiscordMessage validates and constructs the discordgo.MessageSend for a given payload.
func (r *Router) PrepareDiscordMessage(payload *BroadcastPayload, defaultColor int) *discordgo.MessageSend {
	return payload.ToDiscordMessage(defaultColor)
}

func pickDefaultColor(target string) int {
	switch {
	case strings.Contains(target, "alert") || strings.Contains(target, "error") || strings.Contains(target, "fail"):
		return ColorDanger
	case strings.Contains(target, "warn"):
		return ColorWarning
	case strings.Contains(target, "success") || strings.Contains(target, "done") || strings.Contains(target, "pay"):
		return ColorSuccess
	case strings.Contains(target, "info") || strings.Contains(target, "news") || strings.Contains(target, "announc"):
		return ColorInfo
	default:
		return ColorDefault
	}
}
