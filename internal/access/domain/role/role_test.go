package role

import (
 "errors"
 "testing"
 "time"
 "github.com/m8-team/platform/internal/access/domain/permission"
 "github.com/m8-team/platform/internal/access/domain/shared"
)
const testTenant="11111111-1111-4111-8111-111111111111"
const roleID="22222222-2222-4222-8222-222222222222"
const bindingID="33333333-3333-4333-8333-333333333333"

func TestRoleAndBindingIndependence(t *testing.T){
 tenant,_:=shared.NewTenantID(testTenant)
 p,_:=permission.NewPermission("workspace.view")
 r,err:=NewRole(RoleInput{ID:roleID,TenantID:tenant,Name:"viewer",Permissions:[]permission.Permission{p},Revision:1})
 if err!=nil{t.Fatal(err)}
 returned:=r.Permissions()
 returned[0],_=permission.NewPermission("workspace.edit")
 if r.Permissions()[0].Name()!="workspace.view"{t.Fatal("mutable role permission slice")}
 if _,err=r.WithPermissions(0,nil);!errors.Is(err,shared.ErrRevisionConflict){t.Fatal(err)}
 updated,err:=r.WithPermissions(1,[]permission.Permission{p})
 if err!=nil||updated.Revision()!=2||r.Revision()!=1{t.Fatal(err)}
 actor,_:=shared.NewSubject("user","alice")
 scope,_:=shared.NewResource("workspace","workspace-1")
 binding,err:=NewRoleBinding(RoleBindingInput{ID:bindingID,TenantID:tenant,RoleID:roleID,Subject:actor,Scope:scope,Revision:1})
 if err!=nil{t.Fatal(err)}
 if err:=binding.ValidateForRole(r);err!=nil{t.Fatal(err)}
 revoked,err:=binding.Revoke(1,time.Now())
 if err!=nil||revoked.State()!=BindingRevoked||binding.State()!=BindingActive{t.Fatal(err)}
 if _,err=revoked.Revoke(2,time.Now());!errors.Is(err,ErrBindingNotActive){t.Fatal(err)}
}
