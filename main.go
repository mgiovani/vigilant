package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	appruntime "vigilant/internal/app"
	"vigilant/internal/youtube"
)

//go:embed all:frontend/dist all:assets
var assets embed.FS

// VanillaApp wraps the orchestrator app for Wails bindings
type VanillaApp struct {
	orchestrator *appruntime.App
	youtubeService *youtube.Service
	configPath string
}

// NewVanillaApp creates a new Wails-compatible app wrapper
func NewVanillaApp(orchestrator *appruntime.App, youtubeService *youtube.Service, configPath string) *VanillaApp {
	return &VanillaApp{
		orchestrator: orchestrator,
		youtubeService: youtubeService,
		configPath: configPath,
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
				"processes": []string{},
				"websites":  []string{},
				"patterns":  []string{},
			},
		}
	}
	return map[string]interface{}{
		"blocklist": map[string]interface{}{
			"processes": cfg.Blocklist.Processes,
			"websites":  cfg.Blocklist.Websites,
			"patterns":  cfg.Blocklist.Patterns,
		},
	}
}

// GetLofiVideoID returns a working lofi video ID from YouTube API
func (a *VanillaApp) GetLofiVideoID() string {
	if a.youtubeService == nil {
		return "jfKfPfyJRdk" // Fallback to hardcoded ID
	}
	videoID, err := a.youtubeService.GetLofiVideoID(context.Background())
	if err != nil {
		log.Printf("Error getting lofi video ID: %v", err)
		return "jfKfPfyJRdk" // Fallback to hardcoded ID
	}
	return videoID
}

// GetBlocklist returns the current blocklist configuration
func (a *VanillaApp) GetBlocklist() map[string]interface{} {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return map[string]interface{}{
			"processes": []string{},
			"websites":  []string{},
			"patterns":  []string{},
		}
	}
	return map[string]interface{}{
		"processes": cfg.Blocklist.Processes,
		"websites":  cfg.Blocklist.Websites,
		"patterns":  cfg.Blocklist.Patterns,
	}
}

// AddBlocklistEntry adds a new entry to the blocklist
func (a *VanillaApp) AddBlocklistEntry(entryType, value string) error {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return nil
	}
	switch entryType {
	case "processes":
		cfg.Blocklist.Processes = append(cfg.Blocklist.Processes, value)
	case "websites":
		cfg.Blocklist.Websites = append(cfg.Blocklist.Websites, value)
	case "patterns":
		cfg.Blocklist.Patterns = append(cfg.Blocklist.Patterns, value)
	}
	return a.SaveConfig()
}

// RemoveBlocklistEntry removes an entry from the blocklist
func (a *VanillaApp) RemoveBlocklistEntry(entryType, value string) error {
	cfg := a.orchestrator.GetConfig()
	if cfg == nil {
		return nil
	}
	switch entryType {
	case "processes":
		for i, p := range cfg.Blocklist.Processes {
			if p == value {
				cfg.Blocklist.Processes = append(cfg.Blocklist.Processes[:i], cfg.Blocklist.Processes[i+1:]...)
				break
			}
		}
	case "websites":
		for i, w := range cfg.Blocklist.Websites {
			if w == value {
				cfg.Blocklist.Websites = append(cfg.Blocklist.Websites[:i], cfg.Blocklist.Websites[i+1:]...)
				break
			}
		}
	case "patterns":
		for i, p := range cfg.Blocklist.Patterns {
			if p == value {
				cfg.Blocklist.Patterns = append(cfg.Blocklist.Patterns[:i], cfg.Blocklist.Patterns[i+1:]...)
				break
			}
		}
	}
	return a.SaveConfig()
}

// SaveConfig saves the current configuration to disk
func (a *VanillaApp) SaveConfig() error {
	// This will be implemented based on the orchestrator's config saving mechanism
	// For now, return nil as a placeholder
	return nil
}

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Initialize YouTube service (optional, has fallback)
	var youtubeService *youtube.Service
	if apiKey := os.Getenv("YOUTUBE_API_KEY"); apiKey != "" {
		var err error
		youtubeService, err = youtube.NewService(apiKey)
		if err != nil {
			log.Printf("Warning: YouTube service initialization failed, will use fallback: %v", err)
		}
	}

	// Load configuration
	orchestrator, err := appruntime.New("")
	if err != nil {
		log.Fatalf("Failed to create app orchestrator: %v", err)
	}

	// Create Wails-compatible wrapper
	app := NewVanillaApp(orchestrator, youtubeService, "")

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
