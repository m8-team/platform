package workspace

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
