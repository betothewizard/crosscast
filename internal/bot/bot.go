package bot

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/betothewizard/crosscast/internal/config"
	"github.com/bwmarrin/discordgo"
)

// BotStatus represents real-time operational metrics of the Discord session.
type BotStatus struct {
	Online       bool          `json:"online"`
	Username     string        `json:"username,omitempty"`
	GuildsCount  int           `json:"guilds_count"`
	HeartbeatMs  int64         `json:"heartbeat_ms"`
	Uptime       string        `json:"uptime"`
	UptimeSec    int64         `json:"uptime_seconds"`
	MemoryAllocMb float64      `json:"memory_alloc_mb"`
	NumGoroutine int           `json:"num_goroutine"`
}

// Bot manages the 24/7 Discord WebSocket session, presence, and interaction handlers.
type Bot struct {
	cfg       *config.Config
	session   *discordgo.Session
	startTime time.Time
	ready     bool
	cmdIDs    []string
}

// New creates a new Bot instance.
func New(cfg *config.Config) *Bot {
	return &Bot{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// Start opens the Discord WebSocket session and registers listeners.
func (b *Bot) Start(ctx context.Context) error {
	if b.cfg.DiscordToken == "" {
		log.Printf("[bot] warning: DISCORD_BOT_TOKEN is empty; bot gateway will operate in offline/mock mode")
		return nil
	}

	session, err := discordgo.New("Bot " + b.cfg.DiscordToken)
	if err != nil {
		return fmt.Errorf("failed to create discord session: %w", err)
	}
	b.session = session

	// Identify intents: Guilds is sufficient for receiving channel events & presence
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	session.AddHandler(b.handleReady)
	session.AddHandler(b.handleInteraction)

	if err := session.Open(); err != nil {
		return fmt.Errorf("failed to connect to discord gateway: %w", err)
	}

	log.Printf("[bot] connecting to Discord gateway...")
	return nil
}

// Close gracefully disconnects from the Discord gateway and cleans up slash commands.
func (b *Bot) Close() error {
	if b.session == nil {
		return nil
	}
	log.Printf("[bot] disconnecting from Discord gateway...")
	return b.session.Close()
}

// SendMessage sends a pre-constructed Discord message to the designated channel.
func (b *Bot) SendMessage(channelID string, msg *discordgo.MessageSend) (*discordgo.Message, error) {
	if b.session == nil {
		return nil, fmt.Errorf("discord bot session is not initialized (check DISCORD_BOT_TOKEN)")
	}
	return b.session.ChannelMessageSendComplex(channelID, msg)
}

// IsReady returns whether the Discord session is active and ready to send messages.
func (b *Bot) IsReady() bool {
	return b.session != nil
}

// GetStatus returns the operational status and telemetry of the bot.
func (b *Bot) GetStatus() BotStatus {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	status := BotStatus{
		Online:        b.ready && b.session != nil,
		Uptime:        time.Since(b.startTime).Round(time.Second).String(),
		UptimeSec:     int64(time.Since(b.startTime).Seconds()),
		MemoryAllocMb: float64(mem.Alloc) / (1024 * 1024),
		NumGoroutine:  runtime.NumGoroutine(),
	}

	if b.session != nil && b.session.State != nil && b.session.State.User != nil {
		status.Username = b.session.State.User.Username + "#" + b.session.State.User.Discriminator
		status.GuildsCount = len(b.session.State.Guilds)
		status.HeartbeatMs = b.session.HeartbeatLatency().Milliseconds()
	}

	return status
}
