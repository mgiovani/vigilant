package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from multiple sources in priority order:
// 1. If path is provided, load from that path
// 2. User config at ~/.vigilant/config.yaml
// 3. Bundled default config at config/default.yaml
// Returns the first valid config found, or error if none exist
func LoadConfig(path string) (*Config, error) {
	// If path is provided, try it first
	if path != "" {
		if cfg, err := loadFromPath(path); err == nil {
			return cfg, nil
		}
	}

	// Try user config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homeDir, ".vigilant", "config.yaml")
		if cfg, err := loadFromPath(userConfigPath); err == nil {
			return cfg, nil
		}
	}

	// Try bundled default config
	bundledPath := "config/default.yaml"
	if cfg, err := loadFromPath(bundledPath); err == nil {
		return cfg, nil
	}

	// Try relative to executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		bundledPath = filepath.Join(exeDir, "config", "default.yaml")
		if cfg, err := loadFromPath(bundledPath); err == nil {
			return cfg, nil
		}
	}

	// Return error if no config found
	return nil, fmt.Errorf("no configuration file found. tried %s, ~/.vigilant/config.yaml and bundled config/default.yaml", path)
}

// loadFromPath attempts to load and parse config from a specific file path
func loadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from %s: %w", path, err)
	}

	return cfg, nil
}

// LoadConfigFromPath loads configuration from a specific file path
func LoadConfigFromPath(path string) (*Config, error) {
	return loadFromPath(path)
}

// MarshalConfig converts a Config struct back to YAML bytes
func MarshalConfig(cfg *Config) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	return yaml.Marshal(cfg)
}
