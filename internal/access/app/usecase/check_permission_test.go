package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/app/command"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

func TestCheckPermissionHandlerReturnsDeterministicRevisionDecision(t *testing.T) {
	revision := mustRevision(t, "access-model-rev-1")
	engineDecision, err := domain.NewPermissionDecision(domain.DecisionAllow, revision, "direct relationship")
	if err != nil {
		t.Fatalf("NewPermissionDecision() error = %v", err)
	}

	handler := mustHandler(t, fakePermissionEngine{decision: engineDecision})

	result, err := handler.Handle(context.Background(), baseCommand())
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Allowed {
		t.Fatal("Allowed = false, want true")
	}
	if result.Decision != domain.DecisionAllow.String() {
		t.Fatalf("Decision = %q, want %q", result.Decision, domain.DecisionAllow)
	}
	if result.ModelRevision != "access-model-rev-1" {
		t.Fatalf("ModelRevision = %q, want access-model-rev-1", result.ModelRevision)
	}
	if result.Degraded {
		t.Fatal("Degraded = true, want false")
	}
}

func TestCheckPermissionHandlerRejectsRevisionMismatch(t *testing.T) {
	engineDecision, err := domain.NewPermissionDecision(
		domain.DecisionAllow,
		mustRevision(t, "access-model-rev-2"),
		"stale engine result",
	)
	if err != nil {
		t.Fatalf("NewPermissionDecision() error = %v", err)
	}

	handler := mustHandler(t, fakePermissionEngine{decision: engineDecision})

	_, err = handler.Handle(context.Background(), baseCommand())
	if !errors.Is(err, domain.ErrModelRevisionMismatch) {
		t.Fatalf("Handle() error = %v, want %v", err, domain.ErrModelRevisionMismatch)
	}
}

func TestCheckPermissionHandlerFailClosedOnTimeout(t *testing.T) {
	handler := mustHandler(t, fakePermissionEngine{err: ports.ErrPermissionEngineTimeout})

	result, err := handler.Handle(context.Background(), baseCommand())
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if result.Allowed {
		t.Fatal("Allowed = true, want false")
	}
	if result.Decision != domain.DecisionDeny.String() {
		t.Fatalf("Decision = %q, want %q", result.Decision, domain.DecisionDeny)
	}
	if !result.Degraded {
		t.Fatal("Degraded = false, want true")
	}
	if result.FailureKind != domain.EngineFailureTimeout.String() {
		t.Fatalf("FailureKind = %q, want %q", result.FailureKind, domain.EngineFailureTimeout)
	}
}

func TestCheckPermissionHandlerExplicitFailOpenOnUnavailable(t *testing.T) {
	handler := mustHandler(t, fakePermissionEngine{err: ports.ErrPermissionEngineUnavailable})
	cmd := baseCommand()
	cmd.FailMode = domain.FailModeAllow
	cmd.Critical = false
	cmd.FailOpenReference = "ADR-1234"

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !result.Allowed {
		t.Fatal("Allowed = false, want true")
	}
	if result.Decision != domain.DecisionAllow.String() {
		t.Fatalf("Decision = %q, want %q", result.Decision, domain.DecisionAllow)
	}
	if !result.Degraded {
		t.Fatal("Degraded = false, want true")
	}
	if result.FailureKind != domain.EngineFailureUnavailable.String() {
		t.Fatalf("FailureKind = %q, want %q", result.FailureKind, domain.EngineFailureUnavailable)
	}
}

func TestCheckPermissionHandlerRejectsCriticalFailOpenBeforeEngineCall(t *testing.T) {
	engine := &countingPermissionEngine{}
	handler := mustHandler(t, engine)
	cmd := baseCommand()
	cmd.FailMode = domain.FailModeAllow
	cmd.Critical = true
	cmd.FailOpenReference = "ADR-1234"

	_, err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, domain.ErrFailOpenForCriticalCheck) {
		t.Fatalf("Handle() error = %v, want %v", err, domain.ErrFailOpenForCriticalCheck)
	}
	if engine.calls != 0 {
		t.Fatalf("engine calls = %d, want 0", engine.calls)
	}
}

func TestCheckPermissionHandlerReturnsUnknownEngineError(t *testing.T) {
	errEngine := errors.New("engine internal error")
	handler := mustHandler(t, fakePermissionEngine{err: errEngine})

	_, err := handler.Handle(context.Background(), baseCommand())
	if !errors.Is(err, errEngine) {
		t.Fatalf("Handle() error = %v, want wrapped %v", err, errEngine)
	}
}

func TestNewCheckPermissionHandlerRequiresEngine(t *testing.T) {
	_, err := NewCheckPermissionHandler(nil)
	if !errors.Is(err, ErrPermissionEngineRequired) {
		t.Fatalf("NewCheckPermissionHandler() error = %v, want %v", err, ErrPermissionEngineRequired)
	}
}

type fakePermissionEngine struct {
	decision domain.PermissionDecision
	err      error
}

func (f fakePermissionEngine) CheckPermission(
	_ context.Context,
	_ domain.CheckPermissionRequest,
) (domain.PermissionDecision, error) {
	return f.decision, f.err
}

type countingPermissionEngine struct {
	calls int
}

func (f *countingPermissionEngine) CheckPermission(
	_ context.Context,
	_ domain.CheckPermissionRequest,
) (domain.PermissionDecision, error) {
	f.calls++
	return domain.PermissionDecision{}, nil
}

func mustHandler(t *testing.T, engine ports.PermissionEngine) *CheckPermissionHandler {
	t.Helper()

	handler, err := NewCheckPermissionHandler(engine)
	if err != nil {
		t.Fatalf("NewCheckPermissionHandler() error = %v", err)
	}

	return handler
}

func baseCommand() command.CheckPermissionCommand {
	return command.CheckPermissionCommand{
		SubjectType:   "user",
		SubjectID:     "usr_123",
		Permission:    "m8.project.read",
		ResourceType:  "project",
		ResourceID:    "prj_123",
		ModelRevision: "access-model-rev-1",
		FailMode:      domain.FailModeDeny,
		Critical:      true,
	}
}

func mustRevision(t *testing.T, value string) domain.ModelRevision {
	t.Helper()

	revision, err := domain.NewModelRevision(value)
	if err != nil {
		t.Fatalf("NewModelRevision() error = %v", err)
	}

	return revision
}
