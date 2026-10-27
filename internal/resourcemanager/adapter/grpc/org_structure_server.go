package grpcadapter

import (
	"context"
	"crypto/rand"
	"errors"
	"math"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type orgStructureActorKey struct{}

// WithOrgStructureActor is ONLY for trusted authentication middleware and tests.
// Untrusted gRPC metadata, tenant headers and body fields never become identity.
func WithOrgStructureActor(ctx context.Context, actor orgstructure.Actor) context.Context {
	return context.WithValue(ctx, orgStructureActorKey{}, actor)
}

func authenticatedOrgActor(ctx context.Context) (orgstructure.Actor, error) {
	if ctx == nil {
		return orgstructure.Actor{}, status.Error(codes.Unauthenticated, "authentication required")
	}
	actor, ok := ctx.Value(orgStructureActorKey{}).(orgstructure.Actor)
	if !ok || actor.TenantID == "" || actor.PrincipalID == "" {
		return orgstructure.Actor{}, status.Error(codes.Unauthenticated, "authentication required")
	}
	return actor, nil
}

// orgStructureGate is intentionally deny-all unless the server is explicitly
// running in its existing *development-only* AllowUnauthenticated mode.
// Production must replace it with a policy/impact adapter to M8 Access.
type orgStructureGate struct{ allowDevelopment bool }

func (g *orgStructureGate) AuthorizeOrganization(context.Context, orgstructure.Actor, string) error {
	if g == nil || !g.allowDevelopment {
		return ports.ErrPermissionDenied
	}
	return nil
}
func (g *orgStructureGate) CanReadUnit(context.Context, orgstructure.Actor, string, string) (bool, error) {
	if g == nil || !g.allowDevelopment {
		return false, ports.ErrPermissionDenied
	}
	return true, nil
}
func (g *orgStructureGate) Authorize(context.Context, orgstructure.Actor, orgunit.ChangeSet) error {
	if g == nil || !g.allowDevelopment {
		return ports.ErrPermissionDenied
	}
	return nil
}
func (g *orgStructureGate) CheckImpact(context.Context, orgstructure.Actor, orgunit.Structure, orgunit.Structure, orgunit.ChangeSet) error {
	if g == nil || !g.allowDevelopment {
		return ports.ErrPermissionDenied
	}
	return nil
}

// OrgStructureServer is the generated v1 gRPC transport. Its backing store is
// currently process-local and must be replaced with YDB before production.
type OrgStructureServer struct {
	pb.UnimplementedOrgStructureServiceServer
	pb.UnimplementedOrgUnitChangeSetServiceServer
	queries *orgstructure.QueryService
	changes *orgstructure.Service
}

func NewOrgStructureServer(store *memory.OrgStructureStore, cfg resourcemanager.Config) (*OrgStructureServer, error) {
	gate := &orgStructureGate{allowDevelopment: cfg.AllowUnauthenticated}
	key := append([]byte(nil), cfg.PageTokenKey...)
	if len(key) == 0 {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
	}
	reader, err := orgstructure.NewQueryService(store, gate, key)
	if err != nil {
		return nil, err
	}
	writer, err := orgstructure.New(store, gate)
	if err != nil {
		return nil, err
	}
	return &OrgStructureServer{queries: reader, changes: writer}, nil
}

func grpcOrgErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	case errors.Is(err, ports.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, orgstructure.ErrUnitNotVisible), errors.Is(err, orgstructure.ErrStoreNotFound), errors.Is(err, orgunit.ErrUnitNotFound):
		return status.Error(codes.NotFound, "org unit not found")
	case errors.Is(err, orgunit.ErrRevisionConflict), errors.Is(err, orgstructure.ErrIdempotencyConflict):
		return status.Error(codes.Aborted, "organization structure revision or idempotency conflict")
	case errors.Is(err, orgunit.ErrDeferredEffectiveTime), errors.Is(err, orgunit.ErrSourceAuthority), errors.Is(err, orgunit.ErrUnitHasActiveChildren), errors.Is(err, orgunit.ErrParentInactive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orgstructure.ErrInvalidList), errors.Is(err, orgstructure.ErrInvalidPageToken),
		errors.Is(err, orgunit.ErrInvalidUnit), errors.Is(err, orgunit.ErrInvalidChangeSet),
		errors.Is(err, orgunit.ErrMissingParent), errors.Is(err, orgunit.ErrHierarchyCycle),
		errors.Is(err, orgunit.ErrConflictingChanges), errors.Is(err, orgunit.ErrUnitAlreadyExists):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "org structure operation failed")
	}
}

