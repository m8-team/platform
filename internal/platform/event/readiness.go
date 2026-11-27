package event

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"
)

// CheckReadiness checks configured topic metadata using the caller's startup
// deadline. It does not test transactional write permissions.
func (p *Publisher) CheckReadiness(ctx context.Context) error {
	if p == nil || p.db == nil {
		return ErrInvalidConfiguration
	}
	for alias, path := range p.cfg.Topics {
		if err := ctx.Err(); err != nil {
			return err
		}
		description, err := p.db.Topic().Describe(ctx, path)
		if err != nil {
			return fmt.Errorf("describe event topic %q: %w", alias, err)
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
