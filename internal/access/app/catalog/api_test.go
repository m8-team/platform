package catalog

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

type apiIdentity struct {
	subject shared.Subject
	err error
}
func (p apiIdentity) AuthenticatedCatalogPrincipal(context.Context) (shared.Subject,error) {
	return p.subject,p.err
}
type apiAuthorizer struct {
	publish bool
	read bool
}
func (p *apiAuthorizer) CanPublishModule(_ context.Context,_ shared.Subject,moduleID string)(bool,error){
	return p.publish&&moduleID=="finance",nil
}
func (p *apiAuthorizer) CanReadModule(_ context.Context,_ shared.Subject,moduleID string)(bool,error){
	return p.read&&moduleID=="finance",nil
}
type apiRegistry struct {
	head int64
	records map[string]PublishedRecord
	commands map[string]module.Manifest
}
func newAPIRegistry()*apiRegistry{
	return &apiRegistry{records:make(map[string]PublishedRecord),commands:make(map[string]module.Manifest)}
}
func (r *apiRegistry) Publish(_ context.Context,m module.Manifest,expected int64,key string)(PublishedRecord,error){
	if prior,ok:=r.commands[key];ok{
		if !reflect.DeepEqual(prior,m){return PublishedRecord{},ErrCatalogConflict}
		return r.records[prior.Version()],nil
	}
	if expected!=r.head{return PublishedRecord{},shared.ErrRevisionConflict}
	if _,exists:=r.records[m.Version()];exists{return PublishedRecord{},ErrCatalogConflict}
	r.head++
	record:=PublishedRecord{Manifest:m,Revision:r.head}
	r.records[m.Version()]=record
	r.commands[key]=m
	return record,nil
}
func (r *apiRegistry) Get(_ context.Context,_,version string)(PublishedRecord,error){
	record,ok:=r.records[version]
	if !ok{return PublishedRecord{},ErrCatalogNotFound}
	return record,nil
}
func (r *apiRegistry) List(_ context.Context,_ shared.Subject,_ string,size int,token string)([]PublishedRecord,string,error){
	if token!="" {return nil,"",ErrCatalogArgument}
	records:=make([]PublishedRecord,0,len(r.records))
	for _,v:=range r.records {records=append(records,v)}
	if len(records)>size {records=records[:size]}
	return records,"",nil
}

func inputForVersion(t *testing.T,version string)PublishInput{
	t.Helper()
	kind,err:=resource.NewTypeDefinition("finance","finance.invoice",[]authorization.ScopeKind{authorization.ScopeInstance,authorization.ScopeOrganization})
	if err!=nil{t.Fatal(err)}
	action,err:=permission.NewPermission("finance.invoice.view")
	if err!=nil{t.Fatal(err)}
	definition,err:=permission.NewDefinition(action,"finance.invoice","use")
	if err!=nil{t.Fatal(err)}
	template,err:=role.NewTemplate("finance",role.TemplateRef{Name:"finance.accountant",Version:version},
		[]permission.Permission{action},[]authorization.ScopeKind{authorization.ScopeOrganization})
	if err!=nil{t.Fatal(err)}
	return PublishInput{ModuleID:"finance",Version:version,Resources:[]resource.TypeDefinition{kind},
		Permissions:[]permission.Definition{definition},Roles:[]role.Template{template}}
}

func TestCatalogAPI_PublishCASAndIdempotency(t *testing.T){
	subject,_:=shared.NewSubject("service_account","finance-service")
	auth:=&apiAuthorizer{publish:true,read:true}
	registry:=newAPIRegistry()
	api,err:=NewAPI(apiIdentity{subject:subject},auth,registry)
	if err!=nil{t.Fatal(err)}
	first:=inputForVersion(t,"1.0.0")
	created,err:=api.Publish(context.Background(),first,0,"publish-1")
	if err!=nil||created.Revision!=1{t.Fatalf("create: %+v %v",created,err)}
	replayed,err:=api.Publish(context.Background(),first,0,"publish-1")
	if err!=nil||replayed.Revision!=1||registry.head!=1{t.Fatalf("replay: %+v %v",replayed,err)}
	changed:=inputForVersion(t,"1.0.1")
	if _,err=api.Publish(context.Background(),changed,0,"publish-1");!errors.Is(err,ErrCatalogConflict){t.Fatalf("idempotency mismatch accepted: %v",err)}
	if _,err=api.Publish(context.Background(),changed,0,"publish-2");!errors.Is(err,shared.ErrRevisionConflict){t.Fatalf("stale CAS accepted: %v",err)}
	published,err:=api.Publish(context.Background(),changed,1,"publish-2")
	if err!=nil||published.Revision!=2{t.Fatalf("new version: %v",err)}
	loaded,err:=api.Get(context.Background(),"finance","1.0.0")
	if err!=nil||loaded.Revision!=1{t.Fatalf("read prior immutable release: %v",err)}
	got,_,err:=api.List(context.Background(),"finance",10,"")
	if err!=nil||len(got)!=2{t.Fatalf("authorized list: %v %+v",err,got)}
	if _,_,err=api.List(context.Background(),"finance",101,"");!errors.Is(err,ErrCatalogArgument){t.Fatal(err)}
}

func TestCatalogAPI_RejectsUnauthorizedPublishingAndReads(t *testing.T){
	user,_:=shared.NewSubject("user","alice")
	service,_:=shared.NewSubject("service_account","finance-service")
	auth:=&apiAuthorizer{publish:true,read:false}
	registry:=newAPIRegistry()
	api,_:=NewAPI(apiIdentity{subject:user},auth,registry)
	if _,err:=api.Publish(context.Background(),inputForVersion(t,"1.0.0"),0,"key");!errors.Is(err,ErrCatalogForbidden){t.Fatalf("user registered module: %v",err)}
	api,_=NewAPI(apiIdentity{subject:service},auth,registry)
	if _,err:=api.Get(context.Background(),"finance","1.0.0");!errors.Is(err,ErrCatalogForbidden){t.Fatalf("unauthorized get: %v",err)}
	if _,_,err:=api.List(context.Background(),"finance",20,"");!errors.Is(err,ErrCatalogForbidden){t.Fatalf("unauthorized list: %v",err)}
	auth.publish=false
	if _,err:=api.Publish(context.Background(),inputForVersion(t,"1.0.0"),0,"key");!errors.Is(err,ErrCatalogForbidden){t.Fatalf("cross-namespace publish: %v",err)}
	if registry.head!=0{t.Fatal("unauthorized mutation committed")}
	api,_=NewAPI(apiIdentity{err:errors.New("missing credentials")},auth,registry)
	if _,err:=api.Publish(context.Background(),inputForVersion(t,"1.0.0"),0,"key");!errors.Is(err,ErrCatalogUnauthenticated){t.Fatalf("missing identity: %v",err)}
}
