package organization

type State string

const (
	StateUnspecified State = "STATE_UNSPECIFIED"
	StateCreating    State = "CREATING"
	StateActive      State = "ACTIVE"
	StateSuspended   State = "SUSPENDED"
	StateDeleting    State = "DELETING"
	StateDeleted     State = "DELETED"
	StateFailed      State = "FAILED"
)

func (s State) IsZero() bool {
	return s == "" || s == StateUnspecified
}
