package grpcadapter

import (
	"context"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	grpcpresenter "github.com/m8platform/platform/internal/adapter/presenters/grpc/resourcemanager"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrganizationServiceServer struct {
	resourcemanagerv1.UnimplementedOrganizationServiceServer
	Commands  organizationboundary.OrganizationCommandUseCase
	Queries   organizationboundary.OrganizationQueryUseCase
	Presenter grpcpresenter.OrganizationPresenter
}

func (s OrganizationServiceServer) GetOrganization(ctx context.Context, req *resourcemanagerv1.GetOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	output, err := s.Queries.GetOrganization(ctx, organizationboundary.GetOrganizationInput{
		ID:             req.GetId(),
		IncludeDeleted: true,
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentGet(output), nil
}

func (s OrganizationServiceServer) ListOrganizations(ctx context.Context, req *resourcemanagerv1.ListOrganizationsRequest) (*resourcemanagerv1.ListOrganizationsResponse, error) {
	output, err := s.Queries.ListOrganizations(ctx, organizationboundary.ListOrganizationsInput{
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
	output, err := s.Commands.CreateOrganization(ctx, organizationboundary.CreateOrganizationInput{
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
	output, err := s.Commands.UpdateOrganization(ctx, organizationboundary.UpdateOrganizationInput{
		ID:          resource.GetId(),
		ETag:        resource.GetEtag(),
		UpdateMask:  mask,
		Name:        optionalString(mask, "name", resource.GetName()),
		Description: optionalString(mask, "description", resource.GetDescription()),
		Annotations: optionalMap(mask, "annotations", resource.GetAnnotations()),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUpdate(output), nil
}

func (s OrganizationServiceServer) DeleteOrganization(ctx context.Context, req *resourcemanagerv1.DeleteOrganizationRequest) (*emptypb.Empty, error) {
	_, err := s.Commands.DeleteOrganization(ctx, organizationboundary.DeleteOrganizationInput{
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
	output, err := s.Commands.UndeleteOrganization(ctx, organizationboundary.UndeleteOrganizationInput{
		ID: req.GetId(),
	})
	if err != nil {
		return nil, grpcpresenter.PresentError(err)
	}
	return s.Presenter.PresentUndelete(output), nil
}
