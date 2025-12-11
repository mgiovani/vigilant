//go:build darwin
// +build darwin

package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"vigilant/internal/config"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// Helper to get process identifier from NSRunningApplication
int getProcessID(void* app) {
	NSRunningApplication* runningApp = (NSRunningApplication*)app;
	if (!runningApp) {
		return -1;
	}
	return (int)[runningApp processIdentifier];
}

// Helper to get bundle identifier
char* getBundleIdentifier(void* app) {
	NSRunningApplication* runningApp = (NSRunningApplication*)app;
	if (!runningApp) {
		return NULL;
	}
	NSString* bundleID = [runningApp bundleIdentifier];
	if (bundleID) {
		return (char*)[bundleID UTF8String];
	}
	return NULL;
}

// Helper to get localized name
char* getLocalizedName(void* app) {
	NSRunningApplication* runningApp = (NSRunningApplication*)app;
	if (!runningApp) {
		return NULL;
	}
	NSString* name = [runningApp localizedName];
	if (name) {
		return (char*)[name UTF8String];
	}
	return NULL;
}

// Helper function to get the frontmost application
void* getFrontmostApplication(void) {
	return (void*)[NSWorkspace sharedWorkspace].frontmostApplication;
}
*/
import "C"

// DarwinMonitor implements the Monitor interface for macOS (Darwin) systems.
// It uses NSWorkspace to detect active application changes through polling.
type DarwinMonitor struct {
	config    *config.Config
	eventsCh  chan FocusEvent
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	stopped   bool
	previousW *WindowInfo
}

// newPlatformMonitor creates a macOS-specific monitor instance.
func newPlatformMonitor(cfg *config.Config) (Monitor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &DarwinMonitor{
		config:   cfg,
		eventsCh: make(chan FocusEvent, 10),
	}, nil
}

// Start launches the window monitoring goroutine.
func (m *DarwinMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ctx != nil {
		return fmt.Errorf("monitor already started")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.stopped = false

	// Initialize with current window
	current, err := m.getActiveWindowInternal()
	if err == nil && current != nil {
		m.previousW = current
	}

	// Start polling goroutine
	go m.pollLoop()

	return nil
}

// Stop gracefully stops the monitor.
func (m *DarwinMonitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopped || m.cancel == nil {
		return nil
	}

	m.cancel()
	m.stopped = true
	return nil
}

// Events returns the focus events channel.
func (m *DarwinMonitor) Events() <-chan FocusEvent {
	return m.eventsCh
}

// GetActiveWindow returns the currently active window.
func (m *DarwinMonitor) GetActiveWindow() (*WindowInfo, error) {
	return m.getActiveWindowInternal()
}

// pollLoop continuously polls for active window changes.
func (m *DarwinMonitor) pollLoop() {
	ticker := time.NewTicker(m.config.Monitor.PollInterval.Duration())
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			current, err := m.getActiveWindowInternal()
			if err != nil {
				// Continue polling even if we get an error
				continue
			}

			if current == nil {
				continue
			}

			// Check if window changed
			if !m.windowChanged(m.previousW, current) {
				continue
			}

			// Emit event with previous window
			event := FocusEvent{
				Previous:  m.previousW,
				Current:   current,
				Timestamp: time.Now(),
			}

			select {
			case m.eventsCh <- event:
				m.previousW = current
			case <-m.ctx.Done():
				return
			}
		}
	}
}

// getActiveWindowInternal retrieves the currently active window.
func (m *DarwinMonitor) getActiveWindowInternal() (*WindowInfo, error) {
	// Get the frontmost application using NSWorkspace
	appPtr := unsafe.Pointer(C.getFrontmostApplication())
	if appPtr == nil {
		return nil, fmt.Errorf("no active application found")
	}

	// Get process identifier (PID)
	pid := C.getProcessID(appPtr)
	if pid < 0 {
		return nil, fmt.Errorf("could not get process ID")
	}

	// Get application name (localized)
	nameStr := C.getLocalizedName(appPtr)
	if nameStr == nil {
		return nil, fmt.Errorf("could not get application name")
	}
	name := C.GoString(nameStr)

	// Get bundle identifier for process name
	bundleStr := C.getBundleIdentifier(appPtr)
	processName := ""
	if bundleStr != nil {
		bundleID := C.GoString(bundleStr)
		// Extract just the last component of the bundle identifier
		// e.g., com.google.Chrome -> Chrome
		if bundleID != "" {
			parts := extractBundleComponent(bundleID)
			if parts != "" {
				processName = parts
			}
		}
	}

	// Fallback to name if we don't have a process name
	if processName == "" {
		processName = name
	}

	return &WindowInfo{
		PID:       uint32(pid),
		Title:     name,
		Process:   processName,
		Handle:    uintptr(pid), // On macOS, use PID as the handle
		Timestamp: time.Now(),
	}, nil
}

// extractBundleComponent extracts the last component of a bundle identifier
func extractBundleComponent(bundleID string) string {
	// Find the last dot
	lastDot := -1
	for i := len(bundleID) - 1; i >= 0; i-- {
		if bundleID[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 {
		// No dot found, return the whole thing
		return bundleID
	}

	if lastDot == len(bundleID)-1 {
		// Dot at the end, shouldn't happen but handle it
		return bundleID[:lastDot]
	}

	return bundleID[lastDot+1:]
}

// windowChanged detects if the window has actually changed.
func (m *DarwinMonitor) windowChanged(prev, current *WindowInfo) bool {
	if prev == nil && current != nil {
		return true
	}

	if prev == nil && current == nil {
		return false
	}

	if current == nil {
		return false
	}

	// Compare by process ID and name to detect real changes
	return prev.PID != current.PID || prev.Process != current.Process
}
