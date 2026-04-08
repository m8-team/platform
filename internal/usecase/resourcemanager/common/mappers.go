package common

import (
	"github.com/m8platform/platform/internal/entity/resourcemanager/project"
	"github.com/m8platform/platform/internal/entity/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
)

func WorkspaceToBoundary(entity workspace.Entity) boundary.Workspace {
	return boundary.Workspace{
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

func ProjectToBoundary(entity project.Entity) boundary.Project {
	return boundary.Project{
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
