// Package monitor provides real-time window focus monitoring for Windows and macOS.
// It implements cross-platform active window detection with configurable polling intervals
// and grace period debouncing to prevent false positives from brief focus changes.
//
// The Monitor interface abstracts platform-specific implementations, allowing the same
// code to work on Windows (via Win32 API) and macOS (via Cocoa framework through CGO).
//
// The Detector component adds grace period logic on top of raw window events, ensuring
// that brief focus switches (e.g., alt-tabbing) don't immediately trigger actions.
package monitor

import (
	"context"
	"time"

	"vigilant/internal/config"
)

// WindowInfo represents an active window on the system.
// It contains all relevant information about a window including
// its process ID, title, executable name, and when it became active.
type WindowInfo struct {
	// PID is the process ID of the window
	PID uint32

	// Title is the window's display title or name
	Title string

	// Process is the executable name without the full path
	Process string

	// Handle is the platform-specific window handle
	// On Windows: HWND (from win32 API)
	// On macOS: PID (process identifier, as handles are less meaningful)
	Handle uintptr

	// Timestamp marks when this window became active
	Timestamp time.Time
}

// FocusEvent represents a window focus change event.
// This is emitted when the active window changes on the system.
type FocusEvent struct {
	// Previous holds the window that was previously active.
	// This is nil on startup when there's no previous window.
	Previous *WindowInfo

	// Current holds the window that is now active.
	// This is never nil for valid FocusEvent instances.
	Current *WindowInfo

	// Timestamp marks when this focus change occurred
	Timestamp time.Time
}

// Monitor is the interface that all platform-specific window monitoring
// implementations must satisfy. It provides methods to detect active window
// changes and query the current active window.
//
// Implementations must be safe for concurrent use. The Events() channel
// must be created before Start() is called and must not be closed by
// the Monitor implementation (the caller is responsible for cleanup).
type Monitor interface {
	// Start begins monitoring active window changes.
	// It launches the monitoring goroutine(s) and sends FocusEvent objects
	// to the Events() channel whenever the active window changes.
	//
	// The ctx parameter is used to gracefully stop monitoring; when the
	// context is cancelled, the monitor should stop and clean up resources.
	//
	// Returns an error if monitoring cannot be started (e.g., permission denied,
	// system API unavailable). Returns nil on success.
	Start(ctx context.Context) error

	// Stop gracefully stops the monitor and cleans up all resources.
	// This will block until all goroutines have exited and cleanup is complete.
	// It is safe to call Stop() multiple times.
	// Returns an error if graceful shutdown fails.
	Stop() error

	// Events returns a read-only channel of FocusEvent objects.
	// The channel is created when the Monitor is instantiated and is used to
	// communicate focus change events from the monitoring goroutine to the caller.
	// The caller is responsible for reading from this channel; if the channel
	// buffer fills up, the monitor may block or drop events.
	Events() <-chan FocusEvent

	// GetActiveWindow returns information about the currently active window.
	// This method provides immediate access to the active window without waiting
	// for an event, useful for initialization or explicit polling.
	// Returns nil if no window is active or an error if retrieval fails.
	GetActiveWindow() (*WindowInfo, error)
}

// NewMonitor creates and returns a Monitor instance appropriate for the
// current platform (Windows or macOS).
//
// The config parameter contains monitor configuration such as poll interval
// and grace period, which influence how the monitor behaves.
//
// Returns an error if the platform is unsupported or if the monitor cannot
// be created due to system issues (e.g., permission denied).
func NewMonitor(cfg *config.Config) (Monitor, error) {
	return newPlatformMonitor(cfg)
}
