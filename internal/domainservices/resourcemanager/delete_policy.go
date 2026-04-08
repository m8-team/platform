package domainservices

import "github.com/m8platform/platform/internal/entities/resourcemanager/hierarchy"

type DeletePolicy struct{}

func (DeletePolicy) EnsureAllowed(hasActiveChildren bool) error {
	return hierarchy.EnsureDeleteAllowed(hasActiveChildren)
}
