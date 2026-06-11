package organization

import "github.com/m8platform/platform/internal/platform/types"

//go:generate go tool stringer -type=State -linecomment -output=state_string.go

type State int8

const (
	StateUnspecified State = iota // STATE_UNSPECIFIED
	StateCreating                 // CREATING
	StateActive                   // ACTIVE
	StateSuspended                // SUSPENDED
	StateDeleting                 // DELETING
	StateDeleted                  // DELETED
	StateFailed                   // FAILED
)

var _ types.State = StateUnspecified

func (s State) IsValid() bool {
	switch s {
	case StateUnspecified,
		StateCreating,
		StateActive,
		StateSuspended,
		StateDeleting,
		StateDeleted,
		StateFailed:
		return true
	default:
		return false
	}
}
