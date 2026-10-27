package orgstructure

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

type bootstrapFake struct {
    initialized bool
    current orgunit.Structure
    attempts int
}
func(b *bootstrapFake) Load(_ context.Context,tenant,org string)(orgunit.Structure,error) {
    if !b.initialized{return orgunit.Structure{},ErrStoreNotFound}
    return b.current,nil
}
func(b *bootstrapFake) Initialize(_ context.Context,_ string,v orgunit.Structure,_ time.Time)error {
    b.attempts++
    if b.initialized{return errors.New("already initialized")}
    b.current=v;b.initialized=true;return nil
}
func TestBootstrapCreatesEmptyStructureOnce(t *testing.T){
    b:=&bootstrapFake{}
    ctx:=context.Background()
    for i:=0;i<2;i++{
        if err:=EnsureOrganizationInitialized(ctx,b,"tenant","org",time.Now());err!=nil{t.Fatal(err)}
    }
    if b.attempts!=1 || b.current.Revision!=1 || len(b.current.Units)!=0 || len(b.current.Closure)!=0 {
        t.Fatalf("invalid bootstrap: %#v",b)
    }
}
func TestBootstrapDoesNotResetExistingStructure(t *testing.T){
    b:=&bootstrapFake{}
    ctx:=context.Background()
    original,e:=orgunit.Build("org",5,[]orgunit.Unit{{ID:"dept",OrganizationID:"org",State:orgunit.StateActive}})
    if e!=nil{t.Fatal(e)}
    b.initialized=true;b.current=original
    if err:=EnsureOrganizationInitialized(ctx,b,"tenant","org",time.Now());err!=nil{t.Fatal(err)}
    if b.attempts!=0||b.current.Revision!=5{t.Fatal("bootstrap reset existing organization")}
}
func TestBootstrapRequiresScope(t *testing.T){
    if err:=EnsureOrganizationInitialized(context.Background(),&bootstrapFake{},"","org",time.Now());!errors.Is(err,orgunit.ErrInvalidUnit){t.Fatalf("unexpected: %v",err)}
}
