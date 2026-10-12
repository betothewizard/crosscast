package router

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/betothewizard/crosscast/internal/config"
	"github.com/bwmarrin/discordgo"
)

var numericRegex = regexp.MustCompile(`^\d{17,20}$`)

// Standard color palette (Hex)
const (
	ColorDefault = 0x5865F2 // Discord Blurple
	ColorSuccess = 0x10B981 // Emerald Green
	ColorWarning = 0xF59E0B // Amber Warning
	ColorDanger  = 0xEF4444 // Crimson Red
	ColorInfo    = 0x3B82F6 // Blue
)

// TargetInfo holds the resolved channel ID and default color.
type TargetInfo struct {
	ChannelID    string
	DefaultColor int
}

// Router manages routing rules between logical target names and Discord channel IDs.
type Router struct {
	cfg *config.Config
}

// New creates a new Router instance.
func New(cfg *config.Config) *Router {
	return &Router{cfg: cfg}
}

// Resolve maps a logical target string to a Discord channel ID and appropriate default color.
func (r *Router) Resolve(target string) (*TargetInfo, error) {
	t := strings.ToLower(strings.TrimSpace(target))

	// 1. Handle default/empty target
	if t == "" || t == "default" {
		if r.cfg.DefaultChannelID != "" {
			return &TargetInfo{ChannelID: r.cfg.DefaultChannelID, DefaultColor: ColorDefault}, nil
		}
		return nil, fmt.Errorf("target is empty and DISCORD_DEFAULT_CHANNEL_ID is not configured")
	}

	// 2. Check dynamic channel map from config
	if channelID, exists := r.cfg.Channels[t]; exists && channelID != "" {
		color := pickDefaultColor(t)
		return &TargetInfo{ChannelID: channelID, DefaultColor: color}, nil
	}

	// 3. Direct snowflake ID (17-20 digits)
	if numericRegex.MatchString(t) {
		return &TargetInfo{ChannelID: t, DefaultColor: ColorDefault}, nil
	}

	return nil, fmt.Errorf("unknown target '%s': configure via CROSSCAST_CHANNELS or CROSSCAST_CHANNEL_%s, or pass a Discord channel ID",
		target, strings.ToUpper(target))
}

// PrepareMessage validates and constructs the discordgo.MessageSend for a given payload.
func (r *Router) PrepareMessage(payload *BroadcastPayload) (string, *discordgo.MessageSend, error) {
	info, err := r.Resolve(payload.Target)
	if err != nil {
		return "", nil, err
	}

	msg := payload.ToDiscordMessage(info.DefaultColor)
	return info.ChannelID, msg, nil
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
