package lifecycle

import (
	"fmt"
)

type State string

const (
	StateCreated   State = "CREATED"
	StateVerified  State = "VERIFIED"
	StateActive    State = "ACTIVE"
	StateFlagged   State = "FLAGGED"
	StateSuspended State = "SUSPENDED"
	StateArchived  State = "ARCHIVED"
)

type Validator struct {
	allowedMoves map[State][]State
}

func NewValidator() *Validator {
	return &Validator{
		allowedMoves: map[State][]State{
			StateCreated:   {StateVerified, StateArchived},
			StateVerified:  {StateActive, StateFlagged, StateSuspended, StateArchived},
			StateActive:    {StateFlagged, StateSuspended, StateArchived},
			StateFlagged:   {StateActive, StateSuspended, StateArchived},
			StateSuspended: {StateActive, StateArchived},
			StateArchived:  {}, // Terminal state
		},
	}
}

func (v *Validator) ValidateTransition(from, to State) error {
	if from == to {
		return nil
	}

	allowed, ok := v.allowedMoves[from]
	if !ok {
		return fmt.Errorf("unknown source state: %s", from)
	}

	for _, s := range allowed {
		if s == to {
			return nil
		}
	}

	return fmt.Errorf("transition from %s to %s is not permitted", from, to)
}
