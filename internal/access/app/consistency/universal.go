package consistency

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/access/domain"
)

// UniversalOwnerProof MUST be loaded from an authoritative resource owner,
// never from an HTTP header, Kafka event or caller-supplied attributes.
// It carries the live tenant and resource lifecycle revision.
type UniversalOwnerProof struct {
	TenantID     domain.TenantID
	ResourceType string
	ResourceID   string
	Revision     uint64
	Live         bool
}
type UniversalResourceOwner interface {
	ResolveUniversalResource(context.Context, domain.TenantID, string, string) (UniversalOwnerProof, error)
}

// UniversalCatalogReader recognizes only actions explicitly published for an
// exact resource type. It cannot infer permissions from a wildcard module.
type UniversalCatalogReader interface {
	HasCatalogPermission(context.Context, domain.TenantID, string, string) (bool, error)
}

// UniversalGraphPDP MUST issue fully consistent SpiceDB checks for one
// capability tuple; CONDITIONAL caveats and missing graph data are denies.
type UniversalGraphPDP interface {
	CheckUniversalCapability(context.Context, domain.TenantID, domain.Subject, string, string, string) (bool, error)
}

var (
	ErrInvalidUniversalBarrier    = errors.New("universal Access decision dependencies unavailable")
	ErrUniversalResourceInvalid   = errors.New("unverified universal resource or expired owner revision")
	ErrUniversalPermissionUnknown = errors.New("permission is not registered for this resource kind")
)

// UniversalBarrier implements full fail-closed freshness on the new graph path.
// The existing fixed-kind CheckPermission/AuthZEN path remains unchanged.
type UniversalBarrier struct {
	policies PolicyStateReader
	owner    UniversalResourceOwner
	catalog  UniversalCatalogReader
	graph    UniversalGraphPDP
}

func NewUniversalBarrier(policies PolicyStateReader, owner UniversalResourceOwner, catalog UniversalCatalogReader, graph UniversalGraphPDP) (*UniversalBarrier, error) {
	if policies == nil || owner == nil || catalog == nil || graph == nil {
		return nil, ErrInvalidUniversalBarrier
	}
	return &UniversalBarrier{policies: policies, owner: owner, catalog: catalog, graph: graph}, nil
}
func (b *UniversalBarrier) checkedPolicyHead(ctx context.Context, tenant domain.TenantID) (uint64, error) {
	state, err := b.policies.LoadPolicyState(ctx, tenant)
	if err != nil {
		return 0, err
	}
	if state.Head == 0 || state.Applied == 0 || state.Applied > state.Head {
		return 0, ErrPolicyUninitialized
	}
	if state.Head != state.Applied {
		return 0, ErrPolicyProjectionLag
	}
	return state.Head, nil
}
func validateUniversalProof(p UniversalOwnerProof, tenant domain.TenantID, kind, id string) error {
	if p.TenantID != tenant || p.ResourceType != kind || p.ResourceID != id || p.Revision == 0 || !p.Live {
		return ErrUniversalResourceInvalid
	}
	return nil
}

// Check is a trusted internal PDP use case. Callers MUST authenticate subject
// and derive tenant identity before invoking it; this function is not exposed
// as a public endpoint or proxy that trusts unverified user claims.
func (b *UniversalBarrier) Check(ctx context.Context, tenant domain.TenantID, subject domain.Subject, kind, id, action string) (bool, error) {
	if b == nil || b.policies == nil || b.owner == nil || b.catalog == nil || b.graph == nil {
		return false, ErrInvalidUniversalBarrier
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if tenant.Validate() != nil || subject.Validate() != nil || kind == "" || id == "" || action == "" {
		return false, ErrUniversalResourceInvalid
	}
	before, err := b.checkedPolicyHead(ctx, tenant)
	if err != nil {
		return false, err
	}
	registered, err := b.catalog.HasCatalogPermission(ctx, tenant, kind, action)
	if err != nil {
		return false, err
	}
	if !registered {
		return false, ErrUniversalPermissionUnknown
	}
	proof, err := b.owner.ResolveUniversalResource(ctx, tenant, kind, id)
	if err != nil {
		return false, err
	}
	if err := validateUniversalProof(proof, tenant, kind, id); err != nil {
		return false, err
	}
	allowed, err := b.graph.CheckUniversalCapability(ctx, tenant, subject, kind, id, action)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, nil
	}
	after, err := b.checkedPolicyHead(ctx, tenant)
	if err != nil {
		return false, err
	}
	if after != before {
		return false, ErrPolicyHeadChanged
	}
	newProof, err := b.owner.ResolveUniversalResource(ctx, tenant, kind, id)
	if err != nil {
		return false, err
	}
	if err := validateUniversalProof(newProof, tenant, kind, id); err != nil {
		return false, err
	}
	if newProof.Revision != proof.Revision {
		return false, ErrResourceHeadChanged
	}
	return true, nil
}
