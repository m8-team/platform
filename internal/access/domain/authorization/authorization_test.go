package authorization

import (
	"errors"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
	"testing"
)

const tenantID = "11111111-1111-4111-8111-111111111111"

func TestCriticalCheckFailsClosed(t *testing.T) {
	tenant, _ := shared.NewTenantID(tenantID)
	subject, _ := shared.NewSubject("user", "alice")
	resource, _ := shared.NewResource("workspace", "workspace-1")
	p, _ := permission.NewPermission("workspace.view")
	rev, _ := shared.NewModelRevision("v1")
	input := CheckPermissionInput{TenantID: tenant, Subject: subject, Resource: resource, Permission: p, ModelRevision: rev, Critical: true, FailMode: FailAllow, FailOpenReference: "exception-1"}
	if _, err := NewCheckPermissionRequest(input); !errors.Is(err, ErrFailOpenForCriticalCheck) {
		t.Fatal(err)
	}
	input.FailMode = FailDeny
	req, err := NewCheckPermissionRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := NewEngineFailureDecision(req, FailureUnavailable, "unavailable")
	if err != nil || decision.Allowed() || !decision.Degraded() {
		t.Fatalf("failure must deny: %v", err)
	}
	if _, err = NewPermissionDecision(Allow, rev, "ok"); err != nil {
		t.Fatal(err)
	}
}
func TestRelationshipTupleIdentity(t *testing.T) {
	tenant, _ := shared.NewTenantID(tenantID)
	resource, _ := shared.NewResource("workspace", "workspace-1")
	subject, _ := shared.NewSubject("user", "alice")
	base := RelationshipInput{TenantID: tenant, Resource: resource, Subject: subject, Relation: "viewer"}
	a, err := NewRelationship(base)
	if err != nil {
		t.Fatal(err)
	}
	base.ConditionRef = "only_if_valid"
	b, err := NewRelationship(base)
	if err != nil {
		t.Fatal(err)
	}
	if a.Key() != b.Key() {
		t.Fatal("condition must not change tuple identity")
	}
	base.Relation = "Invalid"
	if _, err = NewRelationship(base); !errors.Is(err, ErrInvalidRelationName) {
		t.Fatal(err)
	}
}
