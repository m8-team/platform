package ydb

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
	_ "github.com/ydb-platform/ydb-go-sdk/v3"
)

// TestIntegrationYDBOrgStructure uses a REAL YDB database with the migration
// 001_init.sql + 002_org_structure.sql already applied. It is opt-in because
// the test intentionally writes resources into its scoped tenant namespace.
//
// Run with M8_RM_TEST_YDB_DSN=grpc://localhost:2136/local and
// M8_RM_TEST_YDB_PREFIX=/local/m8 (or the deployed table prefix).
func TestIntegrationYDBOrgStructure(t *testing.T) {
	dsn := os.Getenv("M8_RM_TEST_YDB_DSN")
	if dsn == "" {
		t.Skip("set M8_RM_TEST_YDB_DSN and apply YDB migrations for a live integration test")
	}
	prefix := os.Getenv("M8_RM_TEST_YDB_PREFIX")
	if prefix == "" {
		prefix = "/local/m8"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	db, err := sql.Open("ydb", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var pingErr error
    for attempt := 0; attempt < 30; attempt++ {
        pingErr = db.PingContext(ctx)
        if pingErr == nil {break}
        select {
        case <-ctx.Done(): t.Fatalf("YDB startup: %v",pingErr)
        case <-time.After(2*time.Second):
        }
    }
    if pingErr != nil {t.Fatalf("YDB connection unavailable: %v",pingErr)}
	if err := prepareIntegrationSchema(ctx, db, dsn, prefix); err != nil {
        t.Fatalf("YDB test schema: %v", err)
    }
    store, err := NewStore(db, prefix)
	if err != nil {
		t.Fatal(err)
	}
	tenantID, err := newID()
	if err != nil {
		t.Fatal(err)
	}
	orgID, err := newID()
	if err != nil {
		t.Fatal(err)
	}
	initial, err := orgunit.Build(orgID, 1, []orgunit.Unit{{ID: "root", OrganizationID: orgID, State: orgunit.StateActive, DisplayName: "Root"}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err = store.Initialize(ctx, tenantID, initial, now); err != nil {
		t.Fatalf("YDB Initialize: %v", err)
	}
	loaded, err := store.Load(ctx, tenantID, orgID)
	if err != nil || loaded.Revision != 1 || len(loaded.Closure) != 1 {
		t.Fatalf("YDB Load initial: %#v %v", loaded, err)
	}
	changeID, err := newID()
	if err != nil {
		t.Fatal(err)
	}
	plan := orgunit.ChangeSet{
		ID: changeID, OrganizationID: orgID, ExpectedRevision: 1,
		EffectiveAt: now, RecordedAt: now,
		Changes: []orgunit.Change{{Kind: orgunit.ChangeCreate, UnitID: "dept", ParentID: "root", DisplayName: "Finance"}},
	}
	next, err := orgunit.Apply(loaded, plan)
	if err != nil {
		t.Fatal(err)
	}
	req := orgstructure.CommitRequest{
		Actor: orgstructure.Actor{TenantID: tenantID, PrincipalID: "integration-tester"},
		Plan:  plan, RequestHash: "sha256-stable-integration-request", Snapshot: next,
	}
	got, err := store.Commit(ctx, req)
	if err != nil {
		t.Fatalf("YDB Commit: %v", err)
	}
	if got.Structure.Revision != 2 {
		t.Fatalf("want revision 2, got %#v", got)
	}
	active, err := store.Load(ctx, tenantID, orgID)
	if err != nil || active.Revision != 2 || len(active.Closure) != 3 {
		t.Fatalf("YDB Load published: %#v %v", active, err)
	}
	replay, found, err := store.FindIdempotent(ctx, tenantID, orgID, plan.ID, req.RequestHash)
	if err != nil || !found || !replay.Replayed || replay.Structure.Revision != 2 {
		t.Fatalf("YDB FindIdempotent: %#v found=%v err=%v", replay, found, err)
	}
	_, found, err = store.FindIdempotent(ctx, tenantID, orgID, plan.ID, "different-hash")
	if found || !errors.Is(err, orgstructure.ErrIdempotencyConflict) {
		t.Fatalf("idempotency mismatch accepted: %v", err)
	}
	if _, err = store.Commit(ctx, req); !errors.Is(err, orgunit.ErrRevisionConflict) {
		t.Fatalf("stale revision must fail: %v", err)
	}
}
