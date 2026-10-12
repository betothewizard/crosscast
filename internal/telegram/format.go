package telegram

import (
	"fmt"
	"html"
	"strings"

	"github.com/betothewizard/crosscast/internal/router"
)

// FormatPayloadToHTML converts a BroadcastPayload into Telegram-compatible HTML formatted text.
func FormatPayloadToHTML(p *router.BroadcastPayload) string {
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
