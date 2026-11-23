package policy

import (
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
)
func context(t *testing.T,ip string,mfa bool,hasMFA bool)authorization.EvaluationContext{
	t.Helper()
	address:=netip.Addr{}
	if ip!=""{address=netip.MustParseAddr(ip)}
	c,err:=authorization.NewEvaluationContext(authorization.ContextInput{
		ClientIP:address,MFAKnown:hasMFA,MFAVerified:mfa,
		EvaluatedAt:time.Date(2026,11,23,16,0,0,0,time.UTC),
		ResourceAttributes:map[string]string{"status":"APPROVABLE"},ResourceRevision:5,
	})
	if err!=nil{t.Fatal(err)}
	return c
}
func TestTrustedIPMFAAndBusinessContext(t *testing.T){
	ip,err:=NewIPInCIDR([]string{"10.0.0.0/8","192.168.10.0/24"})
	if err!=nil{t.Fatal(err)}
	status,err:=ResourceAttributeEquals("status","APPROVABLE")
	if err!=nil{t.Fatal(err)}
	before,err:=Before(time.Date(2026,12,1,0,0,0,0,time.UTC))
	if err!=nil{t.Fatal(err)}
	conditions,err:=NewConditionSet([]Condition{ip,RequireMFA(),status,before},nil)
	if err!=nil{t.Fatal(err)}
	ok,err:=conditions.Evaluate(context(t,"10.5.6.7",true,true))
	if err!=nil||!ok{t.Fatalf("expected corporate permit gate: %v",err)}
	ok,err=conditions.Evaluate(context(t,"8.8.8.8",true,true))
	if err!=nil||ok{t.Fatal("public IP authorized")}
	ok,err=conditions.Evaluate(context(t,"10.5.6.7",false,true))
	if err!=nil||ok{t.Fatal("no MFA authorized")}
	if _,err=conditions.Evaluate(context(t,"",true,true));!errors.Is(err,ErrMissingTrustedContext){t.Fatal(err)}
	if _,err=conditions.Evaluate(context(t,"10.5.6.7",false,false));!errors.Is(err,ErrMissingTrustedContext){t.Fatal(err)}
}
func TestPolicyIsRestrictionNotPermission(t *testing.T){
	tenant,_:=shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	scope,_:=authorization.NewScope(authorization.ScopeInput{Kind:authorization.ScopeInstance,TenantID:tenant,
		ResourceType:"finance.invoice",ResourceID:"invoice-123"})
	p,_:=permission.NewPermission("finance.invoice.approve")
	cidr,_:=NewIPInCIDR([]string{"10.0.0.0/8"})
	conditions,_:=NewConditionSet([]Condition{cidr},nil)
	policy,err:=NewPolicy(Input{ID:"22222222-2222-4222-8222-222222222222",TenantID:tenant,Name:"corporate-ip",
		Scope:scope,Actions:[]permission.Permission{p},Effect:EffectRequire,Conditions:conditions})
	if err!=nil{t.Fatal(err)}
	passed,err:=policy.CheckRestriction(context(t,"8.8.8.8",true,true))
	if err!=nil||passed{t.Fatal("mandatory policy should block")}
	denied,err:=NewPolicy(Input{ID:"33333333-3333-4333-8333-333333333333",TenantID:tenant,Name:"deny-corporate",
		Scope:scope,Actions:[]permission.Permission{p},Effect:EffectDenyWhen,Conditions:conditions})
	if err!=nil{t.Fatal(err)}
	passed,err=denied.CheckRestriction(context(t,"10.1.2.3",true,true))
	if err!=nil||passed{t.Fatal("deny policy should block matching context")}
	disabled,err:=policy.Disable(1)
	if err!=nil{t.Fatal(err)}
	if _,err=disabled.CheckRestriction(context(t,"10.1.2.3",true,true));!errors.Is(err,ErrPolicyDisabled){t.Fatal(err)}
}
