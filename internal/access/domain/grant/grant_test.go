package grant

import (
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

const tenantID = "11111111-1111-4111-8111-111111111111"
const grantID = "22222222-2222-4222-8222-222222222222"
const orgID = "33333333-3333-4333-8333-333333333333"

func fixture(t *testing.T) (AccessGrant, shared.Subject, shared.Subject, time.Time) {
	t.Helper()
	tenant, _ := shared.NewTenantID(tenantID)
	user, _ := shared.NewSubject("user", "beneficiary")
	requester, _ := shared.NewSubject("user", "requester")
	approver, _ := shared.NewSubject("user", "approver")
	target, _ := shared.NewResource("environment", "env-1")
	p, _ := permission.NewPermission("environment.deploy")
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	g, err := NewAccessGrant(Input{
		ID: grantID, TenantID: tenant, Subject: user, RequestedBy: requester,
		Target:      Target{OrganizationID: orgID, Resource: target},
		Permissions: []permission.Permission{p}, Reason: ReasonSupportCase, ReasonRef: "CASE-123",
		ValidFrom: now, ExpiresAt: now.Add(4 * time.Hour), CreatedAt: now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	return g, user, approver, now
}
func TestGrantApprovalActivationRevocation(t *testing.T) {
	g, user, approver, now := fixture(t)
	pending, err := g.SubmitForApproval(1, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pending.Approve(2, user, now); !errors.Is(err, ErrGrantSelfApproval) {
		t.Fatal(err)
	}
	approved, err := pending.Approve(2, approver, now)
	if err != nil {
		t.Fatal(err)
	}
	if approved.EffectiveAt(now) {
		t.Fatal("approved grant authorized prematurely")
	}
	if _, err = approved.Activate(3, "", now); !errors.Is(err, ErrGrantProjectionRequired) {
		t.Fatal(err)
	}
	active, err := approved.Activate(3, "verified-checkpoint", now)
	if err != nil || !active.EffectiveAt(now) {
		t.Fatalf("activate failed: %v", err)
	}
	revoked, err := active.Revoke(4, approver, now.Add(time.Minute))
	if err != nil || revoked.EffectiveAt(now.Add(2*time.Minute)) {
		t.Fatalf("revocation failed: %v", err)
	}
	if g.Revision() != 1 || revoked.Revision() != 5 {
		t.Fatal("aggregate snapshots mutated")
	}
	if _, err = revoked.Revoke(5, approver, now.Add(2*time.Minute)); !errors.Is(err, ErrInvalidGrantTransition) {
		t.Fatal(err)
	}
}
func TestGrantExpiryAndContext(t *testing.T) {
	g, _, _, now := fixture(t)
	if _, err := g.Expire(1, now); !errors.Is(err, ErrInvalidGrantTransition) {
		t.Fatal(err)
	}
	expired, err := g.Expire(1, now.Add(4*time.Hour))
	if err != nil || expired.State() != Expired {
		t.Fatal(err)
	}
	if !(Constraints{RequireMFA: true}).RequiresContext() {
		t.Fatal("MFA constraint lost")
	}
}
