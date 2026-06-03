package organization

import (
	"github.com/m8platform/platform/internal/platform/types"
)

const ResourceType = "resourcemanager.organization"

type Organization struct {
	id          types.ID
	state       State
	name        string
	description string

	version types.Version
}
