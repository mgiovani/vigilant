package monitor

import (
	"testing"
	"time"

	"vigilant/internal/config"
)

// TestGracePeriodExpires tests that a distraction event is emitted
// when the grace period expires without user switching away.
func TestGracePeriodExpires(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(100 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)
	window := &WindowInfo{
		PID:       1234,
		Title:     "Distracted Site",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	event := FocusEvent{
		Previous:  nil,
		Current:   window,
		Timestamp: time.Now(),
	}

	// Process a blocked window
	resultCh := detector.ProcessEvent(event, true)

	if resultCh == nil {
		t.Fatal("expected non-nil result channel for blocked window")
	}

	// Wait for the grace period to expire
	select {
	case result := <-resultCh:
		if result.Window.PID != window.PID {
			t.Errorf("expected PID %d, got %d", window.PID, result.Window.PID)
		}
		if !result.Timestamp.After(event.Timestamp) {
			t.Error("distraction timestamp should be after event timestamp")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("grace period did not expire within expected time")
	}

	// Verify state changed to distracted
	if detector.GetCurrentState() != StateDistracted {
		t.Errorf("expected state %s, got %s", StateDistracted, detector.GetCurrentState())
	}
}

// TestGracePeriodCancelledOnSwitch tests that the grace period is cancelled
// when the user switches away from blocked content.
func TestGracePeriodCancelledOnSwitch(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(500 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)

	blockedWindow := &WindowInfo{
		PID:       1234,
		Title:     "Distracted Site",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	// Start with blocked window
	event := FocusEvent{
		Previous:  nil,
		Current:   blockedWindow,
		Timestamp: time.Now(),
	}

	resultCh := detector.ProcessEvent(event, true)

	// User switches away before grace period expires
	time.Sleep(100 * time.Millisecond)

	workWindow := &WindowInfo{
		PID:       5678,
		Title:     "Work Project",
		Process:   "vscode.exe",
		Timestamp: time.Now(),
	}

	switchEvent := FocusEvent{
		Previous:  blockedWindow,
		Current:   workWindow,
		Timestamp: time.Now(),
	}

	// This should cancel the pending distraction
	switchResultCh := detector.ProcessEvent(switchEvent, false)

	if switchResultCh != nil {
		t.Fatal("expected nil result channel for unblocked window")
	}

	// The original result channel should NOT emit within the original grace period
	select {
	case result := <-resultCh:
		// Check if we got a result - we shouldn't
		t.Errorf("unexpected distraction event after switch: %v", result)
	case <-time.After(600 * time.Millisecond):
		// This is expected - no distraction event should be emitted
		// If we get here, the grace period was properly cancelled
	}

	// Verify state switched back to working
	if detector.GetCurrentState() != StateWorking {
		t.Errorf("expected state %s, got %s", StateWorking, detector.GetCurrentState())
	}
}

// TestNonBlockedWindowDoesNotTriggerGracePeriod tests that non-blocked
// windows do not trigger the grace period.
func TestNonBlockedWindowDoesNotTriggerGracePeriod(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(100 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)

	workWindow := &WindowInfo{
		PID:       5678,
		Title:     "Work Project",
		Process:   "vscode.exe",
		Timestamp: time.Now(),
	}

	event := FocusEvent{
		Previous:  nil,
		Current:   workWindow,
		Timestamp: time.Now(),
	}

	// Process a non-blocked window
	resultCh := detector.ProcessEvent(event, false)

	if resultCh != nil {
		t.Fatal("expected nil result channel for non-blocked window")
	}

	// Verify state is working
	if detector.GetCurrentState() != StateWorking {
		t.Errorf("expected state %s, got %s", StateWorking, detector.GetCurrentState())
	}
}

// TestRapidSwitchingDebounce tests that rapid switching between windows
// doesn't cause false positives.
func TestRapidSwitchingDebounce(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(200 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)

	blockedWindow := &WindowInfo{
		PID:       1111,
		Title:     "Distracted 1",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	// Switch to blocked window
	event1 := FocusEvent{
		Current:   blockedWindow,
		Timestamp: time.Now(),
	}
	ch1 := detector.ProcessEvent(event1, true)

	// Rapidly switch to another window
	time.Sleep(50 * time.Millisecond)

	workWindow := &WindowInfo{
		PID:       2222,
		Title:     "Work",
		Process:   "vscode.exe",
		Timestamp: time.Now(),
	}

	event2 := FocusEvent{
		Previous:  blockedWindow,
		Current:   workWindow,
		Timestamp: time.Now(),
	}
	ch2 := detector.ProcessEvent(event2, false)

	// Rapidly switch back to a different blocked window
	time.Sleep(50 * time.Millisecond)

	blockedWindow2 := &WindowInfo{
		PID:       3333,
		Title:     "Distracted 2",
		Process:   "firefox.exe",
		Timestamp: time.Now(),
	}

	event3 := FocusEvent{
		Previous:  workWindow,
		Current:   blockedWindow2,
		Timestamp: time.Now(),
	}
	ch3 := detector.ProcessEvent(event3, true)

	// We should only get a result from ch3 (the current pending), not ch1
	select {
	case <-ch1:
		t.Error("unexpected result from first blocked window")
	case <-time.After(300 * time.Millisecond):
		// This is expected - ch1 should timeout
	}

	// ch2 should be nil (non-blocked)
	if ch2 != nil {
		t.Error("ch2 should be nil for non-blocked window")
	}

	// ch3 should emit a distraction event
	select {
	case result := <-ch3:
		if result.Window.PID != blockedWindow2.PID {
			t.Errorf("expected PID %d, got %d", blockedWindow2.PID, result.Window.PID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ch3 did not emit distraction event within grace period")
	}
}

// TestReset tests that Reset() cancels pending timers and resets state.
func TestReset(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(500 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)

	blockedWindow := &WindowInfo{
		PID:       1234,
		Title:     "Distracted Site",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	event := FocusEvent{
		Current:   blockedWindow,
		Timestamp: time.Now(),
	}

	resultCh := detector.ProcessEvent(event, true)
	if resultCh == nil {
		t.Fatal("expected non-nil result channel")
	}

	// Reset immediately
	detector.Reset()

	// Grace period should not expire
	select {
	case <-resultCh:
		t.Error("unexpected distraction event after reset")
	case <-time.After(600 * time.Millisecond):
		// This is expected
	}

	// Verify state is working
	if detector.GetCurrentState() != StateWorking {
		t.Errorf("expected state %s after reset, got %s", StateWorking, detector.GetCurrentState())
	}
}

// TestCancelPending tests that CancelPending() properly cancels pending events.
func TestCancelPending(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(500 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)

	blockedWindow := &WindowInfo{
		PID:       1234,
		Title:     "Distracted Site",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	event := FocusEvent{
		Current:   blockedWindow,
		Timestamp: time.Now(),
	}

	resultCh := detector.ProcessEvent(event, true)

	// Cancel the pending distraction
	time.Sleep(100 * time.Millisecond)
	detector.CancelPending()

	// Grace period should not expire
	select {
	case <-resultCh:
		t.Error("unexpected distraction event after cancel")
	case <-time.After(600 * time.Millisecond):
		// This is expected
	}

	// Verify state is working
	if detector.GetCurrentState() != StateWorking {
		t.Errorf("expected state %s after cancel, got %s", StateWorking, detector.GetCurrentState())
	}
}

// TestMultipleGracePeriods tests that multiple grace periods can be overlapped
// and only the latest one matters.
func TestMultipleGracePeriods(t *testing.T) {
	cfg := &config.Config{
		Monitor: config.MonitorConfig{
			GracePeriod: config.DurationValue(150 * time.Millisecond),
		},
	}

	detector := NewDetector(cfg)

	window1 := &WindowInfo{
		PID:       1111,
		Title:     "Distracted 1",
		Process:   "chrome.exe",
		Timestamp: time.Now(),
	}

	// Start first grace period
	event1 := FocusEvent{
		Current:   window1,
		Timestamp: time.Now(),
	}
	ch1 := detector.ProcessEvent(event1, true)

	time.Sleep(50 * time.Millisecond)

	window2 := &WindowInfo{
		PID:       2222,
		Title:     "Distracted 2",
		Process:   "firefox.exe",
		Timestamp: time.Now(),
	}

	// Start second grace period (should cancel first)
	event2 := FocusEvent{
		Previous:  window1,
		Current:   window2,
		Timestamp: time.Now(),
	}
	ch2 := detector.ProcessEvent(event2, true)

	// Wait for both to potentially expire
	select {
	case <-ch1:
		t.Error("first window should not emit after being replaced")
	case <-time.After(300 * time.Millisecond):
		// Expected
	}

	// Second window should emit
	select {
	case result := <-ch2:
		if result.Window.PID != window2.PID {
			t.Errorf("expected PID %d, got %d", window2.PID, result.Window.PID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second window did not emit within grace period")
	}
}
