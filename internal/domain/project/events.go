package project

import "time"

const (
	EventCreated      = "project.created"
	EventUpdated      = "project.updated"
	EventStateChanged = "project.state_changed"
	EventArchived     = "project.archived"
	EventUnarchived   = "project.unarchived"
	EventDeleted      = "project.deleted"
	EventUndeleted    = "project.undeleted"
	EventPurged       = "project.purged"
)

type Event struct {
	Type       string
	Aggregate  Project
	OccurredAt time.Time
}
