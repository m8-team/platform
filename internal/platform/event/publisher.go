// Package event provides native YDB transactional event publication.
package event

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"math"
	"sync"
	"sync/atomic"
	"time"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
)

// Publisher publishes with one stable producer lane per StatefulSet slot.
// Use one shared instance per process.
type Publisher struct {
	cfg                                     Config
	stats                                   publishCounters
	laneWait, laneErrors, ownershipRejected atomic.Uint64
	owner                                   *Transactor
	ownership                               PodOwnership
	guard                                   string
	producer                                string
	lane                                    chan struct{}
}

// NewPublisher explicitly enables stable publication for a bound Pod.
// guardTable is a YDB row table with key replica Utf8 and revision Uint64.
// The publisher needs write access to this table and read-only access to the
// separate ownership table. Neither table is created automatically.
func NewPublisher(owner *Transactor, cfg Config, ownership PodOwnership, guardTable string) (*Publisher, error) {
	if owner == nil || owner.db == nil || !validPodUID(ownership.uid) || ownership.generation == 0 || ownership.slot.validate() != nil ||
		!ownerTablePath.MatchString(guardTable) || guardTable == ownership.slot.TablePath {
		return nil, ErrInvalidConfiguration
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.Topics = maps.Clone(cfg.Topics)
	return &Publisher{
		cfg: cfg, owner: owner, ownership: ownership, guard: guardTable,
		producer: fmt.Sprintf("m8-%x", sha256.Sum256([]byte(ownership.slot.key()))),
		lane:     make(chan struct{}, 1),
	}, nil
}

// WithinTransaction verifies ownership and stages one batch per physical topic.
// The SDK owns commit and retries. The callback, TxActor and publish function
// are scoped to an attempt. Handle uncertain outcomes with a command receipt.
func (p *Publisher) WithinTransaction(ctx context.Context, fn func(context.Context, query.TxActor, func(string, ...*eventv1.Event) error) error) error {
	if p == nil || p.owner == nil || p.lane == nil || fn == nil {
		return ErrInvalidConfiguration
	}
	ctx, cancel := context.WithTimeout(ctx, p.owner.timeout)
	defer cancel()
	start := time.Now()
	select {
	case p.lane <- struct{}{}:
		p.laneWait.Add(uint64(time.Since(start)))
		defer func() { <-p.lane }()
	case <-ctx.Done():
		p.laneWait.Add(uint64(time.Since(start)))
		p.laneErrors.Add(1)
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		p.laneErrors.Add(1)
		return err
	}
	return p.owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) (err error) {
		if err := p.ownership.check(ctx, tx); err != nil {
			if errors.Is(err, ErrReplicaOwnership) {
				p.ownershipRejected.Add(1)
			}
			return err
		}
		if err := p.advanceGuard(ctx, tx); err != nil {
			return err
		}
		var mu sync.Mutex
		var closed atomic.Bool
		var publishErr error
		seen := make(map[string]struct{})
		publish := func(topic string, events ...*eventv1.Event) (result error) {
			var stagedBytes int
			defer func() { p.stats.record(result, len(events), stagedBytes) }()
			mu.Lock()
			defer mu.Unlock()
			if closed.Load() {
				return ErrReplicaOwnership
			}
			if publishErr != nil {
				return publishErr
			}
			path := p.cfg.Topics[topic]
			if _, duplicate := seen[path]; duplicate && len(events) > 0 {
				publishErr = fmt.Errorf("%w: duplicate batch for %q", ErrInvalidTopic, topic)
				return publishErr
			}
			stagedBytes, publishErr = p.publish(ctx, tx, topic, events...)
			if len(events) > 0 && publishErr == nil {
				seen[path] = struct{}{}
			}
			return publishErr
		}
		defer func() {
			closed.Store(true)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				err = publishErr
				if err == nil {
					err = ctx.Err()
				}
			}
		}()
		return fn(ctx, struct{ query.TxActor }{tx}, publish)
	})
}

// The guard makes overlapping attempts with the same slot conflict even
// across process restarts where the Pod UID and its generation are unchanged.
// It is independent of YDB's producer sequence numbers.
func (p *Publisher) advanceGuard(ctx context.Context, tx query.TxActor) error {
	revision, err := p.guardRevision(ctx, tx)
	if err != nil {
		return err
	}
	if revision == math.MaxUint64 {
		return ErrReplicaOwnership
	}
	return tx.Exec(ctx, declareReplicaKey+"DECLARE $revision AS Uint64; "+
		"UPSERT INTO `"+p.guard+"` (replica, revision) VALUES ($replica, $revision);",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(p.ownership.slot.key()).Param("$revision").Uint64(revision+1).Build()))
}

func (p *Publisher) guardRevision(ctx context.Context, tx query.TxActor) (uint64, error) {
	key := p.ownership.slot.key()
	row, err := tx.QueryRow(ctx, declareReplicaKey+"SELECT revision FROM `"+p.guard+"` WHERE replica = $replica;",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(key).Build()))
	var revision uint64
	switch {
	case err == nil:
		if err := row.Scan(&revision); err != nil {
			return 0, err
		}
	case !errors.Is(err, query.ErrNoRows):
		return 0, err
	}
	return revision, nil
}

func (p *Publisher) publish(ctx context.Context, tx query.TxActor, topic string, events ...*eventv1.Event) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	path, ok := p.cfg.Topics[topic]
	if !ok {
		return 0, ErrInvalidTopic
	}
	if len(events) == 0 {
		return 0, nil
	}
	messages, size, err := encodeBatch(ctx, p.cfg, events)
	if err != nil {
		return 0, err
	}
	writer, err := p.owner.db.Topic().StartTransactionalWriter(tx, path,
		topicoptions.WithWriterMessageMaxBytesSize(p.cfg.MaxEventBytes),
		topicoptions.WithWriteToManyPartitions(
			topicoptions.WithWriterPartitionByKey(topicoptions.BoundPartitionChooser()),
			topicoptions.WithProducerIDPrefix(p.producer),
		),
	)
	if err != nil {
		return 0, fmt.Errorf("start transactional event writer: %w", err)
	}
	if err := writer.Write(ctx, messages...); err != nil {
		return 0, fmt.Errorf("write transactional events: %w", err)
	}
	return size, nil
}