func orgUnitProto(u orgunit.Unit, actor orgstructure.Actor, revision uint64) *pb.OrgUnit {
	var state pb.OrgUnitState
	switch u.State {
	case orgunit.StatePlanned:
		state = pb.OrgUnitState_ORG_UNIT_STATE_PLANNED
	case orgunit.StateActive:
		state = pb.OrgUnitState_ORG_UNIT_STATE_ACTIVE
	case orgunit.StateClosed:
		state = pb.OrgUnitState_ORG_UNIT_STATE_CLOSED
	}
	mode := pb.OrgUnitManagementMode_ORG_UNIT_MANAGEMENT_MODE_MANUAL
	if u.AuthoritativeSourceID != "" {
		mode = pb.OrgUnitManagementMode_ORG_UNIT_MANAGEMENT_MODE_EXTERNAL
	}
	return &pb.OrgUnit{
		Metadata: &pb.ResourceMetadata{Id: u.ID, Name: u.DisplayName, Revision: int64(revision)},
		TenantId: actor.TenantID,
		OrganizationId: u.OrganizationID,
		ParentOrgUnitId: u.ParentID,
		LifecycleState: state,
		ManagementMode: mode,
		AuthoritativeSourceId: u.AuthoritativeSourceID,
		StructureRevision: int64(revision),
	}
}

func positiveRevision(v int64) (uint64, error) {
	if v < 0 {
		return 0, orgstructure.ErrInvalidList
	}
	return uint64(v), nil
}

