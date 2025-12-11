// Package app provides the central application orchestrator that coordinates all
// Vigilant components. It manages the event flow from window monitoring through
// blocklist matching, state management, player control, and statistics tracking.
//
// The App struct is a Wails-enabled component that bridges between the Go backend
// and the Svelte frontend. It exposes public methods that are automatically bound to
// the frontend via Wails runtime.
//
// The orchestration flow:
//  1. Monitor emits FocusEvent when active window changes
//  2. Detector applies grace period debouncing
//  3. Blocker evaluates if window is blocked
//  4. StateManager updates focus state (Working/Distracted)
//  5. PlayerController switches between lofi/FBI based on state
//  6. StatsTracker accumulates time in each state
//  7. App emits events to frontend to update UI
package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vigilant/internal/blocker"
	"vigilant/internal/config"
	"vigilant/internal/monitor"
	"vigilant/internal/player"
	"vigilant/internal/stats"
)

// App is the central orchestrator that coordinates all components
// of the vigilant application. It manages:
// - Window monitoring and focus detection
// - Process/website blocking logic
// - Stats tracking and triggering
// - Media player state management
// - Event emission to the frontend
type App struct {
	config        *config.Config
	monitor       monitor.Monitor
	blocker       blocker.Blocker
	stateManager  *blocker.StateManager
	playerCtrl    *player.Controller
	statsTracker  *stats.Tracker
	detector      *monitor.Detector
	gracePeriodCh <-chan monitor.ConfirmedDistractionEvent

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc

	// Thread-safe access to app state
	mu sync.RWMutex

	// For Wails context
	wailsCtx context.Context

	// Logger
	logger *log.Logger
}

// New creates and initializes a new App instance
// It loads configuration from the provided path and initializes all components
// Returns error if any component fails to initialize
func New(configPath string) (*App, error) {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create platform-specific monitor
	mon, err := monitor.NewMonitor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}

	// Create blocker from configuration
	blk, err := blocker.NewBlocklistMatcher(cfg.Blocklist, cfg.Exceptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create blocker: %w", err)
	}

	// Create state manager
	stateManager := blocker.NewStateManager(blk)

	// Create player controller
	playerCtrl := player.NewController()

	// Create stats tracker
	statsTracker := stats.NewTracker()

	// Create grace period detector
	detector := monitor.NewDetector(cfg)

	app := &App{
		config:       cfg,
		monitor:      mon,
		blocker:      blk,
		stateManager: stateManager,
		playerCtrl:   playerCtrl,
		statsTracker: statsTracker,
		detector:     detector,
		logger:       log.New(log.Writer(), "[vigilant] ", log.LstdFlags),
	}

	app.logger.Println("App initialized successfully")

	return app, nil
}

// Start begins all background monitoring and orchestration
// It accepts a context that allows the caller (Wails) to control the app lifetime
// Returns error if any component fails to start
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Println("Starting app...")

	// Store Wails context for event emission
	a.wailsCtx = ctx

	// Create cancellable context for internal control
	a.ctx, a.cancel = context.WithCancel(ctx)

	// Start monitor
	if err := a.monitor.Start(a.ctx); err != nil {
		return fmt.Errorf("failed to start monitor: %w", err)
	}
	a.logger.Println("Monitor started")

	// Start stats tracker
	if err := a.statsTracker.Start(a.ctx); err != nil {
		return fmt.Errorf("failed to start stats tracker: %w", err)
	}
	a.logger.Println("Stats tracker started")

	// Launch main orchestration loop
	go a.orchestrationLoop()
	a.logger.Println("Orchestration loop started")

	return nil
}

// orchestrationLoop is the main event processing loop
// It coordinates between monitor events, state changes, and stats tracking
func (a *App) orchestrationLoop() {
	defer a.logger.Println("Orchestration loop stopped")

	// Ticker for periodic stats updates to frontend
	statsTicker := time.NewTicker(time.Second)
	defer statsTicker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Println("Orchestration context cancelled")
			return

		case focusEvent := <-a.monitor.Events():
			a.handleFocusEvent(focusEvent)

		case stateChange := <-a.stateManager.StateChanges():
			a.handleStateChange(stateChange)

		case <-statsTicker.C:
			// Emit stats update to frontend every second
			a.emitStatsUpdate()
		}
	}
}

