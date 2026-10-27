// Package orgstructure orchestrates authorized, revision-guarded org changes.
// Storage must commit the head revision, idempotency record and outbox atomically.
package orgstructure

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "time"

    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

var ErrPolicyNotConfigured = errors.New("org structure policy not configured")
var ErrIdempotencyConflict = errors.New("idempotency key reused for different request")
var ErrStoreNotFound = errors.New("org structure not found")

// Actor must be derived from the authenticated server request context.
type Actor struct { TenantID string; PrincipalID string }
type DecisionGate interface {
    Authorize(ctx context.Context, actor Actor, plan orgunit.ChangeSet) error
    CheckImpact(ctx context.Context, actor Actor, before, after orgunit.Structure, plan orgunit.ChangeSet) error
}
type CommitRequest struct {
    Actor Actor
    Plan orgunit.ChangeSet
    RequestHash string
    Snapshot orgunit.Structure
}
type Result struct { Structure orgunit.Structure; Replayed bool }
type Store interface {
    Load(ctx context.Context, tenantID, organizationID string) (orgunit.Structure, error)
    FindIdempotent(ctx context.Context, tenantID, organizationID, changeSetID, requestHash string) (Result, bool, error)
    // The production implementation must CAS the head revision and persist
    // snapshot, history, idempotency and outbox atomically.
    Commit(ctx context.Context, request CommitRequest) (Result, error)
}
type Service struct { store Store; gate DecisionGate }
func New(store Store, gate DecisionGate) (*Service, error) {
    if store == nil || gate == nil {return nil,ErrPolicyNotConfigured}
    return &Service{store:store,gate:gate},nil
}
func (s *Service) Apply(ctx context.Context, actor Actor, plan orgunit.ChangeSet) (Result,error) {
    if s==nil || s.gate==nil || s.store==nil {return Result{},ErrPolicyNotConfigured}
    if actor.TenantID=="" || actor.PrincipalID=="" || plan.ID=="" || plan.OrganizationID=="" {return Result{},orgunit.ErrInvalidChangeSet}
    if err:=s.gate.Authorize(ctx,actor,plan);err!=nil{return Result{},err}
    digest,err:=requestHash(plan);if err!=nil{return Result{},err}
    if r,found,err:=s.store.FindIdempotent(ctx,actor.TenantID,plan.OrganizationID,plan.ID,digest);err!=nil{return Result{},err}else if found {r.Replayed=true;return r,nil}
    before,err:=s.store.Load(ctx,actor.TenantID,plan.OrganizationID);if err!=nil{return Result{},err}
    next,err:=orgunit.Apply(before,plan);if err!=nil{return Result{},err}
    if err:=s.gate.CheckImpact(ctx,actor,before,next,plan);err!=nil{return Result{},err}
    return s.store.Commit(ctx,CommitRequest{Actor:actor,Plan:plan,RequestHash:digest,Snapshot:next})
}
// ApplyImmediate derives audit and effective timestamps from trusted server time.
// Scheduling and retroactive HR corrections require the temporal workflow.
func (s *Service) ApplyImmediate(ctx context.Context, actor Actor, plan orgunit.ChangeSet) (Result, error) {
    if !plan.EffectiveAt.IsZero() {
        return Result{}, orgunit.ErrDeferredEffectiveTime
    }
    now := time.Now().UTC()
    plan.EffectiveAt, plan.RecordedAt = now, now
    return s.Apply(ctx,actor,plan)
}

func requestHash(plan orgunit.ChangeSet) (string,error) {
    // recorded_at comes from the trusted server clock and is deliberately not
    // part of the logical idempotency fingerprint; retries need not share it.
    payload,err:=json.Marshal(struct{
        ID string
        OrganizationID string
        ExpectedRevision uint64
                SourceID string
        Changes []orgunit.Change
    }{plan.ID,plan.OrganizationID,plan.ExpectedRevision,plan.SourceID,plan.Changes})
    if err!=nil{return "",fmt.Errorf("marshal changeset: %w",err)}
    digest:=sha256.Sum256(payload)
    return hex.EncodeToString(digest[:]),nil
}
