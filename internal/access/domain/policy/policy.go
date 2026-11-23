package policy

import (
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrInvalidPolicy  = errors.New("invalid mandatory access policy")
	ErrPolicyDisabled = errors.New("access policy is disabled")
)

type Effect string

const (
	EffectRequire  Effect = "REQUIRE"
	EffectDenyWhen Effect = "DENY_WHEN"
)

type State string

const (
	StateActive   State = "ACTIVE"
	StateDisabled State = "DISABLED"
)

// Policy applies an additional constraint. REQUIRE or DENY_WHEN can block,
// but neither can grant access without an independent entitlement.
type Policy struct {
	id         string
	tenantID   shared.TenantID
	name       string
	scope      authorization.Scope
	actions    []permission.Permission
	effect     Effect
	conditions ConditionSet
	state      State
	revision   shared.Revision
}
type Input struct {
	ID         string
	TenantID   shared.TenantID
	Name       string
	Scope      authorization.Scope
	Actions    []permission.Permission
	Effect     Effect
	Conditions ConditionSet
}

func NewPolicy(input Input) (Policy, error) {
	p := Policy{id: strings.TrimSpace(input.ID), tenantID: input.TenantID, name: strings.TrimSpace(input.Name),
		scope: input.Scope, actions: append([]permission.Permission(nil), input.Actions...), effect: input.Effect,
		conditions: input.Conditions, state: StateActive, revision: 1}
	return p, p.Validate()
}
func (p Policy) Validate() error {
	if !shared.ValidUUID(p.id) || p.tenantID.Validate() != nil || p.name == "" || len(p.name) > 128 ||
		p.scope.Validate() != nil || p.scope.TenantID() != p.tenantID ||
		permission.ValidateSet(p.actions, true) != nil || p.conditions.Validate() != nil ||
		(p.effect != EffectRequire && p.effect != EffectDenyWhen) ||
		(p.state != StateActive && p.state != StateDisabled) || p.revision.Validate() != nil {
		return ErrInvalidPolicy
	}
	return nil
}

// CheckRestriction can only decide whether THIS policy blocks an already
// established entitlement; it never returns a permit by itself.
func (p Policy) CheckRestriction(ctx authorization.EvaluationContext) (bool, error) {
	if err := p.Validate(); err != nil {
		return false, err
	}
	if p.state == StateDisabled {
		return false, ErrPolicyDisabled
	}
	matches, err := p.conditions.Evaluate(ctx)
	if err != nil {
		return false, err
	}
	if p.effect == EffectRequire {
		return matches, nil
	}
	return !matches, nil
}
func (p Policy) WithConditions(expected int64, next ConditionSet) (Policy, error) {
	if p.state != StateActive {
		return Policy{}, ErrPolicyDisabled
	}
	rev, err := p.revision.Next(expected)
	if err != nil {
		return Policy{}, err
	}
	updated := p
	updated.revision = rev
	updated.conditions = next
	return updated, updated.Validate()
}
func (p Policy) Disable(expected int64) (Policy, error) {
	if p.state != StateActive {
		return Policy{}, ErrPolicyDisabled
	}
	rev, err := p.revision.Next(expected)
	if err != nil {
		return Policy{}, err
	}
	updated := p
	updated.revision = rev
	updated.state = StateDisabled
	return updated, updated.Validate()
}
func (p Policy) ID() string                 { return p.id }
func (p Policy) TenantID() shared.TenantID  { return p.tenantID }
func (p Policy) Name() string               { return p.name }
func (p Policy) Scope() authorization.Scope { return p.scope }
func (p Policy) Actions() []permission.Permission {
	return append([]permission.Permission(nil), p.actions...)
}
func (p Policy) Effect() Effect           { return p.effect }
func (p Policy) Conditions() ConditionSet { return p.conditions }
func (p Policy) State() State             { return p.state }
func (p Policy) Revision() int64          { return p.revision.Int64() }
