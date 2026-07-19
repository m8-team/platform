package command

import (
	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

// CreateOrganization contains client-controlled fields only. Identity,
// lifecycle state, timestamps, and version are assigned by the application.
type CreateOrganization struct {
	Name        string
	Description string
	Labels      map[string]string
}

// UpdateOrganization is a partial update. Nil pointers mean that a field was
// not selected by the update mask; a pointer to a zero value clears it.
type UpdateOrganization struct {
	ID              organization.ID
	ExpectedVersion types.Version
	Name            *string
	Description     *string
	Labels          *map[string]string
}

type DeleteOrganization struct {
	ID              organization.ID
	ExpectedVersion types.Version
	AllowMissing    bool
}

type UndeleteOrganization struct {
	ID organization.ID
}
