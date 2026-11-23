package consistency

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain"
)

type policyStub struct {
	states []PolicyState
	err    error
	reads  int
}

func (p *policyStub) LoadPolicyState(_ context.Context, _ domain.TenantID) (PolicyState, error) {
	p.reads++
	if p.err != nil {
		return PolicyState{}, p.err
	}
	if len(p.states) == 0 {
		return PolicyState{}, nil
	}
	i := p.reads - 1
	if i >= len(p.states) {
		i = len(p.states) - 1
	}
	return p.states[i], nil
}

func policyTestBarrier(t *testing.T, p *policyStub, engine *pdpStub) *PolicyBarrier {
	t.Helper()
	resource, err := New(sourceStub{revision: 4}, progressStub{watermark: Watermark{Revision: 4, ZedToken: "rm-token"}}, engine)
	if err != nil {
		t.Fatal(err)
	}
	barrier, err := NewPolicyBarrier(resource, p)
	if err != nil {
		t.Fatal(err)
	}
	return barrier
}
func TestPolicyBarrierBlocksUninitializedAndIncompleteFanout(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state PolicyState
		want  error
	}{
		{"missing", PolicyState{}, ErrPolicyUninitialized},
		{"unbootstrapped", PolicyState{Head: 1}, ErrPolicyUninitialized},
		{"delayed grant revoke", PolicyState{Head: 7, Applied: 6}, ErrPolicyProjectionLag},
		{"role edit second binding pending", PolicyState{Head: 10, Applied: 9}, ErrPolicyProjectionLag},
		{"invalid progress", PolicyState{Head: 7, Applied: 8}, ErrPolicyUninitialized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &policyStub{states: []PolicyState{tc.state}}
			engine := &pdpStub{}
			b := policyTestBarrier(t, p, engine)
			_, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeAllow))
			if !errors.Is(err, tc.want) || engine.called {
				t.Fatalf("unexpected check: %v engine called=%v", err, engine.called)
			}
		})
	}
}
func TestPolicyBarrierRequiresFullyConsistentCheckAndForcesDenyMode(t *testing.T) {
	engine := &pdpStub{}
	p := &policyStub{states: []PolicyState{{Head: 9, Applied: 9}}}
	b := policyTestBarrier(t, p, engine)
	got, err := b.CheckPermission(context.Background(), checkRequest(t, "caller-token", domain.FailModeAllow))
	if err != nil || !got.Allowed() || !engine.called || engine.token != "" || !engine.critical || engine.failMode != domain.FailModeDeny || p.reads != 2 {
		t.Fatalf("unsafe policy decision: %v %+v engine=%+v reads=%d", err, got, engine, p.reads)
	}
}
func TestPolicyBarrierRejectsRevocationDuringSpiceDBCheck(t *testing.T) {
	p := &policyStub{states: []PolicyState{{Head: 13, Applied: 13}, {Head: 14, Applied: 13}}}
	engine := &pdpStub{}
	b := policyTestBarrier(t, p, engine)
	_, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeDeny))
	if !errors.Is(err, ErrPolicyProjectionLag) || !engine.called {
		t.Fatalf("inflight revoke escaped: %v", err)
	}
}
func TestPolicyBarrierRejectsNewHeadEvenIfAlreadyProjected(t *testing.T) {
	p := &policyStub{states: []PolicyState{{Head: 5, Applied: 5}, {Head: 6, Applied: 6}}}
	b := policyTestBarrier(t, p, &pdpStub{})
	_, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeDeny))
	if !errors.Is(err, ErrPolicyHeadChanged) {
		t.Fatalf("source moved during allow: %v", err)
	}
}
func TestPolicyBarrierRejectsUnavailableAuthority(t *testing.T) {
	p := &policyStub{err: errors.New("Access unavailable")}
	engine := &pdpStub{}
	b := policyTestBarrier(t, p, engine)
	if _, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeDeny)); err == nil || engine.called {
		t.Fatalf("allowed without Access owner: %v", err)
	}
}
func TestPolicyBarrierRequiresBothOwners(t *testing.T) {
	if _, err := NewPolicyBarrier(nil, &policyStub{}); !errors.Is(err, ErrInvalidBarrier) {
		t.Fatal(err)
	}
}
