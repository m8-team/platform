package projection

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
)

// Reconciliation deliberately uses bounded fan-out/fan-in on independent keys.
// Kafka/Outbox event delivery remains sequential, because acknowledging offsets
// out of source order can lose an uncommitted event on consumer restart.
const MaxReconcileWorkers = 32

var ErrInvalidReconcilePool = errors.New("reconciliation workers must be between 1 and 32")

func reconcileKeys(ctx context.Context, keys []Key, workers int, process func(context.Context, Key) error) error {
	if ctx == nil || process == nil || workers <= 0 || workers > MaxReconcileWorkers {
		return ErrInvalidReconcilePool
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Validate and deduplicate before executing any side effects. A single
	// aggregate never runs concurrently within this reconciliation batch.
	seen := make(map[Key]struct{}, len(keys))
	unique := make([]Key, 0, len(keys))
	for _, key := range keys {
		if err := key.Validate(); err != nil {
			return err
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, key)
	}
	if len(unique) == 0 {
		return nil
	}
	if workers > len(unique) {
		workers = len(unique)
	}

	group, workerCtx := errgroup.WithContext(ctx)
	jobs := make(chan Key)
	group.Go(func() error {
		defer close(jobs)
		for _, key := range unique {
			select {
			case <-workerCtx.Done():
				return workerCtx.Err()
			case jobs <- key:
			}
		}
		return nil
	})
	for i := 0; i < workers; i++ {
		group.Go(func() error {
			for {
				select {
				case <-workerCtx.Done():
					return workerCtx.Err()
				case key, ok := <-jobs:
					if !ok {
						return nil
					}
					if err := process(workerCtx, key); err != nil {
						return err
					}
				}
			}
		})
	}
	return group.Wait()
}

// ReconcileKeys parallelizes independent aggregate repairs while preserving
// Worker.ReconcileKey's per-key durable checkpoint and fencing requirements.
func (w *Worker) ReconcileKeys(ctx context.Context, keys []Key, workers int) error {
	if w == nil {
		return ErrInvalidProjection
	}
	return reconcileKeys(ctx, keys, workers, w.ReconcileKey)
}

// ReconcileKeys only accepts Resource Manager-owned graph aggregates. Each
// successful key also commits the independent RM watermark before returning.
func (w *ResourceWorker) ReconcileKeys(ctx context.Context, keys []Key, workers int) error {
	if w == nil || w.worker == nil || w.watermarks == nil {
		return ErrResourceWatermarkStore
	}
	for _, key := range keys {
		if !IsResourceKind(key.Kind) {
			return ErrInvalidEvent
		}
	}
	return reconcileKeys(ctx, keys, workers, w.ReconcileKey)
}
