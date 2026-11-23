package wiring

import (
	"context"
	"net/http"
	"time"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
	rm "github.com/m8-team/platform/internal/access/adapters/resourcemanager"
	"github.com/m8-team/platform/internal/access/adapters/spicedb"
	"github.com/m8-team/platform/internal/access/adapters/ydbcheckpoint"
	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
)

// NewProductionPDP assembles the concrete RM YDB catalog reader, certified YDB
// projection store and official SpiceDB Go SDK. It does not implement caller
// identity verification; authentication middleware supplies the resolver.
func NewProductionPDP(ctx context.Context, principal authzen.PrincipalResolver,
	resourceCatalog rm.Catalog, checkpoint *ydbcheckpoint.Store,
	spicedbClient *spicedb.Client, revision string) (http.Handler, error) {
	if checkpoint == nil || spicedbClient == nil {
		return nil, authzen.ErrInvalidConfiguration
	}
	source, err := rm.New(resourceCatalog)
	if err != nil {
		return nil, err
	}
	return NewAuthZENHandler(ctx, principal, source, checkpoint, checkpoint, spicedbClient, revision)
}

// NewResourceProjectionWorker connects authoritative RM snapshots to fenced
// YDB checkpoint and SpiceDB graph projection. The caller wires the trusted
// event source and starts worker.Run; no event offset is ACKed before YDB
// confirms both checkpoint and RM watermark progress.
func NewResourceProjectionWorker(resourceCatalog rm.Catalog,
	checkpoint *ydbcheckpoint.Store, spicedbClient *spicedb.Client,
	permissions ports.PermissionCatalog, now func() time.Time) (*projection.ResourceWorker, error) {
	if checkpoint == nil || spicedbClient == nil {
		return nil, projection.ErrInvalidProjection
	}
	source, err := rm.New(resourceCatalog)
	if err != nil {
		return nil, err
	}
	projector, err := projection.New(permissions, spicedbClient, now)
	if err != nil {
		return nil, err
	}
	return projection.NewResourceWorker(projector, source, checkpoint, checkpoint, now)
}

// NewAccessPolicyWorker uses the SAME fenced Access grant/role projector as v1,
// but commits a policy journal job before an event delivery may be acknowledged.
// The Access authoritative SnapshotReader and outbox transport are explicitly
// supplied by the owner; there is NO unverified SQL snapshot fallback.
func NewAccessPolicyWorker(source projection.SnapshotReader,
	checkpoint *ydbcheckpoint.Store, spicedbClient *spicedb.Client,
	permissions ports.PermissionCatalog, now func() time.Time) (*policyjournal.Worker, error) {
	if source == nil || checkpoint == nil || spicedbClient == nil {
		return nil, projection.ErrInvalidProjection
	}
	projector, err := projection.New(permissions, spicedbClient, now)
	if err != nil {
		return nil, err
	}
	worker, err := projection.NewWorker(projector, source, checkpoint, now)
	if err != nil {
		return nil, err
	}
	return policyjournal.NewWorker(worker, checkpoint)
}
