package router

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// AuthorPayload represents author metadata in a rich embed.
type AuthorPayload struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// FieldPayload represents a key-value field in a rich embed.
type FieldPayload struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// FooterPayload represents footer metadata in a rich embed.
type FooterPayload struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// BroadcastPayload represents an inbound broadcast request.
type BroadcastPayload struct {
	Platform    string          `json:"platform,omitempty"`    // "all" (default), "discord", or "telegram"
	Target      string          `json:"target"`                // Target alias ("jobs", "alerts") or direct channel ID
	Content     string          `json:"content,omitempty"`     // Raw content or mentions (e.g. "@everyone")
	Title       string          `json:"title,omitempty"`       // Embed Title
	Description string          `json:"description,omitempty"` // Embed Description (Markdown)
	URL         string          `json:"url,omitempty"`         // Embed Title URL
	Color       int             `json:"color,omitempty"`       // Hex integer (e.g. 0x10B981)
	Author      *AuthorPayload  `json:"author,omitempty"`
	Thumbnail   string          `json:"thumbnail,omitempty"`   // URL of thumbnail image
	Image       string          `json:"image,omitempty"`       // URL of banner image
	Fields      []FieldPayload  `json:"fields,omitempty"`
	Footer      *FooterPayload  `json:"footer,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`   // Optional ISO8601 string
}

// ToDiscordMessage converts a BroadcastPayload into a discordgo.MessageSend.
func (p *BroadcastPayload) ToDiscordMessage(defaultColor int) *discordgo.MessageSend {
	msg := &discordgo.MessageSend{
		Content: p.Content,
	}

	// Check if this payload has embed fields
	hasEmbed := p.Title != "" || p.Description != "" || len(p.Fields) > 0 ||
		p.Thumbnail != "" || p.Image != "" || p.Author != nil || p.Footer != nil

	if !hasEmbed {
		return msg
	}

	color := p.Color
	if color == 0 {
		color = defaultColor
	}

	embed := &discordgo.MessageEmbed{
		Title:       p.Title,
		Description: p.Description,
		URL:         p.URL,
		Color:       color,
	}

	if p.Author != nil && p.Author.Name != "" {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    p.Author.Name,
			URL:     p.Author.URL,
			IconURL: p.Author.IconURL,
		}
	}

	if p.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: p.Thumbnail,
		}
	}

	if p.Image != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: p.Image,
		}
	}

	if len(p.Fields) > 0 {
		embed.Fields = make([]*discordgo.MessageEmbedField, 0, len(p.Fields))
		for _, f := range p.Fields {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   f.Name,
				Value:  f.Value,
				Inline: f.Inline,
			})
		}
	}

	if p.Footer != nil && p.Footer.Text != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text:    p.Footer.Text,
			IconURL: p.Footer.IconURL,
		}
	}

	if p.Timestamp != "" {
		embed.Timestamp = p.Timestamp
	} else {
		embed.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	msg.Embeds = []*discordgo.MessageEmbed{embed}
	return msg
}
