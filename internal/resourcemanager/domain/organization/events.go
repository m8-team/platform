package organization

type EventType string

const (
	EventOrganizationCreated   EventType = "m8.resourcemanager.organization.created.v1"
	EventOrganizationUpdated   EventType = "m8.resourcemanager.organization.updated.v1"
	EventOrganizationDeleted   EventType = "m8.resourcemanager.organization.deleted.v1"
	EventOrganizationUndeleted EventType = "m8.resourcemanager.organization.undeleted.v1"
)
