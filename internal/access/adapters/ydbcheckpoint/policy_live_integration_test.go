//go:build integration

package ydbcheckpoint

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

// This test uses a REAL disposable YDB instance started by Access CI and
// executes schema_policy_v2.sql without a hand-rolled substitute.
func livePolicyStore(t *testing.T) *Store {
	t.Helper()
	db, prefix := liveYDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ddl, err := os.ReadFile("schema_policy_v2.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, piece := range strings.Split(string(ddl), ";") {
		var clean []string
		for _, line := range strings.Split(piece, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				clean = append(clean, line)
			}
		}
		statement := strings.TrimSpace(strings.Join(clean, "\n"))
		if statement == "" {
			continue
		}
		if _, err = db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s;", prefix, statement)); err != nil {
			t.Fatalf("apply original policy consistency v2 migration: %v", err)
		}
	}
	for _, table := range []string{"policy_heads", "policy_mutations", "policy_jobs", "policy_request_ids"} {
		var n uint64
		q := fmt.Sprintf("PRAGMA TablePathPrefix(%q); SELECT COUNT(*) FROM %s;", prefix, "`m8-access/"+table+"`")
		if err = db.QueryRowContext(ctx, q).Scan(&n); err != nil || n != 0 {
			t.Fatalf("verify policy table %s: %v count=%d", table, err, n)
		}
	}
	store, err := New(db, prefix, "policy-ci-worker", 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
func publishLivePolicy(t *testing.T, store *Store, m policyjournal.Mutation) (uint64, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	revision, err := store.PublishPolicyMutationTx(ctx, tx, m)
	if err != nil {
		return 0, err
	}
	return revision, tx.Commit()
}
func certifyFakeGraph(t *testing.T, store *Store, key projection.Key, revision int64) {
	t.Helper()
	err := store.WithFenced(context.Background(), key, func(_ context.Context, cp projection.Checkpoint, epochFence ports.GraphFence) (projection.Checkpoint, error) {
		cp.Key = key
		cp.Revision = revision
		cp.EventID = "authoritative-ci-event"
		cp.ZedToken = fmt.Sprintf("ci-checkpoint-%d", revision)
		cp.FenceEpoch = epochFence.Epoch
		cp.UpdatedAt = time.Now().UTC()
		return cp, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLivePolicyJournalBlocksPartialFanoutAndCertifiesContiguousRevision(t *testing.T) {
	store := livePolicyStore(t)
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	first := projection.Key{TenantID: tenant, Kind: projection.KindRoleBinding, AggregateID: "11111111-1111-4111-8111-111111111111"}
	second := projection.Key{TenantID: tenant, Kind: projection.KindRoleBinding, AggregateID: "22222222-2222-4222-8222-222222222222"}
	bootstrap := policyjournal.Mutation{TenantID: tenant, ExpectedHead: 0, ID: "bootstrap-all-roles", Bootstrap: true, Jobs: []policyjournal.Job{{Key: first, RequiredRevision: 1}, {Key: second, RequiredRevision: 1}}}
	if _, err := store.LoadPolicyState(context.Background(), tenant); !errors.Is(err, consistency.ErrPolicyUninitialized) {
		t.Fatalf("non-bootstrapped policy journal was accepted: %v", err)
	}
	revision, err := publishLivePolicy(t, store, bootstrap)
	if err != nil || revision != 1 {
		t.Fatalf("bootstrap: %d %v", revision, err)
	}
	state, err := store.LoadPolicyState(context.Background(), tenant)
	if err != nil || state.Head != 1 || state.Applied != 0 {
		t.Fatalf("bootstrap enabled policy early: %+v %v", state, err)
	}
	if err = store.CommitPolicyJob(context.Background(), 1, first); !errors.Is(err, ErrPolicyCheckpointBehind) {
		t.Fatalf("fabricated uncertified graph completion: %v", err)
	}
	certifyFakeGraph(t, store, first, 1)
	if err = store.CommitPolicyJob(context.Background(), 1, first); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadPolicyState(context.Background(), tenant)
	if err != nil || state.Applied != 0 {
		t.Fatalf("first binding incorrectly certified whole role: %+v", state)
	}
	certifyFakeGraph(t, store, second, 1)
	if err = store.CommitPolicyJob(context.Background(), 1, second); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadPolicyState(context.Background(), tenant)
	if err != nil || state.Head != 1 || state.Applied != 1 {
		t.Fatalf("failed complete bootstrap: %+v %v", state, err)
	}
	change := policyjournal.Mutation{TenantID: tenant, ExpectedHead: 1, ID: "revoke-binding-one", Jobs: []policyjournal.Job{{Key: first, RequiredRevision: 2}}}
	rev2, err := publishLivePolicy(t, store, change)
	if err != nil || rev2 != 2 {
		t.Fatalf("publish revocation: %d %v", rev2, err)
	}
	state, _ = store.LoadPolicyState(context.Background(), tenant)
	if state.Head != 2 || state.Applied != 1 {
		t.Fatalf("revocation not visible before projection: %+v", state)
	}
	if err = store.CommitPolicyJob(context.Background(), 2, first); !errors.Is(err, ErrPolicyCheckpointBehind) {
		t.Fatalf("stale graph incorrectly certified revoke: %v", err)
	}
	certifyFakeGraph(t, store, first, 2)
	if err = store.CommitPolicyJob(context.Background(), 2, first); err != nil {
		t.Fatal(err)
	}
	if err = store.CommitPolicyJob(context.Background(), 2, first); err != nil {
		t.Fatalf("duplicate event failed: %v", err)
	}
	state, _ = store.LoadPolicyState(context.Background(), tenant)
	if state.Head != 2 || state.Applied != 2 {
		t.Fatalf("revocation not certified: %+v", state)
	}
	rev, err := publishLivePolicy(t, store, change)
	if err != nil || rev != 2 {
		t.Fatalf("idempotent policy mutation failed: %d %v", rev, err)
	}
	change.Jobs[0].RequiredRevision = 3
	if _, err := publishLivePolicy(t, store, change); !errors.Is(err, ErrPolicyConflict) {
		t.Fatalf("different payload accepted under reused id: %v", err)
	}
}

func TestLivePolicyJournalNeverSkipsOutOfOrderCompletedRevisions(t *testing.T) {
	store := livePolicyStore(t)
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	first := projection.Key{TenantID: tenant, Kind: projection.KindGrant, AggregateID: "11111111-1111-4111-8111-111111111111"}
	second := projection.Key{TenantID: tenant, Kind: projection.KindGrant, AggregateID: "22222222-2222-4222-8222-222222222222"}
	one := policyjournal.Mutation{TenantID: tenant, ExpectedHead: 0, ID: "bootstrap-grant", Bootstrap: true, Jobs: []policyjournal.Job{{Key: first, RequiredRevision: 1}}}
	two := policyjournal.Mutation{TenantID: tenant, ExpectedHead: 1, ID: "other-grant", Jobs: []policyjournal.Job{{Key: second, RequiredRevision: 1}}}
	if rev, err := publishLivePolicy(t, store, one); err != nil || rev != 1 {
		t.Fatalf("bootstrap: %d %v", rev, err)
	}
	if rev, err := publishLivePolicy(t, store, two); err != nil || rev != 2 {
		t.Fatalf("next: %d %v", rev, err)
	}
	certifyFakeGraph(t, store, second, 1)
	if err := store.CommitPolicyJob(context.Background(), 2, second); err != nil {
		t.Fatal(err)
	}
	state, err := store.LoadPolicyState(context.Background(), tenant)
	if err != nil || state.Head != 2 || state.Applied != 0 {
		t.Fatalf("policy revision jumped gap: %+v %v", state, err)
	}
	certifyFakeGraph(t, store, first, 1)
	if err := store.CommitPolicyJob(context.Background(), 1, first); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadPolicyState(context.Background(), tenant)
	if err != nil || state.Head != 2 || state.Applied != 2 {
		t.Fatalf("policy was not drained contiguously: %+v %v", state, err)
	}
}
