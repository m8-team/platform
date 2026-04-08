package project

type State string

const (
	StateUnspecified State = "STATE_UNSPECIFIED"
	StateCreating    State = "CREATING"
	StateActive      State = "ACTIVE"
	StateArchived    State = "ARCHIVED"
	StateDeleting    State = "DELETING"
	StateDeleted     State = "DELETED"
	StateFailed      State = "FAILED"
)
