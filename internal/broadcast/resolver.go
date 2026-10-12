package broadcast

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/betothewizard/crosscast/internal/config"
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

// Resolver maps logical target aliases to platform-specific destinations.
type Resolver struct {
	cfg *config.Config
}

// NewResolver creates a new target Resolver instance.
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

// Resolve maps a logical target alias and optional platform filter to resolved platform destinations.
func (r *Resolver) Resolve(target string, platform string) (*ResolvedTargets, error) {
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
