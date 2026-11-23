//go:build integration

package ydbcheckpoint

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
	ydbsdk "github.com/ydb-platform/ydb-go-sdk/v3"
)

func liveYDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	if os.Getenv("M8_ACCESS_LIVE_YDB") != "1" {
		t.Skip("set M8_ACCESS_LIVE_YDB=1 for disposable YDB test")
	}
	dsn := os.Getenv("M8_ACCESS_TEST_YDB_DSN")
	prefix := os.Getenv("M8_ACCESS_TEST_YDB_PREFIX")
	if !strings.HasPrefix(prefix, "/local/m8_access_ci") || dsn == "" || !strings.Contains(dsn, "localhost:2136/local") {
		t.Fatal("integration tests only accept a disposable local YDB instance and /local/m8_access_ci prefix")
	}
	prefix = prefix + "_" + strings.ToLower(t.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)
	db, err := sql.Open("ydb", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for i := 0; i < 40; i++ {
		probeCtx, stop := context.WithTimeout(ctx, 3*time.Second)
		err = db.PingContext(probeCtx)
		stop()
		if err == nil {
			break
		}
		if i%5 == 0 {
			t.Logf("YDB readiness probe %d: %v", i+1, err)
		}
		if ctx.Err() != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		t.Fatalf("YDB health: %v", err)
	}
	sdk, err := ydbsdk.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer sdk.Close(ctx)
	for _, path := range []string{os.Getenv("M8_ACCESS_TEST_YDB_PREFIX"), prefix, prefix + "/m8-access"} {
		if err := sdk.Scheme().MakeDirectory(ctx, path); err != nil {
			if _, checkErr := sdk.Scheme().ListDirectory(ctx, path); checkErr != nil {
				t.Fatalf("create test dir %s: %v", path, err)
			}
		}
	}
	source, err := os.ReadFile(filepath.Join("schema.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range strings.Split(string(source), ";") {
		var nonComment []string
		for _, line := range strings.Split(fragment, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				nonComment = append(nonComment, line)
			}
		}
		statement := strings.TrimSpace(strings.Join(nonComment, "\n"))
		if statement == "" {
			continue
		}
		if _, err = db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s;", prefix, statement)); err != nil {
			t.Fatalf("apply exact YDB checkpoint migration: %v", err)
		}
	}
	for _, table := range []string{"projection_checkpoints", "resource_projection_watermarks"} {
		sqlText := fmt.Sprintf("PRAGMA TablePathPrefix(%q); SELECT COUNT(*) FROM "+"`m8-access/%s`;", prefix, table)
		var count uint64
		if err := db.QueryRowContext(ctx, sqlText).Scan(&count); err != nil {
			t.Fatalf("verify %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("unexpected rows in disposable %s", table)
		}
	}
	return db, prefix
}
func liveKey(t *testing.T, kind projection.Kind) projection.Key {
	t.Helper()
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	return projection.Key{TenantID: tenant, Kind: kind, AggregateID: "11111111-1111-4111-8111-111111111111"}
}
func TestLiveYDBMigrationAndTwoReplicaLease(t *testing.T) {
	db, prefix := liveYDB(t)
	ownerA, err := New(db, prefix, "replica-a", 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ownerB, err := New(db, prefix, "replica-b", 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	key := liveKey(t, projection.KindGrant)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		done <- ownerA.WithFenced(ctx, key, func(c context.Context, cp projection.Checkpoint, fence ports.GraphFence) (projection.Checkpoint, error) {
			close(started)
			<-c.Done()
			return cp, c.Err()
		})
	}()
	select {
	case <-started:
	case err := <-done:
		t.Fatalf("replica A did not acquire initial YDB lease: %v", err)
	case <-time.After(25 * time.Second):
		cancel()
		t.Fatal("timed out waiting for initial YDB lease")
	}
	err = ownerB.WithFenced(context.Background(), key, func(_ context.Context, cp projection.Checkpoint, _ ports.GraphFence) (projection.Checkpoint, error) {
		t.Error("replica B entered while A holds lease")
		return cp, nil
	})
	if !errors.Is(err, ErrLeaseBusy) {
		t.Fatalf("second replica must not acquire: %v", err)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel lease: %v", err)
	}
	time.Sleep(4 * time.Second)
	err = ownerB.WithFenced(context.Background(), key, func(_ context.Context, cp projection.Checkpoint, fence ports.GraphFence) (projection.Checkpoint, error) {
		if fence.Epoch < 2 {
			t.Errorf("fencing epoch did not advance: %d", fence.Epoch)
		}
		cp.Revision = 1
		cp.EventID = "lease-test"
		cp.ZedToken = "test-graph-token"
		cp.FenceEpoch = fence.Epoch
		cp.UpdatedAt = time.Now().UTC()
		return cp, nil
	})
	if err != nil {
		t.Fatalf("takeover: %v", err)
	}
	cp, err := ownerB.LoadCheckpoint(context.Background(), key)
	if err != nil || cp.FenceEpoch < 2 || cp.ZedToken == "" {
		t.Fatalf("durable checkpoint: %+v %v", cp, err)
	}
}
func TestLiveYDBarrierRejectsTakeoverUntilCertified(t *testing.T) {
	db, prefix := liveYDB(t)
	a, _ := New(db, prefix, "replica-a", 3*time.Second)
	b, _ := New(db, prefix, "replica-b", 3*time.Second)
	key := liveKey(t, projection.KindRMWorkspace)
	err := a.WithFenced(context.Background(), key, func(_ context.Context, cp projection.Checkpoint, fence ports.GraphFence) (projection.Checkpoint, error) {
		cp.Revision = 1
		cp.EventID = "rm-source-1"
		cp.ZedToken = "rm-zed-1"
		cp.FenceEpoch = fence.Epoch
		cp.UpdatedAt = time.Now().UTC()
		return cp, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.CommitResourceWatermark(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	resource, _ := domain.NewResource("workspace", key.AggregateID)
	w, err := a.LoadWatermark(context.Background(), key.TenantID, resource)
	if err != nil || w.Revision != 1 {
		t.Fatalf("missing watermark: %+v %v", w, err)
	}
	err = b.WithFenced(context.Background(), key, func(_ context.Context, cp projection.Checkpoint, fence ports.GraphFence) (projection.Checkpoint, error) {
		if _, loadErr := b.LoadWatermark(context.Background(), key.TenantID, resource); !errors.Is(loadErr, ErrCheckpointNotConverged) {
			t.Errorf("stale watermark accepted during takeover: %v", loadErr)
		}
		return cp, errors.New("simulate recovery crash")
	})
	if err == nil {
		t.Fatal("simulated crash unexpectedly succeeded")
	}
	if _, err = b.LoadWatermark(context.Background(), key.TenantID, resource); !errors.Is(err, ErrCheckpointNotConverged) {
		t.Fatalf("uncertified checkpoint was used: %v", err)
	}
}
