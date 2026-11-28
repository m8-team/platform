package event

import (
	"context"
	"errors"
	"testing"
	"time"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

func TestPublisherLaneDeadline(t *testing.T) {
	owner, err := NewTransactor(new(ydb.Driver), 30*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig(map[string]string{"resource": "/local/resource"})
	claim := PodOwnership{
		slot: ReplicaSlot{"test", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "publisher-0", "/local/owners"},
		uid:  "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", generation: 1,
	}
	p, err := NewPublisher(owner, cfg, claim, "/local/guards")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Topics["resource"] = "/local/changed"
	if p.cfg.Topics["resource"] != "/local/resource" {
		t.Fatal("publisher retained caller-owned map")
	}
	p.lane <- struct{}{}
	fn := func(context.Context, query.TxActor, func(string, ...*eventv1.Event) error) error {
		t.Fatal("entered the transaction while lane was occupied")
		return nil
	}
	if err := p.WithinTransaction(t.Context(), fn); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	stopped, cancel := context.WithCancel(t.Context())
	cancel()
	if err := p.WithinTransaction(stopped, fn); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if stats := p.Stats(); stats.LaneInUse != 1 || stats.LaneCapacity != 1 || stats.LaneAcquireErrors != 2 || stats.LaneWait <= 0 || owner.Stats().Attempts != 0 {
		t.Fatal("waiting request affected lane or transaction admission")
	}
	<-p.lane
	if err := p.WithinTransaction(stopped, fn); !errors.Is(err, context.Canceled) || len(p.lane) != 0 {
		t.Fatalf("canceled admission leaked lane: %v", err)
	}
	if _, err := NewPublisher(owner, cfg, claim, claim.slot.TablePath); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("accepted shared ownership and guard tables: %v", err)
	}
}

func TestInvalidTransactor(t *testing.T) {
	db := new(ydb.Driver)
	if _, err := NewTransactor(nil, time.Second); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatal(err)
	}
	if _, err := NewTransactor(db, 0); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatal(err)
	}
	tr, err := NewTransactor(db, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := tr.WithinTransaction(t.Context(), nil); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatal(err)
	}
}
