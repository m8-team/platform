package orgunit

import ("errors"; "testing"; "time")

func fixture(t *testing.T)Structure{
    t.Helper()
    s,e:=Build("org",1,[]Unit{
        {ID:"root",OrganizationID:"org",State:StateActive,DisplayName:"Root"},
        {ID:"a",OrganizationID:"org",ParentID:"root",State:StateActive,DisplayName:"A"},
        {ID:"b",OrganizationID:"org",ParentID:"a",State:StateActive,DisplayName:"B"},
    })
    if e!=nil{t.Fatal(e)}
    return s
}
func plan()ChangeSet {
    at:=time.Date(2026,10,27,12,0,0,0,time.UTC)
    return ChangeSet{ID:"cs1",OrganizationID:"org",ExpectedRevision:1,EffectiveAt:at,RecordedAt:at}
}
func has(s Structure,anc,des string,depth uint32)bool {
    for _,r:=range s.Closure{
        if r.AncestorID==anc && r.DescendantID==des && r.Depth==depth{return true}
    }
    return false
}
func TestApplyChangeSet(t *testing.T){
    old:=fixture(t)
    p:=plan()
    p.Changes=[]Change{
        {Kind:ChangeCreate,UnitID:"c",ParentID:"root",DisplayName:"C"},
        {Kind:ChangeMove,UnitID:"a",ParentID:"c"},
    }
    got,err:=Apply(old,p)
    if err!=nil{t.Fatal(err)}
    if got.Revision!=2 || !has(got,"c","b",2) || has(old,"c","a",1) || !has(old,"root","b",2) {t.Fatalf("bad changeset: %#v",got)}
}
func TestApplyRejectsCycleAndInactiveParents(t *testing.T){
    p:=plan();p.Changes=[]Change{{Kind:ChangeMove,UnitID:"a",ParentID:"b"}}
    if _,err:=Apply(fixture(t),p);!errors.Is(err,ErrHierarchyCycle){t.Fatalf("cycle: %v",err)}
    p.Changes=[]Change{{Kind:ChangeClose,UnitID:"a"}}
    if _,err:=Apply(fixture(t),p);!errors.Is(err,ErrUnitHasActiveChildren){t.Fatalf("close parent: %v",err)}
}
func TestApplyRenameAndClose(t *testing.T){
    p:=plan()
    p.Changes=[]Change{{Kind:ChangeRename,UnitID:"a",DisplayName:"Renamed"},{Kind:ChangeClose,UnitID:"b"}}
    got,err:=Apply(fixture(t),p)
    if err!=nil{t.Fatal(err)}
    if got.Revision!=2 || got.Units[0].DisplayName!="Renamed" || got.Units[1].State!=StateClosed {t.Fatalf("unexpected %#v",got.Units)}
}
func TestApplyGuards(t *testing.T){
    tests:=[]struct{name string;mutate func(*ChangeSet);want error}{
        {"revision",func(p *ChangeSet){p.ExpectedRevision=2},ErrRevisionConflict},
        {"backdated",func(p *ChangeSet){p.EffectiveAt=p.EffectiveAt.Add(-time.Hour)},ErrDeferredEffectiveTime},
        {"dupe",func(p *ChangeSet){p.Changes=append(p.Changes,Change{Kind:ChangeRename,UnitID:"b",DisplayName:"Twice"})},ErrConflictingChanges},
        {"external",func(p *ChangeSet){p.SourceID="zup"},ErrSourceAuthority},
        {"missing",func(p *ChangeSet){p.Changes[0].UnitID="unknown"},ErrUnitNotFound},
    }
    for _,tt:=range tests{
        t.Run(tt.name,func(t *testing.T){
            p:=plan();p.Changes=[]Change{{Kind:ChangeRename,UnitID:"b",DisplayName:"Renamed"}}
            tt.mutate(&p)
            _,err:=Apply(fixture(t),p)
            if !errors.Is(err,tt.want){t.Fatalf("want %v, got %v",tt.want,err)}
        })
    }
}
func TestNoOpRevisionAndInputIsolation(t *testing.T){
    old:=fixture(t)
    p:=plan();p.Changes=[]Change{{Kind:ChangeMove,UnitID:"a",ParentID:"root"}}
    got,err:=Apply(old,p)
    if err!=nil||got.Revision!=1{t.Fatalf("no op: %v %#v",err,got)}
    got.Units[0].ParentID="garbage"
    if old.Units[0].ParentID=="garbage"{t.Fatal("aliased input")}
}
func TestExternalCreateRetainsAuthority(t *testing.T){
    p:=plan();p.SourceID="zup";p.Changes=[]Change{{Kind:ChangeCreate,UnitID:"z1",ParentID:"root",DisplayName:"ZUP OrgUnit"}}
    s,err:=Apply(fixture(t),p)
    if err!=nil{t.Fatal(err)}
    var found bool
    for _,u:=range s.Units{if u.ID=="z1"{found=true;if u.AuthoritativeSourceID!="zup"{t.Fatalf("lost source authority: %#v",u)}}}
    if !found{t.Fatal("missing new OrgUnit")}
}
