package resource

import (
	"errors"
	"time"
)

var ErrInvalidTransition = errors.New("invalid resource lifecycle transition")
var ErrTransitionConflict = errors.New("resource lifecycle revision conflict")

// Lifecycle is the state and optimistic revision of a Resource Manager-owned
// aggregate. Adapters control persistence but never redefine transition rules.
type Lifecycle struct {
	State     State
	Revision  int64
	DeletedAt *time.Time
}

func (s State) MutableTarget() bool {
	switch s {
	case StateActive, StateSuspended, StateArchived, StateDeleted:
		return true
	default:
		return false
	}
}

// Transition is pure and deterministic. The caller owns the authoritative
// transaction, child preconditions, expected-owner fencing and outbox write.
// It is intentionally not a replacement for a conditional database update.
func (l Lifecycle) Transition(expectedRevision int64, target State, at time.Time) (Lifecycle, error) {
	if !l.State.Valid() || l.Revision < 1 || expectedRevision < 1 ||
		!target.MutableTarget() || at.IsZero() {
		return Lifecycle{}, ErrInvalidTransition
	}
	if l.Revision != expectedRevision || l.State == StateDeleted {
		return Lifecycle{}, ErrTransitionConflict
	}
	next := l
	next.State = target
	next.Revision++
	if target == StateDeleted {
		deletedAt := at.UTC()
		next.DeletedAt = &deletedAt
	}
	return next, nil
}
