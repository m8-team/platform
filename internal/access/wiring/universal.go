package wiring

import (
	"context"
	"net/http"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
	"github.com/m8-team/platform/internal/access/adapters/spicedb"
	"github.com/m8-team/platform/internal/access/adapters/ydbcatalog"
	"github.com/m8-team/platform/internal/access/adapters/ydbcheckpoint"
	"github.com/m8-team/platform/internal/access/app/consistency"
)

// NewProductionUniversalPDP is the opt-in AuthZEN composition for namespaced
// module resource types (e.g. finance.invoice). All dependencies are mandatory;
// production must supply trusted OIDC/mTLS identity, module installation state,
// complete policy snapshots, and owner-verifiable context. The legacy PDP path
// remains the same, and no client-provided tenant/context is trusted.
func NewProductionUniversalPDP(
	ctx context.Context,
	principal authzen.PrincipalResolver,
	legacyOwner consistency.ResourceRevisionReader,
	policyStore *ydbcheckpoint.Store,
	graph *spicedb.Client,
	universalOwner consistency.UniversalResourceOwner,
	moduleInstallations ydbcatalog.TenantModuleResolver,
	catalogStore *ydbcatalog.Store,
	restrictions consistency.RestrictionSource,
	contextSource consistency.TrustedContextSource,
	modelRevision string,
) (http.Handler, error) {
	if principal == nil || legacyOwner == nil || policyStore == nil || graph == nil ||
		universalOwner == nil || moduleInstallations == nil || catalogStore == nil ||
		restrictions == nil || contextSource == nil {
		return nil, authzen.ErrInvalidConfiguration
	}
	legacy, err := consistency.New(legacyOwner, policyStore, graph)
	if err != nil {
		return nil, err
	}
	legacyBarrier, err := consistency.NewPolicyBarrier(legacy, policyStore)
	if err != nil {
		return nil, err
	}
	catalogReader, err := ydbcatalog.NewUniversalPermissionReader(moduleInstallations, catalogStore)
	if err != nil {
		return nil, err
	}
	universalBase, err := consistency.NewUniversalBarrier(policyStore, universalOwner, catalogReader, graph)
	if err != nil {
		return nil, err
	}
	universal, err := consistency.NewRestrictedUniversalBarrier(universalBase, restrictions, contextSource)
	if err != nil {
		return nil, err
	}
	service, err := authzen.NewWithUniversal(principal, legacyBarrier, universal, modelRevision)
	if err != nil {
		return nil, err
	}
	return authzen.NewHTTPHandler(ctx, service)
}
