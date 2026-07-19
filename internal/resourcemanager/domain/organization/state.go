package organization

import (
	"fmt"

	"github.com/m8-team/platform/internal/platform/types"
)

type State int8

const (
	StateUnspecified State = iota
	StateCreating
	StateActive
	StateSuspended
	StateDeleting
	StateDeleted
	StateFailed
)

var _ types.State = StateUnspecified

// IsValid reports whether the state is valid for a persisted Organization.
func (s State) IsValid() bool {
	switch s {
	case StateCreating, StateActive, StateSuspended, StateDeleting, StateDeleted, StateFailed:
		return true
	default:
		return false
	}
}

func (s State) String() string {
	switch s {
	case StateUnspecified:
		return "STATE_UNSPECIFIED"
	case StateCreating:
		return "CREATING"
	case StateActive:
		return "ACTIVE"
	case StateSuspended:
		return "SUSPENDED"
	case StateDeleting:
		return "DELETING"
	case StateDeleted:
		return "DELETED"
	case StateFailed:
		return "FAILED"
	default:
		return fmt.Sprintf("State(%d)", s)
	}
}
