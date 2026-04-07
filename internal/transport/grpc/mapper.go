package grpctransport

import (
	"time"

	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/domain/workspace"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

func toProtoOrganization(value organization.Organization) *resourcemanagerv1.Organization {
	return &resourcemanagerv1.Organization{
		Id:          value.ID,
		State:       organizationStateToProto(value.State),
		Name:        value.Name,
		Description: value.Description,
		CreateTime:  toProtoTimestamp(value.CreateTime),
		UpdateTime:  toProtoTimestamp(value.UpdateTime),
		DeleteTime:  toProtoTimestampPtr(value.DeleteTime),
		PurgeTime:   toProtoTimestampPtr(value.PurgeTime),
		Etag:        value.ETag,
		Annotations: cloneMap(value.Annotations),
	}
}

func toProtoWorkspace(value workspace.Workspace) *resourcemanagerv1.Workspace {
	return &resourcemanagerv1.Workspace{
		Id:             value.ID,
		OrganizationId: value.OrganizationID,
		State:          workspaceStateToProto(value.State),
		Name:           value.Name,
		Description:    value.Description,
		CreateTime:     toProtoTimestamp(value.CreateTime),
		UpdateTime:     toProtoTimestamp(value.UpdateTime),
		DeleteTime:     toProtoTimestampPtr(value.DeleteTime),
		PurgeTime:      toProtoTimestampPtr(value.PurgeTime),
		Etag:           value.ETag,
		Annotations:    cloneMap(value.Annotations),
	}
}

func toProtoProject(value project.Project) *resourcemanagerv1.Project {
	return &resourcemanagerv1.Project{
		Id:          value.ID,
		WorkspaceId: value.WorkspaceID,
		State:       projectStateToProto(value.State),
		Name:        value.Name,
		Description: value.Description,
		CreateTime:  toProtoTimestamp(value.CreateTime),
		UpdateTime:  toProtoTimestamp(value.UpdateTime),
		DeleteTime:  toProtoTimestampPtr(value.DeleteTime),
		PurgeTime:   toProtoTimestampPtr(value.PurgeTime),
		Etag:        value.ETag,
		Annotations: cloneMap(value.Annotations),
	}
}

func organizationStateToProto(value organization.State) resourcemanagerv1.Organization_State {
	switch value {
	case organization.StateCreating:
		return resourcemanagerv1.Organization_CREATING
	case organization.StateActive:
		return resourcemanagerv1.Organization_ACTIVE
	case organization.StateSuspended:
		return resourcemanagerv1.Organization_SUSPENDED
	case organization.StateDeleting:
		return resourcemanagerv1.Organization_DELETING
	case organization.StateDeleted:
		return resourcemanagerv1.Organization_DELETED
	case organization.StateFailed:
		return resourcemanagerv1.Organization_FAILED
	default:
		return resourcemanagerv1.Organization_STATE_UNSPECIFIED
	}
}

func workspaceStateToProto(value workspace.State) resourcemanagerv1.Workspace_State {
	switch value {
	case workspace.StateCreating:
		return resourcemanagerv1.Workspace_CREATING
	case workspace.StateActive:
		return resourcemanagerv1.Workspace_ACTIVE
	case workspace.StateSuspended:
		return resourcemanagerv1.Workspace_SUSPENDED
	case workspace.StateDeleting:
		return resourcemanagerv1.Workspace_DELETING
	case workspace.StateDeleted:
		return resourcemanagerv1.Workspace_DELETED
	case workspace.StateFailed:
		return resourcemanagerv1.Workspace_FAILED
	default:
		return resourcemanagerv1.Workspace_STATE_UNSPECIFIED
	}
}

func projectStateToProto(value project.State) resourcemanagerv1.Project_State {
	switch value {
	case project.StateCreating:
		return resourcemanagerv1.Project_CREATING
	case project.StateActive:
		return resourcemanagerv1.Project_ACTIVE
	case project.StateArchived:
		return resourcemanagerv1.Project_ARCHIVED
	case project.StateDeleting:
		return resourcemanagerv1.Project_DELETING
	case project.StateDeleted:
		return resourcemanagerv1.Project_DELETED
	case project.StateFailed:
		return resourcemanagerv1.Project_FAILED
	default:
		return resourcemanagerv1.Project_STATE_UNSPECIFIED
	}
}

func updateMaskPaths(mask *fieldmaskpb.FieldMask) []string {
	if mask == nil {
		return nil
	}
	return append([]string(nil), mask.Paths...)
}

func toProtoTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func toProtoTimestampPtr(value *time.Time) *timestamppb.Timestamp {
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
