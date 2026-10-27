package ydb

import (
	"context"
	"database/sql"
	"errors"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
	"strings"
	"testing"
)

func TestStoreConfiguration(t *testing.T) {
	if _, e := NewStore(nil, "/local/m8"); !errors.Is(e, ErrInvalidStore) {
		t.Fatalf("nil DB: %v", e)
	}
	db := new(sql.DB)
	for _, s := range []string{"", "../etc", "/local/m8\nDROP", "/local/m8;foo"} {
		if _, e := NewStore(db, s); !errors.Is(e, ErrInvalidStore) {
			t.Fatalf("path %q accepted", s)
		}
	}
	store, e := NewStore(db, "/local/m8")
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(store.q("SELECT 1"), `PRAGMA TablePathPrefix("/local/m8");`) {
		t.Fatal("wrong prefix")
	}
}
func TestSnapshotIDsUnique(t *testing.T) {
	first, e := newID()
	if e != nil {
		t.Fatal(e)
	}
	next, e := newID()
	if e != nil {
		t.Fatal(e)
	}
	if first == next || len(first) != 32 || len(next) != 32 {
		t.Fatalf("invalid IDs %q %q", first, next)
	}
}
func TestStageEnforcesBoundBeforeDatabaseUse(t *testing.T) {
	store := &Store{}
	units := make([]orgunit.Unit, maxStagedRows+1)
	for i := range units {
		units[i] = orgunit.Unit{ID: "x", OrganizationID: "org", State: orgunit.StateActive}
	}
	_, e := store.Load(context.Background(), "", "")
	if !errors.Is(e, orgunit.ErrInvalidUnit) {
		t.Fatalf("invalid scope: %v", e)
	}
	if e = store.stage(context.Background(), "tenant", "org", "snap", orgunit.Structure{Units: units}); !errors.Is(e, ErrSnapshotTooLarge) {
		t.Fatalf("stage bound: %v", e)
	}
}
