// Package memory provides an atomic deterministic contract-test repository.
// It is NOT a production persistence implementation.
package memory

import (
    "context"
    "errors"
    "reflect"
    "sync"

    "github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

type OrgStructureEvent struct {
    TenantID, OrganizationID, ChangeSetID string
    PreviousRevision, Revision uint64
    ActorPrincipalID string
}
type orgStructureRecord struct { hash string; result orgstructure.Result }
type orgStructureState struct {
    snapshot orgunit.Structure
    idempotency map[string]orgStructureRecord
    events []OrgStructureEvent
}
type OrgStructureStore struct {
    mu sync.Mutex
    data map[string]*orgStructureState
}
func NewOrgStructureStore() *OrgStructureStore {
    return &OrgStructureStore{data:map[string]*orgStructureState{}}
}
func orgStructureKey(tenant,org string)string{return tenant+"/"+org}
func cloneOrgStructure(v orgunit.Structure)orgunit.Structure {
    v.Units=append([]orgunit.Unit(nil),v.Units...)
    v.Closure=append([]orgunit.ClosureRow(nil),v.Closure...)
    return v
}
func (s *OrgStructureStore) Seed(tenant string, v orgunit.Structure)error {
    expected,err:=orgunit.Build(v.OrganizationID,v.Revision,v.Units);if err!=nil{return err}
    if !reflect.DeepEqual(expected.Closure,v.Closure) {return orgunit.ErrInvalidUnit}
    s.mu.Lock();defer s.mu.Unlock()
    id:=orgStructureKey(tenant,v.OrganizationID)
    if _,ok:=s.data[id];ok{return errors.New("org structure already seeded")}
    s.data[id]=&orgStructureState{snapshot:cloneOrgStructure(v),idempotency:map[string]orgStructureRecord{}}
    return nil
}
func (s *OrgStructureStore) Load(_ context.Context,tenant,org string)(orgunit.Structure,error){
    s.mu.Lock();defer s.mu.Unlock()
    v,ok:=s.data[orgStructureKey(tenant,org)]
    if !ok{return orgunit.Structure{},orgstructure.ErrStoreNotFound}
    return cloneOrgStructure(v.snapshot),nil
}
func (s *OrgStructureStore) FindIdempotent(_ context.Context,tenant,org,id,hash string)(orgstructure.Result,bool,error) {
    s.mu.Lock();defer s.mu.Unlock()
    v,ok:=s.data[orgStructureKey(tenant,org)]
    if !ok{return orgstructure.Result{},false,orgstructure.ErrStoreNotFound}
    rec,found:=v.idempotency[id]
    if !found{return orgstructure.Result{},false,nil}
    if rec.hash!=hash{return orgstructure.Result{},false,orgstructure.ErrIdempotencyConflict}
    r:=rec.result;r.Structure=cloneOrgStructure(r.Structure)
    return r,true,nil
}
func (s *OrgStructureStore) Commit(_ context.Context,r orgstructure.CommitRequest)(orgstructure.Result,error){
    s.mu.Lock();defer s.mu.Unlock()
    if r.Actor.TenantID==""||r.Plan.OrganizationID==""||r.Actor.PrincipalID==""||r.Plan.ID==""||r.RequestHash==""{
        return orgstructure.Result{},orgunit.ErrInvalidChangeSet
    }
    v,ok:=s.data[orgStructureKey(r.Actor.TenantID,r.Plan.OrganizationID)]
    if !ok{return orgstructure.Result{},orgstructure.ErrStoreNotFound}
    if rec,found:=v.idempotency[r.Plan.ID];found{
        if rec.hash!=r.RequestHash{return orgstructure.Result{},orgstructure.ErrIdempotencyConflict}
        res:=rec.result;res.Structure=cloneOrgStructure(res.Structure);res.Replayed=true
        return res,nil
    }
    if v.snapshot.Revision!=r.Plan.ExpectedRevision{return orgstructure.Result{},orgunit.ErrRevisionConflict}
    recomputed,err:=orgunit.Apply(v.snapshot,r.Plan);if err!=nil{return orgstructure.Result{},err}
    if !reflect.DeepEqual(recomputed,r.Snapshot){return orgstructure.Result{},orgunit.ErrInvalidChangeSet}
    oldRevision:=v.snapshot.Revision
    v.snapshot=cloneOrgStructure(recomputed)
    result:=orgstructure.Result{Structure:cloneOrgStructure(recomputed)}
    v.idempotency[r.Plan.ID]=orgStructureRecord{hash:r.RequestHash,result:result}
    if recomputed.Revision!=oldRevision {
        v.events=append(v.events,OrgStructureEvent{TenantID:r.Actor.TenantID,OrganizationID:r.Plan.OrganizationID,ChangeSetID:r.Plan.ID,PreviousRevision:oldRevision,Revision:recomputed.Revision,ActorPrincipalID:r.Actor.PrincipalID})
    }
    return result,nil
}
func (s *OrgStructureStore) Events(tenant,org string)[]OrgStructureEvent{
    s.mu.Lock();defer s.mu.Unlock()
    v,ok:=s.data[orgStructureKey(tenant,org)];if !ok{return nil}
    return append([]OrgStructureEvent(nil),v.events...)
}
