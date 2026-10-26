package orgunit

import (
    "errors"
    "reflect"
    "testing"
)

func sample() []Unit {
    return []Unit{
        {ID:"root",OrganizationID:"alpha",State:StateActive},
        {ID:"it",OrganizationID:"alpha",ParentID:"root",State:StateActive},
        {ID:"eng",OrganizationID:"alpha",ParentID:"it",State:StateActive},
        {ID:"backend",OrganizationID:"alpha",ParentID:"eng",State:StateActive},
        {ID:"finance",OrganizationID:"alpha",ParentID:"root",State:StateActive},
    }
}
func mustBuild(t *testing.T, u []Unit) Structure { t.Helper(); s,e:=Build("alpha",1,u);if e!=nil{t.Fatal(e)};return s }
func hasPath(s Structure, ancestor, descendant string, depth uint32) bool {
    for _, r := range s.Closure {if r.AncestorID==ancestor && r.DescendantID==descendant && r.Depth==depth {return true}}
    return false
}
func TestBuildClosureDeterministic(t *testing.T) {
    units:=sample();a:=mustBuild(t,units)
    if len(a.Closure)!=12 {t.Fatalf("got %d closure rows, want 12",len(a.Closure))}
    if !hasPath(a,"root","backend",3) || !hasPath(a,"backend","backend",0) || hasPath(a,"finance","backend",0){t.Fatal("incorrect closure paths")}
    for i,j:=0,len(units)-1;i<j;i,j=i+1,j-1 {units[i],units[j]=units[j],units[i]}
    b:=mustBuild(t,units)
    if !reflect.DeepEqual(a,b){t.Fatal("order-dependent closure")}
}
func TestReparentKeepsStableIDsAndRevisesSubtree(t *testing.T) {
    old:=mustBuild(t,sample())
    next,err:=Reparent(old,"eng","finance",1)
    if err!=nil{t.Fatal(err)}
    if next.Revision!=2 || !hasPath(next,"finance","backend",2) || hasPath(next,"it","backend",2){t.Fatal("bad subtree reparent")}
    if !hasPath(old,"it","backend",2) || old.Revision!=1 {t.Fatal("input mutated")}
    again,err:=Reparent(next,"eng","finance",2)
    if err!=nil || again.Revision!=2 {t.Fatalf("idempotent no-op: %v",err)}
}
func TestRejectInvalidStructures(t *testing.T) {
    tests:=[]struct{name string;modify func([]Unit)[]Unit;want error}{
        {"duplicate",func(u []Unit)[]Unit{return append(u,u[0])},ErrDuplicateUnit},
        {"orphan",func(u []Unit)[]Unit{u[2].ParentID="ghost";return u},ErrMissingParent},
        {"cycle",func(u []Unit)[]Unit{u[1].ParentID="backend";return u},ErrHierarchyCycle},
        {"different organization",func(u []Unit)[]Unit{u[1].OrganizationID="other";return u},ErrInvalidUnit},
        {"self parent",func(u []Unit)[]Unit{u[1].ParentID="it";return u},ErrInvalidUnit},
        {"invalid state",func(u []Unit)[]Unit{u[1].State="UNKNOWN";return u},ErrInvalidUnit},
    }
    for _,tt:=range tests{t.Run(tt.name,func(t *testing.T){_,err:=Build("alpha",1,tt.modify(sample()));if !errors.Is(err,tt.want){t.Fatalf("got %v, want %v",err,tt.want)}})}
}
func TestReparentGuards(t *testing.T){
    tests:=[]struct{name,id,parent string;rev uint64;want error}{
        {"stale revision","eng","finance",3,ErrRevisionConflict},
        {"unknown unit","missing","finance",1,ErrUnitNotFound},
        {"unknown parent","eng","missing",1,ErrMissingParent},
        {"subtree cycle","it","backend",1,ErrHierarchyCycle},
    }
    for _,tt:=range tests{t.Run(tt.name,func(t *testing.T){base:=mustBuild(t,sample());_,err:=Reparent(base,tt.id,tt.parent,tt.rev);if !errors.Is(err,tt.want){t.Fatalf("got %v, want %v",err,tt.want)}})}
}
func TestClosedParentCannotReceiveNewChild(t *testing.T){
    units:=sample();units[4].State=StateClosed; s:=mustBuild(t,units)
    _,err:=Reparent(s,"eng","finance",1)
    if !errors.Is(err,ErrParentInactive){t.Fatalf("got %v",err)}
}
func TestOrganizationAndRevisionRequired(t *testing.T){
    for _,tc:=range []struct{org string;rev uint64}{{"",1},{"alpha",0}}{if _,err:=Build(tc.org,tc.rev,sample());!errors.Is(err,ErrInvalidUnit){t.Fatalf("got %v",err)}}
}
