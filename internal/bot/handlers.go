package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Check Crosscast bot gateway latency and responsiveness",
	},
	{
		Name:        "status",
		Description: "Display operational metrics, uptime, and memory usage of Crosscast",
	},
}

func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	b.ready = true
	log.Printf("[bot] authenticated as %s#%s across %d guilds",
		r.User.Username, r.User.Discriminator, len(r.Guilds))

	// Update presence status (Online + Custom Activity)
	err := s.UpdateCustomStatus(b.cfg.StatusMessage)
	if err != nil {
		log.Printf("[bot] failed to set presence: %v", err)
	}

	// Register slash commands (either guild-scoped for instant registration or global)
	appID := b.cfg.DiscordAppID
	if appID == "" && s.State.User != nil {
		appID = s.State.User.ID
	}

	if appID != "" {
		for _, cmd := range commands {
			_, err := s.ApplicationCommandCreate(appID, b.cfg.DiscordGuildID, cmd)
			if err != nil {
				log.Printf("[bot] failed to register command /%s: %v", cmd.Name, err)
			} else {
				log.Printf("[bot] registered slash command /%s", cmd.Name)
			}
		}
	}
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "ping":
		latency := s.HeartbeatLatency().Milliseconds()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("**Pong!** WebSocket gateway latency: `%dms`", latency),
			},
		})

	case "status":
		status := b.GetStatus()
		embed := &discordgo.MessageEmbed{
			Title:       "Crosscast Gateway Status",
			Description: "Ultra-lightweight multi-tenant social & community broadcasting daemon.",
			Color:       0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Status",
					Value:  "**Online & Connected**",
					Inline: true,
				},
				{
					Name:   "Uptime",
					Value:  fmt.Sprintf("`%s`", status.Uptime),
					Inline: true,
				},
				{
					Name:   "Gateway Ping",
					Value:  fmt.Sprintf("`%dms`", status.HeartbeatMs),
					Inline: true,
				},
				{
					Name:   "Memory Usage",
					Value:  fmt.Sprintf("`%.2f MB`", status.MemoryAllocMb),
					Inline: true,
				},
				{
					Name:   "Goroutines",
					Value:  fmt.Sprintf("`%d`", status.NumGoroutine),
					Inline: true,
				},
				{
					Name:   "Guilds",
					Value:  fmt.Sprintf("`%d`", status.GuildsCount),
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Crosscast Engine • Broadcast Gateway",
			},
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}
}
