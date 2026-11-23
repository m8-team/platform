package consistency

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain"
)

const testTenant = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

type sourceStub struct {
	revision uint64
	err      error
}

func (s sourceStub) AuthoritativeRevision(_ context.Context, _ domain.TenantID, _ domain.Resource) (uint64, error) {
	return s.revision, s.err
}

type progressStub struct {
	watermark Watermark
	err       error
}

func (s progressStub) LoadWatermark(_ context.Context, _ domain.TenantID, _ domain.Resource) (Watermark, error) {
	return s.watermark, s.err
}

type pdpStub struct {
	called   bool
	token    string
	critical bool
	failMode domain.FailMode
}

func (s *pdpStub) CheckPermission(_ context.Context, r domain.CheckPermissionRequest) (domain.PermissionDecision, error) {
	s.called = true
	s.token = r.MinZedToken()
	s.critical = r.Critical()
	s.failMode = r.FailMode()
	return domain.NewPermissionDecision(domain.DecisionAllow, r.ModelRevision(), "allowed")
}
func checkRequest(t *testing.T, token string, mode domain.FailMode) domain.CheckPermissionRequest {
	t.Helper()
	tn, e := domain.NewTenantID(testTenant)
	if e != nil {
		t.Fatal(e)
	}
	user, e := domain.NewSubject("user", "user-a")
	if e != nil {
		t.Fatal(e)
	}
	res, e := domain.NewResource("workspace", "workspace-a")
	if e != nil {
		t.Fatal(e)
	}
	perm, e := domain.NewPermission("m8.workspace.read")
	if e != nil {
		t.Fatal(e)
	}
	revision, e := domain.NewModelRevision("rev-1")
	if e != nil {
		t.Fatal(e)
	}
	r, e := domain.NewCheckPermissionRequest(domain.CheckPermissionInput{TenantID: tn, MinZedToken: token, Subject: user, Resource: res, Permission: perm, ModelRevision: revision, FailMode: mode, FailOpenReference: "ADR-001"})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestBarrierRejectsLagAndNeverInvokesPDP(t *testing.T) {
	p := &pdpStub{}
	b, _ := New(sourceStub{revision: 2}, progressStub{watermark: Watermark{Revision: 1, ZedToken: "z1"}}, p)
	_, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeAllow))
	if !errors.Is(err, ErrProjectionLag) || p.called {
		t.Fatalf("stale allow: %v, called=%v", err, p.called)
	}
}
func TestBarrierRequiresAuthoritativeRevision(t *testing.T) {
	p := &pdpStub{}
	b, _ := New(sourceStub{revision: 0}, progressStub{watermark: Watermark{Revision: 999, ZedToken: "z1"}}, p)
	_, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeDeny))
	if !errors.Is(err, ErrResourceRevisionUnavailable) || p.called {
		t.Fatalf("missing authority: %v", err)
	}
}
func TestBarrierPinsGraphRevisionAndForcesFailClosed(t *testing.T) {
	p := &pdpStub{}
	b, _ := New(sourceStub{revision: 2}, progressStub{watermark: Watermark{Revision: 2, ZedToken: "z1"}}, p)
	decision, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeAllow))
	if err != nil || !decision.Allowed() || p.token != "z1" || !p.critical || p.failMode != domain.FailModeDeny {
		t.Fatalf("unexpected secure PDP input: %v token=%q critical=%v mode=%q", err, p.token, p.critical, p.failMode)
	}
}
func TestBarrierMultipleOpaqueTokensUseFullyConsistent(t *testing.T) {
	p := &pdpStub{}
	b, _ := New(sourceStub{revision: 2}, progressStub{watermark: Watermark{Revision: 2, ZedToken: "z1"}}, p)
	_, err := b.CheckPermission(context.Background(), checkRequest(t, "z2", domain.FailModeDeny))
	if err != nil || p.token != "" {
		t.Fatalf("opaque ZedToken merge unsafe: %v token %q", err, p.token)
	}
}
func TestBarrierFailsClosedOnResolverError(t *testing.T) {
	p := &pdpStub{}
	failure := errors.New("resource manager unavailable")
	b, _ := New(sourceStub{err: failure}, progressStub{watermark: Watermark{Revision: 100, ZedToken: "z1"}}, p)
	_, err := b.CheckPermission(context.Background(), checkRequest(t, "", domain.FailModeDeny))
	if !errors.Is(err, failure) || p.called {
		t.Fatalf("allowed without authority: %v", err)
	}
}
