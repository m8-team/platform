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

type ProjectServer struct {
	resourcemanagerv1.UnimplementedProjectServiceServer

	Get      query.GetProjectHandler
	List     query.ListProjectsHandler
	Create   command.CreateProjectHandler
	Update   command.UpdateProjectHandler
	Delete   command.DeleteProjectHandler
	Undelete command.UndeleteProjectHandler

	ActorResolver ports.ActorResolver
}

func (s *ProjectServer) GetProject(ctx context.Context, req *resourcemanagerv1.GetProjectRequest) (*resourcemanagerv1.Project, error) {
	result, err := s.Get.Handle(ctx, query.GetProject{ID: req.GetId()})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoProject(result), nil
}

func (s *ProjectServer) ListProjects(ctx context.Context, req *resourcemanagerv1.ListProjectsRequest) (*resourcemanagerv1.ListProjectsResponse, error) {
	page, err := s.List.Handle(ctx, query.ListProjects{
		WorkspaceID: req.GetWorkspaceId(),
		PageSize:    req.GetPageSize(),
		PageToken:   req.GetPageToken(),
		Filter:      req.GetFilter(),
		OrderBy:     req.GetOrderBy(),
		ShowDeleted: req.GetShowDeleted(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]*resourcemanagerv1.Project, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toProtoProject(item))
	}
	return &resourcemanagerv1.ListProjectsResponse{
		Projects:      items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}

func (s *ProjectServer) CreateProject(ctx context.Context, req *resourcemanagerv1.CreateProjectRequest) (*resourcemanagerv1.Project, error) {
	if req.GetProject() == nil {
		return nil, mapError(commandErrorInvalidArgument("project is required"))
	}
	result, err := s.Create.Handle(ctx, command.CreateProject{
		Metadata:    middleware.CommandMetadata(ctx, s.ActorResolver),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetProject().GetName(),
		Description: req.GetProject().GetDescription(),
		Annotations: req.GetProject().GetAnnotations(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoProject(result), nil
}

func (s *ProjectServer) UpdateProject(ctx context.Context, req *resourcemanagerv1.UpdateProjectRequest) (*resourcemanagerv1.Project, error) {
	if req.GetProject() == nil {
		return nil, mapError(commandErrorInvalidArgument("project is required"))
	}
	result, err := s.Update.Handle(ctx, command.UpdateProject{
		Metadata:    middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:          req.GetProject().GetId(),
		WorkspaceID: req.GetProject().GetWorkspaceId(),
		ETag:        req.GetProject().GetEtag(),
		UpdateMask:  updateMaskPaths(req.GetUpdateMask()),
		Name:        stringPtr(req.GetProject().GetName()),
		Description: stringPtr(req.GetProject().GetDescription()),
		Annotations: req.GetProject().GetAnnotations(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoProject(result), nil
}

func (s *ProjectServer) DeleteProject(ctx context.Context, req *resourcemanagerv1.DeleteProjectRequest) (*emptypb.Empty, error) {
	if err := s.Delete.Handle(ctx, command.DeleteProject{
		Metadata:     middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:           req.GetId(),
		ETag:         req.GetEtag(),
		AllowMissing: req.GetAllowMissing(),
	}); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ProjectServer) UndeleteProject(ctx context.Context, req *resourcemanagerv1.UndeleteProjectRequest) (*resourcemanagerv1.Project, error) {
	result, err := s.Undelete.Handle(ctx, command.UndeleteProject{
		Metadata: middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:       req.GetId(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoProject(result), nil
}
