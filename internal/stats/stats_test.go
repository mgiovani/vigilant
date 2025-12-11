package stats

import (
	"context"
	"sync"
	"testing"
	"time"

	"vigilant/internal/blocker"
	"vigilant/internal/monitor"
)

// TestNewTracker_InitialState tests that constructor initializes to zero
func TestNewTracker_InitialState(t *testing.T) {
	tracker := NewTracker()

	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}

	stats := tracker.GetStats()
	if stats.FocusedTime != 0 {
		t.Errorf("Expected FocusedTime=0, got %v", stats.FocusedTime)
	}
	if stats.DistractedTime != 0 {
		t.Errorf("Expected DistractedTime=0, got %v", stats.DistractedTime)
	}
	if stats.TriggerCount != 0 {
		t.Errorf("Expected TriggerCount=0, got %d", stats.TriggerCount)
	}
}

// TestNewTracker_SessionStart tests that SessionStart is recent
func TestNewTracker_SessionStart(t *testing.T) {
	before := time.Now()
	tracker := NewTracker()
	after := time.Now()

	stats := tracker.GetStats()
	if stats.SessionStart.Before(before) || stats.SessionStart.After(after.Add(time.Second)) {
		t.Errorf("SessionStart not within expected timeframe: %v vs %v-%v", stats.SessionStart, before, after)
	}
}

// TestStartStop tests Start/Stop lifecycle
func TestStartStop(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	err := tracker.Start(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Start, got %v", err)
	}

	// Give goroutine time to start
	time.Sleep(50 * time.Millisecond)

	err = tracker.Stop()
	if err != nil {
		t.Fatalf("Expected no error from Stop, got %v", err)
	}
}

// TestBackgroundGoroutineUpdatesEverySecond tests that stats update every second
func TestBackgroundGoroutineUpdatesEverySecond(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	// Give goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Record initial LastUpdate
	initialStats := tracker.GetStats()
	initialUpdate := initialStats.LastUpdate

	// Wait for 1.5 seconds (should see at least 1 update)
	time.Sleep(1500 * time.Millisecond)

	// Check that LastUpdate changed
	updatedStats := tracker.GetStats()
	if !updatedStats.LastUpdate.After(initialUpdate) {
		t.Error("Expected LastUpdate to be updated after 1.5 seconds")
	}
}

// TestFocusedTimeAccumulates tests that FocusedTime accumulates correctly
func TestFocusedTimeAccumulates(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	// Initial state is Working, so FocusedTime should accumulate
	time.Sleep(1100 * time.Millisecond) // Let it accumulate for ~1 second

	stats := tracker.GetStats()
	// Should be roughly 1 second (with tolerance)
	if stats.FocusedTime < 900*time.Millisecond || stats.FocusedTime > 1200*time.Millisecond {
		t.Errorf("Expected FocusedTime ~1000ms, got %v", stats.FocusedTime)
	}

	// DistractedTime should still be zero
	if stats.DistractedTime != 0 {
		t.Errorf("Expected DistractedTime=0, got %v", stats.DistractedTime)
	}
}

// TestDistractedTimeAccumulates tests that DistractedTime accumulates correctly
func TestDistractedTimeAccumulates(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	// Transition to Distracted state
	change := blocker.StateChange{
		Previous:  blocker.StateWorking,
		Current:   blocker.StateDistracted,
		Timestamp: time.Now(),
	}
	tracker.ProcessStateChange(change)

	// Now DistractedTime should accumulate
	time.Sleep(1100 * time.Millisecond)

	stats := tracker.GetStats()
	// DistractedTime should be roughly 1 second
	if stats.DistractedTime < 900*time.Millisecond || stats.DistractedTime > 1200*time.Millisecond {
		t.Errorf("Expected DistractedTime ~1000ms, got %v", stats.DistractedTime)
	}
}

// TestStateTransitionsUpdateTracking tests that state transitions update tracking correctly
func TestStateTransitionsUpdateTracking(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	// Let some focused time accumulate (need >1 second for ticker to fire)
	time.Sleep(1100 * time.Millisecond)

	// Transition to Distracted
	change := blocker.StateChange{
		Previous:  blocker.StateWorking,
		Current:   blocker.StateDistracted,
		Timestamp: time.Now(),
	}
	tracker.ProcessStateChange(change)

	// Let some distracted time accumulate (need >1 second for ticker to fire)
	time.Sleep(1100 * time.Millisecond)

	stats := tracker.GetStats()
	// Both should have accumulated (at least 900ms in each state)
	if stats.FocusedTime < 900*time.Millisecond {
		t.Errorf("Expected some FocusedTime, got %v", stats.FocusedTime)
	}
	if stats.DistractedTime < 900*time.Millisecond {
		t.Errorf("Expected some DistractedTime, got %v", stats.DistractedTime)
	}
}

