package grpcadapter

import (
	"context"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	grpcpresenter "github.com/m8platform/platform/internal/adapters/presenters/grpc/resourcemanager"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrganizationServiceServer struct {
	resourcemanagerv1.UnimplementedOrganizationServiceServer
	Commands  boundaries.OrganizationCommandUseCase
	Queries   boundaries.OrganizationQueryUseCase
	Presenter grpcpresenter.OrganizationPresenter
}

func (s OrganizationServiceServer) GetOrganization(ctx context.Context, req *resourcemanagerv1.GetOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	output, err := s.Queries.GetOrganization(ctx, boundaries.GetOrganizationInput{ID: req.GetId()})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentGet(output), nil
}

func (s OrganizationServiceServer) ListOrganizations(ctx context.Context, req *resourcemanagerv1.ListOrganizationsRequest) (*resourcemanagerv1.ListOrganizationsResponse, error) {
	output, err := s.Queries.ListOrganizations(ctx, boundaries.ListOrganizationsInput{
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

func (s OrganizationServiceServer) CreateOrganization(ctx context.Context, req *resourcemanagerv1.CreateOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	resource := req.GetOrganization()
	output, err := s.Commands.CreateOrganization(ctx, boundaries.CreateOrganizationInput{
		Metadata:    requestMetadata(ctx),
		Name:        resource.GetName(),
		Description: resource.GetDescription(),
		Annotations: cloneMap(resource.GetAnnotations()),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentCreate(output), nil
}

func (s OrganizationServiceServer) UpdateOrganization(ctx context.Context, req *resourcemanagerv1.UpdateOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	resource := req.GetOrganization()
	mask := fieldMaskPaths(req)
	output, err := s.Commands.UpdateOrganization(ctx, boundaries.UpdateOrganizationInput{
		Metadata:    requestMetadata(ctx),
		ID:          resource.GetId(),
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

func (s OrganizationServiceServer) DeleteOrganization(ctx context.Context, req *resourcemanagerv1.DeleteOrganizationRequest) (*emptypb.Empty, error) {
	_, err := s.Commands.DeleteOrganization(ctx, boundaries.DeleteOrganizationInput{
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

func (s OrganizationServiceServer) UndeleteOrganization(ctx context.Context, req *resourcemanagerv1.UndeleteOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	output, err := s.Commands.UndeleteOrganization(ctx, boundaries.UndeleteOrganizationInput{
		Metadata: requestMetadata(ctx),
		ID:       req.GetId(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUndelete(output), nil
}
