package organization

type EventType string

const (
	EventOrganizationCreated EventType = "resourcemanager.organization.create"
	EventOrganizationUpdated EventType = "resourcemanager.organization.update"
	EventOrganizationDeleted EventType = "resourcemanager.organization.delete"
)
