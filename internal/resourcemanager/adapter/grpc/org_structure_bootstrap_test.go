package grpcadapter

import (
 "context"
 "testing"

 pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
 "github.com/m8-team/platform/internal/resourcemanager"
 "github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
 "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

func TestCreateOrganizationInitializesOrgStructure(t *testing.T) {
 ctx:=context.Background()
 units:=memory.NewOrgStructureStore()
 service:=NewResourceServer(v1engine.NewStore(),resourcemanager.Config{AllowUnauthenticated:true},units)
 tenant,err:=service.CreateResource(ctx,&pb.CreateResourceRequest{
  Resource:&pb.ResourceRecord{Resource:&pb.ResourceRecord_Tenant{Tenant:&pb.Tenant{Metadata:&pb.ResourceMetadata{Name:"Demo"}}}},
  IdempotencyKey:"tenant-init-1",
 })
 if err!=nil{t.Fatal(err)}
 orgReq:=&pb.CreateResourceRequest{
  Resource:&pb.ResourceRecord{Resource:&pb.ResourceRecord_Organization{Organization:&pb.Organization{
   Metadata:&pb.ResourceMetadata{Name:"Retail"},TenantId:tenant.GetTenant().GetMetadata().GetId(),
  }}},
  IdempotencyKey:"org-init-1",
 }
 organization,err:=service.CreateResource(ctx,orgReq)
 if err!=nil{t.Fatal(err)}
 orgID:=organization.GetOrganization().GetMetadata().GetId()
 tenantID:=tenant.GetTenant().GetMetadata().GetId()
 got,err:=units.Load(ctx,tenantID,orgID)
 if err!=nil{t.Fatal(err)}
 if got.Revision!=1||len(got.Units)!=0||len(got.Closure)!=0{
   t.Fatalf("expected initial empty revision: %#v",got)
 }
 replay,err:=service.CreateResource(ctx,orgReq)
 if err!=nil{t.Fatal(err)}
 if replay.GetOrganization().GetMetadata().GetId()!=orgID{t.Fatal("idempotent organization identity changed")}
}
