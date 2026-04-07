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

type OrganizationServer struct {
	resourcemanagerv1.UnimplementedOrganizationServiceServer

	Get      query.GetOrganizationHandler
	List     query.ListOrganizationsHandler
	Create   command.CreateOrganizationHandler
	Update   command.UpdateOrganizationHandler
	Delete   command.DeleteOrganizationHandler
	Undelete command.UndeleteOrganizationHandler

	ActorResolver ports.ActorResolver
}

func (s *OrganizationServer) GetOrganization(ctx context.Context, req *resourcemanagerv1.GetOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	result, err := s.Get.Handle(ctx, query.GetOrganization{ID: req.GetId()})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoOrganization(result), nil
}

func (s *OrganizationServer) ListOrganizations(ctx context.Context, req *resourcemanagerv1.ListOrganizationsRequest) (*resourcemanagerv1.ListOrganizationsResponse, error) {
	page, err := s.List.Handle(ctx, query.ListOrganizations{
		PageSize:    req.GetPageSize(),
		PageToken:   req.GetPageToken(),
		Filter:      req.GetFilter(),
		OrderBy:     req.GetOrderBy(),
		ShowDeleted: req.GetShowDeleted(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]*resourcemanagerv1.Organization, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toProtoOrganization(item))
	}
	return &resourcemanagerv1.ListOrganizationsResponse{
		Organizations: items,
		NextPageToken: page.NextPageToken,
		TotalSize:     page.TotalSize,
	}, nil
}

func (s *OrganizationServer) CreateOrganization(ctx context.Context, req *resourcemanagerv1.CreateOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	if req.GetOrganization() == nil {
		return nil, mapError(commandErrorInvalidArgument("organization is required"))
	}
	result, err := s.Create.Handle(ctx, command.CreateOrganization{
		Metadata:    middleware.CommandMetadata(ctx, s.ActorResolver),
		Name:        req.GetOrganization().GetName(),
		Description: req.GetOrganization().GetDescription(),
		Annotations: req.GetOrganization().GetAnnotations(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoOrganization(result), nil
}

func (s *OrganizationServer) UpdateOrganization(ctx context.Context, req *resourcemanagerv1.UpdateOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	if req.GetOrganization() == nil {
		return nil, mapError(commandErrorInvalidArgument("organization is required"))
	}
	result, err := s.Update.Handle(ctx, command.UpdateOrganization{
		Metadata:    middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:          req.GetOrganization().GetId(),
		ETag:        req.GetOrganization().GetEtag(),
		UpdateMask:  updateMaskPaths(req.GetUpdateMask()),
		Name:        stringPtr(req.GetOrganization().GetName()),
		Description: stringPtr(req.GetOrganization().GetDescription()),
		Annotations: req.GetOrganization().GetAnnotations(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoOrganization(result), nil
}

func (s *OrganizationServer) DeleteOrganization(ctx context.Context, req *resourcemanagerv1.DeleteOrganizationRequest) (*emptypb.Empty, error) {
	if err := s.Delete.Handle(ctx, command.DeleteOrganization{
		Metadata:     middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:           req.GetId(),
		ETag:         req.GetEtag(),
		AllowMissing: req.GetAllowMissing(),
	}); err != nil {
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *OrganizationServer) UndeleteOrganization(ctx context.Context, req *resourcemanagerv1.UndeleteOrganizationRequest) (*resourcemanagerv1.Organization, error) {
	result, err := s.Undelete.Handle(ctx, command.UndeleteOrganization{
		Metadata: middleware.CommandMetadata(ctx, s.ActorResolver),
		ID:       req.GetId(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProtoOrganization(result), nil
}
