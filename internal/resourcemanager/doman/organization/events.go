package organization

type EventType string

const (
	EventOrganizationCreated EventType = "resource.organization.create"
	EventOrganizationUpdated EventType = "resource.organization.update"
	EventOrganizationDeleted EventType = "resource.organization.delete"
)
