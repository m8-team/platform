package organizationmapper

import (
	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

func ToBoundary(entity organizationentity.Entity) organizationboundary.Organization {
	return organizationboundary.Organization{
		ID:          entity.ID,
		State:       string(entity.State),
		Name:        entity.Name,
		Description: entity.Description,
		CreateTime:  entity.CreateTime,
		UpdateTime:  entity.UpdateTime,
		DeleteTime:  entity.DeleteTime,
		PurgeTime:   entity.PurgeTime,
		ETag:        entity.ETag.String(),
		Annotations: cloneMap(entity.Annotations),
	}
}

func cloneMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
