package blocker

import (
	"testing"
	"time"

	"vigilant/internal/monitor"
)

// MockBlocker is a mock implementation of Blocker for testing
type MockBlocker struct {
	blockedWindows map[string]bool
}

// NewMockBlocker creates a new mock blocker
func NewMockBlocker() *MockBlocker {
	return &MockBlocker{
		blockedWindows: make(map[string]bool),
	}
}

// IsBlocked implements the Blocker interface for testing
func (mb *MockBlocker) IsBlocked(window *monitor.WindowInfo) bool {
	if window == nil {
		return false
	}
	return mb.blockedWindows[window.Title]
}

// GetBlockReason implements the Blocker interface for testing
func (mb *MockBlocker) GetBlockReason(window *monitor.WindowInfo) string {
	if window == nil || !mb.IsBlocked(window) {
		return ""
	}
	return "Mock blocked: " + window.Title
}

// SetBlocked sets whether a window title should be blocked
func (mb *MockBlocker) SetBlocked(title string, blocked bool) {
	mb.blockedWindows[title] = blocked
}

// TestNewStateManager tests the constructor
func TestNewStateManager(t *testing.T) {
	blocker := NewMockBlocker()
	sm := NewStateManager(blocker)

	if sm == nil {
		t.Fatal("Expected non-nil StateManager")
	}

	if sm.GetCurrentState() != StateWorking {
		t.Errorf("Expected initial state to be Working, got %v", sm.GetCurrentState())
	}

	if sm.GetLastWindow() != nil {
		t.Error("Expected nil lastWindow initially")
	}

	if sm.StateChanges() == nil {
		t.Error("Expected non-nil StateChanges channel")
	}
}

// TestNewStateManager_NilBlocker tests StateManager with nil blocker
func TestNewStateManager_NilBlocker(t *testing.T) {
	sm := NewStateManager(nil)

	if sm == nil {
		t.Fatal("Expected non-nil StateManager")
	}

	if sm.GetCurrentState() != StateWorking {
		t.Errorf("Expected initial state to be Working, got %v", sm.GetCurrentState())
	}
}

