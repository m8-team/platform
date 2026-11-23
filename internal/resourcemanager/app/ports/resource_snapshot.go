package ports

import (
	"context"

	"github.com/m8-team/platform/internal/resourcemanager/domain/resource"
)

// ResourceSnapshotRepository is the owner-side read port for strongly
// consistent resource facts. Implementations MUST query authoritative state,
// never a Kafka event, Access projection or client-supplied revision.
type ResourceSnapshotRepository interface {
	Load(ctx context.Context, kind resource.Kind, id string) (resource.Snapshot, error)
}
