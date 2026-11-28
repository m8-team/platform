package event_test

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/platform/event"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

var (
	_ func(*event.Transaction, context.Context, string, *table.QueryParameters) error              = (*event.Transaction).Exec
	_ func(*event.Transaction, context.Context, string, *table.QueryParameters) (query.Row, error) = (*event.Transaction).QueryRow
)

func TestTransactionDoesNotExposeSDKActor(t *testing.T) {
	tx := new(event.Transaction)
	if _, ok := any(tx).(query.Executor); ok {
		t.Fatal("scope exposes the SDK query options")
	}
	if _, ok := any(tx).(query.TxActor); ok {
		t.Fatal("scope exposes the SDK transaction")
	}
	if _, ok := any(tx).(interface{ CommitTx(context.Context) error }); ok {
		t.Fatal("scope exposes commit")
	}
	if _, ok := any(tx).(interface{ Rollback(context.Context) error }); ok {
		t.Fatal("scope exposes rollback")
	}
	if err := tx.Exec(t.Context(), "SELECT 1;", nil); !errors.Is(err, event.ErrTransactionClosed) {
		t.Fatalf("zero-value scope: %v", err)
	}
}
