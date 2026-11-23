package policyjournal

import (
    "context"
    "errors"
    "testing"

    "github.com/m8-team/platform/internal/access/app/projection"
    "github.com/m8-team/platform/internal/access/domain"
)
type fakeProjection struct{calls int;err error}
func(p *fakeProjection)Process(context.Context,projection.Event)error{p.calls++;return p.err}
type fakeCertification struct{calls int;err error}
func(c *fakeCertification)CommitPolicyJob(context.Context,uint64,projection.Key)error{c.calls++;return c.err}
func testPolicyEvent(t *testing.T)Event{
    t.Helper()
    tenant,_:=domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
    return Event{PolicyRevision:3,Projection:projection.Event{
        EventID:"mutation-42-binding-1",
        Key:projection.Key{TenantID:tenant,Kind:projection.KindRoleBinding,AggregateID:"11111111-1111-4111-8111-111111111111"},
        Revision:7,
    }}
}
func TestPolicyWorkerDoesNotAckBeforeDurableCertification(t *testing.T){
    p:=&fakeProjection{err:errors.New("SpiceDB failed")}
    c:=&fakeCertification{}
    w,_:=NewWorker(p,c)
    event:=testPolicyEvent(t)
    if err:=w.Process(context.Background(),event);err==nil || c.calls!=0{t.Fatal("certified unprojected permissions")}
    p.err=nil;c.err=errors.New("YDB checkpoint not durable")
    if err:=w.Process(context.Background(),event);err==nil || c.calls!=1{t.Fatal("acked without policy job completion")}
    c.err=nil
    if err:=w.Process(context.Background(),event);err!=nil || c.calls!=2 {t.Fatal("retry did not converge")}
}
func TestPolicyWorkerRequiresManifestRevision(t *testing.T){
    p:=&fakeProjection{}
    c:=&fakeCertification{}
    w,_:=NewWorker(p,c)
    event:=testPolicyEvent(t)
    event.PolicyRevision=0
    if err:=w.Process(context.Background(),event);!errors.Is(err,ErrInvalidPolicyDelivery)||p.calls!=0{t.Fatal("unmanifested event accepted")}
    event=testPolicyEvent(t)
    event.Projection.Key.Kind=projection.KindRMWorkspace
    if err:=w.Process(context.Background(),event);!errors.Is(err,ErrInvalidPolicyDelivery){t.Fatal("resource event accepted as Access grant")}
}
