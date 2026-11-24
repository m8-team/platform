package trustedcontext

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
)

type peerStub struct {
	ip  netip.Addr
	err error
}

func (p peerStub) VerifiedClientIP(context.Context) (netip.Addr, error) { return p.ip, p.err }

type mfaStub struct {
	verified bool
	known    bool
	err      error
}

func (m mfaStub) VerifiedMFA(context.Context, domain.TenantID, domain.Subject) (bool, bool, error) {
	return m.verified, m.known, m.err
}

type ownerStub struct {
	revision uint64
	attrs    map[string]string
	err      error
}

func (o ownerStub) VerifiedResourceAttributes(context.Context, domain.TenantID, string, string) (uint64, map[string]string, error) {
	return o.revision, o.attrs, o.err
}
func TestTrustedContextDerivesVerifiedNetworkMFASessionAndOwnerRevision(t *testing.T) {
	tenant, _ := domain.NewTenantID("11111111-1111-4111-8111-111111111111")
	subject, _ := domain.NewSubject("user", "alice")
	now := time.Date(2026, 11, 23, 15, 0, 0, 0, time.UTC)
	newContext := func(peer peerStub, mfa mfaStub, owner ownerStub) (int64, bool, error) {
		t.Helper()
		source, err := New(peer, mfa, owner, func() time.Time { return now })
		if err != nil {
			t.Fatal(err)
		}
		ctx, err := source.ResolveTrustedContext(context.Background(), tenant, subject, "finance.invoice", "invoice-1")
		if err != nil {
			return 0, false, err
		}
		ok, known := ctx.MFA()
		return ctx.ResourceRevision(), ok && known, nil
	}
	peer := peerStub{ip: netip.MustParseAddr("10.2.3.4")}
	mfa := mfaStub{verified: true, known: true}
	owner := ownerStub{revision: 7, attrs: map[string]string{"status": "approved"}}
	rev, verified, err := newContext(peer, mfa, owner)
	if err != nil || rev != 7 || !verified {
		t.Fatalf("verified sources lost: rev=%d MFA=%v err=%v", rev, verified, err)
	}
	for _, test := range []struct {
		peer  peerStub
		mfa   mfaStub
		owner ownerStub
	}{
		{peerStub{}, mfa, owner},
		{peer, mfaStub{verified: true, known: false}, owner},
		{peer, mfa, ownerStub{revision: 0}},
		{peer, mfa, ownerStub{revision: 7, attrs: map[string]string{" ": "invalid"}}},
		{peerStub{err: errors.New("unverified IP")}, mfa, owner},
	} {
		if _, _, err := newContext(test.peer, test.mfa, test.owner); !errors.Is(err, ErrUnverifiedEvidence) {
			t.Fatalf("invalid trusted context accepted: %v", err)
		}
	}
}
