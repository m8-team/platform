// Package event provides native YDB transactional event publication.
// It does not own business state, consumers, or command idempotency.
package event

import (
	"context"
	"fmt"
	"maps"
	"reflect"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
)

type Publisher struct {
	db    *ydb.Driver
	cfg   Config
	stats publishCounters
}

func NewPublisher(db *ydb.Driver, cfg Config) (*Publisher, error) {
	if db == nil {
		return nil, ErrInvalidConfiguration
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.Topics = maps.Clone(cfg.Topics)
	return &Publisher{db: db, cfg: cfg}, nil
}

// Publish stages one batch on the caller's transaction. A nil return is
// NOT a commit confirmation. The caller must return every error to the
// transaction coordinator. Events and their IDs are supplied by the caller;
// neither messages nor writers may be mutated/reused across attempts.
func (p *Publisher) Publish(ctx context.Context, tx query.TxActor, topic string, events ...*eventv1.Event) error {
	return p.publish(ctx, tx, topic, "", events...)
}

func (p *Publisher) publish(ctx context.Context, tx query.TxActor, topic, producer string, events ...*eventv1.Event) (result error) {
	var stagedEvents, stagedBytes int
	if p == nil {
		return ErrInvalidConfiguration
	}
	defer func() { p.stats.record(result, stagedEvents, stagedBytes) }()
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.db == nil || tx == nil || reflect.ValueOf(tx).Kind() == reflect.Pointer && reflect.ValueOf(tx).IsNil() {
		return ErrInvalidConfiguration
	}
	topicPath, ok := p.cfg.Topics[topic]
	if !ok {
		return ErrInvalidTopic
	}
	if len(events) == 0 {
		return nil
	}
	messages, stagedBytes, err := encodeBatch(ctx, p.cfg, events)
	if err != nil {
		return err
	}
	options := []topicoptions.MultiWriterOption{topicoptions.WithWriterPartitionByKey(topicoptions.BoundPartitionChooser())}
	if producer != "" {
		options = append(options, topicoptions.WithProducerIDPrefix(producer))
	}
	writer, err := p.db.Topic().StartTransactionalWriter(tx, topicPath,
		topicoptions.WithWriterMessageMaxBytesSize(p.cfg.MaxEventBytes),
		topicoptions.WithWriteToManyPartitions(options...),
	)
	if err != nil {
		return fmt.Errorf("start transactional event writer: %w", err)
	}
	if err := writer.Write(ctx, messages...); err != nil {
		return fmt.Errorf("write transactional events: %w", err)
	}
	stagedEvents = len(messages)
	// SDK flushes before commit and releases the writer on completion.
	return nil
}
