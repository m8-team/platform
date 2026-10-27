package orgstructure

import (
 "context"
 "errors"
 "strings"
 "testing"

 "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)
type hierarchyFixture struct{ value orgunit.Structure }
func (f *hierarchyFixture)Load(_ context.Context,_,_ string)(orgunit.Structure,error){return f.value,nil}
type hierarchyGate struct{ restricted map[string]bool; denied bool }
func(g hierarchyGate)AuthorizeOrganization(_ context.Context,_ Actor,_ string)error{
 if g.denied{return errors.New("not authorized")};return nil
}
func(g hierarchyGate)CanReadUnit(_ context.Context,_ Actor,_,unit string)(bool,error){return !g.restricted[unit],nil}
func makeQueryFixture(t *testing.T)(*hierarchyFixture,Actor){
 t.Helper()
 units:=[]orgunit.Unit{
 {ID:"root",OrganizationID:"org",State:orgunit.StateActive,DisplayName:"Root"},
 {ID:"a",OrganizationID:"org",ParentID:"root",State:orgunit.StateActive,DisplayName:"A"},
 {ID:"b",OrganizationID:"org",ParentID:"root",State:orgunit.StateClosed,DisplayName:"B"},
 {ID:"c",OrganizationID:"org",ParentID:"a",State:orgunit.StateActive,DisplayName:"C"},
 {ID:"d",OrganizationID:"org",ParentID:"a",State:orgunit.StateActive,DisplayName:"D"},
 }
 structure,err:=orgunit.Build("org",3,units);if err!=nil{t.Fatal(err)}
 return &hierarchyFixture{value:structure},Actor{TenantID:"tenant",PrincipalID:"person"}
}
func newQueryFixture(t *testing.T,repo *hierarchyFixture,g hierarchyGate)*QueryService{
 t.Helper();svc,err:=NewQueryService(repo,g,[]byte(strings.Repeat("K",32)));if err!=nil{t.Fatal(err)}
 return svc
}
func TestHierarchyPaginationAuthorizedBeforePage(t *testing.T){
 repo,actor:=makeQueryFixture(t)
 svc:=newQueryFixture(t,repo,hierarchyGate{restricted:map[string]bool{"b":true}})
 first,err:=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:2})
 if err!=nil || len(first.Units)!=2 || first.Units[0].ID!="a" || first.Units[1].ID!="c" || first.NextPageToken=="" {t.Fatalf("first: %#v %v",first,err)}
 second,err:=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:2,PageToken:first.NextPageToken})
 if err!=nil||len(second.Units)!=2||second.Units[0].ID!="d"||second.Units[1].ID!="root"||second.NextPageToken!=""{t.Fatalf("second: %#v %v",second,err)}
}
func TestHierarchyAncestorsAndDescendants(t *testing.T){
 repo,actor:=makeQueryFixture(t);svc:=newQueryFixture(t,repo,hierarchyGate{restricted:map[string]bool{"root":true}})
 ancestors,err:=svc.Ancestors(context.Background(),actor,"org","c",3)
 if err!=nil||len(ancestors.Ancestors)!=2||ancestors.Ancestors[0].AncestorID!="c"||ancestors.Ancestors[0].Depth!=0||ancestors.Ancestors[1].AncestorID!="a"{t.Fatalf("ancestors: %#v %v",ancestors,err)}
 subtree,err:=svc.Descendants(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",AncestorID:"a",PageSize:1})
 if err!=nil||len(subtree.Descendants)!=1||subtree.Descendants[0].DescendantID!="c"||subtree.Descendants[0].Depth!=1||subtree.NextPageToken==""{t.Fatalf("subtree: %#v %v",subtree,err)}
 next,err:=svc.Descendants(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",AncestorID:"a",PageSize:1,PageToken:subtree.NextPageToken})
 if err!=nil||len(next.Descendants)!=1||next.Descendants[0].DescendantID!="d"{t.Fatalf("next: %#v %v",next,err)}
}
func TestHierarchyCursorTamperScopeAndRevision(t *testing.T){
 repo,actor:=makeQueryFixture(t);svc:=newQueryFixture(t,repo,hierarchyGate{})
 first,err:=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:1})
 if err!=nil{t.Fatal(err)}
 token:=first.NextPageToken
 _,err=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:1,PageToken:"Z"+token[1:]})
 if !errors.Is(err,ErrInvalidPageToken){t.Fatalf("tamper: %v",err)}
 _,err=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:1,PageToken:token,AncestorID:"root"})
 if !errors.Is(err,ErrInvalidPageToken){t.Fatalf("filter: %v",err)}
 other:=actor;other.PrincipalID="other"
 _,err=svc.List(context.Background(),other,ListUnitsRequest{OrganizationID:"org",PageSize:1,PageToken:token})
 if !errors.Is(err,ErrInvalidPageToken){t.Fatalf("principal: %v",err)}
 repo.value.Revision=4
 _,err=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:1,PageToken:token})
 if !errors.Is(err,orgunit.ErrRevisionConflict){t.Fatalf("revision: %v",err)}
}
func TestHierarchyGateFailClosed(t *testing.T){
 repo,actor:=makeQueryFixture(t)
 _,err:=NewQueryService(repo,nil,[]byte(strings.Repeat("k",32)))
 if !errors.Is(err,ErrReadNotConfigured){t.Fatal("nil gate accepted")}
 svc:=newQueryFixture(t,repo,hierarchyGate{denied:true})
 if _,err=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org"});err==nil{t.Fatal("denied tenant read")}
 svc=newQueryFixture(t,repo,hierarchyGate{restricted:map[string]bool{"a":true}})
 if _,err=svc.Get(context.Background(),actor,"org","a");!errors.Is(err,ErrUnitNotVisible){t.Fatalf("hidden unit: %v",err)}
 if _,err=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",AncestorID:"a"});!errors.Is(err,ErrUnitNotVisible){t.Fatalf("hidden parent: %v",err)}
}
func TestHierarchyQueryBadInputs(t *testing.T){
 repo,actor:=makeQueryFixture(t);svc:=newQueryFixture(t,repo,hierarchyGate{})
 if _,err:=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",PageSize:201});!errors.Is(err,ErrInvalidList){t.Fatalf("page size: %v",err)}
 if _,err:=svc.List(context.Background(),actor,ListUnitsRequest{OrganizationID:"org",State:"WRONG"});!errors.Is(err,ErrInvalidList){t.Fatalf("state: %v",err)}
 if _,err:=svc.Ancestors(context.Background(),actor,"org","c",2);!errors.Is(err,orgunit.ErrRevisionConflict){t.Fatalf("revision: %v",err)}
}
