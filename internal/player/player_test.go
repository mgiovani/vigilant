package player

import (
	"sync"
	"testing"
)

func TestNewController(t *testing.T) {
	ctrl := NewController()
	if ctrl == nil {
		t.Fatal("NewController returned nil")
	}

	// Should start in lofi state
	if ctrl.GetState() != StateLofi {
		t.Errorf("Expected initial state to be StateLofi, got %v", ctrl.GetState())
	}
}

func TestPlayerStateString(t *testing.T) {
	tests := []struct {
		state    PlayerState
		expected string
	}{
		{StateLofi, "lofi"},
		{StateFBI, "fbi"},
		{PlayerState(99), "unknown"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("PlayerState(%d).String() = %q, want %q", tt.state, tt.state.String(), tt.expected)
		}
	}
}

func TestSwitchToLofi(t *testing.T) {
	ctrl := NewController()
	ctrl.SwitchToFBI()
	if ctrl.GetState() != StateFBI {
		t.Fatal("Failed to switch to FBI state")
	}

	ctrl.SwitchToLofi()
	if ctrl.GetState() != StateLofi {
		t.Errorf("Expected StateLofi after SwitchToLofi, got %v", ctrl.GetState())
	}
}

func TestSwitchToFBI(t *testing.T) {
	ctrl := NewController()
	if ctrl.GetState() != StateLofi {
		t.Fatal("Expected initial state to be StateLofi")
	}

	ctrl.SwitchToFBI()
	if ctrl.GetState() != StateFBI {
		t.Errorf("Expected StateFBI after SwitchToFBI, got %v", ctrl.GetState())
	}
}

func TestGetState(t *testing.T) {
	ctrl := NewController()

	// Test lofi state
	state := ctrl.GetState()
	if state != StateLofi {
		t.Errorf("Expected StateLofi, got %v", state)
	}

	// Test after switch
	ctrl.SwitchToFBI()
	state = ctrl.GetState()
	if state != StateFBI {
		t.Errorf("Expected StateFBI, got %v", state)
	}
}

func TestControllerString(t *testing.T) {
	ctrl := NewController()
	str := ctrl.String()
	if str != "PlayerController{state: lofi}" {
		t.Errorf("String() = %q, want %q", str, "PlayerController{state: lofi}")
	}

	ctrl.SwitchToFBI()
	str = ctrl.String()
	if str != "PlayerController{state: fbi}" {
		t.Errorf("String() = %q, want %q", str, "PlayerController{state: fbi}")
	}
}

func TestControllerThreadSafety(t *testing.T) {
	ctrl := NewController()
	var wg sync.WaitGroup
	numGoroutines := 10

	// Launch concurrent readers and writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		// Writer goroutine
		go func() {
			defer wg.Done()
			if i%2 == 0 {
				ctrl.SwitchToFBI()
			} else {
				ctrl.SwitchToLofi()
			}
		}()

		// Reader goroutine
		go func() {
			defer wg.Done()
			_ = ctrl.GetState()
		}()

		// String goroutine
		go func() {
			defer wg.Done()
			_ = ctrl.String()
		}()
	}

	wg.Wait()

	// Should complete without panic
	state := ctrl.GetState()
	if state != StateLofi && state != StateFBI {
		t.Errorf("Invalid state after concurrent operations: %v", state)
	}
}

func TestStateTransitions(t *testing.T) {
	ctrl := NewController()

	// Test multiple transitions
	transitions := []struct {
		fn       func()
		expected PlayerState
	}{
		{ctrl.SwitchToFBI, StateFBI},
		{ctrl.SwitchToLofi, StateLofi},
		{ctrl.SwitchToFBI, StateFBI},
		{ctrl.SwitchToFBI, StateFBI},
		{ctrl.SwitchToLofi, StateLofi},
	}

	for _, tt := range transitions {
		tt.fn()
		if ctrl.GetState() != tt.expected {
			t.Errorf("After transition, expected %v, got %v", tt.expected, ctrl.GetState())
		}
	}
}
