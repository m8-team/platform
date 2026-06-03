package organization

import (
	"time"

	"github.com/m8platform/platform/internal/platform/types"
)

const ResourceType = "resourcemanager.organization"

type Organization struct {
	id          types.ID
	state       State
	name        string
	description string

	createdAt time.Time
	updatedAt time.Time
	deletedAt time.Time
	purgeAt   time.Time

	version types.Version
	labels  map[string]string
}
