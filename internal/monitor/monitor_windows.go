//go:build windows
// +build windows

package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"

	"vigilant/internal/config"
)

// WindowsMonitor implements the Monitor interface for Windows systems.
// It uses Win32 API calls to detect active window changes through polling.
type WindowsMonitor struct {
	config    *config.Config
	eventsCh  chan FocusEvent
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	stopped   bool
	previousW *WindowInfo
}

// Win32 API declarations
var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")
	psapi    = windows.NewLazySystemDLL("psapi.dll")

	procGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessID   = user32.NewProc("GetWindowThreadProcessId")
	procGetProcessImageFileNameW   = psapi.NewProc("GetProcessImageFileNameW")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procCloseHandle                = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	MAX_PATH                          = 260
)

// newPlatformMonitor creates a Windows-specific monitor instance.
func newPlatformMonitor(cfg *config.Config) (Monitor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &WindowsMonitor{
		config:   cfg,
		eventsCh: make(chan FocusEvent, 10),
	}, nil
}

// Start launches the window monitoring goroutine.
func (m *WindowsMonitor) Start(ctx context.Context) error {
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
func (m *WindowsMonitor) Stop() error {
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
func (m *WindowsMonitor) Events() <-chan FocusEvent {
	return m.eventsCh
}

// GetActiveWindow returns the currently active window.
func (m *WindowsMonitor) GetActiveWindow() (*WindowInfo, error) {
	return m.getActiveWindowInternal()
}

// pollLoop continuously polls for active window changes.
func (m *WindowsMonitor) pollLoop() {
	ticker := time.NewTicker(m.config.Monitor.PollInterval.Duration())
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			current, err := m.getActiveWindowInternal()
			if err != nil {
				// Log error but continue polling
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
func (m *WindowsMonitor) getActiveWindowInternal() (*WindowInfo, error) {
	// Get foreground window handle
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return nil, fmt.Errorf("no active window found")
	}

	// Get window title
	title := m.getWindowTitle(syscall.Handle(hwnd))

	// Get process ID
	var pid uint32
	_, _, _ = procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	if pid == 0 {
		return nil, fmt.Errorf("could not get process ID")
	}

	// Get process name
	processName, err := m.getProcessName(pid)
	if err != nil {
		// Fallback to using just the PID if we can't get the process name
		processName = fmt.Sprintf("pid_%d", pid)
	}

	return &WindowInfo{
		PID:       pid,
		Title:     title,
		Process:   processName,
		Handle:    hwnd,
		Timestamp: time.Now(),
	}, nil
}

// getWindowTitle extracts the window title from a handle.
func (m *WindowsMonitor) getWindowTitle(hwnd syscall.Handle) string {
	// Get window text length (including null terminator)
	const maxLen = 512
	buf := make([]uint16, maxLen)

	len, _, _ := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(maxLen))

	if len == 0 {
		return ""
	}

	// Convert UTF-16 to string
	return windows.UTF16ToString(buf[:len])
}

// getProcessName extracts the executable name from a process ID.
func (m *WindowsMonitor) getProcessName(pid uint32) (string, error) {
	// Open process handle
	handle, _, err := procOpenProcess.Call(uintptr(PROCESS_QUERY_LIMITED_INFORMATION), 0, uintptr(pid))
	if handle == 0 {
		return "", fmt.Errorf("failed to open process: %w", err)
	}
	defer procCloseHandle.Call(handle)

	// Get process image file name
	buf := make([]uint16, MAX_PATH)
	len, _, err := procQueryFullProcessImageNameW.Call(
		handle,
		uintptr(0), // lpExeName: NULL means standard name
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&len)),
	)

	if len == 0 && handle != 0 {
		// Fallback method: try GetProcessImageFileNameW
		buf = make([]uint16, MAX_PATH)
		len, _, _ = procGetProcessImageFileNameW.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len))
		if len == 0 {
			return "", fmt.Errorf("failed to get process image file name")
		}
	}

	if len == 0 {
		return "", fmt.Errorf("failed to query full process image name: %w", err)
	}

	// Convert UTF-16 to string
	fullPath := windows.UTF16ToString(buf[:len])

	// Extract just the executable name from the full path
	parts := strings.Split(fullPath, "\\")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return fullPath, nil
}

// windowChanged detects if the window has actually changed.
func (m *WindowsMonitor) windowChanged(prev, current *WindowInfo) bool {
	if prev == nil && current != nil {
		return true
	}

	if prev == nil && current == nil {
		return false
	}

	if current == nil {
		return false
	}

	// Compare by process ID and title to detect real changes
	// Don't rely on handle since it can change
	return prev.PID != current.PID || prev.Title != current.Title
}
