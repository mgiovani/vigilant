package monitor

import (
	"sync"
	"time"

	"vigilant/internal/config"
)

// FocusState represents the current focus state of the user.
type FocusState string

const (
	// StateWorking indicates the user is focused on productive content
	StateWorking FocusState = "working"

	// StateDistracted indicates the user is engaged with blocked/distracting content
	StateDistracted FocusState = "distracted"
)

// ConfirmedDistractionEvent represents a confirmed distraction event
// that has passed the grace period without being dismissed.
type ConfirmedDistractionEvent struct {
	Window    *WindowInfo
	Timestamp time.Time
}

// Detector processes FocusEvents and applies grace period logic
// to debounce rapid window switching and confirm distractions.
//
// The detector maintains a pending distraction that is only confirmed
// after the grace period elapses without the user switching away from
// the blocked window.
type Detector struct {
	config        *config.Config
	currentState  FocusState
	pendingTimer  *time.Timer
	pendingWindow *WindowInfo
	mu            sync.RWMutex
}

// NewDetector creates a new grace period detector.
func NewDetector(cfg *config.Config) *Detector {
	return &Detector{
		config:       cfg,
		currentState: StateWorking,
	}
}

// ProcessEvent processes a focus event and returns a channel that will
// emit a ConfirmedDistractionEvent if the grace period expires.
//
// The returned channel will be closed if the grace period is cancelled
// (user switched away from blocked content).
//
// If the window is not blocked, the channel is nil and the state is
// set to Working.
func (d *Detector) ProcessEvent(event FocusEvent, isBlocked bool) <-chan ConfirmedDistractionEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	// If there's a pending timer, cancel it
	if d.pendingTimer != nil {
		d.pendingTimer.Stop()
		d.pendingTimer = nil
		d.pendingWindow = nil
	}

	// If the current window is not blocked, we're back to working
	if !isBlocked {
		d.currentState = StateWorking
		return nil
	}

	// Window is blocked, start the grace period
	d.currentState = StateDistracted
	d.pendingWindow = event.Current
	resultCh := make(chan ConfirmedDistractionEvent, 1)

	d.pendingTimer = time.AfterFunc(d.config.Monitor.GracePeriod.Duration(), func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		// Only emit if the window hasn't changed
		if d.pendingWindow != nil && d.pendingWindow.PID == event.Current.PID {
			resultCh <- ConfirmedDistractionEvent{
				Window:    event.Current,
				Timestamp: time.Now(),
			}
		}
		close(resultCh)
	})

	return resultCh
}

// Reset cancels any pending timers and resets the detector to working state.
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.pendingTimer != nil {
		d.pendingTimer.Stop()
		d.pendingTimer = nil
		d.pendingWindow = nil
	}
	d.currentState = StateWorking
}

// GetCurrentState returns the current focus state.
func (d *Detector) GetCurrentState() FocusState {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentState
}

// CancelPending cancels any pending distraction event.
// This is called when the user switches away from blocked content.
func (d *Detector) CancelPending() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.pendingTimer != nil {
		d.pendingTimer.Stop()
		d.pendingTimer = nil
		d.pendingWindow = nil
	}
	d.currentState = StateWorking
}
