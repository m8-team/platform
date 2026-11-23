package rolemanagement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

type identityStub struct{principal shared.Subject;err error}
func(i identityStub) AuthenticatedPrincipal(context.Context)(shared.Subject,error){return i.principal,i.err}
type authStub struct{allow bool}
func(a authStub) CanPublishProfile(context.Context,shared.Subject,role.BundleRef)(bool,error){return a.allow,nil}
func(a authStub) CanManageSubject(context.Context,shared.Subject,shared.TenantID,shared.Subject)(bool,error){return a.allow,nil}
func(a authStub) CanReadSubject(context.Context,shared.Subject,shared.TenantID,shared.Subject)(bool,error){return a.allow,nil}
type ownerStub struct{rev int64}
func(o ownerStub) VerifyScope(context.Context,authorization.Scope)(int64,error){return o.rev,nil}
type rolesStub struct{templates map[string]VerifiedTemplate}
func(r rolesStub) ResolveTemplate(_ context.Context,ref role.TemplateRef)(VerifiedTemplate,error){
	v,ok:=r.templates[ref.Key()]
	if !ok{return VerifiedTemplate{},ErrUnresolvedRole}
	return v,nil
}
type bundlesStub struct{stored map[string]role.Bundle}
func(b *bundlesStub) GetBundle(_ context.Context,ref role.BundleRef)(BundleRecord,error){
	v,ok:=b.stored[ref.Key()]
	if !ok{return BundleRecord{},ErrInvalidProfile}
	return BundleRecord{Bundle:v,Revision:1},nil
}
func(b *bundlesStub) PublishBundle(_ context.Context,bb role.Bundle,_ int64,_ string)(BundleRecord,error){
	b.stored[bb.Ref().Key()]=bb
	return BundleRecord{Bundle:bb,Revision:1},nil
}
type assignmentsStub struct{items map[string]role.ProfileAssignment}
func(a *assignmentsStub) CreateAssignment(_ context.Context,item role.ProfileAssignment,_ string)(role.ProfileAssignment,error){
	a.items[item.ID()]=item
	return item,nil
}
func(a *assignmentsStub) GetAssignment(_ context.Context,_ shared.TenantID,id string)(role.ProfileAssignment,error){
	v,ok:=a.items[id]
	if !ok{return role.ProfileAssignment{},ErrInvalidProfile}
	return v,nil
}
func(a *assignmentsStub) RevokeAssignment(_ context.Context,_ shared.TenantID,id string,expected int64,_ string)(role.ProfileAssignment,error){
	v,err:=a.items[id].Revoke(expected,aTestTime())
	if err!=nil{return role.ProfileAssignment{},err}
	a.items[id]=v
	return v,nil
}
func(a *assignmentsStub) ListBySubject(_ context.Context,_ shared.TenantID,_ shared.Subject,_ string,_ int)([]role.ProfileAssignment,error){
	var result []role.ProfileAssignment
	for _,x:=range a.items{result=append(result,x)}
	return result,nil
}
func aTestTime()time.Time{return time.Date(2026,11,23,12,0,0,0,time.UTC)}

func setup(t *testing.T)(*Service,role.ProfileAssignmentInput,*bundlesStub,*assignmentsStub){
	t.Helper()
	actor,_:=shared.NewSubject("service_account","access-admin")
	tenant,_:=shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject,_:=shared.NewSubject("user","alice")
	finScope,_:=authorization.NewScope(authorization.ScopeInput{
		Kind:authorization.ScopeOrganization,TenantID:tenant,ResourceType:"finance.invoice",
		OrganizationID:"22222222-2222-4222-8222-222222222222",
	})
	docScope,_:=authorization.NewScope(authorization.ScopeInput{
		Kind:authorization.ScopeInstance,TenantID:tenant,ResourceType:"documents.file",ResourceID:"file-42",
	})
	finAction,_:=permission.NewPermission("finance.invoice.view")
	docAction,_:=permission.NewPermission("documents.file.view")
	finRef:=role.TemplateRef{Name:"finance.accountant",Version:"1.0.0"}
	docRef:=role.TemplateRef{Name:"documents.reader",Version:"1.0.0"}
	finRole,_:=role.NewTemplate("finance",finRef,[]permission.Permission{finAction},[]authorization.ScopeKind{authorization.ScopeOrganization})
	docRole,_:=role.NewTemplate("documents",docRef,[]permission.Permission{docAction},[]authorization.ScopeKind{authorization.ScopeInstance})
	resolver:=rolesStub{templates:map[string]VerifiedTemplate{
		finRef.Key():{Template:finRole,Permissions:[]VerifiedPermission{{Action:finAction,ResourceType:"finance.invoice",Scopes:[]authorization.ScopeKind{authorization.ScopeOrganization}}}},
		docRef.Key():{Template:docRole,Permissions:[]VerifiedPermission{{Action:docAction,ResourceType:"documents.file",Scopes:[]authorization.ScopeKind{authorization.ScopeInstance}}}},
	}}
	bundleRef:=role.BundleRef{Name:"profiles.accountant",Version:"1.0.0"}
	bundle,_:=role.NewBundle(bundleRef,[]role.TemplateRef{finRef,docRef},nil)
	bundles:=&bundlesStub{stored:map[string]role.Bundle{bundleRef.Key():bundle}}
	assignments:=&assignmentsStub{items:map[string]role.ProfileAssignment{}}
	s,err:=New(identityStub{principal:actor},authStub{allow:true},resolver,ownerStub{rev:7},bundles,assignments)
	if err!=nil{t.Fatal(err)}
	input:=role.ProfileAssignmentInput{
		ID:"33333333-3333-4333-8333-333333333333",TenantID:tenant,Subject:subject,
		Target:role.AssignmentTarget{Kind:role.AssignmentBundle,Name:bundleRef.Name,Version:bundleRef.Version},
		Scopes:[]authorization.Scope{finScope,docScope},
	}
	return s,input,bundles,assignments
}
func TestCompositeAssignmentMustCoverEveryResourceType(t *testing.T){
	s,input,_,store:=setup(t)
	saved,err:=s.CreateAssignment(context.Background(),input,"request-1")
	if err!=nil||saved.Status()!=role.ProfileAssignmentPending||len(store.items)!=1{t.Fatalf("stage profile: %v",err)}
	input.ID="44444444-4444-4444-8444-444444444444"
	input.Scopes=input.Scopes[:1]
	if _,err:=s.CreateAssignment(context.Background(),input,"request-2");!errors.Is(err,ErrInvalidProfile){t.Fatalf("missing documents scope accepted: %v",err)}
	s.owner=ownerStub{rev:0}
	input.Scopes=saved.Scopes()
	if _,err:=s.CreateAssignment(context.Background(),input,"request-3");!errors.Is(err,ErrUnverifiedScope){t.Fatalf("missing authoritative proof accepted: %v",err)}
}
func TestPublishBundleRequiresPrivilegedIdentityAndPinnedRoles(t *testing.T){
	s,_,bundles,_:=setup(t)
	b:=bundles.stored["profiles.accountant@1.0.0"]
	_,err:=s.PublishBundle(context.Background(),b,0,"published-1")
	if err!=nil{t.Fatal(err)}
	s.auth=authStub{allow:false}
	if _,err:=s.PublishBundle(context.Background(),b,1,"published-2");!errors.Is(err,ErrForbidden){t.Fatal(err)}
	s.identity=identityStub{err:errors.New("unverified identity")}
	if _,err:=s.PublishBundle(context.Background(),b,1,"published-3");!errors.Is(err,ErrUnauthenticated){t.Fatal(err)}
}