// TestIncrementTrigger tests that IncrementTrigger increments counter
func TestIncrementTrigger(t *testing.T) {
	tracker := NewTracker()

	if tracker.GetTriggerCount() != 0 {
		t.Error("Expected initial trigger count to be 0")
	}

	tracker.IncrementTrigger()
	if tracker.GetTriggerCount() != 1 {
		t.Error("Expected trigger count to be 1 after increment")
	}

	tracker.IncrementTrigger()
	tracker.IncrementTrigger()
	if tracker.GetTriggerCount() != 3 {
		t.Error("Expected trigger count to be 3 after 3 increments")
	}
}

// TestGetStats returns accurate snapshot
func TestGetStatsReturnsSnapshot(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	time.Sleep(1100 * time.Millisecond)
	tracker.IncrementTrigger()

	stats1 := tracker.GetStats()
	stats2 := tracker.GetStats()

	// Both snapshots should match
	if stats1.FocusedTime != stats2.FocusedTime {
		t.Error("Expected consistent FocusedTime in snapshots")
	}
	if stats1.TriggerCount != stats2.TriggerCount {
		t.Error("Expected consistent TriggerCount in snapshots")
	}
}

// TestCalculateFocusRateNormal tests focus rate for normal cases
func TestCalculateFocusRateNormal(t *testing.T) {
	tests := []struct {
		name            string
		focusedTime     time.Duration
		distractedTime  time.Duration
		expectedRateMin float64
		expectedRateMax float64
	}{
		{
			name:            "50/50 split",
			focusedTime:     50 * time.Second,
			distractedTime:  50 * time.Second,
			expectedRateMin: 0.49,
			expectedRateMax: 0.51,
		},
		{
			name:            "100% focused",
			focusedTime:     100 * time.Second,
			distractedTime:  0,
			expectedRateMin: 0.99,
			expectedRateMax: 1.01,
		},
		{
			name:            "0% focused",
			focusedTime:     0,
			distractedTime:  100 * time.Second,
			expectedRateMin: -0.01,
			expectedRateMax: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker()
			tracker.mu.Lock()
			tracker.stats.FocusedTime = tt.focusedTime
			tracker.stats.DistractedTime = tt.distractedTime
			tracker.mu.Unlock()

			rate := tracker.CalculateFocusRate()
			if rate < tt.expectedRateMin || rate > tt.expectedRateMax {
				t.Errorf("Expected rate between %f and %f, got %f", tt.expectedRateMin, tt.expectedRateMax, rate)
			}
		})
	}
}

// TestCalculateFocusRateZeroTime tests focus rate edge case with no time tracked
func TestCalculateFocusRateZeroTime(t *testing.T) {
	tracker := NewTracker()

	rate := tracker.CalculateFocusRate()
	if rate != 0.0 {
		t.Errorf("Expected 0.0 focus rate with no time tracked, got %f", rate)
	}
}

// TestThreadSafeConcurrentAccess tests thread-safe concurrent access
func TestThreadSafeConcurrentAccess(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Launch 10 concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_ = tracker.GetStats()
				_ = tracker.GetFocusedTime()
				_ = tracker.GetDistractedTime()
				_ = tracker.GetTriggerCount()
				_ = tracker.CalculateFocusRate()
			}
		}()
	}

	// Launch 5 concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				tracker.IncrementTrigger()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	// Wait for all goroutines
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("Unexpected error during concurrent access: %v", err)
		}
	}
}

// TestGracefulShutdownWithContextCancellation tests graceful shutdown
func TestGracefulShutdownWithContextCancellation(t *testing.T) {
	tracker := NewTracker()
	ctx, cancel := context.WithCancel(context.Background())

	tracker.Start(ctx)

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for shutdown
	time.Sleep(200 * time.Millisecond)

	// Should be able to stop without issues
	err := tracker.Stop()
	if err != nil {
		t.Fatalf("Expected no error during graceful shutdown, got %v", err)
	}
}

// TestTimingAccuracyOver5Seconds tests timing accuracy within acceptable margin
func TestTimingAccuracyOver5Seconds(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	// Let it accumulate for 5 seconds
	time.Sleep(5100 * time.Millisecond)

	stats := tracker.GetStats()
	focusedTime := stats.FocusedTime

	// Should be roughly 5 seconds ±200ms
	minExpected := 4800 * time.Millisecond
	maxExpected := 5200 * time.Millisecond

	if focusedTime < minExpected || focusedTime > maxExpected {
		t.Errorf("Expected FocusedTime ~5000ms ±200ms, got %v", focusedTime)
	}
}

// TestLastUpdateTimestampUpdates tests that LastUpdate timestamp updates correctly
func TestLastUpdateTimestampUpdates(t *testing.T) {
	tracker := NewTracker()

	initialStats := tracker.GetStats()
	initialUpdate := initialStats.LastUpdate

	// Increment trigger and check LastUpdate changed
	time.Sleep(50 * time.Millisecond)
	tracker.IncrementTrigger()

	updatedStats := tracker.GetStats()
	if !updatedStats.LastUpdate.After(initialUpdate) {
		t.Error("Expected LastUpdate to change after IncrementTrigger")
	}
}

