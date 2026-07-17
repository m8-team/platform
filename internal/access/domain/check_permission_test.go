package domain

import (
	"errors"
	"testing"
)

func TestCheckPermissionRequestDefaultsToFailClosed(t *testing.T) {
	request := mustCheckPermissionRequest(t, FailMode(""), true, "")

	if request.FailMode() != FailModeDeny {
		t.Fatalf("FailMode() = %q, want %q", request.FailMode(), FailModeDeny)
	}
}

func TestCheckPermissionRequestRejectsCriticalFailOpen(t *testing.T) {
	_, err := newCheckPermissionRequest(t, FailModeAllow, true, "ADR-0000")

	if !errors.Is(err, ErrFailOpenForCriticalCheck) {
		t.Fatalf("NewCheckPermissionRequest() error = %v, want %v", err, ErrFailOpenForCriticalCheck)
	}
}

func TestCheckPermissionRequestRejectsFailOpenWithoutReference(t *testing.T) {
	_, err := newCheckPermissionRequest(t, FailModeAllow, false, "")

	if !errors.Is(err, ErrFailOpenReferenceRequired) {
		t.Fatalf("NewCheckPermissionRequest() error = %v, want %v", err, ErrFailOpenReferenceRequired)
	}
}

func TestPermissionDecisionRequiresModelRevision(t *testing.T) {
	_, err := NewPermissionDecision(DecisionAllow, ModelRevision{}, "matched")

	if !errors.Is(err, ErrModelRevisionRequired) {
		t.Fatalf("NewPermissionDecision() error = %v, want %v", err, ErrModelRevisionRequired)
	}
}

func TestEngineFailureDecisionUsesFailClosed(t *testing.T) {
	request := mustCheckPermissionRequest(t, FailModeDeny, true, "")

	decision, err := NewEngineFailureDecision(request, EngineFailureTimeout, "deadline exceeded")
	if err != nil {
		t.Fatalf("NewEngineFailureDecision() error = %v", err)
	}

	if decision.Decision() != DecisionDeny {
		t.Fatalf("Decision() = %q, want %q", decision.Decision(), DecisionDeny)
	}
	if !decision.Degraded() {
		t.Fatal("Degraded() = false, want true")
	}
	if decision.FailureKind() != EngineFailureTimeout {
		t.Fatalf("FailureKind() = %q, want %q", decision.FailureKind(), EngineFailureTimeout)
	}
}

func TestEngineFailureDecisionAllowsExplicitNonCriticalFailOpen(t *testing.T) {
	request := mustCheckPermissionRequest(t, FailModeAllow, false, "ADR-1234")

	decision, err := NewEngineFailureDecision(request, EngineFailureUnavailable, "engine unavailable")
	if err != nil {
		t.Fatalf("NewEngineFailureDecision() error = %v", err)
	}

	if decision.Decision() != DecisionAllow {
		t.Fatalf("Decision() = %q, want %q", decision.Decision(), DecisionAllow)
	}
	if !decision.Degraded() {
		t.Fatal("Degraded() = false, want true")
	}
	if decision.FailureKind() != EngineFailureUnavailable {
		t.Fatalf("FailureKind() = %q, want %q", decision.FailureKind(), EngineFailureUnavailable)
	}
}

func TestPermissionDecisionDetectsRevisionMismatch(t *testing.T) {
	expected := mustModelRevision(t, "access-model-rev-1")
	actual := mustModelRevision(t, "access-model-rev-2")
	decision, err := NewPermissionDecision(DecisionAllow, actual, "matched")
	if err != nil {
		t.Fatalf("NewPermissionDecision() error = %v", err)
	}

	if err := decision.EnsureModelRevision(expected); !errors.Is(err, ErrModelRevisionMismatch) {
		t.Fatalf("EnsureModelRevision() error = %v, want %v", err, ErrModelRevisionMismatch)
	}
}

func mustCheckPermissionRequest(
	t *testing.T,
	failMode FailMode,
	critical bool,
	failOpenReference string,
) CheckPermissionRequest {
	t.Helper()

	request, err := newCheckPermissionRequest(t, failMode, critical, failOpenReference)
	if err != nil {
		t.Fatalf("NewCheckPermissionRequest() error = %v", err)
	}

	return request
}

func newCheckPermissionRequest(
	t *testing.T,
	failMode FailMode,
	critical bool,
	failOpenReference string,
) (CheckPermissionRequest, error) {
	t.Helper()

	return NewCheckPermissionRequest(CheckPermissionInput{
		Subject:           mustSubject(t),
		Permission:        mustPermission(t),
		Resource:          mustResource(t),
		ModelRevision:     mustModelRevision(t, "access-model-rev-1"),
		FailMode:          failMode,
		Critical:          critical,
		FailOpenReference: failOpenReference,
	})
}

func mustSubject(t *testing.T) Subject {
	t.Helper()

	subject, err := NewSubject("user", "usr_123")
	if err != nil {
		t.Fatalf("NewSubject() error = %v", err)
	}

	return subject
}

func mustPermission(t *testing.T) Permission {
	t.Helper()

	permission, err := NewPermission("m8.project.read")
	if err != nil {
		t.Fatalf("NewPermission() error = %v", err)
	}

	return permission
}

func mustResource(t *testing.T) Resource {
	t.Helper()

	resource, err := NewResource("project", "prj_123")
	if err != nil {
		t.Fatalf("NewResource() error = %v", err)
	}

	return resource
}

func mustModelRevision(t *testing.T, value string) ModelRevision {
	t.Helper()

	revision, err := NewModelRevision(value)
	if err != nil {
		t.Fatalf("NewModelRevision() error = %v", err)
	}

	return revision
}
