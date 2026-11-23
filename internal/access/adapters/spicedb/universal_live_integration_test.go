//go:build integration

package spicedb

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

// Exercises independent profile assignment grants, exact tenant/action
// isolation and graph fence takeover on real SpiceDB using official SDK.
func TestLiveUniversalProfileIndependentRevokeAndFence(t *testing.T) {
	client,ctx:=liveSpiceDB(t)
	tenant,err:=domain.NewTenantID("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	if err!=nil{t.Fatal(err)}
	user,err:=domain.NewSubject("user","profile-multi-reader")
	if err!=nil{t.Fatal(err)}
	capID,err:=projection.CapabilityObjectID("finance.invoice","invoice-123","finance.invoice.view")
	if err!=nil{t.Fatal(err)}
	capObj,err:=domain.NewResource("access_capability",capID)
	if err!=nil{t.Fatal(err)}
	type owner struct{
		fence ports.GraphFence
		holder,grantee domain.Relationship
	}
	makeAssignment:=func(id string)owner{
		t.Helper()
		fence:=ports.GraphFence{TenantID:tenant,Kind:"PROFILE_ASSIGNMENT",AggregateID:id,Epoch:1}
		node,err:=domain.NewResource("access_profile_assignment",id)
		if err!=nil{t.Fatal(err)}
		holder,err:=domain.NewRelationship(domain.RelationshipInput{TenantID:tenant,
			Resource:node,Relation:"holder",Subject:user})
		if err!=nil{t.Fatal(err)}
		subject,err:=domain.NewSubject("access_profile_assignment",id)
		if err!=nil{t.Fatal(err)}
		grantee,err:=domain.NewRelationship(domain.RelationshipInput{TenantID:tenant,
			Resource:capObj,Relation:"grantee",Subject:subject,SubjectRelation:"holder"})
		if err!=nil{t.Fatal(err)}
		return owner{fence:fence,holder:holder,grantee:grantee}
	}
	one:=makeAssignment("11111111-1111-4111-8111-111111111111")
	two:=makeAssignment("22222222-2222-4222-8222-222222222222")
	check:=func(tenant domain.TenantID,id,action string,want bool){
		t.Helper()
		got,err:=client.CheckUniversalCapability(ctx,tenant,user,"finance.invoice",id,action)
		if err!=nil||got!=want{t.Fatalf("capability check want %v got %v: %v",want,got,err)}
	}
	for _,item:=range []owner{one,two}{
		if _,err:=client.EstablishFence(ctx,item.fence);err!=nil{t.Fatal(err)}
		guarded,err:=ports.WithGraphFence(ctx,item.fence)
		if err!=nil{t.Fatal(err)}
		if _,err:=client.WriteRelationships(guarded,tenant,[]ports.RelationshipChange{
			{Operation:ports.RelationshipTouch,Relationship:item.holder},
			{Operation:ports.RelationshipTouch,Relationship:item.grantee},
		});err!=nil{t.Fatalf("profile graph write: %v",err)}
	}
	check(tenant,"invoice-123","finance.invoice.view",true)
	check(tenant,"invoice-999","finance.invoice.view",false)
	check(tenant,"invoice-123","finance.invoice.edit",false)
	check(otherTenantValue,"invoice-123","finance.invoice.view",false)
	revoke:=func(item owner){
		t.Helper()
		guarded,err:=ports.WithGraphFence(ctx,item.fence)
		if err!=nil{t.Fatal(err)}
		if _,err:=client.WriteRelationships(guarded,tenant,[]ports.RelationshipChange{
			{Operation:ports.RelationshipDelete,Relationship:item.holder},
			{Operation:ports.RelationshipDelete,Relationship:item.grantee},
		});err!=nil{t.Fatal(err)}
	}
	revoke(one)
	check(tenant,"invoice-123","finance.invoice.view",true)
	takeover:=one.fence
	takeover.Epoch=2
	if _,err:=client.EstablishFence(ctx,takeover);err!=nil{t.Fatal(err)}
	stale,err:=ports.WithGraphFence(ctx,one.fence)
	if err!=nil{t.Fatal(err)}
	if _,err:=client.WriteRelationships(stale,tenant,[]ports.RelationshipChange{
		{Operation:ports.RelationshipTouch,Relationship:one.holder},
	});!errors.Is(err,ErrGraphFenceConflict){t.Fatalf("stale worker restored revoked profile: %v",err)}
	revoke(two)
	check(tenant,"invoice-123","finance.invoice.view",false)
}
