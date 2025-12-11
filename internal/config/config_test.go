package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Monitor.PollInterval.Duration() != 100*time.Millisecond {
		t.Errorf("Expected poll_interval 100ms, got %v", cfg.Monitor.PollInterval.Duration())
	}

	if cfg.Monitor.GracePeriod.Duration() != 500*time.Millisecond {
		t.Errorf("Expected grace_period 500ms, got %v", cfg.Monitor.GracePeriod.Duration())
	}

	if cfg.Player.DefaultVolume != 0.5 {
		t.Errorf("Expected default_volume 0.5, got %f", cfg.Player.DefaultVolume)
	}

	if cfg.UI.Mode != "gui" {
		t.Errorf("Expected UI mode 'gui', got '%s'", cfg.UI.Mode)
	}

	if len(cfg.Blocklist.Patterns) == 0 {
		t.Errorf("Expected non-empty patterns list")
	}

	if len(cfg.Exceptions) == 0 {
		t.Errorf("Expected non-empty exceptions list")
	}
}

func TestValidateWithValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.Validate()
	if err != nil {
		t.Errorf("DefaultConfig should be valid, but got error: %v", err)
	}
}

func TestValidateWithInvalidPollInterval(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Monitor.PollInterval = durationValue(0)
	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for poll_interval = 0")
	}
}

func TestValidateWithInvalidGracePeriod(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Monitor.GracePeriod = durationValue(-1 * time.Millisecond)
	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for negative grace_period")
	}
}

func TestValidateWithEmptyLofiPlaylist(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Player.LofiPlaylist = ""
	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for empty lofi_playlist")
	}
}

func TestValidateWithInvalidVolume(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Player.DefaultVolume = 1.5
	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for volume > 1")
	}
}

func TestValidateWithEmptyBlocklist(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Blocklist.Patterns = []string{}
	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected validation error for empty blocklist patterns")
	}
}

func TestLoadConfigFromPath(t *testing.T) {
	// Create a temporary config file
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "test.yaml")

	content := `
blocklist:
  patterns:
    - "discord"
    - "youtube"

exceptions:
  - "youtube music"

player:
  lofi_playlist: "https://www.youtube.com/watch?v=test"
  fbi_meme_source: "embedded://fbi-meme.mp4"
  default_volume: 0.7

monitor:
  poll_interval: 200ms
  grace_period: 600ms

ui:
  mode: "gui"
  second_monitor: true
  always_on_top: false
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config from path: %v", err)
	}

	if cfg.Monitor.PollInterval.Duration() != 200*time.Millisecond {
		t.Errorf("Expected poll_interval 200ms, got %v", cfg.Monitor.PollInterval.Duration())
	}

	if cfg.Monitor.GracePeriod.Duration() != 600*time.Millisecond {
		t.Errorf("Expected grace_period 600ms, got %v", cfg.Monitor.GracePeriod.Duration())
	}

	if cfg.Player.DefaultVolume != 0.7 {
		t.Errorf("Expected default_volume 0.7, got %f", cfg.Player.DefaultVolume)
	}

	if len(cfg.Blocklist.Patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(cfg.Blocklist.Patterns))
	}

	if cfg.Blocklist.Patterns[0] != "discord" {
		t.Errorf("Expected first pattern 'discord', got %s", cfg.Blocklist.Patterns[0])
	}
}

func TestLoadConfigFromPathWithInvalidFile(t *testing.T) {
	_, err := LoadConfigFromPath("/nonexistent/path/config.yaml")
	if err == nil {
		t.Errorf("Expected error when loading from nonexistent path")
	}
}

func TestConfigYAMLParsing(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "test.yaml")

	content := `
blocklist:
  patterns:
    - "discord"
    - "slack"
    - "reddit"
    - "netflix"

exceptions:
  - "youtube music"

player:
  lofi_playlist: "https://example.com/playlist"
  fbi_meme_source: "embedded://fbi.mp4"
  default_volume: 0.5

monitor:
  poll_interval: 150ms
  grace_period: 750ms

ui:
  mode: "gui"
  second_monitor: false
  always_on_top: true
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Valid config failed validation: %v", err)
	}

	if len(cfg.Blocklist.Patterns) != 4 {
		t.Errorf("Expected 4 patterns, got %d", len(cfg.Blocklist.Patterns))
	}

	if !cfg.UI.AlwaysOnTop {
		t.Errorf("Expected always_on_top to be true")
	}

	if cfg.UI.SecondMonitor {
		t.Errorf("Expected second_monitor to be false")
	}
}
