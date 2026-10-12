package config_test

import (
	"os"
	"testing"

	"github.com/betothewizard/crosscast/internal/config"
)

func TestConfig_Load_MultiPlatformChannels(t *testing.T) {
	jsonChannels := `{
		"jobs": "123456789012345678",
		"alerts": "-1001234567890",
		"broadcast": {
			"discord": "999888777666",
			"telegram": "-100999888777"
		}
	}`

	os.Setenv("PORT", "6403")
	os.Setenv("CROSSCAST_CHANNELS", jsonChannels)
	os.Setenv("TELEGRAM_BOT_TOKEN", "123456:ABC-DEF")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CROSSCAST_CHANNELS")
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
	}()

	cfg := config.Load()
	if cfg.Port != 6403 {
		t.Errorf("port = %d, want 6403", cfg.Port)
	}

	// Jobs should route to Discord
	if cfg.Channels["jobs"].Discord != "123456789012345678" {
		t.Errorf("jobs discord = %q, want 123456789012345678", cfg.Channels["jobs"].Discord)
	}
	if cfg.Channels["jobs"].Telegram != "" {
		t.Errorf("jobs telegram should be empty, got %q", cfg.Channels["jobs"].Telegram)
	}

	// Alerts should route to Telegram
	if cfg.Channels["alerts"].Telegram != "-1001234567890" {
		t.Errorf("alerts telegram = %q, want -1001234567890", cfg.Channels["alerts"].Telegram)
	}

	// Broadcast should route to both
	if cfg.Channels["broadcast"].Discord != "999888777666" {
		t.Errorf("broadcast discord = %q, want 999888777666", cfg.Channels["broadcast"].Discord)
	}
	if cfg.Channels["broadcast"].Telegram != "-100999888777" {
		t.Errorf("broadcast telegram = %q, want -100999888777", cfg.Channels["broadcast"].Telegram)
	}
}
