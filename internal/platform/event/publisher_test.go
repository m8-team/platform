package event

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

// Embed only to exercise paths that must finish before reaching the SDK.
type unusedTransaction struct{ query.TxActor }

func TestPublishPreflight(t *testing.T) {
	ctx := context.Background()
	tx := &unusedTransaction{}
	db := new(ydb.Driver)
	cfg := DefaultConfig(map[string]string{"resource": "/local/resource"})
	p, err := NewPublisher(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Topics["resource"] = "/local/replaced"
	cfg.Topics["unconfigured"] = "/local/extra"
	if p.cfg.Topics["resource"] != "/local/resource" {
		t.Fatal("publisher retained caller-owned map")
	}

	var absent *Publisher
	var nilTransaction *unusedTransaction
	for _, tc := range []struct {
		p    *Publisher
		tx   query.TxActor
		want error
	}{
		{absent, tx, ErrInvalidConfiguration},
		{p, nil, ErrInvalidConfiguration},
		{p, nilTransaction, ErrInvalidConfiguration},
	} {
		if err := tc.p.Publish(ctx, tc.tx, "resource"); !errors.Is(err, tc.want) {
			t.Fatalf("preflight: %v", err)
		}
	}
	if p.Stats() != (PublishStats{Attempts: 2, Errors: 2}) {
		t.Fatal("failed calls should count as attempts without staged events")
	}
	if err := p.Publish(ctx, tx, "unconfigured"); !errors.Is(err, ErrInvalidTopic) {
		t.Fatal(err)
	}
	if err := p.Publish(ctx, tx, "resource", nil); !errors.Is(err, ErrInvalidEvent) {
		t.Fatal(err)
	}
	stopped, cancel := context.WithCancel(ctx)
	cancel()
	if err := p.Publish(stopped, tx, "resource"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			if err := p.Publish(ctx, tx, "resource"); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
}

func TestInvalidConfigAndTransactor(t *testing.T) {
	db := new(ydb.Driver)
	for _, cfg := range []Config{
		{},
		DefaultConfig(map[string]string{"bad alias": "/local/events"}),
		{Topics: map[string]string{"events": "/local/events"}, MaxEventBytes: 2, MaxBatchEvents: 1, MaxBatchBytes: 1},
	} {
		if _, err := NewPublisher(db, cfg); !errors.Is(err, ErrInvalidConfiguration) {
			t.Fatalf("accepted invalid configuration: %v", err)
		}
	}
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
	if err := tr.WithinTransaction(context.Background(), nil); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatal(err)
	}
}
