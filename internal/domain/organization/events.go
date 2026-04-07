package organization

import "time"

const (
	EventCreated      = "organization.created"
	EventUpdated      = "organization.updated"
	EventStateChanged = "organization.state_changed"
	EventDeleted      = "organization.deleted"
	EventUndeleted    = "organization.undeleted"
	EventPurged       = "organization.purged"
)

type Event struct {
	Type       string
	Aggregate  Organization
	OccurredAt time.Time
}
