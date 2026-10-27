package grpcadapter

import (
	"context"
	"errors"
	"strings"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ResourceServer is the canonical v1 gRPC adapter. Currently backed by a
// process-local development store. Do not enable anonymous access in production.
type ResourceServer struct {
	pb.UnimplementedResourceManagerServiceServer
	store     ResourceBackend
	durable   bool
	devMode   bool
	bootstrap orgstructure.BootstrapRepository
}

func NewResourceServer(store *v1engine.Store, cfg resourcemanager.Config, bootstrap ...orgstructure.BootstrapRepository) *ResourceServer {
	var init orgstructure.BootstrapRepository
	if len(bootstrap) > 0 {
		init = bootstrap[0]
	}
	return &ResourceServer{store: memoryResourceBackend{store: store}, devMode: cfg.AllowUnauthenticated, bootstrap: init}
}
func NewDurableResourceServer(store ResourceBackend, cfg resourcemanager.Config) *ResourceServer {
	return &ResourceServer{store: store, devMode: cfg.AllowUnauthenticated, durable: true}
}
func (s *ResourceServer) guard() error {
	if !s.devMode {
		return status.Error(codes.PermissionDenied, "M8 Access authorization backend not configured")
	}
	return nil
}
func asErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, v1engine.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, v1engine.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, v1engine.ErrConflict):
		return status.Error(codes.Aborted, err.Error())
	default:
		return status.Error(codes.Internal, "resource manager internal error")
	}
}
func owner(in *pb.OwnerScope) v1engine.Owner {
	if in == nil {
		return v1engine.Owner{}
	}
	switch in.GetOwner().(type) {
	case *pb.OwnerScope_OrganizationId:
		return v1engine.Owner{Kind: v1engine.Organization, ID: in.GetOrganizationId()}
	case *pb.OwnerScope_WorkspaceId:
		return v1engine.Owner{Kind: v1engine.Workspace, ID: in.GetWorkspaceId()}
	case *pb.OwnerScope_ProjectId:
		return v1engine.Owner{Kind: v1engine.Project, ID: in.GetProjectId()}
	}
	return v1engine.Owner{}
}
func ownerProto(o v1engine.Owner) *pb.OwnerScope {
	switch o.Kind {
	case v1engine.Organization:
		return &pb.OwnerScope{Owner: &pb.OwnerScope_OrganizationId{OrganizationId: o.ID}}
	case v1engine.Workspace:
		return &pb.OwnerScope{Owner: &pb.OwnerScope_WorkspaceId{WorkspaceId: o.ID}}
	case v1engine.Project:
		return &pb.OwnerScope{Owner: &pb.OwnerScope_ProjectId{ProjectId: o.ID}}
	}
	return nil
}
func kind(s string) v1engine.Kind { return v1engine.Kind(s) }
func fromRecord(in *pb.ResourceRecord) (v1engine.Record, error) {
	if in == nil {
		return v1engine.Record{}, v1engine.ErrInvalid
	}
	var r v1engine.Record
	var m *pb.ResourceMetadata
	switch {
	case in.GetTenant() != nil:
		r.Kind = v1engine.Tenant
		m = in.GetTenant().GetMetadata()
	case in.GetGroup() != nil:
		x := in.GetGroup()
		r.Kind = v1engine.Group
		m = x.GetMetadata()
		r.Owner = v1engine.Owner{Kind: v1engine.Tenant, ID: x.GetTenantId()}
	case in.GetOrganization() != nil:
		x := in.GetOrganization()
		r.Kind = v1engine.Organization
		m = x.GetMetadata()
		r.Owner = v1engine.Owner{Kind: v1engine.Tenant, ID: x.GetTenantId()}
	case in.GetWorkspace() != nil:
		x := in.GetWorkspace()
		r.Kind = v1engine.Workspace
		m = x.GetMetadata()
		r.Owner = v1engine.Owner{Kind: v1engine.Organization, ID: x.GetOrganizationId()}
	case in.GetProject() != nil:
		x := in.GetProject()
		r.Kind = v1engine.Project
		m = x.GetMetadata()
		r.Owner = owner(x.GetOwner())
	case in.GetEnvironment() != nil:
		x := in.GetEnvironment()
		r.Kind = v1engine.Environment
		m = x.GetMetadata()
		r.Owner = v1engine.Owner{Kind: v1engine.Organization, ID: x.GetOrganizationId()}
		r.EnvironmentKey = x.GetKey()
	case in.GetWorkloadService() != nil:
		x := in.GetWorkloadService()
		r.Kind = v1engine.WorkloadService
		m = x.GetMetadata()
		r.Owner = owner(x.GetOwner())
	default:
		return v1engine.Record{}, v1engine.ErrInvalid // Descriptor registration requires a typed module integration
	}
	if m == nil {
		return v1engine.Record{}, v1engine.ErrInvalid
	}
	r.Name = m.GetName()
	r.Description = m.GetDescription()
	return r, nil
}
func toRecord(r v1engine.Record) *pb.ResourceRecord {
	meta := &pb.ResourceMetadata{Id: r.ID, Name: r.Name, Description: r.Description, Revision: r.Revision,
		State:      pb.ResourceState(pb.ResourceState_value["RESOURCE_STATE_"+r.State]),
		CreateTime: timestamppb.New(r.CreatedAt), UpdateTime: timestamppb.New(r.UpdatedAt)}
	switch r.Kind {
	case v1engine.Tenant:
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_Tenant{Tenant: &pb.Tenant{Metadata: meta}}}
	case v1engine.Group:
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_Group{Group: &pb.Group{Metadata: meta, TenantId: r.TenantID}}}
	case v1engine.Organization:
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_Organization{Organization: &pb.Organization{Metadata: meta, TenantId: r.TenantID}}}
	case v1engine.Workspace:
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_Workspace{Workspace: &pb.Workspace{Metadata: meta, OrganizationId: r.OrganizationID}}}
	case v1engine.Project:
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_Project{Project: &pb.Project{Metadata: meta, Owner: ownerProto(r.Owner), OwningOrganizationId: r.OrganizationID}}}
	case v1engine.Environment:
		envKind := "ENVIRONMENT_KIND_" + strings.ToUpper(r.EnvironmentKey)
		switch envKind {
		case "ENVIRONMENT_KIND_PROD":
			envKind = "ENVIRONMENT_KIND_PRODUCTION"
		case "ENVIRONMENT_KIND_TEST":
			envKind = "ENVIRONMENT_KIND_TESTING"
		case "ENVIRONMENT_KIND_DEV":
			envKind = "ENVIRONMENT_KIND_DEVELOPMENT"
		}
		if pb.EnvironmentKind_value[envKind] == 0 {
			envKind = "ENVIRONMENT_KIND_CUSTOM"
		}
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_Environment{Environment: &pb.Environment{Metadata: meta, OrganizationId: r.OrganizationID, Key: r.EnvironmentKey, Kind: pb.EnvironmentKind(pb.EnvironmentKind_value[envKind]), Isolation: pb.IsolationProfile_ISOLATION_PROFILE_LOGICAL}}}
	case v1engine.WorkloadService:
		return &pb.ResourceRecord{Resource: &pb.ResourceRecord_WorkloadService{WorkloadService: &pb.WorkloadService{Metadata: meta, Owner: ownerProto(r.Owner), OwningOrganizationId: r.OrganizationID}}}
	}
	return nil
}
func (s *ResourceServer) CreateResource(ctx context.Context, req *pb.CreateResourceRequest) (*pb.ResourceRecord, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	r, e := fromRecord(req.GetResource())
	if e != nil {
		return nil, asErr(e)
	}
	created, e := s.store.Create(ctx, r, req.GetIdempotencyKey())
	if e != nil {
		return nil, asErr(e)
	}
	if created.Kind == v1engine.Organization && !s.durable {
		if s.bootstrap == nil {
			return nil, status.Error(codes.FailedPrecondition, "org structure initializer unavailable")
		}
		if err := orgstructure.EnsureOrganizationInitialized(ctx, s.bootstrap, created.TenantID, created.ID, created.CreatedAt); err != nil {
			// A retry with the same idempotency key can repair the initial graph.
			// This memory-backed Organization path is NOT an atomic production
			// creation transaction with YDB; a durable Organization store is required.
			return nil, status.Error(codes.Unavailable, "org structure initialization incomplete; retry same key")
		}
	}
	return toRecord(created), nil
}
func (s *ResourceServer) GetResource(ctx context.Context, req *pb.GetResourceRequest) (*pb.ResourceRecord, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	r, e := s.store.Get(ctx, kind(req.GetKind()), req.GetId())
	if e != nil {
		return nil, asErr(e)
	}
	return toRecord(r), nil
}
func (s *ResourceServer) ListResources(ctx context.Context, req *pb.ListResourcesRequest) (*pb.ListResourcesResponse, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if req.GetPageToken() != "" || req.GetIncludeDescendants() || req.GetShowDeleted() {
		return nil, status.Error(codes.Unimplemented, "advanced listing requires durable projection")
	}
	rr, e := s.store.List(ctx, kind(req.GetKind()), req.GetOrganizationId())
	if e != nil {
		return nil, asErr(e)
	}
	out := &pb.ListResourcesResponse{Resources: make([]*pb.ResourceRecord, 0, len(rr))}
	for _, r := range rr {
		out.Resources = append(out.Resources, toRecord(r))
	}
	return out, nil
}
func (s *ResourceServer) UpdateResource(ctx context.Context, req *pb.UpdateResourceRequest) (*pb.ResourceRecord, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil || req.GetResource() == nil || req.GetUpdateMask() == nil {
		return nil, status.Error(codes.InvalidArgument, "resource and update_mask required")
	}
	r, e := fromRecord(req.GetResource())
	if e != nil {
		return nil, asErr(e)
	}
	for _, p := range req.GetUpdateMask().GetPaths() {
		if p != "metadata.name" && p != "metadata.description" {
			return nil, status.Error(codes.InvalidArgument, "unsupported update_mask path")
		}
	}
	v, e := s.store.Update(ctx, r.Kind, req.GetId(), req.GetExpectedRevision(), r.Name, r.Description)
	if e != nil {
		return nil, asErr(e)
	}
	return toRecord(v), nil
}
func (s *ResourceServer) ChangeResourceState(ctx context.Context, req *pb.ChangeResourceStateRequest) (*pb.ResourceRecord, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	target := strings.TrimPrefix(req.GetTargetState().String(), "RESOURCE_STATE_")
	v, e := s.store.ChangeState(ctx, kind(req.GetKind()), req.GetId(), req.GetExpectedRevision(), target)
	if e != nil {
		return nil, asErr(e)
	}
	return toRecord(v), nil
}
func (s *ResourceServer) MoveProject(ctx context.Context, req *pb.MoveProjectRequest) (*pb.Project, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	v, e := s.store.MoveProject(ctx, req.GetProjectId(), owner(req.GetNewOwner()), req.GetExpectedRevision())
	if e != nil {
		return nil, asErr(e)
	}
	return toRecord(v).GetProject(), nil
}
func (s *ResourceServer) ResolveResourceScope(ctx context.Context, req *pb.ResolveResourceScopeRequest) (*pb.ResolveResourceScopeResponse, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	path, e := s.store.Resolve(ctx, kind(req.GetKind()), req.GetId())
	if e != nil {
		return nil, asErr(e)
	}
	result := &pb.ResolveResourceScopeResponse{Ancestry: make([]*pb.ResourceRef, 0, len(path))}
	for _, r := range path {
		if r.Kind == v1engine.Tenant {
			continue
		}
		result.Ancestry = append(result.Ancestry, &pb.ResourceRef{Kind: string(r.Kind), Id: r.ID})
		if r.Kind == v1engine.Organization {
			result.OwningOrganizationId = r.ID
		}
	}
	return result, nil
}
func (s *ResourceServer) SetGroupOrganizationLink(ctx context.Context, req *pb.SetGroupOrganizationLinkRequest) (*pb.GroupOrganizationLink, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if err := s.store.SetGroupLink(ctx, req.GetGroupId(), req.GetOrganizationId()); err != nil {
		return nil, asErr(err)
	}
	return &pb.GroupOrganizationLink{GroupId: req.GetGroupId(), OrganizationId: req.GetOrganizationId(), CreateTime: timestamppb.Now()}, nil
}
func (s *ResourceServer) RemoveGroupOrganizationLink(ctx context.Context, req *pb.RemoveGroupOrganizationLinkRequest) (*pb.EmptyResponse, error) {
	if err := s.guard(); err != nil {
		return nil, err
	}
	if req == nil || req.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key required")
	}
	if err := s.store.RemoveGroupLink(ctx, req.GetGroupId(), req.GetOrganizationId()); err != nil {
		return nil, asErr(err)
	}
	return &pb.EmptyResponse{}, nil
}
func (s *ResourceServer) ListGroupOrganizations(ctx context.Context, req *pb.ListGroupOrganizationsRequest) (*pb.ListGroupOrganizationsResponse, error) {
	if err := s.guard(); err != nil {
		return nil, err
	}
	if req == nil || req.GetPageToken() != "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request/page token")
	}
	ids, err := s.store.ListGroupOrganizations(ctx, req.GetGroupId())
	if err != nil {
		return nil, asErr(err)
	}
	out := &pb.ListGroupOrganizationsResponse{Links: make([]*pb.GroupOrganizationLink, 0, len(ids))}
	for _, id := range ids {
		out.Links = append(out.Links, &pb.GroupOrganizationLink{GroupId: req.GetGroupId(), OrganizationId: id})
	}
	return out, nil
}
func (s *ResourceServer) SetEnvironmentBinding(ctx context.Context, req *pb.SetEnvironmentBindingRequest) (*pb.EnvironmentBinding, error) {
	if e := s.guard(); e != nil {
		return nil, e
	}
	if req == nil || req.GetBinding() == nil || req.GetBinding().GetResource() == nil {
		return nil, status.Error(codes.InvalidArgument, "binding and resource required")
	}
	b := req.GetBinding()
	out, e := s.store.SetBinding(ctx, v1engine.Owner{Kind: kind(b.GetResource().GetKind()), ID: b.GetResource().GetId()}, b.GetEnvironmentId(), b.GetVersionRef(), b.GetConfigurationRef())
	if e != nil {
		return nil, asErr(e)
	}
	return &pb.EnvironmentBinding{Id: out.ID, Resource: b.GetResource(), EnvironmentId: out.EnvironmentID, VersionRef: out.Version, ConfigurationRef: out.ConfigRef, Revision: out.Revision}, nil
}
func (s *ResourceServer) RemoveEnvironmentBinding(ctx context.Context, req *pb.RemoveEnvironmentBindingRequest) (*pb.EmptyResponse, error) {
	if err := s.guard(); err != nil {
		return nil, err
	}
	if req == nil || req.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key required")
	}
	if err := s.store.RemoveBinding(ctx, req.GetId(), req.GetExpectedRevision()); err != nil {
		return nil, asErr(err)
	}
	return &pb.EmptyResponse{}, nil
}
func (s *ResourceServer) ListEnvironmentBindings(ctx context.Context, req *pb.ListEnvironmentBindingsRequest) (*pb.ListEnvironmentBindingsResponse, error) {
	if err := s.guard(); err != nil {
		return nil, err
	}
	if req == nil || req.GetPageToken() != "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request/page token")
	}
	var subject v1engine.Owner
	if req.GetResource() != nil {
		subject = v1engine.Owner{Kind: kind(req.GetResource().GetKind()), ID: req.GetResource().GetId()}
	}
	bindings, err := s.store.ListBindings(ctx, subject, req.GetEnvironmentId())
	if err != nil {
		return nil, asErr(err)
	}
	out := &pb.ListEnvironmentBindingsResponse{Bindings: make([]*pb.EnvironmentBinding, 0, len(bindings))}
	for _, b := range bindings {
		out.Bindings = append(out.Bindings, &pb.EnvironmentBinding{
			Id: b.ID, Resource: &pb.ResourceRef{Kind: string(b.Resource.Kind), Id: b.Resource.ID},
			EnvironmentId: b.EnvironmentID, VersionRef: b.Version, ConfigurationRef: b.ConfigRef, Revision: b.Revision})
	}
	return out, nil
}
func (s *ResourceServer) RegisterResourceKind(context.Context, *pb.RegisterResourceKindRequest) (*pb.ResourceKindPolicy, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}
