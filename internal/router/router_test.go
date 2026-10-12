package router_test

import (
	"testing"

	"github.com/betothewizard/crosscast/internal/config"
	"github.com/betothewizard/crosscast/internal/router"
)

func TestRouter_Resolve(t *testing.T) {
	cfg := &config.Config{
		DefaultChannelID: "000000000000000000",
		Channels: map[string]string{
			"alerts":        "111111111111111111",
			"announcements": "222222222222222222",
			"deployments":   "333333333333333333",
		},
	}
	r := router.New(cfg)

	tests := []struct {
		name        string
		target      string
		expectedID  string
		expectedCol int
		expectErr   bool
	}{
		{"alerts target", "alerts", "111111111111111111", router.ColorDanger, false},
		{"announcements target", "announcements", "222222222222222222", router.ColorInfo, false},
		{"deployments target", "deployments", "333333333333333333", router.ColorDefault, false},
		{"default target", "default", "000000000000000000", router.ColorDefault, false},
		{"empty target falls back to default", "", "000000000000000000", router.ColorDefault, false},
		{"direct snowflake ID", "999999999999999999", "999999999999999999", router.ColorDefault, false},
		{"unconfigured target", "unknown_channel", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := r.Resolve(tt.target)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error for target %q, got nil", tt.target)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.ChannelID != tt.expectedID {
				t.Errorf("channelID = %v, want %v", info.ChannelID, tt.expectedID)
			}
			if info.DefaultColor != tt.expectedCol {
				t.Errorf("color = %v, want %v", info.DefaultColor, tt.expectedCol)
			}
		})
	}
}

func TestBroadcastPayload_ToDiscordMessage(t *testing.T) {
	payload := &router.BroadcastPayload{
		Target:      "deployments",
		Title:       "Production Release v1.2.0",
		Description: "Deployed successfully to cloud cluster",
		URL:         "https://github.com/example/repo/releases/v1.2.0",
		Author: &router.AuthorPayload{
			Name: "GitHub Actions",
		},
		Fields: []router.FieldPayload{
			{Name: "Environment", Value: "Production", Inline: true},
			{Name: "Status", Value: "Success", Inline: true},
		},
	}

	msg := payload.ToDiscordMessage(router.ColorSuccess)
	if len(msg.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(msg.Embeds))
	}

	embed := msg.Embeds[0]
	if embed.Title != "Production Release v1.2.0" {
		t.Errorf("title = %v, want 'Production Release v1.2.0'", embed.Title)
	}
	if embed.Color != router.ColorSuccess {
		t.Errorf("color = %v, want %v", embed.Color, router.ColorSuccess)
	}
	if len(embed.Fields) != 2 {
		t.Errorf("fields len = %d, want 2", len(embed.Fields))
	}
}
