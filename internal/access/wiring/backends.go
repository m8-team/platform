package wiring

import (
	"context"
	"net/http"
	"time"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
	"github.com/m8-team/platform/internal/access/adapters/spicedb"
	"github.com/m8-team/platform/internal/access/adapters/ydbcheckpoint"
	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/app/consistency"
)

// NewProductionPDP composes Access with a provider-neutral authoritative
// resource revision port, the durable Access checkpoint and SpiceDB.
// Authentication middleware must supply the trusted principal resolver.
func NewProductionPDP(ctx context.Context, principal authzen.PrincipalResolver,
	resourceRevisions consistency.ResourceRevisionReader, checkpoint *ydbcheckpoint.Store,
	spicedbClient *spicedb.Client, revision string) (http.Handler, error) {
	if resourceRevisions == nil || checkpoint == nil || spicedbClient == nil {
		return nil, authzen.ErrInvalidConfiguration
	}
	return NewAuthZENHandler(ctx, principal, resourceRevisions, checkpoint, checkpoint, spicedbClient, revision)
}

// NewResourceProjectionWorker accepts an authoritative snapshot reader through
// an Access-owned port. The external composition root provides its adapter.
// Event offsets are acknowledged only after fenced graph projection and
// durable resource watermark commits.
func NewResourceProjectionWorker(source projection.SnapshotReader,
	checkpoint *ydbcheckpoint.Store, spicedbClient *spicedb.Client,
	permissions ports.PermissionCatalog, now func() time.Time) (*projection.ResourceWorker, error) {
	if source == nil || checkpoint == nil || spicedbClient == nil {
		return nil, projection.ErrInvalidProjection
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
