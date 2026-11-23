package resource

import (
	"errors"
	"testing"
)

func TestAuthoritativeSnapshotValidation(t *testing.T) {
	valid := Snapshot{Kind: Workspace, ID: "r1", TenantID: "t1", State: StateActive, Revision: 2, OwnerKind: "organization", OwnerID: "o1"}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	invalid := []Snapshot{
		{Kind: "unknown", ID: valid.ID, TenantID: valid.TenantID, State: valid.State, Revision: 1},
		{Kind: valid.Kind, ID: "", TenantID: valid.TenantID, State: valid.State, Revision: 1},
		{Kind: valid.Kind, ID: valid.ID, TenantID: "", State: valid.State, Revision: 1},
		{Kind: valid.Kind, ID: valid.ID, TenantID: valid.TenantID, State: "UNSPECIFIED", Revision: 1},
		{Kind: valid.Kind, ID: valid.ID, TenantID: valid.TenantID, State: valid.State, Revision: 0},
		{Kind: valid.Kind, ID: valid.ID, TenantID: valid.TenantID, State: valid.State, Revision: 1, OwnerKind: "organization"},
	}
	for i, snapshot := range invalid {
		if err := snapshot.Validate(); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("case %d: got %v", i, err)
		}
	}
	if !valid.Active() {
		t.Fatal("active snapshot not active")
	}
	valid.State = StateDeleted
	if err := valid.Validate(); err != nil || valid.Active() {
		t.Fatalf("tombstone lost: %v", err)
	}
}
