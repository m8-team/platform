package event

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
	"google.golang.org/protobuf/proto"
)

// TestIntegrationAtomicPublication requires a disposable real YDB database.
// It provisions isolated resources, never a shared production topic/table.
func TestIntegrationAtomicPublication(t *testing.T) {
	dsn, prefix := os.Getenv("M8_EVENT_TEST_YDB_DSN"), os.Getenv("M8_EVENT_TEST_YDB_PREFIX")
	if dsn == "" || prefix == "" {
		t.Skip("set M8_EVENT_TEST_YDB_DSN and M8_EVENT_TEST_YDB_PREFIX for disposable YDB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	db, err := ydb.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Close(closeCtx)
	}()

	// Validate the explicitly provided parent path before using it in YQL.
	cfg := DefaultConfig(map[string]string{"resource": prefix + "/event_" + uuid.NewString()})
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	topicPath := cfg.Topics["resource"]
	tablePath := topicPath + "_state"
	// YQL quoted identifiers are safe only if the prefix has no backticks.
	for _, c := range tablePath {
		if !(c == '/' || c == '_' || c == '-' || c == '.' ||
			c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			t.Fatal("invalid test prefix")
		}
	}
	table := "`" + tablePath + "`"
	if err := db.Query().Exec(ctx, "CREATE TABLE "+table+" (id Utf8 NOT NULL, PRIMARY KEY (id));"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Query().Exec(cleanup, "DROP TABLE "+table+";"); err != nil {
			t.Errorf("drop test table: %v", err)
		}
	}()
	const consumer = "event-test"
	if err := db.Topic().Create(ctx, topicPath,
		topicoptions.CreateWithMinActivePartitions(1),
		topicoptions.CreateWithConsumer(topictypes.Consumer{Name: consumer}),
	); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Topic().Drop(cleanup, topicPath); err != nil {
			t.Errorf("drop test topic: %v", err)
		}
	}()

	cfg.Topics["missing"] = topicPath + "_missing"
	cfg.TransactionTimeout = 12 * time.Second
	p, err := NewPublisher(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTransactor(db, cfg.TransactionTimeout)
	if err != nil {
		t.Fatal(err)
	}
	writeState := func(ctx context.Context, tx query.TxActor, id string) error {
		return tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+table+" (id) VALUES ($id);",
			query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(id).Build()))
	}
	assertCount := func(want uint64) {
		t.Helper()
		row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table+";")
		if err != nil {
			t.Fatal(err)
		}
		var got uint64
		if err := row.Scan(&got); err != nil || got != want {
			t.Fatalf("state count got=%d want=%d err=%v", got, want, err)
		}
	}
	makeEvent := func() *eventv1.Event {
		e := testEvent(t)
		e.Id = uuid.NewString()
		return e
	}
	rollback := errors.New("intentional rollback")
	rolledBack := makeEvent()
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, rolledBack.Id); err != nil {
			return err
		}
		if err := p.Publish(ctx, tx, "resource", rolledBack); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("expected rollback after staged event, got %v", err)
	}
	assertCount(0)

	bad := makeEvent()
	bad.Type = "invalid"
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, bad.Id); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "resource", bad)
	})
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected validation failure, got %v", err)
	}
	assertCount(0)

	missing := makeEvent()
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, missing.Id); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "missing", missing)
	})
	if err == nil {
		t.Fatal("unexpected success for nonexistent topic")
	}
	assertCount(0)

	first, second := makeEvent(), makeEvent()
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := writeState(ctx, tx, first.Id); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "resource", first, second)
	})
	if err != nil {
		t.Fatalf("commit state and events: %v", err)
	}
	assertCount(1)

	reader, err := db.Topic().StartReader(consumer, topicoptions.ReadTopic(topicPath))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = reader.Close(cleanup)
	}()
	expected := map[string]*eventv1.Event{first.Id: first, second.Id: second}
	consume := func(ctx context.Context, tx query.TxActor, fail bool) error {
		batch, err := reader.PopMessagesBatchTx(ctx, tx)
		if err != nil {
			return err
		}
		for _, msg := range batch.Messages {
			data, err := io.ReadAll(msg)
			if err != nil {
				return err
			}
			got := new(eventv1.Event)
			if err := proto.Unmarshal(data, got); err != nil {
				return err
			}
			if !proto.Equal(expected[got.Id], got) {
				return errors.New("unexpected or rolled-back event visible")
			}
			if err := writeState(ctx, tx, "consumed-"+got.Id); err != nil {
				return err
			}
		}
		if fail {
			return rollback
		}
		return nil
	}
	if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		return consume(ctx, tx, true)
	}); !errors.Is(err, rollback) {
		t.Fatalf("consumer rollback: %v", err)
	}
	assertCount(1)
	for i := 0; i < 2; i++ {
		row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+table+";")
		if err != nil {
			t.Fatal(err)
		}
		var count uint64
		if err := row.Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count == 3 {
			break
		}
		if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			return consume(ctx, tx, false)
		}); err != nil {
			t.Fatal(err)
		}
	}
	assertCount(3)
	// A successful transactional consumer must have advanced its offset.
	// An additional visible message (including rolled-back events) is a failure.
	readCtx, stop := context.WithTimeout(ctx, time.Second)
	defer stop()
	if msg, err := reader.ReadMessage(readCtx); err == nil || msg != nil {
		t.Fatal("unexpected extra event")
	} else if readCtx.Err() != context.DeadlineExceeded {
		t.Fatalf("unexpected reader failure: %v", err)
	}
}
