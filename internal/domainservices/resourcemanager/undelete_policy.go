package domainservices

import "github.com/m8platform/platform/internal/entity/resourcemanager/hierarchy"

type UndeletePolicy struct{}

func (UndeletePolicy) EnsureParentAllowsUndelete(exists bool, deleted bool) error {
	return hierarchy.EnsureUndeleteAllowed(exists, deleted)
}
