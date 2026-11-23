// Package wiring composes the authenticated AuthZEN handler with the mandatory
// Resource Manager freshness barrier and the production SpiceDB permission port.
package wiring

import (
	"context"
	"net/http"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
	accesshttp "github.com/m8-team/platform/internal/access/adapters/http"
	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/app/ports"
)

// NewAuthZENHandler returns the actual HTTP/gRPC-gateway PDP evaluation routes.
// All dependencies are mandatory; the principal resolver must be supplied by
// verified OIDC/mTLS middleware, never by request headers. No AllowAll fallback.
func NewAuthZENHandler(ctx context.Context, principal authzen.PrincipalResolver,
	resourceManager consistency.ResourceRevisionReader, watermarks consistency.WatermarkReader,
	policyState consistency.PolicyStateReader, spiceDB ports.PermissionEngine, modelRevision string) (http.Handler, error) {
	barrier, err := consistency.New(resourceManager, watermarks, spiceDB)
	if err != nil {
		return nil, err
	}
	policyBarrier, err := consistency.NewPolicyBarrier(barrier, policyState)
	if err != nil {
		return nil, err
	}
	service, err := authzen.New(principal, policyBarrier, modelRevision)
	if err != nil {
		return nil, err
	}
	return accesshttp.NewHandler(ctx, service)
}
