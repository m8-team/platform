// Package event provides native YDB transactional event publication.
// It does not own business state, consumers, or command idempotency.
package event

import (
	"context"
	"fmt"
	"maps"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
)

type Publisher struct {
	db  *ydb.Driver
	cfg Config
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
	if err := ctx.Err(); err != nil {
		return err
	}
	if p == nil || p.db == nil || tx == nil {
		return ErrInvalidConfiguration
	}
	topicPath, ok := p.cfg.Topics[topic]
	if !ok {
		return ErrInvalidTopic
	}
	if len(events) == 0 {
		return nil
	}
	messages, err := encodeBatch(ctx, p.cfg, events)
	if err != nil {
		return err
	}
	// Unique writer-attempt identity avoids shared producer fencing across
	// concurrent transactions and replicas. It is not logical deduplication.
	producer, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("event producer identity: %w", err)
	}
	writer, err := p.db.Topic().StartTransactionalWriter(tx, topicPath,
		topicoptions.WithWriterMessageMaxBytesSize(p.cfg.MaxEventBytes),
		topicoptions.WithWriteToManyPartitions(
			topicoptions.WithProducerIDPrefix(producer.String()),
			topicoptions.WithWriterPartitionByKey(topicoptions.BoundPartitionChooser()),
		),
	)
	if err != nil {
		return fmt.Errorf("start transactional event writer: %w", err)
	}
	if err := writer.Write(ctx, messages...); err != nil {
		return fmt.Errorf("write transactional events: %w", err)
	}
	// SDK flushes before commit and releases the writer on completion.
	return nil
}
