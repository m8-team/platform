package workspace

import "time"

const (
	EventCreated      = "workspace.created"
	EventUpdated      = "workspace.updated"
	EventStateChanged = "workspace.state_changed"
	EventDeleted      = "workspace.deleted"
	EventUndeleted    = "workspace.undeleted"
	EventPurged       = "workspace.purged"
)

type Event struct {
	Type       string
	Entity     Entity
	OccurredAt time.Time
}
