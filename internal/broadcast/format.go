package broadcast

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/betothewizard/crosscast/internal/platform/telegram"
	"github.com/bwmarrin/discordgo"
)

// ToDiscordMessage converts a BroadcastPayload into a discordgo.MessageSend.
func ToDiscordMessage(p *BroadcastPayload, defaultColor int) *discordgo.MessageSend {
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

// FormatPayloadToHTML converts a BroadcastPayload into Telegram-compatible HTML formatted text.
func FormatPayloadToHTML(p *BroadcastPayload) string {
	var sb strings.Builder

	if p.Title != "" {
		sb.WriteString("<b>" + html.EscapeString(p.Title) + "</b>\n\n")
	}

	if p.Description != "" {
		sb.WriteString(html.EscapeString(p.Description) + "\n\n")
	}

	if p.Content != "" {
		sb.WriteString(html.EscapeString(p.Content) + "\n\n")
	}

	if len(p.Fields) > 0 {
		for _, f := range p.Fields {
			sb.WriteString(fmt.Sprintf("- <b>%s:</b> %s\n", html.EscapeString(f.Name), html.EscapeString(f.Value)))
		}
		sb.WriteString("\n")
	}

	if p.Footer != nil && p.Footer.Text != "" {
		sb.WriteString("<i>" + html.EscapeString(p.Footer.Text) + "</i>")
	}

	return strings.TrimSpace(sb.String())
}

// ToTelegramMessage converts a BroadcastPayload into a telegram.Message.
func ToTelegramMessage(p *BroadcastPayload) *telegram.Message {
	text := FormatPayloadToHTML(p)
	if text == "" {
		text = p.Content
	}

	imageURL := p.Image
	if imageURL == "" {
		imageURL = p.Thumbnail
	}

	btnText := ""
	if p.URL != "" {
		btnText = "View Details"
	}

	return &telegram.Message{
		Text:       text,
		ImageURL:   imageURL,
		ButtonURL:  p.URL,
		ButtonText: btnText,
	}
}
