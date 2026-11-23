package authzen

import (
    "context"
    "errors"
    "testing"

    "github.com/m8-team/platform/internal/access/app/consistency"
    "github.com/m8-team/platform/internal/access/domain"
    accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

const testTenant="aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
type testPrincipalResolver struct { principal Principal; err error }
func (p testPrincipalResolver)AuthenticatedPrincipal(context.Context)(Principal,error){return p.principal,p.err}
type fakeRMRevision struct{revision uint64;err error}
func (f fakeRMRevision)AuthoritativeRevision(context.Context,domain.TenantID,domain.Resource)(uint64,error){return f.revision,f.err}
type fakeWatermark struct{revision uint64;err error}
func (f fakeWatermark)LoadWatermark(context.Context,domain.TenantID,domain.Resource)(consistency.Watermark,error){
    return consistency.Watermark{Revision:f.revision,ZedToken:"zed"},f.err
}
type fakePDP struct{calls int}
func (f *fakePDP)CheckPermission(_ context.Context,req domain.CheckPermissionRequest)(domain.PermissionDecision,error){
    f.calls++
    decision,err:=domain.NewPermissionDecision(domain.DecisionAllow,req.ModelRevision(),"allowed")
    if err!=nil{return domain.PermissionDecision{},err}
    return decision.WithZedToken("fresh")
}
func makeService(t *testing.T,revision uint64,sourceError error)(*Service,*fakePDP){
    t.Helper()
    tenant,_:=domain.NewTenantID(testTenant)
    subject,_:=domain.NewSubject("user","alice")
    checker:=&fakePDP{}
    barrier,_:=consistency.New(fakeRMRevision{revision:revision,err:sourceError},fakeWatermark{revision:1},checker)
    service,err:=New(testPrincipalResolver{principal:Principal{TenantID:tenant,Subject:subject}},barrier,"model-v1")
    if err!=nil{t.Fatal(err)}
    return service,checker
}
func eval(subject string)*accesspb.EvaluateAccessRequest{
    return &accesspb.EvaluateAccessRequest{
        Subject:&accesspb.Subject{Type:"user",Id:subject},
        Action:&accesspb.Action{Name:"m8.workspace.read"},
        Resource:&accesspb.Resource{Type:"workspace",Id:"11111111-1111-4111-8111-111111111111"},
    }
}
func TestAuthZENUsesTrustedBarrier(t *testing.T){
    service,checker:=makeService(t,1,nil)
    got,err:=service.EvaluateAccess(context.Background(),eval("alice"))
    if err!=nil || !got.Decision || checker.calls!=1{t.Fatalf("barrier bypassed: %+v %v calls=%d",got,err,checker.calls)}
    got,err=service.EvaluateAccess(context.Background(),eval("bob"))
    if status.Code(err)!=codes.PermissionDenied || got!=nil || checker.calls!=1{t.Fatalf("delegation allowed: %v",err)}
    service,checker=makeService(t,2,nil)
    got,err=service.EvaluateAccess(context.Background(),eval("alice"))
    if got!=nil || status.Code(err)!=codes.Unavailable || checker.calls!=0{t.Fatalf("stale RM allowed: %+v %v",got,err)}
    service,checker=makeService(t,1,errors.New("RM timeout"))
    got,err=service.EvaluateAccess(context.Background(),eval("alice"))
    if got!=nil || status.Code(err)!=codes.Unavailable || checker.calls!=0{t.Fatalf("source timeout allowed: %+v %v",got,err)}
}
func TestAuthZENRejectsUnknownCaller(t *testing.T){
    service,checker:=makeService(t,1,nil)
    service.resolver=testPrincipalResolver{err:errors.New("bad token")}
    _,err:=service.EvaluateAccess(context.Background(),eval("alice"))
    if status.Code(err)!=codes.Unauthenticated || checker.calls!=0{t.Fatalf("unauthenticated allowed %v",err)}
}
func TestAuthZENBatchRespectsShortCircuit(t *testing.T){
    service,checker:=makeService(t,1,nil)
    option:="permit_on_first_permit"
    response,err:=service.EvaluateAccessBatch(context.Background(),&accesspb.EvaluateAccessBatchRequest{
        Subject:eval("alice").Subject,Action:eval("alice").Action,Resource:eval("alice").Resource,
        Options:&accesspb.EvaluateAccessBatchOptions{EvaluationsSemantic:&option},
        Evaluations:[]*accesspb.Evaluation{{},{}},
    })
    if err!=nil || len(response.Evaluations)!=1 || checker.calls!=1{t.Fatalf("short-circuit failed %+v %v",response,err)}
}
