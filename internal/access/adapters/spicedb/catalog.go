package spicedb

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

// StaticPermissionCatalog is an immutable bootstrap registry. It is NOT a
// replacement for a published model/version registry in the Control Plane.
type StaticPermissionCatalog struct {
	definitions map[string]domain.PermissionDefinition
}

func catalogKey(resourceType, permission string) string { return resourceType + ":" + permission }

func NewStaticPermissionCatalog(definitions []domain.PermissionDefinition) (*StaticPermissionCatalog, error) {
	result := &StaticPermissionCatalog{definitions: make(map[string]domain.PermissionDefinition, len(definitions))}
	if len(definitions) == 0 {
		return nil, ErrInvalidConfig
	}
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return nil, err
		}
		if !graphName.MatchString(definition.ResourceType()) || !graphName.MatchString(definition.GraphPermission()) {
			return nil, ErrInvalidGraphReference
		}
		key := catalogKey(definition.ResourceType(), definition.Permission().Name())
		if _, exists := result.definitions[key]; exists {
			return nil, fmt.Errorf("%w: duplicated catalog entry", ErrInvalidConfig)
		}
		result.definitions[key] = definition
	}
	return result, nil
}

func (c *StaticPermissionCatalog) ResolvePermission(ctx context.Context, tenant domain.TenantID, permission domain.Permission, resourceType string) (domain.PermissionDefinition, error) {
	if err := ctx.Err(); err != nil {
		return domain.PermissionDefinition{}, err
	}
	if err := tenant.Validate(); err != nil {
		return domain.PermissionDefinition{}, err
	}
	if err := permission.Validate(); err != nil {
		return domain.PermissionDefinition{}, err
	}
	if c == nil {
		return domain.PermissionDefinition{}, ports.ErrPermissionNotRegistered
	}
	definition, exists := c.definitions[catalogKey(resourceType, permission.Name())]
	if !exists {
		return domain.PermissionDefinition{}, ports.ErrPermissionNotRegistered
	}
	return definition, nil
}
