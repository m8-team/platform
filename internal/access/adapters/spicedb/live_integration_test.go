//go:build integration

package spicedb

import (
    "context"
    "errors"
    "os"
    "testing"
    "time"

    pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
    "github.com/m8-team/platform/internal/access/app/ports"
    "github.com/m8-team/platform/internal/access/domain"
)

// Exercises the real SpiceDB datastore and gRPC preconditions, not mocks.
func TestLiveSchemaFencingIndependentGrants(t *testing.T){
    endpoint:=os.Getenv("M8_ACCESS_TEST_SPICEDB_ENDPOINT")
    token:=os.Getenv("M8_ACCESS_TEST_SPICEDB_TOKEN")
    if endpoint=="" || token==""{t.Skip("requires disposable SpiceDB gRPC test service")}
    p,_:=domain.NewPermission("m8.workspace.read")
    def,_:=domain.NewPermissionDefinition(p,"workspace","view")
    catalog,err:=NewStaticPermissionCatalog([]domain.PermissionDefinition{def})
    if err!=nil{t.Fatal(err)}
    client,err:=NewClient(Config{Endpoint:endpoint,BearerToken:token,ModelRevision:"test-model-v1",AllowInsecureLoopback:true},catalog)
    if err!=nil{t.Fatal(err)}
    defer client.Close()
    ctx,cancel:=context.WithTimeout(context.Background(),90*time.Second)
    defer cancel()
    schema,err:=os.ReadFile("../../schema/m8.zed")
    if err!=nil{t.Fatal(err)}
    var wrote bool
    for i:=0;i<25;i++ {
        _,err=client.sdk.WriteSchema(ctx,&pb.WriteSchemaRequest{Schema:string(schema)})
        if err==nil{wrote=true;break}
        time.Sleep(400*time.Millisecond)
    }
    if !wrote{t.Fatalf("real SpiceDB schema write: %v",err)}
    tenant,_:=domain.NewTenantID(tenantValue)
    workspace,_:=domain.NewResource("workspace","same-id")
    subject,_:=domain.NewSubject("user","same-id")
    mkGrant:=func(id string)(ports.GraphFence,domain.Relationship,domain.Relationship){
        fence:=ports.GraphFence{TenantID:tenant,Kind:"ACCESS_GRANT",AggregateID:id,Epoch:1}
        node,_:=domain.NewResource("access_grant",id)
        holder,_:=domain.NewRelationship(domain.RelationshipInput{TenantID:tenant,Resource:node,Relation:"holder",Subject:subject,ExpiresAt:time.Now().UTC().Add(time.Hour)})
        indirect,_:=domain.NewSubject("access_grant",id)
        viewer,_:=domain.NewRelationship(domain.RelationshipInput{TenantID:tenant,Resource:workspace,Relation:"viewer",Subject:indirect,SubjectRelation:"holder"})
        return fence,holder,viewer
    }
    g1,a1,b1:=mkGrant("11111111-1111-4111-8111-111111111111")
    g2,a2,b2:=mkGrant("22222222-2222-4222-8222-222222222222")
    for _,item:=range []struct{f ports.GraphFence;holder,viewer domain.Relationship}{{g1,a1,b1},{g2,a2,b2}}{
        if _,err:=client.EstablishFence(ctx,item.f);err!=nil{t.Fatalf("establish graph epoch: %v",err)}
        guarded,err:=ports.WithGraphFence(ctx,item.f)
        if err!=nil{t.Fatal(err)}
        _,err=client.WriteRelationships(guarded,tenant,[]ports.RelationshipChange{
            {Operation:ports.RelationshipTouch,Relationship:item.holder},
            {Operation:ports.RelationshipTouch,Relationship:item.viewer},
        })
        if err!=nil{t.Fatalf("write grant tuple: %v",err)}
    }
    evaluate:=func(allowed bool){
        t.Helper()
        rev,_:=domain.NewModelRevision("test-model-v1")
        req,_:=domain.NewCheckPermissionRequest(domain.CheckPermissionInput{TenantID:tenant,Subject:subject,
            Permission:p,Resource:workspace,ModelRevision:rev,Critical:true})
        result,err:=client.CheckPermission(ctx,req)
        if err!=nil || result.Allowed()!=allowed{t.Fatalf("permission should be %v: %+v %v",allowed,result,err)}
    }
    evaluate(true)
    c1,_:=ports.WithGraphFence(ctx,g1)
    if _,err:=client.WriteRelationships(c1,tenant,[]ports.RelationshipChange{
        {Operation:ports.RelationshipDelete,Relationship:a1},{Operation:ports.RelationshipDelete,Relationship:b1},
    });err!=nil{t.Fatalf("revoke first grant: %v",err)}
    evaluate(true) // independent grant #2 remains effective
    newest:=g1
    newest.Epoch=2
    if _,err:=client.EstablishFence(ctx,newest);err!=nil{t.Fatalf("take over graph fence: %v",err)}
    if _,err:=client.EstablishFence(ctx,g1);!errors.Is(err,ErrGraphFenceConflict){
        t.Fatalf("old graph epoch reinstalled: %v",err)
    }
    if _,err:=client.WriteRelationships(c1,tenant,[]ports.RelationshipChange{{Operation:ports.RelationshipTouch,Relationship:a1}});!errors.Is(err,ErrGraphFenceConflict){
        t.Fatalf("stale graph writer accepted: %v",err)
    }
    c2,_:=ports.WithGraphFence(ctx,g2)
    if _,err:=client.WriteRelationships(c2,tenant,[]ports.RelationshipChange{
        {Operation:ports.RelationshipDelete,Relationship:a2},{Operation:ports.RelationshipDelete,Relationship:b2},
    });err!=nil{t.Fatalf("revoke second grant: %v",err)}
    evaluate(false)
}
