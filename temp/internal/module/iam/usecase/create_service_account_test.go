package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8platform/platform/iam/internal/module/iam/entity"
	"github.com/m8platform/platform/iam/internal/module/iam/model"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type serviceAccountRepositoryFake struct {
	saved []entity.ServiceAccount
	err   error
}

func (f *serviceAccountRepositoryFake) Save(_ context.Context, account entity.ServiceAccount) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, account)
	return nil
}

type provisionerFake struct {
	keycloakClientID string
	err              error
}

func (f provisionerFake) CreateConfidentialClient(_ context.Context, _, _, _ string, _ bool) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.keycloakClientID, nil
}

type createWorkflowStarterFake struct {
	workflows []model.CreateServiceAccountWorkflow
	err       error
}

func (f *createWorkflowStarterFake) StartCreateServiceAccount(_ context.Context, workflow model.CreateServiceAccountWorkflow) error {
	f.workflows = append(f.workflows, workflow)
	return f.err
}

type serviceAccountEventPublisherFake struct {
	events []model.ServiceAccountCreatedEvent
	err    error
}

func (f *serviceAccountEventPublisherFake) PublishServiceAccountCreated(_ context.Context, event model.ServiceAccountCreatedEvent) error {
	f.events = append(f.events, event)
	return f.err
}

func TestCreateServiceAccountUseCaseExecute(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	repository := &serviceAccountRepositoryFake{}
	workflows := &createWorkflowStarterFake{err: errors.New("temporal unavailable")}
	publisher := &serviceAccountEventPublisherFake{}

	useCase := NewCreateServiceAccountUseCase(
		fixedClock{now: now},
		repository,
		provisionerFake{keycloakClientID: "kc-client-1", err: errors.New("keycloak unavailable")},
		workflows,
		publisher,
	)

	result, err := useCase.Execute(context.Background(), model.CreateServiceAccountCommand{
		ServiceAccountID: "sa-1",
		TenantID:         "tenant-1",
		DisplayName:      "Platform Bot",
		Description:      "automation",
		PerformedBy:      "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(result.Warnings))
	}
	if len(repository.saved) != 1 {
		t.Fatalf("expected 1 saved account, got %d", len(repository.saved))
	}
	if repository.saved[0].KeycloakClientID != "" {
		t.Fatalf("expected keycloak client id to stay empty on degraded provisioning, got %q", repository.saved[0].KeycloakClientID)
	}
	if len(workflows.workflows) != 1 {
		t.Fatalf("expected workflow to be requested once, got %d", len(workflows.workflows))
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(publisher.events))
	}
	if publisher.events[0].Account.ID != "sa-1" {
		t.Fatalf("expected event account id sa-1, got %q", publisher.events[0].Account.ID)
	}
}
