package module

import (
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)
func TestModuleManifestNamespacesAndRoleCatalog(t *testing.T){
	publisher,_:=shared.NewSubject("service_account","finance-service")
	resourceType,err:=resource.NewTypeDefinition("finance","finance.invoice",[]authorization.ScopeKind{
		authorization.ScopeInstance,authorization.ScopeOrganization,
	})
	if err!=nil{t.Fatal(err)}
	action,err:=permission.NewPermission("finance.invoice.view")
	if err!=nil{t.Fatal(err)}
	definition,err:=permission.NewDefinition(action,"finance.invoice","use")
	if err!=nil{t.Fatal(err)}
	template,err:=role.NewTemplate("finance",role.TemplateRef{Name:"finance.accountant",Version:"1.0.0"},
	[]permission.Permission{action},[]authorization.ScopeKind{authorization.ScopeOrganization})
	if err!=nil{t.Fatal(err)}
	in:=ManifestInput{ModuleID:"finance",Version:"1.0.0",Publisher:publisher,
		Resources:[]resource.TypeDefinition{resourceType},
		Permissions:[]permission.Definition{definition},Roles:[]role.Template{template}}
	manifest,err:=NewManifest(in)
	if err!=nil{t.Fatal(err)}
	if manifest.ModuleID()!="finance"||len(manifest.Roles())!=1{t.Fatal("unexpected registered module")}
	in.Permissions=append(in.Permissions,definition)
	if _,err=NewManifest(in);!errors.Is(err,ErrInvalidManifest){t.Fatalf("duplicate permission accepted: %v",err)}
	in.Permissions=[]permission.Definition{definition}
	foreign,_:=permission.NewPermission("security.users.manage")
	in.Permissions=nil
	foreignDef,_:=permission.NewDefinition(foreign,"finance.invoice","use")
	in.Permissions=[]permission.Definition{foreignDef}
	if _,err=NewManifest(in);!errors.Is(err,ErrInvalidManifest){t.Fatalf("foreign action accepted: %v",err)}
	in.Permissions=[]permission.Definition{definition}
	broadTemplate,_:=role.NewTemplate("finance",role.TemplateRef{Name:"finance.accountant",Version:"1.0.0"},
		[]permission.Permission{action},[]authorization.ScopeKind{authorization.ScopeTenant})
	in.Roles=[]role.Template{broadTemplate}
	if _,err=NewManifest(in);!errors.Is(err,ErrInvalidManifest){t.Fatalf("unsupported role scope accepted: %v",err)}
}
