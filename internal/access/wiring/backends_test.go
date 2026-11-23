package wiring

import (
    "context"
    "testing"
    "time"

    "github.com/m8-team/platform/internal/access/adapters/authzen"
    "github.com/m8-team/platform/internal/access/app/consistency"
    "github.com/m8-team/platform/internal/access/domain"
)

func TestConcreteWiringRejectsMissingBackends(t *testing.T){
    tenant,_:=domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
    principal,_:=domain.NewSubject("user","alice")
    resolver:=testPrincipal{tenant:tenant,subject:principal}
    if _,err:=NewProductionPDP(context.Background(),resolver,nil,nil,nil,"model-v1");err==nil {
        t.Fatal("PDP initialized without YDB and SpiceDB")
    }
    if _,err:=NewResourceProjectionWorker(nil,nil,nil,nil,time.Now);err==nil {
        t.Fatal("projection initialized without durable store and graph writer")
    }
}

type testPrincipal struct {
    tenant domain.TenantID
    subject domain.Subject
}
var _ authzen.PrincipalResolver=testPrincipal{}
var _ consistency.ResourceRevisionReader=source{}
func(p testPrincipal)AuthenticatedPrincipal(context.Context)(authzen.Principal,error) {
    return authzen.Principal{TenantID:p.tenant,Subject:p.subject},nil
}
