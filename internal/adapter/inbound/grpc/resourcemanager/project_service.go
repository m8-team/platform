package grpcadapter

import (
	"context"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	grpcpresenter "github.com/m8platform/platform/internal/adapter/presenters/grpc/resourcemanager"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProjectServiceServer struct {
	resourcemanagerv1.UnimplementedProjectServiceServer
	Commands  boundary.ProjectCommandUseCase
	Queries   boundary.ProjectQueryUseCase
	Presenter grpcpresenter.ProjectPresenter
}

func (s ProjectServiceServer) GetProject(ctx context.Context, req *resourcemanagerv1.GetProjectRequest) (*resourcemanagerv1.Project, error) {
	output, err := s.Queries.GetProject(ctx, boundary.GetProjectInput{ID: req.GetId()})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentGet(output), nil
}

func (s ProjectServiceServer) ListProjects(ctx context.Context, req *resourcemanagerv1.ListProjectsRequest) (*resourcemanagerv1.ListProjectsResponse, error) {
	output, err := s.Queries.ListProjects(ctx, boundary.ListProjectsInput{
		WorkspaceID: req.GetWorkspaceId(),
		PageSize:    req.GetPageSize(),
		PageToken:   req.GetPageToken(),
		Filter:      req.GetFilter(),
		OrderBy:     req.GetOrderBy(),
		ShowDeleted: req.GetShowDeleted(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentList(output), nil
}

func (s ProjectServiceServer) CreateProject(ctx context.Context, req *resourcemanagerv1.CreateProjectRequest) (*resourcemanagerv1.Project, error) {
	resource := req.GetProject()
	output, err := s.Commands.CreateProject(ctx, boundary.CreateProjectInput{
		Metadata:    requestMetadata(ctx),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        resource.GetName(),
		Description: resource.GetDescription(),
		Annotations: cloneMap(resource.GetAnnotations()),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentCreate(output), nil
}

func (s ProjectServiceServer) UpdateProject(ctx context.Context, req *resourcemanagerv1.UpdateProjectRequest) (*resourcemanagerv1.Project, error) {
	resource := req.GetProject()
	mask := fieldMaskPathsProject(req)
	output, err := s.Commands.UpdateProject(ctx, boundary.UpdateProjectInput{
		Metadata:    requestMetadata(ctx),
		ID:          resource.GetId(),
		WorkspaceID: resource.GetWorkspaceId(),
		ETag:        resource.GetEtag(),
		UpdateMask:  mask,
		Name:        optionalString(mask, "name", resource.GetName()),
		Description: optionalString(mask, "description", resource.GetDescription()),
		Annotations: cloneMap(resource.GetAnnotations()),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUpdate(output), nil
}

func (s ProjectServiceServer) DeleteProject(ctx context.Context, req *resourcemanagerv1.DeleteProjectRequest) (*emptypb.Empty, error) {
	_, err := s.Commands.DeleteProject(ctx, boundary.DeleteProjectInput{
		Metadata:     requestMetadata(ctx),
		ID:           req.GetId(),
		ETag:         req.GetEtag(),
		AllowMissing: req.GetAllowMissing(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s ProjectServiceServer) UndeleteProject(ctx context.Context, req *resourcemanagerv1.UndeleteProjectRequest) (*resourcemanagerv1.Project, error) {
	output, err := s.Commands.UndeleteProject(ctx, boundary.UndeleteProjectInput{
		Metadata: requestMetadata(ctx),
		ID:       req.GetId(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUndelete(output), nil
}
