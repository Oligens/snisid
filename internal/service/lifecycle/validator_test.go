package lifecycle

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if len(v.allowedMoves) != 6 {
		t.Errorf("States count = %d, want 6", len(v.allowedMoves))
	}
}

func TestValidateTransition_Valid(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		from State
		to   State
	}{
		{StateCreated, StateVerified},
		{StateCreated, StateArchived},
		{StateVerified, StateActive},
		{StateVerified, StateFlagged},
		{StateActive, StateSuspended},
		{StateFlagged, StateActive},
		{StateSuspended, StateActive},
		{StateSuspended, StateArchived},
	}

	for _, tt := range tests {
		err := v.ValidateTransition(tt.from, tt.to)
		if err != nil {
			t.Errorf("Transition %s -> %s should be allowed: %v", tt.from, tt.to, err)
		}
	}
}

func TestValidateTransition_Invalid(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		from State
		to   State
	}{
		{StateCreated, StateSuspended},
		{StateCreated, StateFlagged},
		{StateCreated, StateActive},
		{StateArchived, StateActive}, // Terminal state
	}

	for _, tt := range tests {
		err := v.ValidateTransition(tt.from, tt.to)
		if err == nil {
			t.Errorf("Transition %s -> %s should be blocked", tt.from, tt.to)
		}
	}
}

func TestValidateTransition_SameState(t *testing.T) {
	v := NewValidator()

	states := []State{StateCreated, StateVerified, StateActive, StateFlagged, StateSuspended, StateArchived}
	for _, s := range states {
		err := v.ValidateTransition(s, s)
		if err != nil {
			t.Errorf("Same-state transition for %s should be allowed: %v", s, err)
		}
	}
}

func TestValidateTransition_UnknownState(t *testing.T) {
	v := NewValidator()
	err := v.ValidateTransition("UNKNOWN", StateActive)
	if err == nil {
		t.Error("Expected error for unknown source state")
	}
}

func TestValidateTransition_ArchivedTerminal(t *testing.T) {
	v := NewValidator()
	err := v.ValidateTransition(StateArchived, StateVerified)
	if err == nil {
		t.Error("Archived state should not allow any outgoing transition")
	}
}

func TestStateConstants(t *testing.T) {
	if StateCreated != "CREATED" {
		t.Errorf("StateCreated = %s", StateCreated)
	}
	if StateArchived != "ARCHIVED" {
		t.Errorf("StateArchived = %s", StateArchived)
	}
}
