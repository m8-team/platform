package workspace

type EventType string

const (
	EventWorkspaceCreated EventType = "resourcemanager.workspace.create"
	EventWorkspaceUpdated EventType = "resourcemanager.workspace.update"
	EventWorkspaceDeleted EventType = "resourcemanager.workspace.delete"
)
