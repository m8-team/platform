package event

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicreader"
)

var ErrTransactionClosed = errors.New("event transaction scope has ended")

// Transaction is valid for one callback attempt. Use it synchronously and
// return when the callback context is canceled. Queries must be trusted YQL.
type Transaction struct {
	attempt *transactionAttempt
}

// Keep lifecycle state shared even if application code copies Transaction.
type transactionAttempt struct {
	mu        sync.Mutex
	closed    atomic.Bool
	ctx       context.Context
	native    query.TxActor
	publisher *Publisher
	failure   error
	published map[string]struct{}
}

func (tx *Transaction) lock() (*transactionAttempt, error) {
	if tx == nil || tx.attempt == nil || tx.attempt.closed.Load() {
		return nil, ErrTransactionClosed
	}
	a := tx.attempt
	a.mu.Lock()
	if a.closed.Load() {
		a.mu.Unlock()
		return nil, ErrTransactionClosed
	}
	return a, nil
}

func (a *transactionAttempt) finish(err error) error {
	a.closed.Store(true)
	if !a.mu.TryLock() {
		a.mu.Lock()
		if err == nil {
			err = ErrTransactionClosed
		}
	}
	defer a.mu.Unlock()
	if err != nil {
		return err
	}
	if a.failure != nil {
		return a.failure
	}
	return a.ctx.Err()
}

// Exec executes a statement in this attempt. A nil params value means no parameters.
func (tx *Transaction) Exec(ctx context.Context, statement string, params *table.QueryParameters) error {
	a, err := tx.lock()
	if err != nil {
		return err
	}
	defer a.mu.Unlock()
	ctx, cancel := a.operationContext(ctx)
	defer cancel()
	return a.native.Exec(ctx, statement, query.WithParameters(params))
}

// QueryRow returns one materialized row. The SDK closes the result stream before returning.
func (tx *Transaction) QueryRow(ctx context.Context, statement string, params *table.QueryParameters) (query.Row, error) {
	a, err := tx.lock()
	if err != nil {
		return nil, err
	}
	defer a.mu.Unlock()
	ctx, cancel := a.operationContext(ctx)
	defer cancel()
	return a.native.QueryRow(ctx, statement, query.WithParameters(params))
}

func (a *transactionAttempt) operationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	operation, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(a.ctx, cancel)
	if a.ctx.Err() != nil {
		cancel()
	}
	return operation, func() {
		stop()
		cancel()
	}
}

// Publish stages one nonempty batch per physical topic. Any publication error
// aborts the attempt even if the callback ignores it.
func (tx *Transaction) Publish(topic string, events ...*eventv1.Event) error {
	a, err := tx.lock()
	if err != nil {
		return err
	}
	defer a.mu.Unlock()
	p := a.publisher
	var size int
	defer func() { p.stats.record(a.failure, len(events), size) }()
	if a.failure != nil {
		return a.failure
	}
	path := p.cfg.Topics[topic]
	if _, duplicate := a.published[path]; duplicate && len(events) > 0 {
		a.failure = fmt.Errorf("%w: duplicate batch for %q", ErrInvalidTopic, topic)
		return a.failure
	}
	size, a.failure = p.publish(a.ctx, a.native, topic, events...)
	if len(events) > 0 && a.failure == nil {
		if a.published == nil {
			a.published = make(map[string]struct{})
		}
		a.published[path] = struct{}{}
	}
	return a.failure
}

// ReadBatch stages input offsets in this attempt. Use a reader dedicated to
// transactional consumption and never acknowledge the returned messages manually.
// wait bounds lane occupation while waiting for input.
func (tx *Transaction) ReadBatch(ctx context.Context, reader *topicreader.Reader, wait time.Duration, opts ...topicreader.ReadBatchOption) (*topicreader.Batch, error) {
	a, err := tx.lock()
	if err != nil {
		return nil, err
	}
	defer a.mu.Unlock()
	if a.failure != nil {
		return nil, a.failure
	}
	if reader == nil || wait <= 0 {
		a.failure = ErrInvalidConfiguration
		return nil, a.failure
	}
	ctx, cancel := a.operationContext(ctx)
	defer cancel()
	ctx, stop := context.WithTimeout(ctx, wait)
	defer stop()
	batch, err := reader.PopMessagesBatchTx(ctx, a.native, opts...)
	a.failure = err
	return batch, err
}