// TestFocusState_String tests the String method of FocusState
func TestFocusState_String(t *testing.T) {
	tests := []struct {
		state    FocusState
		expected string
	}{
		{StateWorking, "Working"},
		{StateDistracted, "Distracted"},
		{FocusState(999), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.state.String()
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}

// TestOnFocusChange_WorkingToDistracted tests transition from Working to Distracted
func TestOnFocusChange_WorkingToDistracted(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	sm := NewStateManager(blocker)

	event := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       1234,
			Title:     "Discord",
			Process:   "Discord.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	err := sm.OnFocusChange(event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sm.GetCurrentState() != StateDistracted {
		t.Errorf("Expected Distracted state, got %v", sm.GetCurrentState())
	}

	// Check that state change was emitted
	select {
	case change := <-sm.StateChanges():
		if change.Previous != StateWorking {
			t.Errorf("Expected previous state Working, got %v", change.Previous)
		}
		if change.Current != StateDistracted {
			t.Errorf("Expected current state Distracted, got %v", change.Current)
		}
		if change.Window.Title != "Discord" {
			t.Errorf("Expected window title Discord, got %s", change.Window.Title)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected state change event, got timeout")
	}
}

// TestOnFocusChange_DistractedToWorking tests transition from Distracted to Working
func TestOnFocusChange_DistractedToWorking(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	blocker.SetBlocked("VSCode", false)

	sm := NewStateManager(blocker)

	// First, transition to Distracted
	event1 := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       1234,
			Title:     "Discord",
			Process:   "Discord.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}
	sm.OnFocusChange(event1)
	_ = <-sm.StateChanges() // Consume the first change

	// Then, transition back to Working
	event2 := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       5678,
			Title:     "VSCode",
			Process:   "code.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	err := sm.OnFocusChange(event2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sm.GetCurrentState() != StateWorking {
		t.Errorf("Expected Working state, got %v", sm.GetCurrentState())
	}

	// Check that state change was emitted
	select {
	case change := <-sm.StateChanges():
		if change.Previous != StateDistracted {
			t.Errorf("Expected previous state Distracted, got %v", change.Previous)
		}
		if change.Current != StateWorking {
			t.Errorf("Expected current state Working, got %v", change.Current)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected state change event, got timeout")
	}
}

// TestOnFocusChange_DuplicateStateNoEvent tests that duplicate state changes don't emit event
func TestOnFocusChange_DuplicateStateNoEvent(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	sm := NewStateManager(blocker)

	event := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       1234,
			Title:     "Discord",
			Process:   "Discord.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	// First call - should transition and emit
	sm.OnFocusChange(event)
	_ = <-sm.StateChanges() // Consume the first change

	// Second call - same state, should not emit
	sm.OnFocusChange(event)

	// Try to receive from channel - should timeout
	select {
	case <-sm.StateChanges():
		t.Error("Expected no state change event for duplicate state")
	case <-time.After(50 * time.Millisecond):
		// Expected - no event was sent
	}
}

// TestOnFocusChange_UpdatesLastWindow tests that lastWindow is updated
func TestOnFocusChange_UpdatesLastWindow(t *testing.T) {
	sm := NewStateManager(NewMockBlocker())

	window := &monitor.WindowInfo{
		PID:       1234,
		Title:     "Test Window",
		Process:   "test.exe",
		Timestamp: time.Now(),
	}

	event := &monitor.FocusEvent{
		Current:   window,
		Timestamp: time.Now(),
	}

	sm.OnFocusChange(event)

	lastWindow := sm.GetLastWindow()
	if lastWindow == nil {
		t.Fatal("Expected non-nil lastWindow")
	}

	if lastWindow.Title != window.Title {
		t.Errorf("Expected lastWindow title %s, got %s", window.Title, lastWindow.Title)
	}
}

// TestOnFocusChange_NilEvent tests handling of nil event
func TestOnFocusChange_NilEvent(t *testing.T) {
	sm := NewStateManager(NewMockBlocker())

	err := sm.OnFocusChange(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sm.GetCurrentState() != StateWorking {
		t.Error("Expected state to remain unchanged")
	}
}

// TestOnFocusChange_NilCurrentWindow tests handling of event with nil Current
func TestOnFocusChange_NilCurrentWindow(t *testing.T) {
	sm := NewStateManager(NewMockBlocker())

	event := &monitor.FocusEvent{
		Current: nil,
	}

	err := sm.OnFocusChange(event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sm.GetCurrentState() != StateWorking {
		t.Error("Expected state to remain unchanged")
	}
}

// TestGetCurrentState_ThreadSafe tests thread safety of GetCurrentState
func TestGetCurrentState_ThreadSafe(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	blocker.SetBlocked("VSCode", false)

	sm := NewStateManager(blocker)

	// Channel to collect results from goroutines
	results := make(chan FocusState, 10)

	// Launch multiple goroutines reading state
	for i := 0; i < 5; i++ {
		go func() {
			results <- sm.GetCurrentState()
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		state := <-results
		if state != StateWorking {
			t.Errorf("Expected Working state in goroutine, got %v", state)
		}
	}
}

// TestGetLastWindow_ThreadSafe tests thread safety of GetLastWindow
func TestGetLastWindow_ThreadSafe(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	sm := NewStateManager(blocker)

	window := &monitor.WindowInfo{
		PID:     1234,
		Title:   "Discord",
		Process: "Discord.exe",
	}

	event := &monitor.FocusEvent{
		Current:   window,
		Timestamp: time.Now(),
	}

	sm.OnFocusChange(event)

	// Try to consume state change with timeout
	select {
	case <-sm.StateChanges():
		// State change received
	case <-time.After(100 * time.Millisecond):
		// No state change (that's fine)
	}

	// Channel to collect results
	results := make(chan *monitor.WindowInfo, 10)

	// Launch multiple goroutines reading last window
	for i := 0; i < 5; i++ {
		go func() {
			results <- sm.GetLastWindow()
		}()
	}

	// Collect results with timeout
	for i := 0; i < 5; i++ {
		select {
		case w := <-results:
			if w == nil {
				t.Error("Expected non-nil window in goroutine")
			}
			if w.Title != "Discord" {
				t.Errorf("Expected title Discord, got %s", w.Title)
			}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for window result")
		}
	}
}

// TestOnFocusChange_MultipleTransitions tests multiple state changes in sequence
func TestOnFocusChange_MultipleTransitions(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	blocker.SetBlocked("Reddit", true)
	blocker.SetBlocked("VSCode", false)

	sm := NewStateManager(blocker)

	transitions := []struct {
		title   string
		blocked bool
	}{
		{"Discord", true},
		{"VSCode", false},
		{"Reddit", true},
		{"VSCode", false},
	}

	expectedStates := []FocusState{
		StateDistracted,
		StateWorking,
		StateDistracted,
		StateWorking,
	}

	for i, transition := range transitions {
		event := &monitor.FocusEvent{
			Current: &monitor.WindowInfo{
				PID:       uint32(1000 + i),
				Title:     transition.title,
				Process:   "test.exe",
				Timestamp: time.Now(),
			},
			Timestamp: time.Now(),
		}

		sm.OnFocusChange(event)

		// Consume the state change if it exists
		select {
		case <-sm.StateChanges():
			// State changed
		case <-time.After(10 * time.Millisecond):
			// State didn't change
		}

		if sm.GetCurrentState() != expectedStates[i] {
			t.Errorf("At transition %d: expected state %v, got %v", i, expectedStates[i], sm.GetCurrentState())
		}
	}
}

// TestStateChange_Structure tests that StateChange contains correct information
func TestStateChange_Structure(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord - Main", true)
	sm := NewStateManager(blocker)

	window := &monitor.WindowInfo{
		PID:       9999,
		Title:     "Discord - Main",
		Process:   "Discord.exe",
		Timestamp: time.Now(),
	}

	event := &monitor.FocusEvent{
		Current:   window,
		Timestamp: time.Now(),
	}

	sm.OnFocusChange(event)

	select {
	case change := <-sm.StateChanges():
		if change.Previous != StateWorking {
			t.Errorf("Expected Previous=Working, got %v", change.Previous)
		}
		if change.Current != StateDistracted {
			t.Errorf("Expected Current=Distracted, got %v", change.Current)
		}
		if change.Window != window {
			t.Error("Expected Window to be the same window from event")
		}
		if change.Timestamp.IsZero() {
			t.Error("Expected non-zero Timestamp")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Expected state change event")
	}
}

// TestOnFocusChange_WithBlockedWindow tests OnFocusChange with blocked window
func TestOnFocusChange_WithBlockedWindow(t *testing.T) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Reddit - Front Page", true)
	sm := NewStateManager(blocker)

	event := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       1234,
			Title:     "Reddit - Front Page",
			Process:   "chrome.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	err := sm.OnFocusChange(event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sm.GetCurrentState() != StateDistracted {
		t.Errorf("Expected Distracted state with blocked window, got %v", sm.GetCurrentState())
	}
}

// TestOnFocusChange_WithAllowedWindow tests OnFocusChange with allowed window
func TestOnFocusChange_WithAllowedWindow(t *testing.T) {
	sm := NewStateManager(NewMockBlocker())

	event := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       1234,
			Title:     "VSCode - my-project",
			Process:   "code.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	err := sm.OnFocusChange(event)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sm.GetCurrentState() != StateWorking {
		t.Errorf("Expected Working state with allowed window, got %v", sm.GetCurrentState())
	}
}

// BenchmarkOnFocusChange benchmarks the OnFocusChange method
func BenchmarkOnFocusChange(b *testing.B) {
	blocker := NewMockBlocker()
	blocker.SetBlocked("Discord", true)
	sm := NewStateManager(blocker)

	event := &monitor.FocusEvent{
		Current: &monitor.WindowInfo{
			PID:       1234,
			Title:     "VSCode",
			Process:   "code.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.OnFocusChange(event)
		// Non-blocking receive to prevent channel buffer from filling
		select {
		case <-sm.StateChanges():
		default:
		}
	}
}

// BenchmarkGetCurrentState benchmarks the GetCurrentState method
func BenchmarkGetCurrentState(b *testing.B) {
	sm := NewStateManager(NewMockBlocker())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.GetCurrentState()
	}
}
