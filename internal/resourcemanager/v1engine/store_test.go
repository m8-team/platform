package v1engine

import (
 "errors"
 "testing"
)
func must(t *testing.T,s *Store,r Record,k string)Record{t.Helper();v,e:=s.Create(r,k);if e!=nil{t.Fatal(e)};return v}

func TestOrganizationHierarchyAndEnvironmentIsolation(t *testing.T){
 s:=NewStore();tenant:=must(t,s,Record{Kind:Tenant,Name:"Account"},"tenant")
 a:=must(t,s,Record{Kind:Organization,Name:"Retail",Owner:Owner{Tenant,tenant.ID}},"orga")
 b:=must(t,s,Record{Kind:Organization,Name:"Logistics",Owner:Owner{Tenant,tenant.ID}},"orgb")
 group:=must(t,s,Record{Kind:Group,Name:"Holding",Owner:Owner{Tenant,tenant.ID}},"group")
 if err:=s.SetGroupLink(group.ID,a.ID);err!=nil{t.Fatal(err)}
 project:=must(t,s,Record{Kind:Project,Name:"OMS",Owner:Owner{Organization,a.ID}},"project")
 ws:=must(t,s,Record{Kind:Workspace,Name:"Engineering",Owner:Owner{Organization,a.ID}},"ws")
 _=must(t,s,Record{Kind:Project,Name:"Pricing",Owner:Owner{Workspace,ws.ID}},"project2")
 if _,err:=s.Create(Record{Kind:Environment,Name:"Invalid",Owner:Owner{Workspace,ws.ID},EnvironmentKey:"testing"},"bad");!errors.Is(err,ErrInvalid){t.Fatalf("invalid parent: %v",err)}
 envA:=must(t,s,Record{Kind:Environment,Name:"Testing",Owner:Owner{Organization,a.ID},EnvironmentKey:"testing"},"enva")
 envB:=must(t,s,Record{Kind:Environment,Name:"Testing",Owner:Owner{Organization,b.ID},EnvironmentKey:"testing"},"envb")
 agent:=must(t,s,Record{Kind:Descriptor,Name:"HR Agent",Owner:Owner{Organization,a.ID}},"agent")
 if _,err:=s.SetBinding(Owner{Descriptor,agent.ID},envB.ID,"v1","");!errors.Is(err,ErrInvalid){t.Fatalf("cross-organization binding: %v",err)}
 if _,err:=s.SetBinding(Owner{Descriptor,agent.ID},envA.ID,"v1","");err!=nil{t.Fatal(err)}
 if _,err:=s.MoveProject(project.ID,Owner{Workspace,ws.ID},1);err!=nil{t.Fatal(err)}
 if _,err:=s.MoveProject(project.ID,Owner{Organization,b.ID},2);!errors.Is(err,ErrInvalid){t.Fatalf("cross-organization move: %v",err)}
 if _,err:=s.ChangeState(Organization,a.ID,1,"DELETED");!errors.Is(err,ErrConflict){t.Fatalf("parent deletion: %v",err)}
 items,err:=s.List("",a.ID);if err!=nil||len(items)<4{t.Fatalf("list: %d %v",len(items),err)}
 if _,err:=s.List("","");!errors.Is(err,ErrInvalid){t.Fatalf("unscoped list: %v",err)}
 ancestry,err:=s.Resolve(Project,project.ID);if err!=nil||len(ancestry)!=4{t.Fatalf("ancestry: %+v %v",ancestry,err)}
 duplicate,err:=s.Create(Record{Kind:Tenant,Name:"Different"},"tenant");if err!=nil||duplicate.ID!=tenant.ID{t.Fatalf("idempotency: %+v %v",duplicate,err)}
}
func TestRevisionAndState(t *testing.T){
 s:=NewStore();r:=must(t,s,Record{Kind:Tenant,Name:"Account"},"1")
 if _,err:=s.Update(Tenant,r.ID,0,"Updated","");!errors.Is(err,ErrInvalid){t.Fatal(err)}
 next,err:=s.Update(Tenant,r.ID,1,"Updated","");if err!=nil||next.Revision!=2{t.Fatal(next,err)}
 if _,err:=s.Update(Tenant,r.ID,1,"Stale","");!errors.Is(err,ErrConflict){t.Fatal(err)}
}
