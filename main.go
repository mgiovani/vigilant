package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	appruntime "vigilant/internal/app"
	"vigilant/internal/config"
	"vigilant/internal/youtube"
)

//go:embed all:frontend/dist
var assets embed.FS

// VanillaApp wraps the orchestrator app for Wails bindings
type VanillaApp struct {
	orchestrator *appruntime.App
	embedProxy   *youtube.EmbedProxy
	configPath   string
	wailsCtx     context.Context
}

// extractVideoID extracts YouTube video ID from various URL formats
// Supports: youtube.com/watch?v=ID, youtu.be/ID, youtube.com/embed/ID, or plain ID
func extractVideoID(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// Pattern for youtube.com/watch?v=ID
	watchPattern := regexp.MustCompile(`(?:youtube\.com/watch\?.*v=)([a-zA-Z0-9_-]{11})`)
	if matches := watchPattern.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	// Pattern for youtu.be/ID
	shortPattern := regexp.MustCompile(`(?:youtu\.be/)([a-zA-Z0-9_-]{11})`)
	if matches := shortPattern.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	// Pattern for youtube.com/embed/ID
	embedPattern := regexp.MustCompile(`(?:youtube\.com/embed/)([a-zA-Z0-9_-]{11})`)
	if matches := embedPattern.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}

	// If it looks like a plain video ID (11 chars, alphanumeric with _ and -)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{11}$`, input); matched {
		return input
	}

	// Return as-is if no pattern matches (might be a live stream or other format)
	return input
}

// NewVanillaApp creates a new Wails-compatible app wrapper
func NewVanillaApp(orchestrator *appruntime.App, embedProxy *youtube.EmbedProxy, configPath string) *VanillaApp {
	return &VanillaApp{
		orchestrator: orchestrator,
		embedProxy:   embedProxy,
		configPath:   configPath,
	}
}

// GetStats returns current statistics
func (a *VanillaApp) GetStats() interface{} {
	return a.orchestrator.GetStats()
}

// GetPlayerState returns the current player state
func (a *VanillaApp) GetPlayerState() string {
	return a.orchestrator.GetPlayerState()
}

// GetFocusState returns the current focus state
func (a *VanillaApp) GetFocusState() string {
	return a.orchestrator.GetFocusState()
}

// SetVolume sets the player volume
func (a *VanillaApp) SetVolume(level float64) error {
	return a.orchestrator.SetVolume(level)
}

// ManualTrigger manually triggers the FBI meme
func (a *VanillaApp) ManualTrigger() error {
	return a.orchestrator.ManualTrigger()
}

// GetConfig returns the current configuration
func (a *VanillaApp) GetConfig() map[string]interface{} {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return map[string]interface{}{
			"blocklist": map[string]interface{}{
				"patterns": []string{},
			},
			"exceptions": []string{},
		}
	}
	return map[string]interface{}{
		"blocklist": map[string]interface{}{
			"patterns": cfg.Blocklist.Patterns,
		},
		"exceptions": cfg.Exceptions,
	}
}

// GetLofiVideoID returns the lofi video ID from config
func (a *VanillaApp) GetLofiVideoID() string {
	cfg := a.orchestrator.GetConfig()
	if cfg != nil && cfg.Player.LofiPlaylist != "" {
		if videoID := extractVideoID(cfg.Player.LofiPlaylist); videoID != "" {
			return videoID
		}
	}
	return "jfKfPfyJRdk" // Lofi Girl fallback
}

// GetLofiEmbedURL returns the HTTP localhost URL for the YouTube embed proxy
// This provides a proper http:// origin that YouTube accepts (fixes Error 153)
func (a *VanillaApp) GetLofiEmbedURL() string {
	videoID := a.GetLofiVideoID()
	if a.embedProxy != nil {
		return a.embedProxy.GetEmbedURL(videoID)
	}
	// Fallback to direct YouTube embed (won't work in native app but works in browser)
	return "https://www.youtube-nocookie.com/embed/" + videoID + "?autoplay=1&controls=1&modestbranding=1&rel=0&playsinline=1"
}

