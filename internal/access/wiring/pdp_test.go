package wiring

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/domain"
)

type principal struct {
	tenant  domain.TenantID
	subject domain.Subject
}

func (p principal) AuthenticatedPrincipal(context.Context) (authzen.Principal, error) {
	return authzen.Principal{TenantID: p.tenant, Subject: p.subject}, nil
}

type source struct{}

func (source) AuthoritativeRevision(context.Context, domain.TenantID, domain.Resource) (uint64, error) {
	return 2, nil
}

type checkpoint struct{}

func (checkpoint) LoadWatermark(context.Context, domain.TenantID, domain.Resource) (consistency.Watermark, error) {
	return consistency.Watermark{Revision: 1, ZedToken: "old"}, nil
}

type readyPolicyState struct{}

func (readyPolicyState) LoadPolicyState(context.Context, domain.TenantID) (consistency.PolicyState, error) {
	return consistency.PolicyState{Head: 4, Applied: 4}, nil
}

type allowEngine struct{ calls int }

func (e *allowEngine) CheckPermission(_ context.Context, req domain.CheckPermissionRequest) (domain.PermissionDecision, error) {
	e.calls++
	return domain.NewPermissionDecision(domain.DecisionAllow, req.ModelRevision(), "allowed")
}
func TestHTTPIsWiredToBarrierAndRejectsStalePositive(t *testing.T) {
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	subject, _ := domain.NewSubject("user", "alice")
	engine := &allowEngine{}
	handler, err := NewAuthZENHandler(context.Background(), principal{tenant, subject}, source{}, checkpoint{}, readyPolicyState{}, engine, "model-v1")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/access/v1/evaluation", strings.NewReader(
		`{"subject":{"type":"user","id":"alice"},"action":{"name":"m8.workspace.read"},"resource":{"type":"workspace","id":"11111111-1111-4111-8111-111111111111"}}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code == 200 || engine.calls != 0 || strings.Contains(response.Body.String(), "\"decision\":true") {
		t.Fatalf("stale RM projection bypassed PDP: HTTP %d body=%s engine=%d", response.Code, response.Body.String(), engine.calls)
	}
}
func TestNoAuthZENHandlerWithoutTrustedComponents(t *testing.T) {
	if _, err := NewAuthZENHandler(context.Background(), nil, nil, nil, nil, nil, "model-v1"); err == nil {
		t.Fatal("untrusted handler was constructed")
	}
}
