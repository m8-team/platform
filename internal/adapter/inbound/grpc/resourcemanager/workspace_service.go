package grpcadapter

import (
	"context"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	grpcpresenter "github.com/m8platform/platform/internal/adapter/presenters/grpc/resourcemanager"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	"google.golang.org/protobuf/types/known/emptypb"
)

type WorkspaceServiceServer struct {
	resourcemanagerv1.UnimplementedWorkspaceServiceServer
	Commands  boundary.WorkspaceCommandUseCase
	Queries   boundary.WorkspaceQueryUseCase
	Presenter grpcpresenter.WorkspacePresenter
}

func (s WorkspaceServiceServer) GetWorkspace(ctx context.Context, req *resourcemanagerv1.GetWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	output, err := s.Queries.GetWorkspace(ctx, boundary.GetWorkspaceInput{ID: req.GetId()})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentGet(output), nil
}

func (s WorkspaceServiceServer) ListWorkspaces(ctx context.Context, req *resourcemanagerv1.ListWorkspacesRequest) (*resourcemanagerv1.ListWorkspacesResponse, error) {
	output, err := s.Queries.ListWorkspaces(ctx, boundary.ListWorkspacesInput{
		OrganizationID: req.GetOrganizationId(),
		PageSize:       req.GetPageSize(),
		PageToken:      req.GetPageToken(),
		Filter:         req.GetFilter(),
		OrderBy:        req.GetOrderBy(),
		ShowDeleted:    req.GetShowDeleted(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentList(output), nil
}

func (s WorkspaceServiceServer) CreateWorkspace(ctx context.Context, req *resourcemanagerv1.CreateWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	resource := req.GetWorkspace()
	output, err := s.Commands.CreateWorkspace(ctx, boundary.CreateWorkspaceInput{
		Metadata:       requestMetadata(ctx),
		OrganizationID: req.GetOrganizationId(),
		Name:           resource.GetName(),
		Description:    resource.GetDescription(),
		Annotations:    cloneMap(resource.GetAnnotations()),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentCreate(output), nil
}

func (s WorkspaceServiceServer) UpdateWorkspace(ctx context.Context, req *resourcemanagerv1.UpdateWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	resource := req.GetWorkspace()
	mask := fieldMaskPathsWorkspace(req)
	output, err := s.Commands.UpdateWorkspace(ctx, boundary.UpdateWorkspaceInput{
		Metadata:       requestMetadata(ctx),
		ID:             resource.GetId(),
		OrganizationID: resource.GetOrganizationId(),
		ETag:           resource.GetEtag(),
		UpdateMask:     mask,
		Name:           optionalString(mask, "name", resource.GetName()),
		Description:    optionalString(mask, "description", resource.GetDescription()),
		Annotations:    cloneMap(resource.GetAnnotations()),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUpdate(output), nil
}

func (s WorkspaceServiceServer) DeleteWorkspace(ctx context.Context, req *resourcemanagerv1.DeleteWorkspaceRequest) (*emptypb.Empty, error) {
	_, err := s.Commands.DeleteWorkspace(ctx, boundary.DeleteWorkspaceInput{
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

func (s WorkspaceServiceServer) UndeleteWorkspace(ctx context.Context, req *resourcemanagerv1.UndeleteWorkspaceRequest) (*resourcemanagerv1.Workspace, error) {
	output, err := s.Commands.UndeleteWorkspace(ctx, boundary.UndeleteWorkspaceInput{
		Metadata: requestMetadata(ctx),
		ID:       req.GetId(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUndelete(output), nil
}
