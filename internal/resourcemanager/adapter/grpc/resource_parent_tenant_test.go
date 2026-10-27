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

// A missing parent Tenant produces gRPC NotFound (HTTP 404 through the
// transcoder), not an unregistered HTTP route. The UI should verify the
// Tenant before submitting CreateOrganization or explicitly provision it.
func TestCreateOrganizationRequiresAnExistingTenant(t *testing.T) {
    ctx := context.Background()
    service := NewResourceServer(
        v1engine.NewStore(),
        resourcemanager.Config{AllowUnauthenticated: true},
        memory.NewOrgStructureStore(),
    )
    organization := &pb.CreateResourceRequest{
        Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Organization{
            Organization: &pb.Organization{
                TenantId: "c6c8dc67-f749-42d6-b7c8-4e0000000000",
                Metadata: &pb.ResourceMetadata{Name: "Retail"},
            },
        }},
        IdempotencyKey: "missing-tenant",
    }
    _, err := service.CreateResource(ctx, organization)
    if status.Code(err) != codes.NotFound {
        t.Fatalf("want NotFound for unknown parent Tenant, got %v", err)
    }
    _, err = service.GetResource(ctx, &pb.GetResourceRequest{
        Kind: "tenant", Id: organization.Resource.GetOrganization().GetTenantId(),
    })
    if status.Code(err) != codes.NotFound {
        t.Fatalf("want NotFound for tenant preflight, got %v", err)
    }

    tenant, err := service.CreateResource(ctx, &pb.CreateResourceRequest{
        Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Tenant{
            Tenant: &pb.Tenant{Metadata: &pb.ResourceMetadata{Name: "Holding"}},
        }},
        IdempotencyKey: "create-parent-tenant",
    })
    if err != nil { t.Fatal(err) }
    tenantID := tenant.GetTenant().GetMetadata().GetId()
    organization.Resource.GetOrganization().TenantId = tenantID
    organization.IdempotencyKey = "create-child-organization"
    _, err = service.GetResource(ctx, &pb.GetResourceRequest{Kind:"tenant", Id:tenantID})
    if err != nil { t.Fatalf("tenant preflight should succeed: %v",err) }
    result, err := service.CreateResource(ctx, organization)
    if err != nil { t.Fatalf("creation with existing tenant failed: %v",err) }
    if result.GetOrganization().GetTenantId() != tenantID {
        t.Fatalf("unexpected parent ID: %#v",result.GetOrganization())
    }
}
