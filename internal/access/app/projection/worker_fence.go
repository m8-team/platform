package projection

import (
	"context"

	"github.com/m8-team/platform/internal/access/app/ports"
)

// FencedCheckpointStore delivers a renewable lease-scoped context. A failed or
// expired lease cancels this context and MUST reject a checkpoint commit. The
// caller must never use the unfenced WithLocked method for distributed writes.
type FencedCheckpointStore interface {
	WithFenced(context.Context, Key, func(context.Context, Checkpoint, ports.GraphFence) (Checkpoint, error)) error
}
