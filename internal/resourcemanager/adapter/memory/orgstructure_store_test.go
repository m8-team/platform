package memory

import (
    "context"
    "errors"
    "sync"
    "testing"
    "time"
    "github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)
type allowOrgGate struct{}
func(allowOrgGate)Authorize(context.Context,orgstructure.Actor,orgunit.ChangeSet)error{return nil}
func(allowOrgGate)CheckImpact(context.Context,orgstructure.Actor,orgunit.Structure,orgunit.Structure,orgunit.ChangeSet)error{return nil}
type denyOrgGate struct{}
func(denyOrgGate)Authorize(context.Context,orgstructure.Actor,orgunit.ChangeSet)error{return errors.New("denied")}
func(denyOrgGate)CheckImpact(context.Context,orgstructure.Actor,orgunit.Structure,orgunit.Structure,orgunit.ChangeSet)error{return nil}
func setupOrgStructure(t *testing.T)(*OrgStructureStore,*orgstructure.Service){
    t.Helper();s:=NewOrgStructureStore()
    v,e:=orgunit.Build("org",1,[]orgunit.Unit{
        {ID:"root",OrganizationID:"org",State:orgunit.StateActive,DisplayName:"Root"},
        {ID:"a",OrganizationID:"org",ParentID:"root",State:orgunit.StateActive,DisplayName:"A"},
    })
    if e!=nil{t.Fatal(e)}
    if e=s.Seed("tenant",v);e!=nil{t.Fatal(e)}
    svc,e:=orgstructure.New(s,allowOrgGate{});if e!=nil{t.Fatal(e)}
    return s,svc
}
func renamePlan(id,name string)orgunit.ChangeSet{
    now:=time.Date(2026,10,27,10,0,0,0,time.UTC)
    return orgunit.ChangeSet{ID:id,OrganizationID:"org",ExpectedRevision:1,EffectiveAt:now,RecordedAt:now,Changes:[]orgunit.Change{{Kind:orgunit.ChangeRename,UnitID:"a",DisplayName:name}}}
}
func TestOrgStoreAtomicReplayAndEvent(t *testing.T){
    s,svc:=setupOrgStructure(t);actor:=orgstructure.Actor{TenantID:"tenant",PrincipalID:"admin"}
    r,e:=svc.Apply(context.Background(),actor,renamePlan("cs1","New A"))
    if e!=nil||r.Structure.Revision!=2{t.Fatalf("apply %v %#v",e,r)}
    again,e:=svc.Apply(context.Background(),actor,renamePlan("cs1","New A"))
    if e!=nil||!again.Replayed||again.Structure.Revision!=2{t.Fatalf("replay %v %#v",e,again)}
    if len(s.Events("tenant","org"))!=1{t.Fatal("duplicate outbox")}
    _,e=svc.Apply(context.Background(),actor,renamePlan("cs1","Other"))
    if !errors.Is(e,orgstructure.ErrIdempotencyConflict){t.Fatalf("hash mismatch: %v",e)}
}
func TestOrgStoreConcurrentCAS(t *testing.T){
    s,svc:=setupOrgStructure(t);actor:=orgstructure.Actor{TenantID:"tenant",PrincipalID:"admin"}
    var wg sync.WaitGroup
    results:=make(chan error,2)
    for _,pair:=range [][2]string{{"x","X"},{"y","Y"}}{
        wg.Add(1)
        go func(id,name string){defer wg.Done();_,e:=svc.Apply(context.Background(),actor,renamePlan(id,name));results<-e}(pair[0],pair[1])
    }
    wg.Wait();close(results)
    ok,conflicts:=0,0
    for e:=range results{
        if e==nil{ok++}else if errors.Is(e,orgunit.ErrRevisionConflict){conflicts++}else{t.Fatal(e)}
    }
    if ok!=1||conflicts!=1||len(s.Events("tenant","org"))!=1{t.Fatalf("ok=%d conflicts=%d events=%d",ok,conflicts,len(s.Events("tenant","org")))}
}
func TestOrgStorePolicyFailClosed(t *testing.T){
    s,_:=setupOrgStructure(t);svc,_:=orgstructure.New(s,denyOrgGate{})
    _,e:=svc.Apply(context.Background(),orgstructure.Actor{TenantID:"tenant",PrincipalID:"p"},renamePlan("cs1","A2"))
    if e==nil{t.Fatal("unauthorized mutation")}
    v,_:=s.Load(context.Background(),"tenant","org")
    if v.Revision!=1{t.Fatal("unauthorized mutation committed")}
    if _,e=orgstructure.New(s,nil);!errors.Is(e,orgstructure.ErrPolicyNotConfigured){t.Fatal("policy bypass")}
}
func TestOrgStoreSnapshotIsolation(t *testing.T){
    s,_:=setupOrgStructure(t)
    v,_:=s.Load(context.Background(),"tenant","org")
    v.Units[0].DisplayName="mutated"
    v2,_:=s.Load(context.Background(),"tenant","org")
    if v2.Units[0].DisplayName=="mutated"{t.Fatal("aliased")}
}


func TestOrgStoreReplayIgnoresServerRecordingTime(t *testing.T) {
    _, svc := setupOrgStructure(t)
    actor := orgstructure.Actor{TenantID: "tenant", PrincipalID: "admin"}
    first := renamePlan("cs1", "New A")
    if _, err := svc.Apply(context.Background(), actor, first); err != nil { t.Fatal(err) }
    retry := renamePlan("cs1", "New A")
    retry.RecordedAt = retry.RecordedAt.Add(time.Minute)
    result, err := svc.Apply(context.Background(), actor, retry)
    if err != nil || !result.Replayed {t.Fatalf("retry: %#v %v", result, err)}
}
func TestOrgStoreTenantIsolation(t *testing.T) {
    s, svc := setupOrgStructure(t)
    actor := orgstructure.Actor{TenantID: "different-tenant", PrincipalID: "admin"}
    _, err := svc.Apply(context.Background(), actor, renamePlan("cs1", "New A"))
    if !errors.Is(err, orgstructure.ErrStoreNotFound) {t.Fatalf("tenant isolation: %v", err)}
    if len(s.Events("tenant", "org"))!=0 {t.Fatal("unexpected event")}
}
