package broadcast_test

import (
	"strings"
	"testing"

	"github.com/betothewizard/crosscast/internal/broadcast"
)

func TestFormatPayloadToHTML(t *testing.T) {
	payload := &broadcast.BroadcastPayload{
		Title:       "Test Alert <&>",
		Description: "This is a test description with <b>raw</b> tag",
		Fields: []broadcast.FieldPayload{
			{Name: "Status", Value: "OK"},
			{Name: "Target", Value: "Cluster-1"},
		},
		Footer: &broadcast.FooterPayload{
			Text: "Sent via Crosscast",
		},
	}

	html := broadcast.FormatPayloadToHTML(payload)

	// Check escaped HTML characters
	if !strings.Contains(html, "&lt;&amp;&gt;") {
		t.Errorf("expected escaped title, got %q", html)
	}
	if !strings.Contains(html, "&lt;b&gt;raw&lt;/b&gt;") {
		t.Errorf("expected escaped description tags, got %q", html)
	}
	if !strings.Contains(html, "- <b>Status:</b> OK") {
		t.Errorf("expected formatted field, got %q", html)
	}
	if !strings.Contains(html, "<i>Sent via Crosscast</i>") {
		t.Errorf("expected formatted footer, got %q", html)
	}
}

func TestToDiscordMessage(t *testing.T) {
	payload := &broadcast.BroadcastPayload{
		Content:     "@everyone",
		Title:       "Notice",
		Description: "Maintenance window",
		Fields: []broadcast.FieldPayload{
			{Name: "Duration", Value: "10m", Inline: true},
		},
	}

	msg := broadcast.ToDiscordMessage(payload, broadcast.ColorWarning)

	if msg.Content != "@everyone" {
		t.Errorf("content = %q, want @everyone", msg.Content)
	}
	if len(msg.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(msg.Embeds))
	}
	embed := msg.Embeds[0]
	if embed.Title != "Notice" {
		t.Errorf("embed.Title = %q, want Notice", embed.Title)
	}
	if embed.Color != broadcast.ColorWarning {
		t.Errorf("embed.Color = %d, want %d", embed.Color, broadcast.ColorWarning)
	}
}

func TestToTelegramMessage(t *testing.T) {
	payload := &broadcast.BroadcastPayload{
		Title:       "Title",
		Description: "Desc",
		URL:         "https://example.com",
		Image:       "https://example.com/image.png",
	}

	tgMsg := broadcast.ToTelegramMessage(payload)

	if !strings.Contains(tgMsg.Text, "<b>Title</b>") {
		t.Errorf("expected bold title, got %q", tgMsg.Text)
	}
	if tgMsg.ImageURL != "https://example.com/image.png" {
		t.Errorf("imageURL = %q, want https://example.com/image.png", tgMsg.ImageURL)
	}
	if tgMsg.ButtonURL != "https://example.com" {
		t.Errorf("buttonURL = %q, want https://example.com", tgMsg.ButtonURL)
	}
	if tgMsg.ButtonText != "View Details" {
		t.Errorf("buttonText = %q, want 'View Details'", tgMsg.ButtonText)
	}
}
