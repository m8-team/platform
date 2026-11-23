package resource

import (
	"errors"
	"testing"
	"time"
)

func TestLifecycleTransitionCAS(t *testing.T) {
	now := time.Date(2026, time.November, 23, 10, 0, 0, 0, time.UTC)
	state := Lifecycle{State: StateActive, Revision: 4}
	next, err := state.Transition(4, StateSuspended, now)
	if err != nil || next.State != StateSuspended || next.Revision != 5 {
		t.Fatalf("transition: %+v %v", next, err)
	}
	if state.State != StateActive || state.Revision != 4 {
		t.Fatal("domain operation mutated the input")
	}
	if _, err := state.Transition(3, StateDeleted, now); !errors.Is(err, ErrTransitionConflict) {
		t.Fatalf("stale revision allowed: %v", err)
	}
	deleted, err := state.Transition(4, StateDeleted, now)
	if err != nil || deleted.DeletedAt == nil || !deleted.DeletedAt.Equal(now) {
		t.Fatalf("delete timestamp lost: %+v %v", deleted, err)
	}
	if _, err := deleted.Transition(5, StateActive, now); !errors.Is(err, ErrTransitionConflict) {
		t.Fatalf("deleted aggregate resurrected: %v", err)
	}
	for _, target := range []State{StateCreating, StateDeleting, StateFailed, "UNKNOWN"} {
		if _, err := state.Transition(4, target, now); !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("invalid target %q accepted: %v", target, err)
		}
	}
}
