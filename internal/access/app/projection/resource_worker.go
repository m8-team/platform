package projection

import (
	"context"
	"errors"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
)

var ErrResourceWatermarkStore = errors.New("resource watermark commit store required")

// ResourceWatermarkCommitter derives progress from a fenced and durable
// checkpoint, never from caller-supplied revisions or untrusted event bodies.
type ResourceWatermarkCommitter interface {
	CommitResourceWatermark(context.Context, Key) error
}

// ResourceWorker couples graph projection and its trusted YDB watermark.
// ACK must happen only after BOTH durable operations have succeeded.
type ResourceWorker struct {
	worker     *Worker
	watermarks ResourceWatermarkCommitter
}

func NewResourceWorker(projector *Projector, source SnapshotReader, store CheckpointStore,
	watermarks ResourceWatermarkCommitter, now func() time.Time) (*ResourceWorker, error) {
	if watermarks == nil {
		return nil, ErrResourceWatermarkStore
	}
	worker, err := NewWorker(projector, source, store, now)
	if err != nil {
		return nil, err
	}
	if _, ok := store.(FencedCheckpointStore); !ok {
		return nil, ports.ErrInvalidGraphFence
	}
	return &ResourceWorker{worker: worker, watermarks: watermarks}, nil
}
func (w *ResourceWorker) Process(ctx context.Context, event Event) error {
	if w == nil || !IsResourceKind(event.Key.Kind) {
		return ErrInvalidEvent
	}
	if err := w.worker.Process(ctx, event); err != nil {
		return err
	}
	return w.watermarks.CommitResourceWatermark(ctx, event.Key)
}
func (w *ResourceWorker) ReconcileKey(ctx context.Context, key Key) error {
	if w == nil || !IsResourceKind(key.Kind) {
		return ErrInvalidEvent
	}
	if err := w.worker.ReconcileKey(ctx, key); err != nil {
		return err
	}
	return w.watermarks.CommitResourceWatermark(ctx, key)
}
func (w *ResourceWorker) Run(ctx context.Context, feed EventFeed) error {
	if w == nil || feed == nil {
		return ErrInvalidEvent
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		delivery, err := feed.Next(ctx)
		if err != nil {
			return err
		}
		if delivery == nil {
			return ErrInvalidEvent
		}
		if err := w.Process(ctx, delivery.Event()); err != nil {
			return err
		}
		if err := delivery.Ack(ctx); err != nil {
			return err
		}
	}
}
