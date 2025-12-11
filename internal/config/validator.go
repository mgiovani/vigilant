package config

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Validate checks the configuration for validity and returns any errors
// It also logs warnings for non-critical issues
func (c *Config) Validate() error {
	var errs []string

	// Validate monitor settings
	if c.Monitor.PollInterval.Duration() <= 0 {
		errs = append(errs, "monitor.poll_interval must be greater than 0")
	}

	if c.Monitor.GracePeriod.Duration() < 0 {
		errs = append(errs, "monitor.grace_period must be non-negative")
	}

	// Validate player settings
	if strings.TrimSpace(c.Player.LofiPlaylist) == "" {
		errs = append(errs, "player.lofi_playlist must not be empty")
	}

	if c.Player.DefaultVolume < 0 || c.Player.DefaultVolume > 1 {
		errs = append(errs, "player.default_volume must be between 0 and 1")
	}

	// Validate blocklist
	if len(c.Blocklist.Patterns) == 0 {
		errs = append(errs, "blocklist.patterns must contain at least one pattern")
	}

	// Validate UI mode
	validModes := map[string]bool{"gui": true, "tui": true}
	if !validModes[strings.ToLower(c.UI.Mode)] {
		log.Printf("WARNING: unknown ui.mode '%s', defaulting to 'gui'", c.UI.Mode)
		c.UI.Mode = "gui"
	}

	// Log warnings for non-critical issues
	if c.Player.DefaultVolume < 0.1 {
		log.Printf("WARNING: player.default_volume is very low (%.1f), audio may be inaudible", c.Player.DefaultVolume)
	}

	if c.Monitor.PollInterval.Duration() > 1*time.Second {
		log.Printf("WARNING: monitor.poll_interval is very high (%v), responsiveness may suffer", c.Monitor.PollInterval.Duration())
	}

	if len(c.Exceptions) > 0 && len(c.Blocklist.Patterns) == 0 {
		log.Printf("WARNING: exceptions are defined but no blocklist patterns exist")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - " + strings.Join(errs, "\n  - "))
	}

	return nil
}