func (s *OrgStructureServer) GetOrgUnit(ctx context.Context, req *pb.GetOrgUnitRequest) (*pb.OrgUnit, error) {
	actor, err := authenticatedOrgActor(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.GetOrganizationId() == "" || req.GetOrgUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id and org_unit_id are required")
	}
	result, err := s.queries.Get(ctx, actor, req.GetOrganizationId(), req.GetOrgUnitId())
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	return orgUnitProto(result.Unit, actor, result.StructureRevision), nil
}
func (s *OrgStructureServer) ListOrgUnits(ctx context.Context, req *pb.ListOrgUnitsRequest) (*pb.ListOrgUnitsResponse, error) {
	actor, err := authenticatedOrgActor(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	result, err := s.queries.List(ctx, actor, orgstructure.ListUnitsRequest{
		OrganizationID: req.GetOrganizationId(), PageSize: int(req.GetPageSize()), PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	resp := &pb.ListOrgUnitsResponse{NextPageToken: result.NextPageToken, StructureRevision: int64(result.StructureRevision)}
	for _, u := range result.Units {
		resp.OrgUnits = append(resp.OrgUnits, orgUnitProto(u, actor, result.StructureRevision))
	}
	return resp, nil
}
func (s *OrgStructureServer) ListOrgUnitAncestors(ctx context.Context, req *pb.ListOrgUnitAncestorsRequest) (*pb.ListOrgUnitAncestorsResponse, error) {
	actor, err := authenticatedOrgActor(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	rev, err := positiveRevision(req.GetStructureRevision())
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	result, err := s.queries.Ancestors(ctx, actor, req.GetOrganizationId(), req.GetOrgUnitId(), rev)
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	resp := &pb.ListOrgUnitAncestorsResponse{StructureRevision: int64(result.StructureRevision)}
	for _, ancestor := range result.Ancestors {
		resp.Ancestors = append(resp.Ancestors, &pb.OrgUnitAncestor{OrgUnitId: ancestor.AncestorID, Depth: int32(ancestor.Depth)})
	}
	return resp, nil
}
func (s *OrgStructureServer) ListOrgUnitDescendants(ctx context.Context, req *pb.ListOrgUnitDescendantsRequest) (*pb.ListOrgUnitDescendantsResponse, error) {
	actor, err := authenticatedOrgActor(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	rev, err := positiveRevision(req.GetStructureRevision())
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	result, err := s.queries.Descendants(ctx, actor, orgstructure.ListUnitsRequest{
		OrganizationID: req.GetOrganizationId(), AncestorID: req.GetOrgUnitId(),
		ExpectedRevision: rev, PageSize: int(req.GetPageSize()), PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	resp := &pb.ListOrgUnitDescendantsResponse{StructureRevision: int64(result.StructureRevision), NextPageToken: result.NextPageToken}
	for _, row := range result.Descendants {
		resp.Descendants = append(resp.Descendants, &pb.OrgUnitAncestor{OrgUnitId: row.DescendantID, Depth: int32(row.Depth)})
	}
	return resp, nil
}
func (*OrgStructureServer) ListOrgUnitVersions(context.Context, *pb.ListOrgUnitVersionsRequest) (*pb.ListOrgUnitVersionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "bitemporal OrgUnit history is not implemented")
}

func changeSetPlan(source *pb.OrgUnitChangeSet) (orgunit.ChangeSet, error) {
	if source == nil || source.GetId() == "" || source.GetOrganizationId() == "" || source.GetExpectedStructureRevision() <= 0 {
		return orgunit.ChangeSet{}, orgunit.ErrInvalidChangeSet
	}
	if source.GetEffectiveAt() != nil {
		return orgunit.ChangeSet{}, orgunit.ErrDeferredEffectiveTime
	}
	result := orgunit.ChangeSet{ID: source.GetId(), OrganizationID: source.GetOrganizationId(), ExpectedRevision: uint64(source.GetExpectedStructureRevision()), SourceID: source.GetSourceId()}
	for _, mutation := range source.GetMutations() {
		if mutation == nil {
			return orgunit.ChangeSet{}, orgunit.ErrInvalidChangeSet
		}
		m := orgunit.Change{UnitID: mutation.GetOrgUnitId(), ParentID: mutation.GetParentOrgUnitId(), DisplayName: mutation.GetDisplayName(), AuthoritativeSourceID: mutation.GetAuthoritativeSourceId()}
		switch mutation.GetKind() {
		case pb.OrgUnitMutationKind_ORG_UNIT_MUTATION_KIND_CREATE:
			m.Kind = orgunit.ChangeCreate
		case pb.OrgUnitMutationKind_ORG_UNIT_MUTATION_KIND_MOVE:
			m.Kind = orgunit.ChangeMove
		case pb.OrgUnitMutationKind_ORG_UNIT_MUTATION_KIND_RENAME:
			m.Kind = orgunit.ChangeRename
		case pb.OrgUnitMutationKind_ORG_UNIT_MUTATION_KIND_CLOSE:
			m.Kind = orgunit.ChangeClose
		default:
			return orgunit.ChangeSet{}, orgunit.ErrInvalidChangeSet
		}
		result.Changes = append(result.Changes, m)
	}
	return result, nil
}
func (s *OrgStructureServer) ValidateOrgUnitChangeSet(ctx context.Context, req *pb.ValidateOrgUnitChangeSetRequest) (*pb.ValidateOrgUnitChangeSetResponse, error) {
	actor, err := authenticatedOrgActor(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	plan, err := changeSetPlan(req.GetChangeSet())
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	revision, err := s.changes.ValidateImmediate(ctx, actor, plan)
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	if revision > math.MaxInt64 {
		return nil, status.Error(codes.Internal, "revision out of range")
	}
	return &pb.ValidateOrgUnitChangeSetResponse{Valid: true, ProposedStructureRevision: int64(revision)}, nil
}
func (s *OrgStructureServer) ApplyOrgUnitChangeSet(ctx context.Context, req *pb.ApplyOrgUnitChangeSetRequest) (*pb.ApplyOrgUnitChangeSetResponse, error) {
	actor, err := authenticatedOrgActor(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.GetChangeSet() == nil || req.GetIdempotencyKey() == "" || req.GetIdempotencyKey() != req.GetChangeSet().GetId() {
		return nil, status.Error(codes.InvalidArgument, "idempotency_key must equal change_set.id in v1")
	}
	plan, err := changeSetPlan(req.GetChangeSet())
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	result, err := s.changes.ApplyImmediate(ctx, actor, plan)
	if err != nil {
		return nil, grpcOrgErr(err)
	}
	if result.Structure.Revision > math.MaxInt64 {
		return nil, status.Error(codes.Internal, "revision out of range")
	}
	return &pb.ApplyOrgUnitChangeSetResponse{StructureRevision: int64(result.Structure.Revision), Replayed: result.Replayed}, nil
}

var _ pb.OrgStructureServiceServer = (*OrgStructureServer)(nil)
var _ pb.OrgUnitChangeSetServiceServer = (*OrgStructureServer)(nil)
