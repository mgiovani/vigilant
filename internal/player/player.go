// Package player provides media player state management for switching between lofi music
// and FBI meme video playback. It coordinates frontend updates through Wails event emission.
//
// The Controller maintains the current player state (lofi or FBI) in a thread-safe manner.
// State changes are managed by the App orchestrator, but the Controller provides the interface
// for querying and updating the player state.
package player

import (
	"fmt"
	"sync"
)

// PlayerState represents the current media player state
type PlayerState int

const (
	// StateLofi indicates lofi beats music is playing
	StateLofi PlayerState = iota

	// StateFBI indicates FBI meme is playing
	StateFBI
)

// String implements the Stringer interface for PlayerState
func (s PlayerState) String() string {
	switch s {
	case StateLofi:
		return "lofi"
	case StateFBI:
		return "fbi"
	default:
		return "unknown"
	}
}

// Controller manages the current player state
// It is thread-safe for concurrent reads and writes
type Controller struct {
	currentState PlayerState
	mu           sync.RWMutex
}

// NewController creates a new PlayerController instance
// Starts in lofi state by default
func NewController() *Controller {
	return &Controller{
		currentState: StateLofi,
	}
}

// SwitchToLofi changes the player state to lofi beats
func (c *Controller) SwitchToLofi() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentState = StateLofi
}

// SwitchToFBI changes the player state to FBI meme
func (c *Controller) SwitchToFBI() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentState = StateFBI
}

// GetState returns the current player state (thread-safe)
func (c *Controller) GetState() PlayerState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentState
}

// String returns a string representation of the controller state
func (c *Controller) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("PlayerController{state: %s}", c.currentState.String())
}
