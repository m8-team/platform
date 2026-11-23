package consistency

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/policy"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

type restrictionsStub struct {
	snapshot CompleteRestrictionSnapshot
	err error
}
func (s restrictionsStub) LoadRestrictions(context.Context,domain.TenantID,string,string,string)(CompleteRestrictionSnapshot,error){
	return s.snapshot,s.err
}
type contextualStub struct{
	value authorization.EvaluationContext
	err error
}
func(s contextualStub) ResolveTrustedContext(context.Context,domain.TenantID,domain.Subject,string,string)(authorization.EvaluationContext,error){
	return s.value,s.err
}
func TestRestrictedUniversalRequiresCompleteSameHeadAndTrustedContext(t *testing.T){
	tenant,_:=domain.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject,_:=domain.NewSubject("user","alice")
	tid,_:=shared.NewTenantID(string(tenant))
	scope,_:=authorization.NewScope(authorization.ScopeInput{
		Kind:authorization.ScopeInstance,TenantID:tid,ResourceType:"finance.invoice",ResourceID:"invoice-123",
	})
	action,_:=permission.NewPermission("finance.invoice.view")
	cidr,_:=policy.NewIPInCIDR([]string{"10.0.0.0/8"})
	attr,_:=policy.ResourceAttributeEquals("status","approved")
	conditions,_:=policy.NewConditionSet([]policy.Condition{cidr,policy.RequireMFA(),attr},nil)
	p,err:=policy.NewPolicy(policy.Input{
		ID:"33333333-3333-4333-8333-333333333333",TenantID:tid,Name:"Verified payment authorizer",
		Scope:scope,Actions:[]permission.Permission{action},
		Effect:policy.EffectRequire,Conditions:conditions,
	})
	if err!=nil{t.Fatal(err)}
	owner:=&universalOwnerStub{proof:UniversalOwnerProof{TenantID:tenant,ResourceType:"finance.invoice",ResourceID:"invoice-123",Revision:7,Live:true}}
	graph:=&universalGraphStub{allowed:true}
	base,err:=NewUniversalBarrier(&policyStub{states:[]PolicyState{{Head:5,Applied:5}}},owner,universalCatalogStub{registered:true},graph)
	if err!=nil{t.Fatal(err)}
	buildCtx:=func(ip string,mfa bool,revision int64,status string)authorization.EvaluationContext{
		t.Helper()
		value,err:=authorization.NewEvaluationContext(authorization.ContextInput{
			ClientIP:netip.MustParseAddr(ip),MFAKnown:true,MFAVerified:mfa,
			EvaluatedAt:time.Now().UTC(),ResourceRevision:revision,
			ResourceAttributes:map[string]string{"status":status},
		})
		if err!=nil{t.Fatal(err)}
		return value
	}
	check:=func(source CompleteRestrictionSnapshot,verified authorization.EvaluationContext)(bool,error){
		t.Helper()
		barrier,err:=NewRestrictedUniversalBarrier(base,restrictionsStub{snapshot:source},contextualStub{value:verified})
		if err!=nil{t.Fatal(err)}
		return barrier.Check(context.Background(),tenant,subject,"finance.invoice","invoice-123","finance.invoice.view")
	}
	source:=CompleteRestrictionSnapshot{PolicyHead:5,Complete:true,Policies:[]policy.Policy{p}}
	ok,err:=check(source,buildCtx("10.3.4.5",true,7,"approved"))
	if err!=nil||!ok{t.Fatalf("verified mandatory conditions denied: %v",err)}
	for _,tc:=range []struct{ip string;mfa bool;revision int64;status string}{
		{"192.168.1.1",true,7,"approved"},
		{"10.3.4.5",false,7,"approved"},
		{"10.3.4.5",true,7,"suspended"},
	}{
		ok,err=check(source,buildCtx(tc.ip,tc.mfa,tc.revision,tc.status))
		if err!=nil||ok{t.Fatalf("mandatory restriction bypassed: %+v allowed=%v err=%v",tc,ok,err)}
	}
	ok,err=check(source,buildCtx("10.3.4.5",true,8,"approved"))
	if ok||!errors.Is(err,ErrResourceHeadChanged){t.Fatalf("unmatched owner attributes accepted: %v",err)}
	noPolicy:=CompleteRestrictionSnapshot{PolicyHead:5,Complete:true}
	ok,err=check(noPolicy,buildCtx("10.3.4.5",true,7,"approved"))
	if err!=nil||!ok{t.Fatalf("complete empty restriction snapshot should allow base entitlement: %v",err)}
	noPolicy.Complete=false
	ok,err=check(noPolicy,buildCtx("10.3.4.5",true,7,"approved"))
	if ok||!errors.Is(err,ErrContextualPolicyConflict){t.Fatalf("incomplete policy inventory accepted: %v",err)}
	noPolicy.Complete=true
	noPolicy.PolicyHead=4
	ok,err=check(noPolicy,buildCtx("10.3.4.5",true,7,"approved"))
	if ok||!errors.Is(err,ErrContextualPolicyConflict){t.Fatalf("stale policy inventory accepted: %v",err)}
}
func TestRestrictedUniversalConstructorRequiresFullInputs(t *testing.T){
	if _,err:=NewRestrictedUniversalBarrier(nil,nil,nil);!errors.Is(err,ErrContextualPolicyUnavailable){t.Fatal(err)}
}
