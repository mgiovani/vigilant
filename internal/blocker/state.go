package blocker

import (
	"log"
	"sync"
	"time"

	"vigilant/internal/monitor"
)

// FocusState represents the user's current focus state
type FocusState int

const (
	// StateWorking indicates the user is working (no distracting window is active)
	StateWorking FocusState = iota

	// StateDistracted indicates the user is distracted (a blocked window is active)
	StateDistracted
)

// String implements the Stringer interface for FocusState
func (s FocusState) String() string {
	switch s {
	case StateWorking:
		return "Working"
	case StateDistracted:
		return "Distracted"
	default:
		return "Unknown"
	}
}

// StateChange represents a change in focus state
type StateChange struct {
	// Previous holds the previous focus state
	Previous FocusState

	// Current holds the new focus state
	Current FocusState

	// Window is the window that triggered the state change
	Window *monitor.WindowInfo

	// Timestamp marks when this state change occurred
	Timestamp time.Time
}

// StateManager tracks the user's focus state based on window activity
type StateManager struct {
	blocker      Blocker             // Inject blocker for checking if windows are blocked
	currentState FocusState          // Current focus state
	lastWindow   *monitor.WindowInfo // Last tracked window
	changesCh    chan StateChange    // Buffered channel for state changes
	mu           sync.RWMutex        // Thread safety
}

// NewStateManager creates a new StateManager with the provided blocker
// It initializes the state to Working and creates a buffered channel for state changes
func NewStateManager(blocker Blocker) *StateManager {
	if blocker == nil {
		log.Println("Warning: StateManager created with nil blocker")
	}

	return &StateManager{
		blocker:      blocker,
		currentState: StateWorking,
		lastWindow:   nil,
		changesCh:    make(chan StateChange, 1), // Buffer size 1 for non-blocking sends
	}
}

// OnFocusChange handles a focus change event from the monitor
// It determines the new state, emits a StateChange event if the state changed,
// and updates the lastWindow tracking
func (sm *StateManager) OnFocusChange(event *monitor.FocusEvent) error {
	if event == nil || event.Current == nil {
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Determine if the new window is blocked
	isBlocked := false
	if sm.blocker != nil {
		isBlocked = sm.blocker.IsBlocked(event.Current)
	}

	// Determine the new state
	newState := StateWorking
	if isBlocked {
		newState = StateDistracted
	}

	// Only emit StateChange if state actually changed
	if newState != sm.currentState {
		previousState := sm.currentState
		sm.currentState = newState

		// Log state transition
		log.Printf("State transition: %s → %s (Window: %s)", previousState.String(), newState.String(), event.Current.Title)

		// Create and send state change event (non-blocking)
		stateChange := StateChange{
			Previous:  previousState,
			Current:   newState,
			Window:    event.Current,
			Timestamp: time.Now(),
		}

		select {
		case sm.changesCh <- stateChange:
			// Sent successfully
		default:
			// Channel buffer full, log a warning
			log.Println("Warning: StateManager changesCh buffer full, dropping event")
		}
	}

	// Update lastWindow
	sm.lastWindow = event.Current

	return nil
}

// GetCurrentState returns the current focus state (thread-safe)
func (sm *StateManager) GetCurrentState() FocusState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// GetLastWindow returns the last tracked window (thread-safe)
// Returns nil if no window has been tracked yet
func (sm *StateManager) GetLastWindow() *monitor.WindowInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.lastWindow
}

// StateChanges returns a read-only channel for state changes
// Callers should read from this channel to be notified of state transitions
func (sm *StateManager) StateChanges() <-chan StateChange {
	return sm.changesCh
}
