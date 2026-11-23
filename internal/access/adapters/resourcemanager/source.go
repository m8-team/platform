// Package resourcemanager reads authoritative resource revisions through the
// Resource Manager's exported catalog interface, not through Access-owned SQL.
package resourcemanager

import (
    "context"
    "errors"

    "github.com/m8-team/platform/internal/access/app/consistency"
    "github.com/m8-team/platform/internal/access/app/projection"
    "github.com/m8-team/platform/internal/access/domain"
    "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

var ErrUntrustedResource = errors.New("resource manager resource tenant, kind or lifecycle mismatch")

// Catalog is an owner-controlled Resource Manager read interface. An adapter
// implementation must read its authoritative storage (not cached client data).
type Catalog interface {
    Get(context.Context,v1engine.Kind,string)(v1engine.Record,error)
}
type Source struct { catalog Catalog }
func New(catalog Catalog)(*Source,error) {
    if catalog==nil{return nil,ErrUntrustedResource}
    return &Source{catalog:catalog},nil
}
var _ consistency.ResourceRevisionReader=(*Source)(nil)
var _ projection.SnapshotReader=(*Source)(nil)

func (s *Source) load(ctx context.Context, tenant domain.TenantID, resource domain.Resource)(v1engine.Record,error){
    if s==nil || s.catalog==nil || ctx.Err()!=nil || tenant.Validate()!=nil || resource.Validate()!=nil{
        if ctx.Err()!=nil{return v1engine.Record{},ctx.Err()}
        return v1engine.Record{},ErrUntrustedResource
    }
    if _,err:=projection.KindForResourceType(resource.Type());err!=nil{return v1engine.Record{},ErrUntrustedResource}
    rec,err:=s.catalog.Get(ctx,v1engine.Kind(resource.Type()),resource.ID())
    if err!=nil{return v1engine.Record{},err}
    if rec.Kind!=v1engine.Kind(resource.Type()) || rec.ID!=resource.ID() ||
        rec.TenantID!=string(tenant) || rec.Revision<1 {
        return v1engine.Record{},ErrUntrustedResource
    }
    return rec,nil
}

// AuthoritativeRevision refuses access to non-ACTIVE resources, even if
// SpiceDB still contains graph tuples from an earlier lifecycle.
func (s *Source) AuthoritativeRevision(ctx context.Context, tenant domain.TenantID, resource domain.Resource)(uint64,error){
    record,err:=s.load(ctx,tenant,resource)
    if err!=nil{return 0,err}
    if record.State!="ACTIVE" {return 0,ErrUntrustedResource}
    return uint64(record.Revision),nil
}
func (s *Source) LoadCurrent(ctx context.Context, key projection.Key)(projection.Snapshot,error){
    if err:=key.Validate();err!=nil{return projection.Snapshot{},err}
    kind,err:=projection.ResourceTypeForKind(key.Kind)
    if err!=nil{return projection.Snapshot{},err}
    res,err:=domain.NewResource(kind,key.AggregateID)
    if err!=nil{return projection.Snapshot{},err}
    rec,err:=s.load(ctx,key.TenantID,res)
    if err!=nil{return projection.Snapshot{},err}
    state:=&projection.ResourceState{
        TenantID:key.TenantID,Resource:res,Revision:uint64(rec.Revision),Lifecycle:rec.State,
    }
    return projection.Snapshot{Key:key,Revision:rec.Revision,Resource:state},nil
}
