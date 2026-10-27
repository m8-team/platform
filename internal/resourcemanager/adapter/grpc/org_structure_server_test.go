package grpcadapter

import (
	"context"
	"testing"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func orgGRPCFixture(t *testing.T, devMode bool) *OrgStructureServer {
	t.Helper()
	store := memory.NewOrgStructureStore()
	structure, err := orgunit.Build("org", 1, []orgunit.Unit{
		{ID: "root", OrganizationID: "org", State: orgunit.StateActive, DisplayName: "Head Office"},
		{ID: "a", OrganizationID: "org", ParentID: "root", State: orgunit.StateActive, DisplayName: "Sales"},
		{ID: "b", OrganizationID: "org", ParentID: "a", State: orgunit.StateActive, DisplayName: "Regional Sales"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Seed("tenant", structure); err != nil {
		t.Fatal(err)
	}
	server, err := NewOrgStructureServer(store, resourcemanager.Config{AllowUnauthenticated: devMode})
	if err != nil {
		t.Fatal(err)
	}
	return server
}
func orgContext() context.Context {
	return WithOrgStructureActor(context.Background(), orgstructure.Actor{TenantID: "tenant", PrincipalID: "admin"})
}
func TestOrgStructureGRPCRequiresTrustedActorAndAccessGate(t *testing.T) {
	s := orgGRPCFixture(t, true)
	if _, err := s.GetOrgUnit(context.Background(), &pb.GetOrgUnitRequest{OrganizationId: "org", OrgUnitId: "a"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated: %v", err)
	}
	if _, err := s.GetOrgUnit(orgContext(), &pb.GetOrgUnitRequest{OrganizationId: "org", OrgUnitId: "a"}); err != nil {
		t.Fatal(err)
	}
	denied := orgGRPCFixture(t, false)
	if _, err := denied.ListOrgUnits(orgContext(), &pb.ListOrgUnitsRequest{OrganizationId: "org"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied: %v", err)
	}
}
func TestOrgStructureGRPCReadHierarchy(t *testing.T) {
	s := orgGRPCFixture(t, true)
	ctx := orgContext()
	got, err := s.GetOrgUnit(ctx, &pb.GetOrgUnitRequest{OrganizationId: "org", OrgUnitId: "b"})
	if err != nil || got.GetMetadata().GetName() != "Regional Sales" || got.GetParentOrgUnitId() != "a" {
		t.Fatalf("Get: %v %#v", err, got)
	}
	units, err := s.ListOrgUnits(ctx, &pb.ListOrgUnitsRequest{OrganizationId: "org", PageSize: 2})
	if err != nil || len(units.GetOrgUnits()) != 2 || units.GetNextPageToken() == "" {
		t.Fatalf("List page1: %v %#v", err, units)
	}
	page2, err := s.ListOrgUnits(ctx, &pb.ListOrgUnitsRequest{OrganizationId: "org", PageSize: 2, PageToken: units.GetNextPageToken()})
	if err != nil || len(page2.GetOrgUnits()) != 1 {
		t.Fatalf("List page2: %v %#v", err, page2)
	}
	ancestors, err := s.ListOrgUnitAncestors(ctx, &pb.ListOrgUnitAncestorsRequest{OrganizationId: "org", OrgUnitId: "b", StructureRevision: 1})
	if err != nil || len(ancestors.GetAncestors()) != 3 || ancestors.GetAncestors()[2].GetDepth() != 2 {
		t.Fatalf("Ancestors: %v %#v", err, ancestors)
	}
	descendants, err := s.ListOrgUnitDescendants(ctx, &pb.ListOrgUnitDescendantsRequest{OrganizationId: "org", OrgUnitId: "root"})
	if err != nil || len(descendants.GetDescendants()) != 2 {
		t.Fatalf("Descendants: %v %#v", err, descendants)
	}
	if _, err = s.ListOrgUnitVersions(ctx, &pb.ListOrgUnitVersionsRequest{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("should not invent history: %v", err)
	}
}
func TestOrgStructureGRPCChangeSetValidateApplyReplay(t *testing.T) {
	s := orgGRPCFixture(t, true)
	ctx := orgContext()
	plan := &pb.OrgUnitChangeSet{
		Id: "cs_001", OrganizationId: "org", ExpectedStructureRevision: 1,
		Mutations: []*pb.OrgUnitMutation{
			{Kind: pb.OrgUnitMutationKind_ORG_UNIT_MUTATION_KIND_RENAME, OrgUnitId: "a", DisplayName: "Sales Division"},
		},
	}
	checked, err := s.ValidateOrgUnitChangeSet(ctx, &pb.ValidateOrgUnitChangeSetRequest{ChangeSet: plan})
	if err != nil || !checked.GetValid() || checked.GetProposedStructureRevision() != 2 {
		t.Fatalf("Validate: %v %#v", err, checked)
	}
	cmd := &pb.ApplyOrgUnitChangeSetRequest{ChangeSet: plan, IdempotencyKey: "cs_001"}
	applied, err := s.ApplyOrgUnitChangeSet(ctx, cmd)
	if err != nil || applied.GetStructureRevision() != 2 || applied.GetReplayed() {
		t.Fatalf("Apply: %v %#v", err, applied)
	}
	replayed, err := s.ApplyOrgUnitChangeSet(ctx, cmd)
	if err != nil || !replayed.GetReplayed() || replayed.GetStructureRevision() != 2 {
		t.Fatalf("Replay: %v %#v", err, replayed)
	}
	got, err := s.GetOrgUnit(ctx, &pb.GetOrgUnitRequest{OrganizationId: "org", OrgUnitId: "a"})
	if err != nil || got.GetMetadata().GetName() != "Sales Division" {
		t.Fatalf("persisted: %v %#v", err, got)
	}
	if _, err := s.ApplyOrgUnitChangeSet(ctx, &pb.ApplyOrgUnitChangeSetRequest{ChangeSet: plan, IdempotencyKey: "wrong"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid key: %v", err)
	}
}
func TestOrgStructureGRPCRegistration(t *testing.T) {
	s := grpc.NewServer()
	err := newOrgStructureRegistration(orgGRPCFixture(t, false))(s)
	if err != nil {
		t.Fatal(err)
	}
	info := s.GetServiceInfo()
	if _, ok := info["m8.platform.resourcemanager.v1.OrgStructureService"]; !ok {
		t.Fatal("missing OrgStructureService registration")
	}
	if _, ok := info["m8.platform.resourcemanager.v1.OrgUnitChangeSetService"]; !ok {
		t.Fatal("missing OrgUnitChangeSetService registration")
	}
}
