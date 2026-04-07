package grpctransport

import (
	"context"

	"github.com/m8platform/platform/internal/application/command"
	"github.com/m8platform/platform/internal/application/query"
	"github.com/m8platform/platform/internal/ports"
	"github.com/m8platform/platform/internal/transport/middleware"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

type WorkspaceServer struct {
	resourcemanagerv1.UnimplementedWorkspaceServiceServer

	Get      query.GetWorkspaceHandler
	List     query.ListWorkspacesHandler
	Create   command.CreateWorkspaceHandler
	Update   command.UpdateWorkspaceHandler
	Delete   command.DeleteWorkspaceHandler
	Undelete command.UndeleteWorkspaceHandler

	ActorResolver ports.ActorResolver
}

func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *resourcemanagerv1.GetWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	result, err := s.Get.Handle(ctx, query.GetWorkspace{ID: req.GetId()})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoWorkspace(result), nil
}

func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, req *resourcemanagerv1.ListWorkspacesRequest) (*resourcemanagerv1.ListWorkspacesResponse, error) {
	page, err := s.List.Handle(ctx, query.ListWorkspaces{
		OrganizationID: req.GetOrganizationId(),
		PageSize:       req.GetPageSize(),
		PageToken:      req.GetPageToken(),
		Filter:         req.GetFilter(),
		OrderBy:        req.GetOrderBy(),
		ShowDeleted:    req.GetShowDeleted(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]*resourcemanagerv1.Workspace, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toProtoWorkspace(item))
	}
	return &resourcemanagerv1.ListWorkspacesResponse{
		Workspaces:    items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}

func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, req *resourcemanagerv1.CreateWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	if req.GetWorkspace() == nil {
		return nil, mapError(commandErrorInvalidArgument("workspace is required"))
	}
	result, err := s.Create.Handle(ctx, command.CreateWorkspace{
		Metadata:       middleware.CommandMetadata(ctx, s.ActorResolver),
		OrganizationID: req.GetOrganizationId(),
		Name:           req.GetWorkspace().GetName(),
		Description:    req.GetWorkspace().GetDescription(),
		Annotations:    req.GetWorkspace().GetAnnotations(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoWorkspace(result), nil
}

func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, req *resourcemanagerv1.UpdateWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	if req.GetWorkspace() == nil {
		return nil, mapError(commandErrorInvalidArgument("workspace is required"))
	}
	result, err := s.Update.Handle(ctx, command.UpdateWorkspace{
		Metadata:       middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:             req.GetWorkspace().GetId(),
		OrganizationID: req.GetWorkspace().GetOrganizationId(),
		ETag:           req.GetWorkspace().GetEtag(),
		UpdateMask:     updateMaskPaths(req.GetUpdateMask()),
		Name:           stringPtr(req.GetWorkspace().GetName()),
		Description:    stringPtr(req.GetWorkspace().GetDescription()),
		Annotations:    req.GetWorkspace().GetAnnotations(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoWorkspace(result), nil
}

func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, req *resourcemanagerv1.DeleteWorkspaceRequest) (*emptypb.Empty, error) {
	if err := s.Delete.Handle(ctx, command.DeleteWorkspace{
		Metadata:     middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:           req.GetId(),
		ETag:         req.GetEtag(),
		AllowMissing: req.GetAllowMissing(),
	}); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *WorkspaceServer) UndeleteWorkspace(ctx context.Context, req *resourcemanagerv1.UndeleteWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	result, err := s.Undelete.Handle(ctx, command.UndeleteWorkspace{
		Metadata: middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:       req.GetId(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoWorkspace(result), nil
}
