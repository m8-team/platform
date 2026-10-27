package grpcadapter

import (
	"context"
	"testing"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListResourcesOrganizationsRequiresTenantScope(t *testing.T) {
	ctx := context.Background()
	svc := NewResourceServer(v1engine.NewStore(), resourcemanager.Config{AllowUnauthenticated: true}, memory.NewOrgStructureStore())
	newTenant := func(name string) string {
		t.Helper()
		r, err := svc.CreateResource(ctx, &pb.CreateResourceRequest{
			IdempotencyKey: "tenant-" + name,
			Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Tenant{Tenant: &pb.Tenant{
				Metadata: &pb.ResourceMetadata{Name: name},
			}}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return r.GetTenant().GetMetadata().GetId()
	}
	first, second := newTenant("First"), newTenant("Second")
	newOrg := func(parent, name string) string {
		t.Helper()
		r, err := svc.CreateResource(ctx, &pb.CreateResourceRequest{
			IdempotencyKey: "org-" + name,
			Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Organization{Organization: &pb.Organization{
				TenantId: parent, Metadata: &pb.ResourceMetadata{Name: name},
			}}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return r.GetOrganization().GetMetadata().GetId()
	}
	one, two := newOrg(first, "Alpha"), newOrg(second, "Beta")
	result, err := svc.ListResources(ctx, &pb.ListResourcesRequest{Kind: "organization", TenantId: first})
	if err != nil || len(result.GetResources()) != 1 {
		t.Fatalf("Tenant-scoped Organizations: %#v err=%v", result, err)
	}
	if result.GetResources()[0].GetOrganization().GetMetadata().GetId() != one {
		t.Fatal("organization from another tenant was returned")
	}
	result, err = svc.ListResources(ctx, &pb.ListResourcesRequest{Kind: "organization", TenantId: second})
	if err != nil || len(result.GetResources()) != 1 ||
		result.GetResources()[0].GetOrganization().GetMetadata().GetId() != two {
		t.Fatalf("second tenant was not isolated: %#v err=%v", result, err)
	}
	for _, request := range []*pb.ListResourcesRequest{
		{Kind: "organization"},
		{Kind: "tenant", TenantId: first},
		{Kind: "organization", TenantId: first, OrganizationId: one},
	} {
		_, err := svc.ListResources(ctx, request)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid tenant query %#v: got %v", request, err)
		}
	}
	_, err = svc.ListResources(ctx, &pb.ListResourcesRequest{Kind: "organization", TenantId: "00000000-0000-4000-8000-000000000001"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("missing Tenant should not return global list: %v", err)
	}
}

func TestListResourcesGroupsIsolatedPerTenant(t *testing.T) {
	ctx := context.Background()
	svc := NewResourceServer(v1engine.NewStore(), resourcemanager.Config{AllowUnauthenticated: true}, memory.NewOrgStructureStore())
	createTenant := func(name string) string {
		t.Helper()
		r, err := svc.CreateResource(ctx, &pb.CreateResourceRequest{
			IdempotencyKey: "tenant-group-" + name,
			Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Tenant{Tenant: &pb.Tenant{
				Metadata: &pb.ResourceMetadata{Name: name},
			}}},
		})
		if err != nil { t.Fatal(err) }
		return r.GetTenant().GetMetadata().GetId()
	}
	createGroup := func(tenantID, name string) string {
		t.Helper()
		r, err := svc.CreateResource(ctx, &pb.CreateResourceRequest{
			IdempotencyKey: "group-" + name,
			Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Group{Group: &pb.Group{
				TenantId: tenantID, Metadata: &pb.ResourceMetadata{Name: name},
			}}},
		})
		if err != nil { t.Fatal(err) }
		return r.GetGroup().GetMetadata().GetId()
	}
	first, second := createTenant("holding-a"), createTenant("holding-b")
	one, two := createGroup(first, "First group"), createGroup(second, "Second group")
	for _, tc := range []struct{ tenantID, wanted string }{{first, one}, {second, two}} {
		result, err := svc.ListResources(ctx, &pb.ListResourcesRequest{Kind: "group", TenantId: tc.tenantID})
		if err != nil || len(result.GetResources()) != 1 ||
			result.GetResources()[0].GetGroup().GetMetadata().GetId() != tc.wanted {
			t.Fatalf("Group tenant isolation failed: %#v err=%v", result, err)
		}
	}
	for _, req := range []*pb.ListResourcesRequest{
		{Kind: "group"},
		{Kind: "group", TenantId: first, OrganizationId: one},
		{Kind: "group", TenantId: first, Owner: &pb.OwnerScope{Owner: &pb.OwnerScope_OrganizationId{OrganizationId: one}}},
	} {
		_, err := svc.ListResources(ctx, req)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("unsafe group scope %#v: got %v", req, err)
		}
	}
	_, err := svc.ListResources(ctx, &pb.ListResourcesRequest{
		Kind: "group", TenantId: "00000000-0000-4000-8000-000000000001",
	})
	if status.Code(err) != codes.NotFound {t.Fatalf("unknown Tenant: %v", err)}
}