// GetBlocklist returns the current blocklist configuration
func (a *VanillaApp) GetBlocklist() map[string]interface{} {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return map[string]interface{}{
			"patterns":   []string{},
			"exceptions": []string{},
		}
	}
	return map[string]interface{}{
		"patterns":   cfg.Blocklist.Patterns,
		"exceptions": cfg.Exceptions,
	}
}

// AddBlocklistEntry adds a new pattern to the blocklist or exceptions
func (a *VanillaApp) AddBlocklistEntry(entryType, value string) error {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return nil
	}
	switch entryType {
	case "patterns":
		cfg.Blocklist.Patterns = append(cfg.Blocklist.Patterns, value)
	case "exceptions":
		cfg.Exceptions = append(cfg.Exceptions, value)
	}
	return a.SaveConfig()
}

// RemoveBlocklistEntry removes a pattern from the blocklist or exceptions
func (a *VanillaApp) RemoveBlocklistEntry(entryType, value string) error {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return nil
	}
	switch entryType {
	case "patterns":
		for i, p := range cfg.Blocklist.Patterns {
			if p == value {
				cfg.Blocklist.Patterns = append(cfg.Blocklist.Patterns[:i], cfg.Blocklist.Patterns[i+1:]...)
				break
			}
		}
	case "exceptions":
		for i, e := range cfg.Exceptions {
			if e == value {
				cfg.Exceptions = append(cfg.Exceptions[:i], cfg.Exceptions[i+1:]...)
				break
			}
		}
	}
	return a.SaveConfig()
}

// GetLofiPlaylist returns the current lofi playlist URL from config
func (a *VanillaApp) GetLofiPlaylist() string {
	cfg := a.orchestrator.GetConfig()
	if cfg != nil && cfg.Player.LofiPlaylist != "" {
		return cfg.Player.LofiPlaylist
	}
	return "https://www.youtube.com/watch?v=jfKfPfyJRdk"
}

// SetLofiPlaylist updates the lofi playlist URL and saves to config
func (a *VanillaApp) SetLofiPlaylist(url string) error {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return nil
	}
	cfg.Player.LofiPlaylist = url
	if err := a.SaveConfig(); err != nil {
		return err
	}
	// Emit event for frontend to reload player
	if a.wailsCtx != nil {
		runtime.EventsEmit(a.wailsCtx, "lofi:url-changed", url)
	}
	return nil
}

// SetWailsContext stores the Wails context for event emission
func (a *VanillaApp) SetWailsContext(ctx context.Context) {
	a.wailsCtx = ctx
}

// SaveConfig saves the current configuration to disk
func (a *VanillaApp) SaveConfig() error {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return nil
	}
	data, err := config.MarshalConfig(cfg)
	if err != nil {
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(homeDir, ".vigilant", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Start embed proxy server for YouTube (provides HTTP localhost origin)
	// This fixes Error 153 by giving YouTube a proper http:// referrer
	embedProxy := youtube.NewEmbedProxy()
	if err := embedProxy.Start(); err != nil {
		log.Printf("Warning: Failed to start embed proxy: %v", err)
	}

	// Load configuration
	orchestrator, err := appruntime.New("")
	if err != nil {
		log.Fatalf("Failed to create app orchestrator: %v", err)
	}

	// Create Wails-compatible wrapper
	app := NewVanillaApp(orchestrator, embedProxy, "")

	// Create application with options
	err = wails.Run(&options.App{
		Title:     "Vigilant - Stay Focused",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Mac: &mac.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		OnStartup: func(ctx context.Context) {
			// Store Wails context for event emission
			app.SetWailsContext(ctx)
			// Start the orchestrator with the Wails context
			if err := orchestrator.Start(ctx); err != nil {
				log.Printf("Failed to start orchestrator: %v", err)
			}
		},
		OnShutdown: func(ctx context.Context) {
			// Gracefully shutdown the orchestrator
			if err := orchestrator.Stop(); err != nil {
				log.Printf("Error stopping orchestrator: %v", err)
			}
			// Stop the embed proxy server
			if embedProxy != nil {
				if err := embedProxy.Stop(); err != nil {
					log.Printf("Error stopping embed proxy: %v", err)
				}
			}
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Printf("Wails error: %v", err)
		os.Exit(1)
	}
}
