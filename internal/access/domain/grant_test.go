package domain

import (
	"errors"
	"testing"
	"time"
)

func grantTestInput(t *testing.T) AccessGrantInput {
	t.Helper()
	base := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	return AccessGrantInput{
		ID: testGrantID, TenantID: TenantID(testTenantID),
		Subject: domainSubject(t), RequestedBy: domainSubject(t),
		Target:      GrantTarget{OrganizationID: testOrganizationID, Resource: domainResource(t)},
		Permissions: []Permission{domainPermission(t, "m8.workspace.read")},
		Reason:      GrantReasonSupportCase, ReasonRef: "case-123",
		ValidFrom: base.Add(time.Hour), ExpiresAt: base.Add(24 * time.Hour), CreatedAt: base,
		Constraints: GrantConstraints{RequireMFA: true, RequireActiveCase: true},
	}
}

func TestGrantApprovalCannotBypassActivation(t *testing.T) {
	input := grantTestInput(t)
	g, err := NewAccessGrant(input)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := NewSubject("user", "approver")
	if err != nil {
		t.Fatal(err)
	}
	g, err = g.SubmitForApproval(1, input.CreatedAt.Add(5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Approve(2, input.RequestedBy, input.CreatedAt.Add(10*time.Minute)); !errors.Is(err, ErrGrantSelfApproval) {
		t.Fatalf("self approval: %v", err)
	}
	g, err = g.Approve(2, actor, input.CreatedAt.Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if g.State() != GrantStateApproved || g.EffectiveAt(input.ValidFrom) {
		t.Fatal("approval must not imply access")
	}
	if _, err := g.Activate(3, "", input.ValidFrom); !errors.Is(err, ErrGrantProjectionRequired) {
		t.Fatalf("missing projection evidence: %v", err)
	}
	if _, err := g.Activate(3, "revision", input.ValidFrom.Add(-time.Second)); !errors.Is(err, ErrGrantNotYetValid) {
		t.Fatalf("early activation: %v", err)
	}
	g, err = g.Activate(3, "projection:123", input.ValidFrom)
	if err != nil {
		t.Fatal(err)
	}
	if !g.EffectiveAt(input.ValidFrom) || g.EffectiveAt(input.ExpiresAt) {
		t.Fatal("effective period must be half-open")
	}
	g, err = g.Revoke(4, actor, input.ValidFrom.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if g.EffectiveAt(input.ValidFrom.Add(2*time.Minute)) || g.State() != GrantStateRevoked {
		t.Fatal("revoked access must be denied")
	}
	if _, err := g.Activate(5, "proof", input.ValidFrom.Add(time.Hour)); !errors.Is(err, ErrInvalidGrantTransition) {
		t.Fatalf("revoked grant must stay terminal: %v", err)
	}
}

func TestGrantValidatesTenantScopeAndTime(t *testing.T) {
	input := grantTestInput(t)
	input.TenantID = ""
	if _, err := NewAccessGrant(input); !errors.Is(err, ErrInvalidTenantID) {
		t.Fatalf("no tenant: %v", err)
	}
	input = grantTestInput(t)
	input.ExpiresAt = input.ValidFrom
	if _, err := NewAccessGrant(input); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("bad interval: %v", err)
	}
	input = grantTestInput(t)
	input.Target.OrganizationID = ""
	if _, err := NewAccessGrant(input); !errors.Is(err, ErrInvalidGrant) {
		t.Fatalf("no organization: %v", err)
	}
	input = grantTestInput(t)
	input.Permissions = nil
	if _, err := NewAccessGrant(input); !errors.Is(err, ErrPermissionRequired) {
		t.Fatalf("empty permissions: %v", err)
	}
}

func TestGrantRejectsStaleRevisionAndAllowsExpiry(t *testing.T) {
	input := grantTestInput(t)
	g, err := NewAccessGrant(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.SubmitForApproval(100, input.CreatedAt.Add(time.Minute)); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale rev: %v", err)
	}
	if _, err := g.Expire(1, input.ExpiresAt.Add(-time.Nanosecond)); !errors.Is(err, ErrInvalidGrantTransition) {
		t.Fatalf("early expire: %v", err)
	}
	expired, err := g.Expire(1, input.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	if expired.State() != GrantStateExpired || expired.Revision() != 2 || g.State() != GrantStateRequested {
		t.Fatal("expire must produce new snapshot")
	}
	if _, err := expired.SubmitForApproval(2, input.ExpiresAt.Add(time.Second)); !errors.Is(err, ErrGrantExpired) {
		t.Fatalf("cannot revive expired grant: %v", err)
	}
}

func TestGrantPermissionsAreImmutable(t *testing.T) {
	input := grantTestInput(t)
	g, err := NewAccessGrant(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Permissions[0] = domainPermission(t, "delete")
	if g.Permissions()[0].Name() != "m8.workspace.read" {
		t.Fatal("input slice leaked")
	}
	out := g.Permissions()
	out[0] = domainPermission(t, "delete")
	if g.Permissions()[0].Name() != "m8.workspace.read" {
		t.Fatal("output slice leaked")
	}
}

// An administrator may request access on behalf of a user. That user cannot
// then self-approve the request even though the requester is someone else.
func TestGrantBeneficiaryCannotSelfApprove(t *testing.T) {
	input := grantTestInput(t)
	admin, err := NewSubject("user", "administrator")
	if err != nil {
		t.Fatal(err)
	}
	input.RequestedBy = admin
	grant, err := NewAccessGrant(input)
	if err != nil {
		t.Fatal(err)
	}
	grant, err = grant.SubmitForApproval(1, input.CreatedAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	_, err = grant.Approve(2, input.Subject, input.CreatedAt.Add(2*time.Minute))
	if !errors.Is(err, ErrGrantSelfApproval) {
		t.Fatalf("beneficiary self approval: %v", err)
	}
}
