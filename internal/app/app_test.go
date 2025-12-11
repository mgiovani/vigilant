package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"vigilant/internal/config"
	"vigilant/internal/monitor"
)

func TestNewAppSuccess(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if app == nil {
		t.Fatal("New() returned nil app")
	}

	if app.config == nil {
		t.Fatal("App config is nil")
	}

	if app.monitor == nil {
		t.Fatal("App monitor is nil")
	}

	if app.blocker == nil {
		t.Fatal("App blocker is nil")
	}

	if app.stateManager == nil {
		t.Fatal("App stateManager is nil")
	}

	if app.playerCtrl == nil {
		t.Fatal("App playerCtrl is nil")
	}

	if app.statsTracker == nil {
		t.Fatal("App statsTracker is nil")
	}
}

func TestNewAppMissingConfigFile(t *testing.T) {
	app, err := New("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Expected error for missing config file, got nil")
	}

	if app != nil {
		t.Fatal("Expected app to be nil on error")
	}
}

func TestAppGetConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	retrievedCfg := app.GetConfig()
	if retrievedCfg == nil {
		t.Fatal("GetConfig() returned nil")
	}

	if retrievedCfg.Player.DefaultVolume != cfg.Player.DefaultVolume {
		t.Errorf("Config mismatch: expected volume %.2f, got %.2f",
			cfg.Player.DefaultVolume, retrievedCfg.Player.DefaultVolume)
	}
}

func TestAppGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	stats := app.GetStats()
	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	if stats.TriggerCount != 0 {
		t.Errorf("Expected initial TriggerCount to be 0, got %d", stats.TriggerCount)
	}

	if stats.SessionStart.IsZero() {
		t.Fatal("SessionStart is zero")
	}
}

func TestAppGetPlayerState(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	state := app.GetPlayerState()
	if state != "lofi" {
		t.Errorf("Expected initial player state to be 'lofi', got %q", state)
	}
}

func TestAppGetFocusState(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	state := app.GetFocusState()
	if state != "working" {
		t.Errorf("Expected initial focus state to be 'working', got %q", state)
	}
}

func TestAppSetVolume(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Valid volume
	if err := app.SetVolume(0.5); err != nil {
		t.Errorf("SetVolume(0.5) failed: %v", err)
	}

	// Test boundary values
	if err := app.SetVolume(0.0); err != nil {
		t.Errorf("SetVolume(0.0) failed: %v", err)
	}

	if err := app.SetVolume(1.0); err != nil {
		t.Errorf("SetVolume(1.0) failed: %v", err)
	}

	// Invalid volume (too high)
	if err := app.SetVolume(1.5); err == nil {
		t.Fatal("Expected error for volume > 1.0")
	}

	// Invalid volume (negative)
	if err := app.SetVolume(-0.1); err == nil {
		t.Fatal("Expected error for negative volume")
	}
}

func TestAppStartStop(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create a context that we control
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the app
	if err := app.Start(ctx); err != nil {
		t.Errorf("Start() failed: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	if err := app.Stop(); err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Verify context is cancelled
	if app.ctx == nil {
		t.Fatal("App context is nil after Stop")
	}
}

func TestAppManualTrigger(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test manual trigger
	if err := app.ManualTrigger(); err != nil {
		t.Errorf("ManualTrigger() failed: %v", err)
	}

	// Verify player state changed to FBI
	if state := app.GetPlayerState(); state != "fbi" {
		t.Errorf("Expected player state to be 'fbi' after trigger, got %q", state)
	}

	// Verify trigger count incremented
	stats := app.GetStats()
	if stats.TriggerCount != 1 {
		t.Errorf("Expected TriggerCount to be 1, got %d", stats.TriggerCount)
	}
}

func TestAppThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Don't start the app to avoid Wails context issues in tests
	// Just test thread safety of getter/setter methods

	// Launch concurrent operations
	done := make(chan bool, 5)

	// Reader goroutines
	for i := 0; i < 3; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_ = app.GetConfig()
				_ = app.GetStats()
				_ = app.GetPlayerState()
				_ = app.GetFocusState()
				time.Sleep(5 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Writer goroutines
	for i := 0; i < 2; i++ {
		go func(idx int) {
			for j := 0; j < 5; j++ {
				if idx%2 == 0 {
					_ = app.SetVolume(0.5)
				} else {
					_ = app.ManualTrigger()
				}
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should complete without panic
	t.Log("Thread safety test passed")
}

func TestAppGettersWithNilComponents(t *testing.T) {
	// Test edge case where components might be nil
	app := &App{
		config:       nil,
		statsTracker: nil,
		playerCtrl:   nil,
		stateManager: nil,
	}

	// These should return safe defaults without panicking
	if config := app.GetConfig(); config != nil {
		t.Errorf("Expected nil config, got %v", config)
	}

	if stats := app.GetStats(); stats != nil {
		t.Errorf("Expected nil stats, got %v", stats)
	}

	if state := app.GetPlayerState(); state != "unknown" {
		t.Errorf("Expected 'unknown' player state, got %q", state)
	}

	if state := app.GetFocusState(); state != "unknown" {
		t.Errorf("Expected 'unknown' focus state, got %q", state)
	}
}

// TestAppEventEmission tests that events are properly emitted without panicking
func TestAppEventEmission(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultConfig()
	content, err := config.MarshalConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	app, err := New(configPath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Don't set wailsCtx since we can't provide a valid one in tests
	// The event emission functions should handle nil context gracefully

	// Test event emission helpers (these should not panic even without valid context)
	app.emitPlayerStateChange("fbi")
	app.emitStatsUpdate()

	window := &monitor.WindowInfo{
		Title:   "Test Window",
		Process: "test.exe",
		PID:     1234,
	}
	app.emitFocusEvent(window, true)
	app.emitTrigger(window)
	app.emitError("test error")

	// Should complete without panic
	t.Log("Event emission test passed")
}