// handleFocusEvent processes a window focus change event
func (a *App) handleFocusEvent(event monitor.FocusEvent) {
	a.mu.RLock()
	blocker := a.blocker
	playerCtrl := a.playerCtrl
	stateManager := a.stateManager
	statsTracker := a.statsTracker
	a.mu.RUnlock()

	if blocker == nil || playerCtrl == nil || stateManager == nil {
		return
	}

	// Check if the window is blocked
	isBlocked := blocker.IsBlocked(event.Current)

	// Emit focus event details to frontend
	a.emitFocusEvent(event.Current, isBlocked)

	// Process through state manager
	if err := stateManager.OnFocusChange(&event); err != nil {
		a.logger.Printf("Error processing focus change: %v", err)
		a.emitError(fmt.Sprintf("Focus processing error: %v", err))
	}

	// If blocked, start grace period
	if isBlocked {
		a.logger.Printf("Blocked window detected: %s (%s)", event.Current.Title, event.Current.Process)

		// Start grace period detector
		confirmCh := a.detector.ProcessEvent(
			monitor.FocusEvent{
				Current:   event.Current,
				Timestamp: time.Now(),
			},
			isBlocked,
		)

		// Wait for grace period confirmation in background
		if confirmCh != nil {
			go a.handleGracePeriod(confirmCh, statsTracker, playerCtrl)
		}
	} else {
		// Cancel grace period if switching away
		a.detector.CancelPending()
		// Switch back to lofi
		playerCtrl.SwitchToLofi()
		a.emitPlayerStateChange("lofi")
	}
}

// handleGracePeriod waits for grace period confirmation and triggers FBI meme
func (a *App) handleGracePeriod(
	confirmCh <-chan monitor.ConfirmedDistractionEvent,
	statsTracker *stats.Tracker,
	playerCtrl *player.Controller,
) {
	for event := range confirmCh {
		a.logger.Printf("Grace period confirmed for: %s", event.Window.Title)

		// Increment trigger count
		statsTracker.IncrementTrigger()

		// Switch to FBI meme
		playerCtrl.SwitchToFBI()

		// Emit events to frontend
		a.emitPlayerStateChange("fbi")
		a.emitStatsUpdate()
		a.emitTrigger(event.Window)
	}
}

// handleStateChange processes focus state changes
func (a *App) handleStateChange(change blocker.StateChange) {
	a.mu.RLock()
	wailsCtx := a.wailsCtx
	statsTracker := a.statsTracker
	a.mu.RUnlock()

	// Forward state change to stats tracker for time accumulation
	if statsTracker != nil {
		statsTracker.ProcessStateChange(change)
	}

	// Determine focus state string
	focusState := "working"
	if change.Current == blocker.StateDistracted {
		focusState = "distracted"
	}

	a.logger.Printf("Focus state changed: %s → %s", change.Previous.String(), change.Current.String())

	// Emit focus state change
	if wailsCtx != nil {
		runtime.EventsEmit(wailsCtx, "focus:state-change", map[string]interface{}{
			"state":     focusState,
			"window":    change.Window.Title,
			"process":   change.Window.Process,
			"timestamp": change.Timestamp.Unix(),
		})
	}
}

// Stop gracefully shuts down all components
func (a *App) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Println("Stopping app...")

	// Cancel orchestration context to stop all goroutines
	if a.cancel != nil {
		a.cancel()
	}

	// Allow time for goroutines to finish
	time.Sleep(100 * time.Millisecond)

	// Stop stats tracker
	if a.statsTracker != nil {
		if err := a.statsTracker.Stop(); err != nil {
			a.logger.Printf("Error stopping stats tracker: %v", err)
		}
	}

	// Stop monitor
	if a.monitor != nil {
		if err := a.monitor.Stop(); err != nil {
			a.logger.Printf("Error stopping monitor: %v", err)
		}
	}

	a.logger.Println("App stopped")
	return nil
}

// GetConfig returns the application configuration
func (a *App) GetConfig() *config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// GetStats returns a snapshot of current statistics
func (a *App) GetStats() *stats.Stats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.statsTracker == nil {
		return nil
	}

	s := a.statsTracker.GetStats()
	return &s
}

