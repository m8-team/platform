package authzen

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type universalOwnerStub struct{ calls int }

func (f *universalOwnerStub) ResolveUniversalResource(_ context.Context, tenant domain.TenantID, typ, id string) (consistency.UniversalOwnerProof, error) {
	f.calls++
	return consistency.UniversalOwnerProof{TenantID: tenant, ResourceType: typ, ResourceID: id, Revision: 7, Live: true}, nil
}

type universalCatalogStub struct {
	ok    bool
	calls int
}

func (f *universalCatalogStub) HasCatalogPermission(context.Context, domain.TenantID, string, string) (bool, error) {
	f.calls++
	return f.ok, nil
}

type universalGraphStub struct {
	called  int
	allowed bool
}

func (f *universalGraphStub) CheckUniversalCapability(context.Context, domain.TenantID, domain.Subject, string, string, string) (bool, error) {
	f.called++
	return f.allowed, nil
}

type restrictionsStub struct{ complete bool }

func (f restrictionsStub) LoadRestrictions(context.Context, domain.TenantID, string, string, string) (consistency.CompleteRestrictionSnapshot, error) {
	return consistency.CompleteRestrictionSnapshot{PolicyHead: 7, Complete: f.complete}, nil
}

type evidenceStub struct{}

func (evidenceStub) ResolveTrustedContext(context.Context, domain.TenantID, domain.Subject, string, string) (authorization.EvaluationContext, error) {
	return authorization.NewEvaluationContext(authorization.ContextInput{
		ClientIP: netip.MustParseAddr("10.0.0.5"), MFAKnown: true, MFAVerified: true,
		EvaluatedAt: time.Now().UTC(), ResourceRevision: 7,
	})
}

func universalRequest(subject string) *accesspb.EvaluateAccessRequest {
	return &accesspb.EvaluateAccessRequest{
		Subject:  &accesspb.Subject{Type: "user", Id: subject},
		Action:   &accesspb.Action{Name: "finance.invoice.view"},
		Resource: &accesspb.Resource{Type: "finance.invoice", Id: "invoice-42"},
	}
}
func TestAuthZENUniversalRoutingDoesNotFallbackToLegacy(t *testing.T) {
	legacy, oldPDP := makeService(t, 1, nil)
	universal, err := legacy.EvaluateAccess(context.Background(), universalRequest("alice"))
	if universal != nil || status.Code(err) != codes.Unavailable || oldPDP.calls != 0 {
		t.Fatalf("namespaced resource fell back to legacy: response=%+v err=%v calls=%d", universal, err, oldPDP.calls)
	}
	tenant, _ := domain.NewTenantID(testTenant)
	owner := &universalOwnerStub{}
	catalog := &universalCatalogStub{ok: true}
	graph := &universalGraphStub{allowed: true}
	base, err := consistency.NewUniversalBarrier(fakePolicyState{state: consistency.PolicyState{Head: 7, Applied: 7}}, owner, catalog, graph)
	if err != nil {
		t.Fatal(err)
	}
	restricted, err := consistency.NewRestrictedUniversalBarrier(base, restrictionsStub{complete: true}, evidenceStub{})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewWithUniversal(testPrincipalResolver{principal: Principal{TenantID: tenant, Subject: mustSubject(t, "user", "alice")}}, legacy.barrier, restricted, "model-v1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.EvaluateAccess(context.Background(), universalRequest("alice"))
	if err != nil || !result.Decision || graph.called != 1 || oldPDP.calls != 0 || owner.calls < 2 || catalog.calls < 2 {
		t.Fatalf("universal PDP not used: %+v %v graph=%d legacy=%d owner=%d catalog=%d", result, err, graph.called, oldPDP.calls, owner.calls, catalog.calls)
	}
	result, err = service.EvaluateAccess(context.Background(), universalRequest("bob"))
	if result != nil || status.Code(err) != codes.PermissionDenied || graph.called != 1 {
		t.Fatalf("unauthorized delegated subject checked: %+v %v", result, err)
	}
	query := universalRequest("alice")
	query.Context, _ = structpb.NewStruct(map[string]any{"mfa": true})
	result, err = service.EvaluateAccess(context.Background(), query)
	if result != nil || status.Code(err) != codes.InvalidArgument || graph.called != 1 {
		t.Fatalf("untrusted MFA passed: %+v %v", result, err)
	}
	catalog.ok = false
	result, err = service.EvaluateAccess(context.Background(), universalRequest("alice"))
	if result != nil || status.Code(err) != codes.Unavailable || graph.called != 1 {
		t.Fatalf("disabled module allowed: %+v %v", result, err)
	}
}
func mustSubject(t *testing.T, kind, id string) domain.Subject {
	t.Helper()
	result, err := domain.NewSubject(kind, id)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func TestUniversalAuthZENRequiresRestrictedChecker(t *testing.T) {
	s, _ := makeService(t, 1, nil)
	if _, err := NewWithUniversal(s.resolver, s.barrier, nil, "model-v1"); !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatal(err)
	}
}
