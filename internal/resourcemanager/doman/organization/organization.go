package organization

import (
	"github.com/m8platform/platform/internal/platform/doman/types"
)

const ResourceType = "resourcemanager.organization"

type Organization struct {
	id          types.ID
	state       State
	name        string
	description string
}

func (o *Organization) Id() types.ID {
	return o.id
}

func (o *Organization) State() State {
	return o.state
}

func (o *Organization) Name() string {
	return o.name
}

func (o *Organization) Description() string {
	return o.description
}

func NewOrganization(id, state string, name, description string) *Organization {
	return &Organization{
		id:          types.NewID(),
		state:       StateCreating,
		name:        name,
		description: description,
	}
}

func NewOrganizationFrom(id, state string, name, description string) *Organization {
	return &Organization{
		state:       StateCreating,
		name:        name,
		description: description,
	}
}

func (o *Organization) Update(name, description string) {

}
