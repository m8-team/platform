package consistency

import (
	"context"
	"errors"
	"math"

	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/policy"
)

var (
	ErrContextualPolicyUnavailable = errors.New("trusted contextual authorization data unavailable")
	ErrContextualPolicyConflict    = errors.New("contextual restriction snapshot is not current")
	ErrUnsupportedPolicyScope      = errors.New("contextual policy scope lacks authoritative matching evidence")
)

// CompleteRestrictionSnapshot is an exhaustive strong read of all active
// mandatory restrictions for an exact tenant/resource/action, pinned to the
// SAME PolicyHead as the committed policy mutation journal. A cache or partial
// search cannot legitimately return Complete=true.
type CompleteRestrictionSnapshot struct {
	PolicyHead uint64
	Complete   bool
	Policies   []policy.Policy
}
type RestrictionSource interface {
	LoadRestrictions(context.Context, domain.TenantID, string, string, string) (CompleteRestrictionSnapshot, error)
}

// TrustedContextSource constructs inputs ONLY from authenticated gateway peer,
// verified MFA/session and owner-authoritative resource attributes. It must
// never use the untrusted AuthZEN context or forwarded IP headers.
type TrustedContextSource interface {
	ResolveTrustedContext(context.Context, domain.TenantID, domain.Subject, string, string) (authorization.EvaluationContext, error)
}

// RestrictedUniversalBarrier is the only universal PDP recommended for
// deployment. The existing UniversalBarrier remains available for internal
// decisions without contextual policy integration; wiring must use this one.
type RestrictedUniversalBarrier struct {
	base         *UniversalBarrier
	restrictions RestrictionSource
	context      TrustedContextSource
}

func NewRestrictedUniversalBarrier(base *UniversalBarrier, restrictions RestrictionSource, trusted TrustedContextSource) (*RestrictedUniversalBarrier, error) {
	if base == nil || base.policies == nil || base.owner == nil || base.catalog == nil || base.graph == nil || restrictions == nil || trusted == nil {
		return nil, ErrContextualPolicyUnavailable
	}
	return &RestrictedUniversalBarrier{base: base, restrictions: restrictions, context: trusted}, nil
}
func (b *RestrictedUniversalBarrier) Check(ctx context.Context, tenant domain.TenantID, subject domain.Subject, kind, id, action string) (bool, error) {
	if b == nil || b.base == nil || b.restrictions == nil || b.context == nil {
		return false, ErrContextualPolicyUnavailable
	}
	// Every entitlement is checked first by the fully consistent SpiceDB and
	// owner/policy barrier. Restrictions can only subtract permits.
	allowed, err := b.base.Check(ctx, tenant, subject, kind, id, action)
	if err != nil || !allowed {
		return false, err
	}
	head, err := b.base.checkedPolicyHead(ctx, tenant)
	if err != nil {
		return false, err
	}
	snapshot, err := b.restrictions.LoadRestrictions(ctx, tenant, kind, id, action)
	if err != nil {
		return false, err
	}
	if !snapshot.Complete || snapshot.PolicyHead != head {
		return false, ErrContextualPolicyConflict
	}
	proof, err := b.base.owner.ResolveUniversalResource(ctx, tenant, kind, id)
	if err != nil {
		return false, err
	}
	if err := validateUniversalProof(proof, tenant, kind, id); err != nil {
		return false, err
	}
	if proof.Revision > math.MaxInt64 {
		return false, ErrContextualPolicyConflict
	}
	verified, err := b.context.ResolveTrustedContext(ctx, tenant, subject, kind, id)
	if err != nil || verified.Validate() != nil {
		return false, ErrContextualPolicyUnavailable
	}
	if verified.ResourceRevision() != int64(proof.Revision) {
		return false, ErrResourceHeadChanged
	}
	for _, restriction := range snapshot.Policies {
		if err := restriction.Validate(); err != nil {
			return false, ErrContextualPolicyConflict
		}
		if restriction.State() == policy.StateDisabled {
			continue
		}
		if string(restriction.TenantID()) != string(tenant) {
			return false, ErrContextualPolicyConflict
		}
		scope := restriction.Scope()
		// Hierarchical matching must not be emulated by comparing resource
		// strings or caller attributes. Broader restrictions fail closed.
		if scope.Kind() != authorization.ScopeInstance {
			return false, ErrUnsupportedPolicyScope
		}
		if scope.ResourceType() != kind || scope.ResourceID() != id {
			return false, ErrContextualPolicyConflict
		}
		targeted := false
		for _, p := range restriction.Actions() {
			if p.Name() == action {
				targeted = true
			}
		}
		if !targeted {
			return false, ErrContextualPolicyConflict
		}
		permit, err := restriction.CheckRestriction(verified)
		if err != nil {
			return false, err
		}
		if !permit {
			return false, nil
		}
	}
	// A restrictive policy or module uninstall committed during the above
	// processing must block even though the original graph result was fresh.
	after, err := b.base.checkedPolicyHead(ctx, tenant)
	if err != nil {
		return false, err
	}
	if after != head {
		return false, ErrPolicyHeadChanged
	}
	latest, err := b.base.owner.ResolveUniversalResource(ctx, tenant, kind, id)
	if err != nil {
		return false, err
	}
	if err := validateUniversalProof(latest, tenant, kind, id); err != nil {
		return false, err
	}
	if latest.Revision != proof.Revision {
		return false, ErrResourceHeadChanged
	}
	stillInstalled, err := b.base.catalog.HasCatalogPermission(ctx, tenant, kind, action)
	if err != nil {
		return false, err
	}
	if !stillInstalled {
		return false, ErrUniversalPermissionUnknown
	}
	return true, nil
}
