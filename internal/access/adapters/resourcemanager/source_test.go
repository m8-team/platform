package resourcemanager

import (
    "context"
    "errors"
    "testing"

    "github.com/m8-team/platform/internal/access/app/projection"
    "github.com/m8-team/platform/internal/access/domain"
    "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

type fakeCatalog struct {record v1engine.Record; err error}
func (f fakeCatalog)Get(context.Context,v1engine.Kind,string)(v1engine.Record,error){return f.record,f.err}
func TestSourceFailClosedOnForeignTenantAndInactiveState(t *testing.T){
    tenant,_:=domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
    resource,_:=domain.NewResource("workspace","11111111-1111-4111-8111-111111111111")
    rec:=v1engine.Record{Kind:v1engine.Workspace,ID:resource.ID(),TenantID:string(tenant),State:"ACTIVE",Revision:3}
    source,_:=New(fakeCatalog{record:rec})
    got,err:=source.AuthoritativeRevision(context.Background(),tenant,resource)
    if err!=nil || got!=3{t.Fatalf("revision %d %v",got,err)}
    key:=projection.Key{TenantID:tenant,Kind:projection.KindRMWorkspace,AggregateID:resource.ID()}
    snapshot,err:=source.LoadCurrent(context.Background(),key)
    if err!=nil || snapshot.Resource==nil || snapshot.Revision!=3 || snapshot.Validate()!=nil {
        t.Fatalf("bad trusted source snapshot %+v %v",snapshot,err)
    }
    rec.TenantID="bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
    source,_=New(fakeCatalog{record:rec})
    if _,err:=source.AuthoritativeRevision(context.Background(),tenant,resource);!errors.Is(err,ErrUntrustedResource) {t.Fatalf("cross tenant resource: %v",err)}
    rec.TenantID=string(tenant)
    rec.State="DELETED"
    source,_=New(fakeCatalog{record:rec})
    if _,err:=source.AuthoritativeRevision(context.Background(),tenant,resource);!errors.Is(err,ErrUntrustedResource){t.Fatalf("deleted resource allowed %v",err)}
    snapshot,err=source.LoadCurrent(context.Background(),key)
    if err!=nil || snapshot.Resource.Lifecycle!="DELETED"{t.Fatalf("tombstone missing: %+v %v",snapshot,err)}
}
