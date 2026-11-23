package consistency

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain"
)

type universalOwnerStub struct{proof UniversalOwnerProof;calls int;mutate bool;err error}
func(s *universalOwnerStub) ResolveUniversalResource(_ context.Context,_ domain.TenantID,_,_ string)(UniversalOwnerProof,error){
	s.calls++
	if s.err!=nil{return UniversalOwnerProof{},s.err}
	p:=s.proof
	if s.mutate&&s.calls>1{p.Revision++}
	return p,nil
}
type universalCatalogStub struct{registered bool;err error}
func(s universalCatalogStub) HasCatalogPermission(context.Context,domain.TenantID,string,string)(bool,error){return s.registered,s.err}
type universalGraphStub struct{allowed bool;called int;err error}
func(s *universalGraphStub) CheckUniversalCapability(context.Context,domain.TenantID,domain.Subject,string,string,string)(bool,error){
	s.called++
	return s.allowed,s.err
}
func TestUniversalBarrierFailClosedConsistencyAndOwnership(t *testing.T){
	const kind="finance.invoice"
	const resource="invoice-123"
	const action="finance.invoice.view"
	tenant,err:=domain.NewTenantID("11111111-1111-4111-8111-111111111111")
	if err!=nil{t.Fatal(err)}
	user,err:=domain.NewSubject("user","alice")
	if err!=nil{t.Fatal(err)}
	newBarrier:=func(policies []PolicyState,registered,allow bool)(*UniversalBarrier,*universalOwnerStub,*universalGraphStub){
		owner:=&universalOwnerStub{proof:UniversalOwnerProof{TenantID:tenant,ResourceType:kind,ResourceID:resource,Revision:7,Live:true}}
		graph:=&universalGraphStub{allowed:allow}
		barrier,err:=NewUniversalBarrier(&policyStub{states:policies},owner,universalCatalogStub{registered:registered},graph)
		if err!=nil{t.Fatal(err)}
		return barrier,owner,graph
	}
	for _,state:=range []PolicyState{{Head:2,Applied:1},{Head:0,Applied:0}}{
		b,_,graph:=newBarrier([]PolicyState{state},true,true)
		if _,err:=b.Check(context.Background(),tenant,user,kind,resource,action);err==nil||graph.called!=0{
			t.Fatalf("lag/uninit state escaped: %v",err)
		}
	}
	b,owner,graph:=newBarrier([]PolicyState{{Head:1,Applied:1}},false,true)
	if _,err:=b.Check(context.Background(),tenant,user,kind,resource,action);!errors.Is(err,ErrUniversalPermissionUnknown)||owner.calls!=0||graph.called!=0{t.Fatal("unregistered permission escaped",err)}
	b,owner,graph=newBarrier([]PolicyState{{Head:1,Applied:1}},true,true)
	owner.proof.Live=false
	if _,err:=b.Check(context.Background(),tenant,user,kind,resource,action);!errors.Is(err,ErrUniversalResourceInvalid)||graph.called!=0{t.Fatal("deleted owner escaped",err)}
	b,owner,graph=newBarrier([]PolicyState{{Head:1,Applied:1}},true,true)
	owner.mutate=true
	if allowed,err:=b.Check(context.Background(),tenant,user,kind,resource,action);allowed||!errors.Is(err,ErrResourceHeadChanged){t.Fatal("owner mutation escaped",err)}
	b,_,graph=newBarrier([]PolicyState{{Head:2,Applied:2},{Head:3,Applied:2}},true,true)
	if allowed,err:=b.Check(context.Background(),tenant,user,kind,resource,action);allowed||!errors.Is(err,ErrPolicyProjectionLag){t.Fatal("in-flight revoke escaped",err)}
	b,_,graph=newBarrier([]PolicyState{{Head:2,Applied:2},{Head:3,Applied:3}},true,true)
	if allowed,err:=b.Check(context.Background(),tenant,user,kind,resource,action);allowed||!errors.Is(err,ErrPolicyHeadChanged){t.Fatal("new policy head escaped",err)}
	b,owner,graph=newBarrier([]PolicyState{{Head:2,Applied:2}},true,true)
	if allowed,err:=b.Check(context.Background(),tenant,user,kind,resource,action);err!=nil||!allowed||graph.called!=1||owner.calls!=2{
		t.Fatalf("converged fully-consistent permission not allowed: %v",err)
	}
	b,_,graph=newBarrier([]PolicyState{{Head:2,Applied:2}},true,false)
	if allowed,err:=b.Check(context.Background(),tenant,user,kind,resource,action);err!=nil||allowed||graph.called!=1{t.Fatal("negative graph result",err)}
}
