package project

type EventType string

const (
	EventProjectCreated EventType = "resourcemanager.project.create"
	EventProjectUpdated EventType = "resourcemanager.project.update"
	EventProjectDeleted EventType = "resourcemanager.project.delete"
)
