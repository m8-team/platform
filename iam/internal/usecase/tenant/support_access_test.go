package tenant

import (
	"context"
	"errors"
	"testing"
	"time"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type supportFixedClock struct {
	now time.Time
}

func (c supportFixedClock) Now() time.Time {
	return c.now
}

type supportGrantRepositoryFake struct {
	byID map[string]tenantentity.SupportGrant
	list []tenantentity.SupportGrant
	err  error
}

func (f *supportGrantRepositoryFake) Save(_ context.Context, grant tenantentity.SupportGrant) error {
	if f.err != nil {
		return f.err
	}
	if f.byID == nil {
		f.byID = make(map[string]tenantentity.SupportGrant)
	}
	f.byID[grant.ID] = grant
	return nil
}

func (f *supportGrantRepositoryFake) GetByID(_ context.Context, supportGrantID string) (tenantentity.SupportGrant, error) {
	if f.err != nil {
		return tenantentity.SupportGrant{}, f.err
	}
	grant, ok := f.byID[supportGrantID]
	if !ok {
		return tenantentity.SupportGrant{}, errors.New("not found")
	}
	return grant, nil
}

func (f *supportGrantRepositoryFake) ListByTenant(_ context.Context, _ string, _ int, _ string) ([]tenantentity.SupportGrant, string, error) {
	return f.list, "", f.err
}

type supportGrantEventPublisherFake struct {
	created []model.SupportGrantCreatedEvent
	revoked []model.SupportGrantRevokedEvent
	err     error
}

func (f *supportGrantEventPublisherFake) PublishSupportGrantCreated(_ context.Context, event model.SupportGrantCreatedEvent) error {
	f.created = append(f.created, event)
	return f.err
}

func (f *supportGrantEventPublisherFake) PublishSupportGrantRevoked(_ context.Context, event model.SupportGrantRevokedEvent) error {
	f.revoked = append(f.revoked, event)
	return f.err
}

type supportGrantWorkflowStarterFake struct {
	workflows []model.SupportGrantExpiryWorkflow
	err       error
}

func (f *supportGrantWorkflowStarterFake) StartSupportGrantExpiry(_ context.Context, workflow model.SupportGrantExpiryWorkflow) error {
	f.workflows = append(f.workflows, workflow)
	return f.err
}

func TestSupportAccessUseCaseGrantApproveRevoke(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 16, 0, 0, 0, time.UTC)
	repository := &supportGrantRepositoryFake{}
	events := &supportGrantEventPublisherFake{}
	workflows := &supportGrantWorkflowStarterFake{err: errors.New("temporal unavailable")}
	useCase := NewSupportAccessUseCase(supportFixedClock{now: now}, repository, events, workflows)

	granted, err := useCase.Grant(context.Background(), model.GrantSupportAccessCommand{
		TenantID: "tenant-1",
		Subject:  authzentity.SubjectRef{Type: "SUBJECT_TYPE_USER_ACCOUNT", ID: "user-1"},
		Resource: authzentity.ResourceRef{Type: "RESOURCE_TYPE_PROJECT", ID: "project-1"},
		RoleID:   "support-operator",
		TTL:      5 * time.Minute,
		Reason:   "incident triage",
	})
	if err != nil {
		t.Fatalf("grant failed: %v", err)
	}
	if granted.Grant.Status != tenantentity.SupportGrantStatusPendingApproval {
		t.Fatalf("expected pending status, got %s", granted.Grant.Status)
	}
	if len(events.created) != 1 {
		t.Fatalf("expected 1 created event, got %d", len(events.created))
	}

	approved, err := useCase.Approve(context.Background(), model.ApproveSupportAccessCommand{
		SupportGrantID: granted.Grant.ID,
		ApprovalTicket: "TICKET-1",
		Reason:         "approved",
		ApprovedBy:     "lead-1",
	})
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if approved.Grant.Status != tenantentity.SupportGrantStatusActive {
		t.Fatalf("expected active status, got %s", approved.Grant.Status)
	}
	if len(approved.Warnings) != 1 {
		t.Fatalf("expected workflow warning, got %d", len(approved.Warnings))
	}

	revoked, err := useCase.Revoke(context.Background(), model.RevokeSupportAccessCommand{
		SupportGrantID: granted.Grant.ID,
		Reason:         "session complete",
		RevokedBy:      "lead-1",
	})
	if err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if revoked.Grant.Status != tenantentity.SupportGrantStatusRevoked {
		t.Fatalf("expected revoked status, got %s", revoked.Grant.Status)
	}
	if len(events.revoked) != 1 {
		t.Fatalf("expected 1 revoked event, got %d", len(events.revoked))
	}
}
