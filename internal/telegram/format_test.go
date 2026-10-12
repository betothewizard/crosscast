package telegram_test

import (
	"strings"
	"testing"

	"github.com/betothewizard/crosscast/internal/router"
	"github.com/betothewizard/crosscast/internal/telegram"
)

func TestFormatPayloadToHTML(t *testing.T) {
	payload := &router.BroadcastPayload{
		Title:       "Senior Go Developer <Remote>",
		Description: "Looking for Go & Docker specialists.",
		Fields: []router.FieldPayload{
			{Name: "Salary", Value: "$3000 & Equity"},
			{Name: "Location", Value: "Global"},
		},
		Footer: &router.FooterPayload{
			Text: "Crosscast Alert",
		},
	}

	html := telegram.FormatPayloadToHTML(payload)

	if !strings.Contains(html, "<b>Senior Go Developer &lt;Remote&gt;</b>") {
		t.Errorf("title HTML escaping failed: %s", html)
	}

	if !strings.Contains(html, "Looking for Go &amp; Docker specialists.") {
		t.Errorf("description HTML escaping failed: %s", html)
	}

	if !strings.Contains(html, "▫️ <b>Salary:</b> $3000 &amp; Equity") {
		t.Errorf("field HTML escaping failed: %s", html)
	}

	if !strings.Contains(html, "<i>Crosscast Alert</i>") {
		t.Errorf("footer italic failed: %s", html)
	}
}
