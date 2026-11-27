package event

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

// Transactor is a small native Query SDK boundary for modules not yet using
// a query.TxActor-based unit of work. It does not own the driver.
type Transactor struct {
	db      *ydb.Driver
	timeout time.Duration
}

func NewTransactor(db *ydb.Driver, timeout time.Duration) (*Transactor, error) {
	if db == nil || timeout <= 0 {
		return nil, ErrInvalidConfiguration
	}
	return &Transactor{db: db, timeout: timeout}, nil
}

// WithinTransaction uses the SDK retry policy. Ambiguous commit errors are
// not blindly retried as idempotent. Resolve uncertain outcomes using a
// durable application command receipt before repeating a logical operation.
func (t *Transactor) WithinTransaction(ctx context.Context, fn func(context.Context, query.TxActor) error) error {
	if t == nil || t.db == nil || fn == nil {
		return ErrInvalidConfiguration
	}
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.db.Query().DoTx(ctx, fn,
		query.WithTxSettings(query.TxSettings(query.WithSerializableReadWrite())),
		query.WithIdempotent(false),
	)
}