// TestProcessStateChangeIgnoresDuplicates tests that duplicate state changes are ignored
func TestProcessStateChangeIgnoresDuplicates(t *testing.T) {
	tracker := NewTracker()

	// Transition to Distracted
	change1 := blocker.StateChange{
		Previous:  blocker.StateWorking,
		Current:   blocker.StateDistracted,
		Timestamp: time.Now(),
	}
	tracker.ProcessStateChange(change1)

	stats1 := tracker.GetStats()
	update1 := stats1.LastUpdate

	// Try to transition to Distracted again (no state change)
	time.Sleep(50 * time.Millisecond)
	change2 := blocker.StateChange{
		Previous:  blocker.StateDistracted,
		Current:   blocker.StateDistracted,
		Timestamp: time.Now(),
	}
	tracker.ProcessStateChange(change2)

	stats2 := tracker.GetStats()
	update2 := stats2.LastUpdate

	// LastUpdate should not change for duplicate state (since condition is if change.Current != t.currentState)
	if update2.After(update1.Add(10 * time.Millisecond)) {
		t.Error("Expected LastUpdate to not change for duplicate state change")
	}
}

// TestStateChangesChannel tests that StateChanges returns read-only channel
func TestStateChangesChannel(t *testing.T) {
	tracker := NewTracker()

	ch := tracker.StateChanges()
	if ch == nil {
		t.Fatal("Expected non-nil channel from StateChanges")
	}

	// Verify it's read-only by trying to use it as receive-only
	select {
	case <-ch:
		// Would receive if there were a change
	default:
		// No change yet is fine
	}
}

// TestMultipleTriggerIncrements tests multiple rapid trigger increments
func TestMultipleTriggerIncrements(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	// Rapid fire triggers
	for i := 0; i < 100; i++ {
		tracker.IncrementTrigger()
	}

	stats := tracker.GetStats()
	if stats.TriggerCount != 100 {
		t.Errorf("Expected TriggerCount=100, got %d", stats.TriggerCount)
	}
}

// TestGettersConsistency tests that all getter methods are consistent
func TestGettersConsistency(t *testing.T) {
	tracker := NewTracker()
	ctx := context.Background()

	tracker.Start(ctx)
	defer tracker.Stop()

	time.Sleep(1100 * time.Millisecond)
	tracker.IncrementTrigger()

	// Get values via individual methods
	focusedTime1 := tracker.GetFocusedTime()
	distractedTime1 := tracker.GetDistractedTime()
	triggerCount1 := tracker.GetTriggerCount()

	// Get values via snapshot
	stats := tracker.GetStats()
	focusedTime2 := stats.FocusedTime
	distractedTime2 := stats.DistractedTime
	triggerCount2 := stats.TriggerCount

	// All should match
	if focusedTime1 != focusedTime2 {
		t.Errorf("FocusedTime mismatch: %v vs %v", focusedTime1, focusedTime2)
	}
	if distractedTime1 != distractedTime2 {
		t.Errorf("DistractedTime mismatch: %v vs %v", distractedTime1, distractedTime2)
	}
	if triggerCount1 != triggerCount2 {
		t.Errorf("TriggerCount mismatch: %d vs %d", triggerCount1, triggerCount2)
	}
}

// TestStateChangeWithWindowInfo tests state change with complete window info
func TestStateChangeWithWindowInfo(t *testing.T) {
	tracker := NewTracker()

	change := blocker.StateChange{
		Previous: blocker.StateWorking,
		Current:  blocker.StateDistracted,
		Window: &monitor.WindowInfo{
			PID:       1234,
			Title:     "Discord",
			Process:   "Discord.exe",
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Should not panic
	tracker.ProcessStateChange(change)

	stats := tracker.GetStats()
	if stats.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set after ProcessStateChange")
	}
}

// BenchmarkGetStats benchmarks the GetStats method
func BenchmarkGetStats(b *testing.B) {
	tracker := NewTracker()
	ctx := context.Background()
	tracker.Start(ctx)
	defer tracker.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.GetStats()
	}
}

// BenchmarkIncrementTrigger benchmarks the IncrementTrigger method
func BenchmarkIncrementTrigger(b *testing.B) {
	tracker := NewTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.IncrementTrigger()
	}
}

// BenchmarkCalculateFocusRate benchmarks the CalculateFocusRate method
func BenchmarkCalculateFocusRate(b *testing.B) {
	tracker := NewTracker()
	tracker.mu.Lock()
	tracker.stats.FocusedTime = 100 * time.Second
	tracker.stats.DistractedTime = 50 * time.Second
	tracker.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.CalculateFocusRate()
	}
}

// BenchmarkProcessStateChange benchmarks the ProcessStateChange method
func BenchmarkProcessStateChange(b *testing.B) {
	tracker := NewTracker()

	change := blocker.StateChange{
		Previous:  blocker.StateWorking,
		Current:   blocker.StateDistracted,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.ProcessStateChange(change)
	}
}
