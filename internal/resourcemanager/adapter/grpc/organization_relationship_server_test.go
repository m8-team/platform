package grpcadapter

import (
	"context"
	"testing"
	"time"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestOrganizationRelationshipServiceLifecycle(t *testing.T) {
	ctx:=context.Background()
	resource:=NewResourceServer(v1engine.NewStore(),resourcemanager.Config{AllowUnauthenticated:true},
		memory.NewOrgStructureStore())
	service:=NewOrganizationRelationshipServer(resource)
	tenant,err:=resource.CreateResource(ctx,&pb.CreateResourceRequest{
		IdempotencyKey:"rel-tenant",
		Resource:&pb.ResourceRecord{Resource:&pb.ResourceRecord_Tenant{Tenant:&pb.Tenant{
			Metadata:&pb.ResourceMetadata{Name:"Corporate tenant"},
		}}},
	})
	if err!=nil {t.Fatal(err)}
	tenantID:=tenant.GetTenant().GetMetadata().GetId()
	createOrg:=func(key,name string) string {
		t.Helper()
		result,err:=resource.CreateResource(ctx,&pb.CreateResourceRequest{
			IdempotencyKey:key,
			Resource:&pb.ResourceRecord{Resource:&pb.ResourceRecord_Organization{Organization:&pb.Organization{
				TenantId:tenantID,Metadata:&pb.ResourceMetadata{Name:name},
			}}},
		})
		if err!=nil {t.Fatal(err)}
		return result.GetOrganization().GetMetadata().GetId()
	}
	from,to,third:=createOrg("rel-a","Parent"),createOrg("rel-b","Child"),createOrg("rel-c","Provider")
	start:=timestamppb.New(time.Now().Add(-time.Hour))
	req:=&pb.CreateOrganizationRelationshipRequest{
		IdempotencyKey:"rel-create",
		Relationship:&pb.OrganizationRelationship{
			TenantId:tenantID,FromOrganizationId:from,ToOrganizationId:to,
			Type:pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_PARENT_SUBSIDIARY,
			OwnershipBasisPoints:5100,ValidFrom:start,
		},
	}
	created,err:=service.CreateOrganizationRelationship(ctx,req)
	if err!=nil || created.GetRevision()!=1 {t.Fatalf("create: %#v %v",created,err)}
	replayed,err:=service.CreateOrganizationRelationship(ctx,req)
	if err!=nil || replayed.GetId()!=created.GetId() {t.Fatalf("idempotent replay: %#v %v",replayed,err)}
	_,err=service.CreateOrganizationRelationship(ctx,&pb.CreateOrganizationRelationshipRequest{
		IdempotencyKey:"rel-cycle",
		Relationship:&pb.OrganizationRelationship{
			TenantId:tenantID,FromOrganizationId:to,ToOrganizationId:from,
			Type:pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_PARENT_SUBSIDIARY,
			ValidFrom:start,
		},
	})
	if status.Code(err)!=codes.FailedPrecondition {t.Fatalf("cycle must fail: %v",err)}
	_,err=service.GetOrganizationRelationship(ctx,&pb.GetOrganizationRelationshipRequest{
		TenantId:"outside",Id:created.GetId(),
	})
	if status.Code(err)!=codes.NotFound {t.Fatalf("cross-Tenant read must not work: %v",err)}
	_,err=service.CreateOrganizationRelationship(ctx,&pb.CreateOrganizationRelationshipRequest{
		IdempotencyKey:"rel-partner",
		Relationship:&pb.OrganizationRelationship{
			TenantId:tenantID,FromOrganizationId:third,ToOrganizationId:from,
			Type:pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_SERVICE_PROVIDER,
			ValidFrom:start,
		},
	})
	if err!=nil {t.Fatal(err)}
	firstPage,err:=service.ListOrganizationRelationships(ctx,&pb.ListOrganizationRelationshipsRequest{
		TenantId:tenantID,PageSize:1,
	})
	if err!=nil || len(firstPage.GetRelationships())!=1 || firstPage.GetNextPageToken()=="" {
		t.Fatalf("first page: %#v %v",firstPage,err)
	}
	nextPage,err:=service.ListOrganizationRelationships(ctx,&pb.ListOrganizationRelationshipsRequest{
		TenantId:tenantID,PageSize:1,PageToken:firstPage.GetNextPageToken(),
	})
	if err!=nil || len(nextPage.GetRelationships())!=1 {t.Fatalf("next page: %#v %v",nextPage,err)}
	if _,err=service.ListOrganizationRelationships(ctx,&pb.ListOrganizationRelationshipsRequest{
		TenantId:"outside",PageToken:firstPage.GetNextPageToken(),
	});status.Code(err)!=codes.InvalidArgument {
		t.Fatalf("scope-bound pagination: %v",err)
	}
	updated,err:=service.UpdateOrganizationRelationship(ctx,&pb.UpdateOrganizationRelationshipRequest{
		TenantId:tenantID,Id:created.GetId(),ExpectedRevision:1,
		OwnershipBasisPoints:5200,EvidenceRef:"document/2",IdempotencyKey:"rel-update",
	})
	if err!=nil || updated.GetRevision()!=2 {t.Fatalf("update: %#v %v",updated,err)}
	if _,err=service.UpdateOrganizationRelationship(ctx,&pb.UpdateOrganizationRelationshipRequest{
		TenantId:tenantID,Id:created.GetId(),ExpectedRevision:1,
		OwnershipBasisPoints:5500,IdempotencyKey:"rel-stale",
	});status.Code(err)!=codes.Aborted {t.Fatalf("stale version must fail: %v",err)}
	closed,err:=service.CloseOrganizationRelationship(ctx,&pb.CloseOrganizationRelationshipRequest{
		TenantId:tenantID,Id:created.GetId(),ExpectedRevision:2,
		IdempotencyKey:"rel-close",ClosedAt:timestamppb.Now(),
	})
	if err!=nil || closed.GetRevision()!=3 || closed.GetState()!=
		pb.OrganizationRelationshipState_ORGANIZATION_RELATIONSHIP_STATE_CLOSED {
		t.Fatalf("close: %#v %v",closed,err)
	}
	history,err:=service.ListOrganizationRelationshipVersions(ctx,&pb.ListOrganizationRelationshipVersionsRequest{
		TenantId:tenantID,Id:created.GetId(),
	})
	if err!=nil || len(history.GetVersions())!=3 || history.GetVersions()[0].GetRevision()!=1 {
		t.Fatalf("history: %#v %v",history,err)
	}
}

func TestOrganizationRelationshipServiceProductionGuard(t *testing.T) {
	svc:=NewOrganizationRelationshipServer(NewResourceServer(v1engine.NewStore(),resourcemanager.Config{
		AllowUnauthenticated:false,
	}))
	_,err:=svc.ListOrganizationRelationships(context.Background(),&pb.ListOrganizationRelationshipsRequest{
		TenantId:"any",
	})
	if status.Code(err)!=codes.PermissionDenied {t.Fatalf("fail-closed guard: %v",err)}
}
