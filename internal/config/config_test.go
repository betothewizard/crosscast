package config_test

import (
	"os"
	"testing"

	"github.com/betothewizard/crosscast/internal/config"
)

func TestConfig_Load_JSONChannels(t *testing.T) {
	os.Setenv("PORT", "6403")
	os.Setenv("CROSSCAST_CHANNELS", `{"jobs":"123456789012345678","alerts":"987654321098765432"}`)
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CROSSCAST_CHANNELS")
	}()

	cfg := config.Load()
	if cfg.Port != 6403 {
		t.Errorf("port = %d, want 6403", cfg.Port)
	}

	if cfg.Channels["jobs"] != "123456789012345678" {
		t.Errorf("jobs channel = %q, want 123456789012345678", cfg.Channels["jobs"])
	}

	if cfg.Channels["alerts"] != "987654321098765432" {
		t.Errorf("alerts channel = %q, want 987654321098765432", cfg.Channels["alerts"])
	}
}

func TestConfig_Load_CommaSeparatedFallback(t *testing.T) {
	os.Setenv("CROSSCAST_CHANNELS", "jobs:111,alerts:222")
	defer os.Unsetenv("CROSSCAST_CHANNELS")

	cfg := config.Load()
	if cfg.Channels["jobs"] != "111" {
		t.Errorf("jobs channel = %q, want 111", cfg.Channels["jobs"])
	}
	if cfg.Channels["alerts"] != "222" {
		t.Errorf("alerts channel = %q, want 222", cfg.Channels["alerts"])
	}
}
