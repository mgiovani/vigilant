// Package stats provides real-time session statistics tracking including focused time,
// distracted time, and trigger counts. It uses background goroutines to accumulate time
// and maintains thread-safe access to all metrics.
//
// The Tracker listens to state change events and automatically accumulates time in each
// state. It updates stats periodically (every second) rather than on every event to reduce
// overhead and provide stable metrics.
//
// All public methods are thread-safe and can be called concurrently from multiple goroutines.
package stats

import (
	"context"
	"sync"
	"time"

	"vigilant/internal/blocker"
)

// Stats represents current productivity statistics
type Stats struct {
	// FocusedTime represents total time spent in Working state
	FocusedTime time.Duration

	// DistractedTime represents total time spent in Distracted state
	DistractedTime time.Duration

	// TriggerCount represents number of times FBI meme was triggered
	TriggerCount int

	// SessionStart marks when tracking session started
	SessionStart time.Time

	// LastUpdate marks when stats were last updated
	LastUpdate time.Time
}

// Tracker is a thread-safe stats tracker that monitors focus state changes
// and accumulates time spent in each state
type Tracker struct {
	stats        Stats
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	currentState blocker.FocusState
	stateStart   time.Time
	trackingCh   chan blocker.StateChange
}

// NewTracker creates a new Tracker with zero stats and current time as session start
func NewTracker() *Tracker {
	return &Tracker{
		stats: Stats{
			FocusedTime:    0,
			DistractedTime: 0,
			TriggerCount:   0,
			SessionStart:   time.Now(),
			LastUpdate:     time.Now(),
		},
		currentState: blocker.StateWorking,
		stateStart:   time.Now(),
		trackingCh:   make(chan blocker.StateChange, 10), // Buffered channel
	}
}

// Start launches the background goroutine that accumulates time statistics
// It updates stats every second based on the current focus state and respects
// context cancellation for clean shutdown
func (t *Tracker) Start(ctx context.Context) error {
	t.mu.Lock()
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.mu.Unlock()

	go t.trackingLoop()
	return nil
}

// trackingLoop runs in background and accumulates time based on focus state
func (t *Tracker) trackingLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return

		case change := <-t.trackingCh:
			t.ProcessStateChange(change)

		case <-ticker.C:
			t.mu.Lock()
			// Accumulate time based on current state
			elapsed := time.Since(t.stateStart)

			switch t.currentState {
			case blocker.StateWorking:
				t.stats.FocusedTime += elapsed
			case blocker.StateDistracted:
				t.stats.DistractedTime += elapsed
			}

			// Reset state start time for next interval
			t.stateStart = time.Now()

			// Update last update timestamp
			t.stats.LastUpdate = time.Now()
			t.mu.Unlock()
		}
	}
}

// Stop cancels the background goroutine and finalizes time accumulation
func (t *Tracker) Stop() error {
	t.mu.Lock()
	if t.cancel != nil {
		t.cancel()
	}
	t.mu.Unlock()

	// Wait briefly for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	return nil
}

// StateChanges returns a read-only channel for receiving state changes
func (t *Tracker) StateChanges() <-chan blocker.StateChange {
	return t.trackingCh
}

// ProcessStateChange handles a state change event
// Updates currentState and resets stateStart for time tracking
func (t *Tracker) ProcessStateChange(change blocker.StateChange) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Only process if state actually changed
	if change.Current != t.currentState {
		// Accumulate time in previous state
		elapsed := time.Since(t.stateStart)
		switch t.currentState {
		case blocker.StateWorking:
			t.stats.FocusedTime += elapsed
		case blocker.StateDistracted:
			t.stats.DistractedTime += elapsed
		}

		// Update to new state and reset timer
		t.currentState = change.Current
		t.stateStart = time.Now()
		t.stats.LastUpdate = time.Now()
	}
}

// GetStats returns a thread-safe snapshot of current statistics
func (t *Tracker) GetStats() Stats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy of the stats
	return Stats{
		FocusedTime:    t.stats.FocusedTime,
		DistractedTime: t.stats.DistractedTime,
		TriggerCount:   t.stats.TriggerCount,
		SessionStart:   t.stats.SessionStart,
		LastUpdate:     t.stats.LastUpdate,
	}
}

// GetFocusedTime returns total focused time
func (t *Tracker) GetFocusedTime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats.FocusedTime
}

// GetDistractedTime returns total distracted time
func (t *Tracker) GetDistractedTime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats.DistractedTime
}

// GetTriggerCount returns the number of times trigger was activated
func (t *Tracker) GetTriggerCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats.TriggerCount
}

// CalculateFocusRate calculates the focus rate as a percentage (0.0-1.0)
// Formula: FocusedTime / (FocusedTime + DistractedTime)
// Returns 0.0 if no time has been tracked yet
func (t *Tracker) CalculateFocusRate() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	total := t.stats.FocusedTime + t.stats.DistractedTime
	if total == 0 {
		return 0.0
	}

	return float64(t.stats.FocusedTime) / float64(total)
}

// IncrementTrigger increments the trigger count (thread-safe)
// Called when FBI meme is triggered
func (t *Tracker) IncrementTrigger() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stats.TriggerCount++
	t.stats.LastUpdate = time.Now()
}
