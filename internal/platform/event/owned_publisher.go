package event

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"sync"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

// OwnedPublisher publishes with one stable producer lane per StatefulSet slot.
// Use one shared instance per process.
type OwnedPublisher struct {
	publisher *Publisher
	owner     *Transactor
	ownership PodOwnership
	guard     string
	producer  string
	lane      chan struct{}
}

// NewOwnedPublisher explicitly enables stable publication for a bound Pod.
// guardTable is a YDB row table with key replica Utf8 and revision Uint64.
// The publisher needs write access to this table and read-only access to the
// separate ownership table. Neither table is created automatically.
func NewOwnedPublisher(owner *Transactor, cfg Config, ownership PodOwnership, guardTable string) (*OwnedPublisher, error) {
	if owner == nil || ownership.generation == 0 || ownership.slot.validate() != nil ||
		!ownerTablePath.MatchString(guardTable) || guardTable == ownership.slot.TablePath {
		return nil, ErrInvalidConfiguration
	}
	publisher, err := NewPublisher(owner.db, cfg)
	if err != nil {
		return nil, err
	}
	return &OwnedPublisher{
		publisher: publisher, owner: owner, ownership: ownership, guard: guardTable,
		producer: fmt.Sprintf("m8-%x", sha256.Sum256([]byte(ownership.slot.key()))),
		lane:     make(chan struct{}, 1),
	}, nil
}

// WithinTransaction verifies ownership and stages one batch per physical topic.
// The SDK owns commit and retries. The callback, TxActor and publish function
// are scoped to an attempt. Handle uncertain outcomes with a command receipt.
func (p *OwnedPublisher) WithinTransaction(ctx context.Context, fn func(context.Context, query.TxActor, func(string, ...*eventv1.Event) error) error) error {
	if p == nil || p.owner == nil || fn == nil {
		return ErrInvalidConfiguration
	}
	select {
	case p.lane <- struct{}{}:
		defer func() { <-p.lane }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) (err error) {
		if err := p.ownership.Check(ctx, tx); err != nil {
			return err
		}
		if err := p.advanceGuard(ctx, tx); err != nil {
			return err
		}
		var mu sync.Mutex
		var closed bool
		var publishErr error
		seen := make(map[string]struct{})
		publish := func(topic string, events ...*eventv1.Event) error {
			mu.Lock()
			defer mu.Unlock()
			if closed {
				return ErrReplicaOwnership
			}
			if publishErr != nil {
				return publishErr
			}
			path := p.publisher.cfg.Topics[topic]
			if _, duplicate := seen[path]; duplicate && len(events) > 0 {
				publishErr = fmt.Errorf("%w: duplicate batch for %q", ErrInvalidTopic, topic)
				return publishErr
			}
			publishErr = p.publisher.publish(ctx, tx, topic, p.producer, events...)
			if len(events) > 0 && publishErr == nil {
				seen[path] = struct{}{}
			}
			return publishErr
		}
		defer func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if err == nil {
				err = publishErr
			}
		}()
		return fn(ctx, tx, publish)
	})
}

// The guard makes overlapping attempts with the same slot conflict even
// across process restarts where the Pod UID and its generation are unchanged.
// It is independent of YDB's producer sequence numbers.
func (p *OwnedPublisher) advanceGuard(ctx context.Context, tx query.TxActor) error {
	key := p.ownership.slot.key()
	row, err := tx.QueryRow(ctx, declareReplicaKey+"SELECT revision FROM `"+p.guard+"` WHERE replica = $replica;",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(key).Build()))
	var revision uint64
	switch {
	case err == nil:
		if err := row.Scan(&revision); err != nil {
			return err
		}
	case !errors.Is(err, query.ErrNoRows):
		return err
	}
	if revision == math.MaxUint64 {
		return ErrReplicaOwnership
	}
	return tx.Exec(ctx, declareReplicaKey+"DECLARE $revision AS Uint64; "+
		"UPSERT INTO `"+p.guard+"` (replica, revision) VALUES ($replica, $revision);",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(key).Param("$revision").Uint64(revision+1).Build()))
}
