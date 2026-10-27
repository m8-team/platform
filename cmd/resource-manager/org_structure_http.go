package main

import (
	"context"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	grpcadapter "github.com/m8-team/platform/internal/resourcemanager/adapter/grpc"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// localOrgStructureHTTPServer is ONLY attached in the explicit, loopback-bound
// unauthenticated development mode. It derives the Tenant from the backend's
// Organization resource, not from an HTTP tenant header or user-provided claim.
// Production deliberately receives no synthetic actor and remains fail-closed
// until a real authenticated M8 Access gateway is wired.
type localOrgStructureHTTPServer struct {
	pb.OrgStructureServiceServer
	pb.OrgUnitChangeSetServiceServer
	resources pb.ResourceManagerServiceServer
}

func (s *localOrgStructureHTTPServer) actorContext(ctx context.Context, organizationID string) (context.Context, error) {
	if organizationID == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}
	record, err := s.resources.GetResource(ctx, &pb.GetResourceRequest{
		Kind: "organization", Id: organizationID,
	})
	if err != nil {
		return nil, err
	}
	organization := record.GetOrganization()
	if organization == nil || organization.GetMetadata().GetId() != organizationID || organization.GetTenantId() == "" {
		return nil, status.Error(codes.NotFound, "organization not found")
	}
	return grpcadapter.WithOrgStructureActor(ctx, orgstructure.Actor{
		TenantID: organization.GetTenantId(), PrincipalID: "local-http-development",
	}), nil
}

func (s *localOrgStructureHTTPServer) GetOrgUnit(ctx context.Context, req *pb.GetOrgUnitRequest) (*pb.OrgUnit, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	scoped, err := s.actorContext(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgStructureServiceServer.GetOrgUnit(scoped, req)
}

func (s *localOrgStructureHTTPServer) ListOrgUnits(ctx context.Context, req *pb.ListOrgUnitsRequest) (*pb.ListOrgUnitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	scoped, err := s.actorContext(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgStructureServiceServer.ListOrgUnits(scoped, req)
}

func (s *localOrgStructureHTTPServer) ListOrgUnitAncestors(ctx context.Context, req *pb.ListOrgUnitAncestorsRequest) (*pb.ListOrgUnitAncestorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	scoped, err := s.actorContext(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgStructureServiceServer.ListOrgUnitAncestors(scoped, req)
}

func (s *localOrgStructureHTTPServer) ListOrgUnitDescendants(ctx context.Context, req *pb.ListOrgUnitDescendantsRequest) (*pb.ListOrgUnitDescendantsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	scoped, err := s.actorContext(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgStructureServiceServer.ListOrgUnitDescendants(scoped, req)
}

func (s *localOrgStructureHTTPServer) ListOrgUnitVersions(ctx context.Context, req *pb.ListOrgUnitVersionsRequest) (*pb.ListOrgUnitVersionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	scoped, err := s.actorContext(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgStructureServiceServer.ListOrgUnitVersions(scoped, req)
}

func (s *localOrgStructureHTTPServer) ValidateOrgUnitChangeSet(ctx context.Context, req *pb.ValidateOrgUnitChangeSetRequest) (*pb.ValidateOrgUnitChangeSetResponse, error) {
	if req == nil || req.GetChangeSet() == nil {
		return nil, status.Error(codes.InvalidArgument, "change_set is required")
	}
	scoped, err := s.actorContext(ctx, req.GetChangeSet().GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgUnitChangeSetServiceServer.ValidateOrgUnitChangeSet(scoped, req)
}

func (s *localOrgStructureHTTPServer) ApplyOrgUnitChangeSet(ctx context.Context, req *pb.ApplyOrgUnitChangeSetRequest) (*pb.ApplyOrgUnitChangeSetResponse, error) {
	if req == nil || req.GetChangeSet() == nil {
		return nil, status.Error(codes.InvalidArgument, "change_set is required")
	}
	scoped, err := s.actorContext(ctx, req.GetChangeSet().GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return s.OrgUnitChangeSetServiceServer.ApplyOrgUnitChangeSet(scoped, req)
}

var _ pb.OrgStructureServiceServer = (*localOrgStructureHTTPServer)(nil)
var _ pb.OrgUnitChangeSetServiceServer = (*localOrgStructureHTTPServer)(nil)
