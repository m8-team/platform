package ydbroles

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

func testBundle(t *testing.T,version string)role.Bundle{
	t.Helper()
	b,err:=role.NewBundle(role.BundleRef{Name:"profiles.accountant",Version:version},
		[]role.TemplateRef{{Name:"finance.accountant",Version:"1.0.0"}},
		[]role.BundleRef{{Name:"profiles.employee",Version:"1.0.0"}})
	if err!=nil{t.Fatal(err)}
	return b
}
func testAssignment(t *testing.T,id string)role.ProfileAssignment{
	t.Helper()
	tenant,_:=shared.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject,_:=shared.NewSubject("user","alice")
	finance,_:=authorization.NewScope(authorization.ScopeInput{Kind:authorization.ScopeOrganization,
		TenantID:tenant,ResourceType:"finance.invoice",OrganizationID:"22222222-2222-4222-8222-222222222222"})
	docs,_:=authorization.NewScope(authorization.ScopeInput{Kind:authorization.ScopeInstance,
		TenantID:tenant,ResourceType:"documents.file",ResourceID:"file-42"})
	a,err:=role.NewProfileAssignment(role.ProfileAssignmentInput{
		ID:id,TenantID:tenant,Subject:subject,
		Target:role.AssignmentTarget{Kind:role.AssignmentBundle,Name:"profiles.accountant",Version:"1.0.0"},
		Scopes:[]authorization.Scope{finance,docs},
	})
	if err!=nil{t.Fatal(err)}
	return a
}

func TestBundleRoundTripAndDigest(t *testing.T){
	b:=testBundle(t,"1.0.0")
	raw,err:=encodeBundle(b)
	if err!=nil{t.Fatal(err)}
	restored,err:=decodeBundle(raw)
	if err!=nil{t.Fatal(err)}
	if restored.Ref()!=b.Ref()||len(restored.Roles())!=1||len(restored.Includes())!=1{t.Fatal("lost nested role references")}
	again,_:=encodeBundle(restored)
	if raw!=again{t.Fatal("noncanonical bundle representation")}
	if commandDigest("bundle",0,raw)==commandDigest("bundle",1,raw){t.Fatal("CAS not bound to digest")}
	if _,err:=decodeBundle("{not-json");!errors.Is(err,ErrInvalidSnapshot){t.Fatal(err)}
}
func TestAssignmentRoundTripIncludesIndependentScopesAndRevoke(t *testing.T){
	a:=testAssignment(t,"33333333-3333-4333-8333-333333333333")
	raw,err:=encodeAssignment(a)
	if err!=nil{t.Fatal(err)}
	loaded,err:=decodeAssignment(raw)
	if err!=nil||len(loaded.Scopes())!=2||loaded.Status()!=role.ProfileAssignmentPending{t.Fatalf("bad staged snapshot: %v",err)}
	revoked,err:=loaded.Revoke(1,time.Date(2026,11,23,10,0,0,0,time.UTC))
	if err!=nil{t.Fatal(err)}
	raw,err=encodeAssignment(revoked)
	if err!=nil{t.Fatal(err)}
	restored,err:=decodeAssignment(raw)
	if err!=nil||restored.Status()!=role.ProfileAssignmentRevoked||restored.Revision()!=2||restored.RevokedAt().IsZero(){t.Fatalf("lost revocation: %v",err)}
	if _,err:=decodeAssignment(strings.Replace(raw,"REVOKED","ACTIVE",1));!errors.Is(err,ErrInvalidSnapshot){t.Fatal("accepted unrecognized effective assignment status")}
}
func TestStoreRejectsInvalidConfig(t *testing.T){
	if _,err:=New(nil,"/local/access");!errors.Is(err,ErrInvalidStore){t.Fatal(err)}
	if _,err:=New(nil,"/local/../escape");!errors.Is(err,ErrInvalidStore){t.Fatal(err)}
}
