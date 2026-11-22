package filecheckpoint

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

func TestDurableCheckpointReloadAndRejectRollback(t *testing.T) {
	tn, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	k := projection.Key{TenantID: tn, Kind: projection.KindGrant, AggregateID: "11111111-1111-4111-8111-111111111111"}
	directory := t.TempDir()
	s, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	err = s.WithLocked(context.Background(), k, func(cp projection.Checkpoint) (projection.Checkpoint, error) {
		if cp.Revision != 0 {
			t.Fatal("not initial")
		}
		return projection.Checkpoint{Key: k, Revision: 1, EventID: "event-a", ZedToken: "zed-1", UpdatedAt: time.Now()}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	err = restarted.WithLocked(context.Background(), k, func(cp projection.Checkpoint) (projection.Checkpoint, error) {
		if cp.Revision != 1 || cp.ZedToken != "zed-1" {
			t.Fatalf("not durable: %+v", cp)
		}
		return projection.Checkpoint{Key: k, Revision: 0}, nil
	})
	if !errors.Is(err, projection.ErrInvalidCheckpoint) {
		t.Fatalf("expected rollback protection, got %v", err)
	}
}
func TestLockIsContextCancelable(t *testing.T) {
	tn, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	k := projection.Key{TenantID: tn, Kind: projection.KindGrant, AggregateID: "11111111-1111-4111-8111-111111111111"}
	store, _ := New(t.TempDir())
	held := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error)
	go func() {
		done <- store.WithLocked(context.Background(), k, func(cp projection.Checkpoint) (projection.Checkpoint, error) { close(held); <-release; return cp, nil })
	}()
	<-held
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()
	err := store.WithLocked(ctx, k, func(cp projection.Checkpoint) (projection.Checkpoint, error) {
		t.Fatal("concurrent lock entered")
		return cp, nil
	})
	close(release)
	if got := <-done; got != nil {
		t.Fatal(got)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestCheckpointReaderSeesDurableRevision(t *testing.T) {
	tn, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	k := projection.Key{TenantID: tn, Kind: projection.KindGrant, AggregateID: "11111111-1111-4111-8111-111111111111"}
	store, _ := New(t.TempDir())
	if err := store.WithLocked(context.Background(), k, func(c projection.Checkpoint) (projection.Checkpoint, error) {
		return projection.Checkpoint{Key: k, Revision: 1, EventID: "event", ZedToken: "zed-after-write"}, nil
	}); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadCheckpoint(context.Background(), k)
	if err != nil || got.Revision != 1 || got.ZedToken != "zed-after-write" {
		t.Fatalf("load: %+v error=%v", got, err)
	}
}
