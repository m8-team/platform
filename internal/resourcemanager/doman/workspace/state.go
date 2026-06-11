package workspace

type State string

const (
	StateCreating  State = "CREATING"
	StateActive    State = "ACTIVE"
	StateSuspended State = "SUSPENDED"
	StateDeleting  State = "DELETING"
	StateDeleted   State = "DELETED"
	StateFailed    State = "FAILED"
)
