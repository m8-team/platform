package event

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicreader"
	"google.golang.org/protobuf/proto"
)

// The harness stops a real test node between staging and the commit call.
// The process conducting the disruption controls when the callback resumes.
func TestDisruptionYDB(t *testing.T) {
	control := os.Getenv("M8_EVENT_TEST_DISRUPTION_DIR")
	if control == "" {
		t.Skip("use testdata/ydb-disruption.sh for a disposable YDB")
	}
	mode := os.Getenv("M8_EVENT_TEST_DISRUPTION_MODE")
	if mode != "restart" && mode != "failover" {
		t.Fatal("expected restart or failover")
	}
	ctx, db, path := liveDatabase(t)
	table := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTopic(t, ctx, db, path, 1)
	p, err := NewPublisher(db, DefaultConfig(map[string]string{"events": path}))
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTransactor(db, 90*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	write := func(ctx context.Context, tx query.TxActor, e *eventv1.Event) error {
		params := ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()
		if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+table+" (id) VALUES ($id);",
			query.WithParameters(params)); err != nil {
			return err
		}
		return p.Publish(ctx, tx, "events", e)
	}
	next := func() *eventv1.Event {
		event := testEvent(t)
		event.Id = uuid.NewString()
		return event
	}
	publish := func(e *eventv1.Event) {
		t.Helper()
		if err := tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			return write(ctx, tx, e)
		}); err != nil {
			t.Fatal(err)
		}
	}
	reader := func() *topicreader.Reader {
		t.Helper()
		result, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	readID := func(msg *topicreader.Message) (string, error) {
		data, err := io.ReadAll(msg)
		if err != nil {
			return "", err
		}
		e := new(eventv1.Event)
		err = proto.Unmarshal(data, e)
		return e.GetId(), err
	}
	first, pending, interrupted, sentinel := next(), next(), next(), next()
	publish(first)
	before := reader()
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		batch, err := before.PopMessagesBatchTx(ctx, tx)
		if err != nil {
			return err
		}
		if len(batch.Messages) != 1 {
			return fmt.Errorf("expected a single pre-disruption event, got %d", len(batch.Messages))
		}
		id, err := readID(batch.Messages[0])
		if err != nil || id != first.Id {
			return fmt.Errorf("invalid pre-disruption event: %s, %v", id, err)
		}
		return nil
	})
	if closeErr := before.Close(ctx); closeErr != nil {
		t.Error(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	publish(pending)
	location := func() (uint32, error) {
		desc, err := db.Topic().Describe(ctx, path, topicoptions.IncludeLocation())
		if err != nil {
			return 0, err
		}
		if len(desc.Partitions) != 1 || desc.Partitions[0].Location.NodeID == 0 {
			return 0, errors.New("cannot determine active tablet node")
		}
		return desc.Partitions[0].Location.NodeID, nil
	}
	var oldNode uint32
	attempts := 0
	// Leave the original transaction unresolved while the harness removes
	// its target node. An SDK callback retry would generate a different
	// observation, so the fixture stops it and reconciles authoritative state.
	err = tr.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		attempts++
		if attempts > 1 {
			return errors.New("stopped retry of interrupted transaction for reconciliation")
		}
		if err := write(ctx, tx, interrupted); err != nil {
			return err
		}
		var err error
		oldNode, err = location()
		if err != nil {
			return err
		}
		ready := filepath.Join(control, "ready")
		tmp := ready + ".tmp"
		if err := os.WriteFile(tmp, []byte(strconv.FormatUint(uint64(oldNode), 10)), 0o600); err != nil {
			return err
		}
		if err := os.Rename(tmp, ready); err != nil {
			return err
		}
		return awaitDisruption(ctx, filepath.Join(control, "resume"))
	})
	if _, statErr := os.Stat(filepath.Join(control, "resume")); statErr != nil {
		t.Fatalf("harness did not complete disruption: %v, transaction error=%v", statErr, err)
	}
	confirmed := err == nil
	if mode == "restart" && confirmed {
		t.Fatal("transaction opened before restart unexpectedly committed")
	}
	t.Logf("mode=%s original_partition_node=%d callback_attempts=%d confirmed_commit=%t", mode, oldNode, attempts, confirmed)

	publish(sentinel)
	newNode, err := location()
	if err != nil {
		t.Fatal(err)
	}
	if mode == "failover" && newNode == oldNode {
		t.Fatalf("topic tablet stayed on stopped node %d", oldNode)
	}
	t.Logf("partition node after recovery=%d", newNode)
	checkState := func(e *eventv1.Event) bool {
		t.Helper()
		params := ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()
		row, err := db.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT COUNT(*) FROM "+table+" WHERE id = $id;",
			query.WithParameters(params))
		var count uint64
		if err != nil {
			t.Fatal(err)
		}
		if err := row.Scan(&count); err != nil || count > 1 {
			t.Fatalf("invalid state count %d, %v", count, err)
		}
		return count == 1
	}
	hadInterruptedState := checkState(interrupted)
	if !checkState(first) || !checkState(pending) || !checkState(sentinel) ||
		(mode == "restart" && hadInterruptedState) || (confirmed && !hadInterruptedState) {
		t.Fatal("unexpected state after recovery")
	}
	after := reader()
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := after.Close(closeCtx); err != nil {
			t.Error(err)
		}
	}()
	expected := []string{pending.Id}
	if hadInterruptedState {
		expected = append(expected, interrupted.Id)
	}
	expected = append(expected, sentinel.Id)
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for _, want := range expected {
		msg, err := after.ReadMessage(readCtx)
		if err != nil {
			t.Fatal(err)
		}
		if id, err := readID(msg); err != nil || id != want {
			t.Fatalf("lost, partial or duplicate event after recovery: got=%s, want=%s, err=%v", id, want, err)
		}
	}
	offset, _, err := topicEnds(ctx, db, path)
	if err != nil || offset != int64(1+len(expected)) {
		t.Fatalf("durable topic end after %s: %d, %v", mode, offset, err)
	}
	t.Logf("durable topic end=%d, restarted consumer resumed after committed offset, interrupted event committed=%t",
		offset, hadInterruptedState)
}

func awaitDisruption(ctx context.Context, path string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
