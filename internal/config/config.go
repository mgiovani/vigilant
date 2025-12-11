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

// BlocklistConfig defines blocked processes, websites, and patterns
type BlocklistConfig struct {
	Processes []string `yaml:"processes"`
	Websites  []string `yaml:"websites"`
	Patterns  []string `yaml:"patterns"`
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
	Exceptions []string        `yaml:"exceptions"`
	Player     PlayerConfig    `yaml:"player"`
	Monitor    MonitorConfig   `yaml:"monitor"`
	UI         UIConfig        `yaml:"ui"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Blocklist: BlocklistConfig{
			Processes: []string{
				"Discord",
				"Discord.exe",
				"Slack",
				"Slack.exe",
				"Steam",
				"Steam.exe",
				"Battle.net",
			},
			Websites: []string{
				"youtube.com",
				"twitter.com",
				"x.com",
				"reddit.com",
				"instagram.com",
				"tiktok.com",
				"facebook.com",
				"twitch.tv",
			},
			Patterns: []string{
				".*Netflix.*",
				".*Prime Video.*",
				".*Disney\\+.*",
				".*Hulu.*",
			},
		},
		Exceptions: []string{
			"YouTube Music",
			"Work - YouTube",
			"YouTube Studio",
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
