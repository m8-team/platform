package grpcpresenter

import (
	"time"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

func timestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func timestampPtr(value *time.Time) *timestamppb.Timestamp {
	if value == nil || value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
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

func organizationState(value string) resourcemanagerv1.Organization_State {
	switch value {
	case "CREATING":
		return resourcemanagerv1.Organization_CREATING
	case "ACTIVE":
		return resourcemanagerv1.Organization_ACTIVE
	case "SUSPENDED":
		return resourcemanagerv1.Organization_SUSPENDED
	case "DELETING":
		return resourcemanagerv1.Organization_DELETING
	case "DELETED":
		return resourcemanagerv1.Organization_DELETED
	case "FAILED":
		return resourcemanagerv1.Organization_FAILED
	default:
		return resourcemanagerv1.Organization_STATE_UNSPECIFIED
	}
}

func workspaceState(value string) resourcemanagerv1.Workspace_State {
	switch value {
	case "CREATING":
		return resourcemanagerv1.Workspace_CREATING
	case "ACTIVE":
		return resourcemanagerv1.Workspace_ACTIVE
	case "SUSPENDED":
		return resourcemanagerv1.Workspace_SUSPENDED
	case "DELETING":
		return resourcemanagerv1.Workspace_DELETING
	case "DELETED":
		return resourcemanagerv1.Workspace_DELETED
	case "FAILED":
		return resourcemanagerv1.Workspace_FAILED
	default:
		return resourcemanagerv1.Workspace_STATE_UNSPECIFIED
	}
}

func projectState(value string) resourcemanagerv1.Project_State {
	switch value {
	case "CREATING":
		return resourcemanagerv1.Project_CREATING
	case "ACTIVE":
		return resourcemanagerv1.Project_ACTIVE
	case "ARCHIVED":
		return resourcemanagerv1.Project_ARCHIVED
	case "DELETING":
		return resourcemanagerv1.Project_DELETING
	case "DELETED":
		return resourcemanagerv1.Project_DELETED
	case "FAILED":
		return resourcemanagerv1.Project_FAILED
	default:
		return resourcemanagerv1.Project_STATE_UNSPECIFIED
	}
}

func mapOrganization(value organizationboundary.Organization) *resourcemanagerv1.Organization {
	return &resourcemanagerv1.Organization{
		Id:          value.ID,
		State:       organizationState(value.State),
		Name:        value.Name,
		Description: value.Description,
		CreateTime:  timestamp(value.CreateTime),
		UpdateTime:  timestamp(value.UpdateTime),
		DeleteTime:  timestampPtr(value.DeleteTime),
		PurgeTime:   timestampPtr(value.PurgeTime),
		Etag:        value.ETag,
		Annotations: cloneMap(value.Annotations),
	}
}

func mapWorkspace(value boundary.Workspace) *resourcemanagerv1.Workspace {
	return &resourcemanagerv1.Workspace{
		Id:             value.ID,
		OrganizationId: value.OrganizationID,
		State:          workspaceState(value.State),
		Name:           value.Name,
		Description:    value.Description,
		CreateTime:     timestamp(value.CreateTime),
		UpdateTime:     timestamp(value.UpdateTime),
		DeleteTime:     timestampPtr(value.DeleteTime),
		PurgeTime:      timestampPtr(value.PurgeTime),
		Etag:           value.ETag,
		Annotations:    cloneMap(value.Annotations),
	}
}

func mapProject(value boundary.Project) *resourcemanagerv1.Project {
	return &resourcemanagerv1.Project{
		Id:          value.ID,
		WorkspaceId: value.WorkspaceID,
		State:       projectState(value.State),
		Name:        value.Name,
		Description: value.Description,
		CreateTime:  timestamp(value.CreateTime),
		UpdateTime:  timestamp(value.UpdateTime),
		DeleteTime:  timestampPtr(value.DeleteTime),
		PurgeTime:   timestampPtr(value.PurgeTime),
		Etag:        value.ETag,
		Annotations: cloneMap(value.Annotations),
	}
}
