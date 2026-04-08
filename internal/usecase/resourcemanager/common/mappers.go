package common

import (
	"github.com/m8platform/platform/internal/entities/resourcemanager/organization"
	"github.com/m8platform/platform/internal/entities/resourcemanager/project"
	"github.com/m8platform/platform/internal/entities/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

func OrganizationToBoundary(entity organization.Entity) boundaries.Organization {
	return boundaries.Organization{
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

func WorkspaceToBoundary(entity workspace.Entity) boundaries.Workspace {
	return boundaries.Workspace{
		ID:             entity.ID,
		OrganizationID: entity.OrganizationID,
		State:          string(entity.State),
		Name:           entity.Name,
		Description:    entity.Description,
		CreateTime:     entity.CreateTime,
		UpdateTime:     entity.UpdateTime,
		DeleteTime:     entity.DeleteTime,
		PurgeTime:      entity.PurgeTime,
		ETag:           entity.ETag.String(),
		Annotations:    cloneMap(entity.Annotations),
	}
}

func ProjectToBoundary(entity project.Entity) boundaries.Project {
	return boundaries.Project{
		ID:          entity.ID,
		WorkspaceID: entity.WorkspaceID,
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
	for k, v := range input {
		out[k] = v
	}
	return out
}
