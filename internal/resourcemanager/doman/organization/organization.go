package organization

import "github.com/m8platform/platform/internal/platform/types"

const ResourceType = "resourcemanager.organization"

type Organization struct {
	id          types.ID
	state       State
	name        string
	description string
}

func NewOrganization(name, description string) *Organization {
	return NewOrganizationFrom(types.NewID(), StateCreating, name, description)
}

func NewOrganizationFrom(id types.ID, state State, name, description string) *Organization {
	return &Organization{
		id:          id,
		state:       state,
		name:        name,
		description: description,
	}
}

func (org *Organization) ID() types.ID {
	return org.id
}

func (org *Organization) State() State {
	return org.state
}

func (org *Organization) Name() string {
	return org.name
}

func (org *Organization) Description() string {
	return org.description
}

func (org *Organization) Update(name, description string) {
	org.name = name
	org.description = description
}
