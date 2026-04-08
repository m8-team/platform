package domainservices

import "github.com/m8platform/platform/internal/entities/resourcemanager/hierarchy"

type HierarchyPolicy struct{}

func (HierarchyPolicy) EnsureParentActive(exists bool, deleted bool) error {
	return hierarchy.EnsureParentActive(exists, deleted)
}
