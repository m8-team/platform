package consistency

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrPolicyUninitialized = errors.New("authoritative Access policy journal has not been bootstrapped")
	ErrPolicyProjectionLag = errors.New("Access grants or role bindings have not converged to SpiceDB")
	ErrPolicyHeadChanged   = errors.New("Access policy changed while the authorization check was in flight")
	ErrResourceHeadChanged = errors.New("Resource Manager revision changed while the authorization check was in flight")
)

// PolicyState is read from the authoritative Access transaction journal, NOT
// from Kafka, a client header, a cached projection, or SpiceDB itself.
// Head advances in the SAME YDB transaction as the underlying Access mutation
// and its complete fan-out manifest; Applied only advances after all entries
// have a certified graph checkpoint.
type PolicyState struct {
	Head    uint64
	Applied uint64
}

// PolicyStateReader MUST load Head and Applied from one consistent read of the
// authoritative Access store. A missing row is NOT an empty policy.
type PolicyStateReader interface {
	LoadPolicyState(context.Context, domain.TenantID) (PolicyState, error)
}

// PolicyBarrier guards a Resource Manager barrier with the independent,
// tenant-wide Access policy revision. The intentionally coarse tenant scope
// fails closed on ANY delayed revoke or incomplete RoleBinding fanout.
// v2 never merges independent opaque ZedTokens and requires a fully consistent
// SpiceDB check after both owners' durable state has converged.
type PolicyBarrier struct {
	resources *Barrier
	policies  PolicyStateReader
}

func NewPolicyBarrier(resources *Barrier, policies PolicyStateReader) (*PolicyBarrier, error) {
	if resources == nil || resources.source == nil || resources.progress == nil || resources.pdp == nil || policies == nil {
		return nil, ErrInvalidBarrier
	}
	return &PolicyBarrier{resources: resources, policies: policies}, nil
}

func validatePolicyState(s PolicyState) error {
	if s.Head == 0 || s.Applied == 0 || s.Applied > s.Head {
		return ErrPolicyUninitialized
	}
	if s.Applied != s.Head {
		return ErrPolicyProjectionLag
	}
	return nil
}

func (b *PolicyBarrier) CheckPermission(ctx context.Context, req domain.CheckPermissionRequest) (domain.PermissionDecision, error) {
	if b == nil || b.policies == nil || b.resources == nil {
		return domain.PermissionDecision{}, ErrInvalidBarrier
	}
	if err := req.Validate(); err != nil {
		return domain.PermissionDecision{}, err
	}
	if err := ctx.Err(); err != nil {
		return domain.PermissionDecision{}, err
	}

	before, err := b.policies.LoadPolicyState(ctx, req.TenantID())
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	if err := validatePolicyState(before); err != nil {
		return domain.PermissionDecision{}, err
	}

	// FullyConsistent is required, because Resource Manager and Access
	// checkpoints may contain independent, unordered opaque ZedTokens.
	decision, resourceRevision, err := b.resources.checkResource(ctx, req, true)
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	if !decision.Allowed() {
		return decision, nil
	}

	// A positive authorization result must not escape if a new revocation or
	// role change became authoritative during the remote SpiceDB evaluation.
	after, err := b.policies.LoadPolicyState(ctx, req.TenantID())
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	if err := validatePolicyState(after); err != nil {
		return domain.PermissionDecision{}, err
	}
	if after.Head != before.Head || after.Applied != before.Applied {
		return domain.PermissionDecision{}, ErrPolicyHeadChanged
	}

	rmRevision, err := b.resources.source.AuthoritativeRevision(ctx, req.TenantID(), req.Resource())
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	if rmRevision != resourceRevision {
		return domain.PermissionDecision{}, ErrResourceHeadChanged
	}
	rmWatermark, err := b.resources.progress.LoadWatermark(ctx, req.TenantID(), req.Resource())
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	if rmWatermark.Revision < rmRevision || rmWatermark.ZedToken == "" {
		return domain.PermissionDecision{}, ErrProjectionLag
	}
	return decision, nil
}
