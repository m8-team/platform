package domainservices

import "github.com/m8platform/platform/internal/entity/resourcemanager/hierarchy"

type DeletePolicy struct{}

func (DeletePolicy) EnsureAllowed(hasActiveChildren bool) error {
	return hierarchy.EnsureDeleteAllowed(hasActiveChildren)
}