// GetPlayerState returns the current player state as a string
func (a *App) GetPlayerState() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.playerCtrl == nil {
		return "unknown"
	}

	return a.playerCtrl.GetState().String()
}

// GetFocusState returns the current focus state as a string
func (a *App) GetFocusState() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.stateManager == nil {
		return "unknown"
	}

	state := a.stateManager.GetCurrentState()
	if state == blocker.StateWorking {
		return "working"
	}
	return "distracted"
}

// SetVolume sets the player volume (0.0-1.0)
// Currently a stub as Wails doesn't have built-in audio control
func (a *App) SetVolume(level float64) error {
	if level < 0.0 || level > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0, got %f", level)
	}

	a.logger.Printf("Volume set to %.2f", level)

	// Emit to frontend for implementation
	a.mu.RLock()
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if wailsCtx != nil {
		runtime.EventsEmit(wailsCtx, "player:volume-change", level)
	}

	return nil
}

// ManualTrigger manually triggers the FBI meme (for testing)
func (a *App) ManualTrigger() error {
	a.mu.RLock()
	playerCtrl := a.playerCtrl
	statsTracker := a.statsTracker
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if playerCtrl == nil || statsTracker == nil {
		return fmt.Errorf("components not initialized")
	}

	a.logger.Println("Manual trigger activated")

	// Switch to FBI
	playerCtrl.SwitchToFBI()

	// Increment trigger count
	statsTracker.IncrementTrigger()

	// Emit events
	a.emitPlayerStateChange("fbi")
	a.emitStatsUpdate()

	if wailsCtx != nil {
		runtime.EventsEmit(wailsCtx, "trigger:manual", map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})
	}

	return nil
}

// Event emission helper methods

// emitPlayerStateChange emits a player state change event to the frontend
func (a *App) emitPlayerStateChange(state string) {
	a.mu.RLock()
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if wailsCtx != nil {
		runtime.EventsEmit(wailsCtx, "player:state-change", map[string]interface{}{
			"state":     state,
			"timestamp": time.Now().Unix(),
		})
	}
}

// emitStatsUpdate emits a stats update event to the frontend
func (a *App) emitStatsUpdate() {
	a.mu.RLock()
	statsTracker := a.statsTracker
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if statsTracker == nil || wailsCtx == nil {
		return
	}

	stats := statsTracker.GetStats()
	runtime.EventsEmit(wailsCtx, "stats:update", map[string]interface{}{
		"focusedTime":    stats.FocusedTime.String(),
		"distractedTime": stats.DistractedTime.String(),
		"triggerCount":   stats.TriggerCount,
		"focusRate":      statsTracker.CalculateFocusRate(),
		"sessionStart":   stats.SessionStart.Unix(),
		"lastUpdate":     stats.LastUpdate.Unix(),
	})
}

// emitFocusEvent emits a focus event to the frontend
func (a *App) emitFocusEvent(window *monitor.WindowInfo, isBlocked bool) {
	a.mu.RLock()
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if wailsCtx == nil || window == nil {
		return
	}

	runtime.EventsEmit(wailsCtx, "window:focus-change", map[string]interface{}{
		"title":     window.Title,
		"process":   window.Process,
		"pid":       window.PID,
		"blocked":   isBlocked,
		"timestamp": time.Now().Unix(),
	})
}

// emitTrigger emits a trigger event to the frontend
func (a *App) emitTrigger(window *monitor.WindowInfo) {
	a.mu.RLock()
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if wailsCtx == nil || window == nil {
		return
	}

	runtime.EventsEmit(wailsCtx, "trigger:confirmed", map[string]interface{}{
		"window":    window.Title,
		"process":   window.Process,
		"timestamp": time.Now().Unix(),
	})
}

// emitError emits an error event to the frontend
func (a *App) emitError(errorMsg string) {
	a.mu.RLock()
	wailsCtx := a.wailsCtx
	a.mu.RUnlock()

	if wailsCtx != nil {
		runtime.EventsEmit(wailsCtx, "app:error", map[string]interface{}{
			"message":   errorMsg,
			"timestamp": time.Now().Unix(),
		})
	}

	a.logger.Printf("Error event: %s", errorMsg)
}
