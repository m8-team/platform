package accessrm

import (
	"context"
	"net/http"
	"time"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
	"github.com/m8-team/platform/internal/access/adapters/spicedb"
	"github.com/m8-team/platform/internal/access/adapters/ydbcheckpoint"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
	accesswiring "github.com/m8-team/platform/internal/access/wiring"
	rmcatalog "github.com/m8-team/platform/internal/resourcemanager/adapter/resourcecatalog"
)

// NewProductionPDP composes Access with a Resource Manager authoritative
// query at the integration boundary, never inside the Access module itself.
func NewProductionPDP(ctx context.Context, principal authzen.PrincipalResolver,
	catalog rmcatalog.Catalog, checkpoint *ydbcheckpoint.Store,
	client *spicedb.Client, modelRevision string) (http.Handler, error) {
	if checkpoint == nil || client == nil {
		return nil, authzen.ErrInvalidConfiguration
	}
	source, err := NewSource(catalog)
	if err != nil {
		return nil, err
	}
	return accesswiring.NewProductionPDP(ctx, principal, source, checkpoint, client, modelRevision)
}

// NewResourceProjectionWorker supplies Access's projection port from an
// external, owner-authoritative Resource Manager read model.
func NewResourceProjectionWorker(catalog rmcatalog.Catalog,
	checkpoint *ydbcheckpoint.Store, client *spicedb.Client,
	permissions ports.PermissionCatalog, now func() time.Time) (*projection.ResourceWorker, error) {
	if checkpoint == nil || client == nil {
		return nil, projection.ErrInvalidProjection
	}
	source, err := NewSource(catalog)
	if err != nil {
		return nil, err
	}
	return accesswiring.NewResourceProjectionWorker(source, checkpoint, client, permissions, now)
}
