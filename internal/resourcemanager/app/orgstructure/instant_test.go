package orgstructure_test

import (
    "context"
    "testing"

    "github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
    "github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

type allowImmediate struct{}
func (allowImmediate) Authorize(context.Context, orgstructure.Actor, orgunit.ChangeSet) error { return nil }
func (allowImmediate) CheckImpact(context.Context, orgstructure.Actor, orgunit.Structure, orgunit.Structure, orgunit.ChangeSet) error { return nil }

func TestApplyImmediateServerTimeAndReplay(t *testing.T) {
    store := memory.NewOrgStructureStore()
    baseline, err := orgunit.Build("org", 1, []orgunit.Unit{
        {ID: "root", OrganizationID: "org", State: orgunit.StateActive, DisplayName: "Root"},
    })
    if err != nil { t.Fatal(err) }
    if err := store.Seed("t", baseline); err != nil { t.Fatal(err) }
    svc, err := orgstructure.New(store, allowImmediate{})
    if err != nil { t.Fatal(err) }
    actor := orgstructure.Actor{TenantID: "t", PrincipalID: "admin"}
    plan := orgunit.ChangeSet{
        ID: "cs1", OrganizationID: "org", ExpectedRevision: 1,
        Changes: []orgunit.Change{{Kind: orgunit.ChangeRename, UnitID: "root", DisplayName: "New"}},
    }
    first, err := svc.ApplyImmediate(context.Background(), actor, plan)
    if err != nil || first.Structure.Revision != 2 { t.Fatalf("first: %v %#v", err, first) }
    retry, err := svc.ApplyImmediate(context.Background(), actor, plan)
    if err != nil || !retry.Replayed || retry.Structure.Revision != 2 { t.Fatalf("replay: %v %#v", err, retry) }
}
