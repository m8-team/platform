package event

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
)

// CheckReadiness inspects ownership, guard schema and topics within the
// operation deadline. Transactional write permissions require deployment tests.
func (p *Publisher) CheckReadiness(ctx context.Context) error {
	if p == nil || p.owner == nil || p.lane == nil {
		return ErrInvalidConfiguration
	}
	ctx, cancel := context.WithTimeout(ctx, p.owner.timeout)
	defer cancel()
	if err := p.owner.db.Table().Do(ctx, func(ctx context.Context, session table.Session) error {
		description, err := session.DescribeTable(ctx, p.guard)
		if err != nil {
			return err
		}
		if !slices.Equal(description.PrimaryKey, []string{"replica"}) || description.StoreType == options.StoreTypeColumn ||
			description.TimeToLiveSettings != nil {
			return ErrInvalidConfiguration
		}
		var replica, revision bool
		for _, column := range description.Columns {
			switch column.Name {
			case "replica":
				replica = types.Equal(column.Type, types.TypeUTF8)
			case "revision":
				revision = types.Equal(column.Type, types.TypeUint64)
			}
		}
		if !replica || !revision {
			return ErrInvalidConfiguration
		}
		return nil
	}); err != nil {
		return fmt.Errorf("describe publication guard: %w", err)
	}
	if err := p.owner.db.Query().DoTx(ctx, func(ctx context.Context, tx query.TxActor) error {
		if err := p.ownership.check(ctx, tx); err != nil {
			return fmt.Errorf("check event ownership: %w", err)
		}
		revision, err := p.guardRevision(ctx, tx)
		if err != nil {
			return fmt.Errorf("read publication guard: %w", err)
		}
		if revision == math.MaxUint64 {
			return ErrReplicaOwnership
		}
		return nil
	}, query.WithTxSettings(query.TxSettings(query.WithSerializableReadWrite()))); err != nil {
		return err
	}
	for alias, path := range p.cfg.Topics {
		if err := ctx.Err(); err != nil {
			return err
		}
		description, err := p.owner.db.Topic().Describe(ctx, path)
		if err != nil {
			return fmt.Errorf("describe event topic %q: %w", alias, err)
		}
		if description.RetentionPeriod < p.cfg.MinTopicRetention {
			return fmt.Errorf("%w: retention below minimum for %q", ErrInvalidTopic, alias)
		}
		if err := checkPartitionBounds(description.Partitions); err != nil {
			return fmt.Errorf("event topic %q: %w", alias, err)
		}
	}
	return ctx.Err()
}

func checkPartitionBounds(partitions []topictypes.PartitionInfo) error {
	active := make([]topictypes.PartitionInfo, 0, len(partitions))
	for _, partition := range partitions {
		if partition.Active {
			active = append(active, partition)
		}
	}
	if len(active) == 0 {
		return fmt.Errorf("%w: no active partitions", ErrInvalidTopic)
	}
	slices.SortFunc(active, func(a, b topictypes.PartitionInfo) int {
		return bytes.Compare(a.FromBound, b.FromBound)
	})
	if len(active[0].FromBound) != 0 || len(active[len(active)-1].ToBound) != 0 {
		return fmt.Errorf("%w: partition bounds do not cover the full key space", ErrInvalidTopic)
	}
	for i := 1; i < len(active); i++ {
		prev, current := active[i-1], active[i]
		if len(current.FromBound) == 0 ||
			bytes.Compare(prev.FromBound, current.FromBound) >= 0 ||
			!bytes.Equal(prev.ToBound, current.FromBound) {
			return fmt.Errorf("%w: missing or incompatible partition bounds", ErrInvalidTopic)
		}
	}
	return nil
}
