// Package config provides YAML-based configuration loading and validation.
// It supports both bundled defaults and user overrides, with comprehensive validation
// of all settings including blocklist rules and monitor parameters.
//
// Configuration is hierarchical:
// 1. Bundled defaults (config/default.yaml)
// 2. User overrides (~/.vigilant/config.yaml)
// 3. Command-line flags (future enhancement)
//
// All configuration loading is thread-safe. The Config struct should be accessed
// via methods that validate input rather than modified directly.
package config

import (
	"encoding/json"
	"time"
)

// BlocklistConfig defines blocked patterns (all regex-based, case-insensitive)
// All patterns are compiled with (?i) flag for case-insensitive matching
// Patterns match against both window title and process name
type BlocklistConfig struct {
	// Patterns are regex patterns that match window titles or process names
	// Examples: "discord", "youtube", "reddit\\.com", "netflix"
	Patterns []string `yaml:"patterns"`
}

// PlayerConfig defines media player settings
type PlayerConfig struct {
	LofiPlaylist  string  `yaml:"lofi_playlist"`
	FBIMemeSource string  `yaml:"fbi_meme_source"`
	DefaultVolume float64 `yaml:"default_volume"`
}

// MonitorConfig defines monitoring behavior
type MonitorConfig struct {
	PollInterval DurationValue `yaml:"poll_interval"`
	GracePeriod  DurationValue `yaml:"grace_period"`
}

// DurationValue is a custom type that unmarshals both strings and time.Duration
type DurationValue time.Duration

// Backward compatibility alias
type durationValue = DurationValue

// UnmarshalJSON handles JSON unmarshaling
func (d *durationValue) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	return d.unmarshal(v)
}

// UnmarshalYAML handles YAML unmarshaling
func (d *durationValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	return d.unmarshal(v)
}

func (d *durationValue) unmarshal(v interface{}) error {
	switch val := v.(type) {
	case float64:
		*d = DurationValue(time.Duration(val) * time.Nanosecond)
	case string:
		dur, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		*d = DurationValue(dur)
	}
	return nil
}

// Duration returns the underlying time.Duration
func (d durationValue) Duration() time.Duration {
	return time.Duration(d)
}

// MarshalJSON marshals DurationValue to JSON
func (d durationValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// MarshalYAML marshals DurationValue to YAML
func (d durationValue) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// UIConfig defines UI/window settings
type UIConfig struct {
	Mode          string `yaml:"mode"`
	SecondMonitor bool   `yaml:"second_monitor"`
	AlwaysOnTop   bool   `yaml:"always_on_top"`
}

// Config is the root configuration struct
type Config struct {
	Blocklist  BlocklistConfig `yaml:"blocklist"`
	Exceptions []string        `yaml:"exceptions"` // Regex patterns for exceptions (case-insensitive)
	Player     PlayerConfig    `yaml:"player"`
	Monitor    MonitorConfig   `yaml:"monitor"`
	UI         UIConfig        `yaml:"ui"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Blocklist: BlocklistConfig{
			Patterns: []string{
				// Apps (matches process name or window title)
				"discord",
				"steam",
				"battle\\.net",
				// Websites (matches browser tab titles)
				"youtube",
				"twitter",
				"reddit",
				"instagram",
				"tiktok",
				"facebook",
				"twitch",
				// Streaming services
				"netflix",
				"prime video",
				"disney\\+",
				"hulu",
			},
		},
		Exceptions: []string{
			"youtube music",
			"youtube studio",
		},
		Player: PlayerConfig{
			LofiPlaylist:  "https://www.youtube.com/watch?v=jfKfPfyJRdk",
			FBIMemeSource: "embedded://fbi-meme.mp4",
			DefaultVolume: 0.5,
		},
		Monitor: MonitorConfig{
			PollInterval: DurationValue(100 * time.Millisecond),
			GracePeriod:  DurationValue(500 * time.Millisecond),
		},
		UI: UIConfig{
			Mode:          "gui",
			SecondMonitor: true,
			AlwaysOnTop:   false,
		},
	}
}
