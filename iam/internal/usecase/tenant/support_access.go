package tenant

import (
	"context"
	"fmt"

	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	"github.com/m8platform/platform/iam/internal/usecase/port"
)

type SupportAccessUseCase struct {
	clock      port.Clock
	repository port.SupportGrantRepository
	events     port.SupportGrantEventPublisher
	workflows  port.SupportGrantWorkflowStarter
}

func NewSupportAccessUseCase(
	clock port.Clock,
	repository port.SupportGrantRepository,
	events port.SupportGrantEventPublisher,
	workflows port.SupportGrantWorkflowStarter,
) *SupportAccessUseCase {
	return &SupportAccessUseCase{
		clock:      clock,
		repository: repository,
		events:     events,
		workflows:  workflows,
	}
}

func (u *SupportAccessUseCase) Grant(ctx context.Context, cmd model.GrantSupportAccessCommand) (model.SupportGrantResult, error) {
	now := u.clock.Now().UTC()
	grant, err := tenantentity.NewSupportGrant(tenantentity.NewSupportGrantParams{
		TenantID: cmd.TenantID,
		Subject:  cmd.Subject,
		Resource: cmd.Resource,
		RoleID:   cmd.RoleID,
		TTL:      cmd.TTL,
		Reason:   cmd.Reason,
		Now:      now,
	})
	if err != nil {
		return model.SupportGrantResult{}, err
	}
	if err := u.repository.Save(ctx, grant); err != nil {
		return model.SupportGrantResult{}, err
	}
	if u.events != nil {
		if err := u.events.PublishSupportGrantCreated(ctx, model.SupportGrantCreatedEvent{
			EventID:     grant.ID,
			OccurredAt:  now,
			RequestedBy: cmd.RequestedBy,
			Grant:       grant,
		}); err != nil {
			return model.SupportGrantResult{}, err
		}
	}
	return model.SupportGrantResult{Grant: grant}, nil
}

func (u *SupportAccessUseCase) Approve(ctx context.Context, cmd model.ApproveSupportAccessCommand) (model.SupportGrantResult, error) {
	grant, err := u.repository.GetByID(ctx, cmd.SupportGrantID)
	if err != nil {
		return model.SupportGrantResult{}, err
	}
	now := u.clock.Now().UTC()
	grant = grant.Approve(cmd.ApprovalTicket, now)
	if err := u.repository.Save(ctx, grant); err != nil {
		return model.SupportGrantResult{}, err
	}

	warnings := make([]error, 0, 1)
	if u.workflows != nil {
		if err := u.workflows.StartSupportGrantExpiry(ctx, model.SupportGrantExpiryWorkflow{
			SupportGrantID: grant.ID,
			TenantID:       grant.TenantID,
			RequestedBy:    cmd.ApprovedBy,
			Reason:         cmd.Reason,
			TTL:            grant.TTL,
		}); err != nil {
			warnings = append(warnings, err)
		}
	}

	return model.SupportGrantResult{
		Grant:    grant,
		Warnings: warnings,
	}, nil
}

func (u *SupportAccessUseCase) Revoke(ctx context.Context, cmd model.RevokeSupportAccessCommand) (model.SupportGrantResult, error) {
	grant, err := u.repository.GetByID(ctx, cmd.SupportGrantID)
	if err != nil {
		return model.SupportGrantResult{}, err
	}
	now := u.clock.Now().UTC()
	grant = grant.Revoke(now)
	if err := u.repository.Save(ctx, grant); err != nil {
		return model.SupportGrantResult{}, err
	}
	if u.events != nil {
		if err := u.events.PublishSupportGrantRevoked(ctx, model.SupportGrantRevokedEvent{
			EventID:    fmt.Sprintf("revoke-%d", now.UnixNano()),
			OccurredAt: now,
			RevokedBy:  cmd.RevokedBy,
			Reason:     cmd.Reason,
			Grant:      grant,
		}); err != nil {
			return model.SupportGrantResult{}, err
		}
	}
	return model.SupportGrantResult{Grant: grant}, nil
}

func (u *SupportAccessUseCase) List(ctx context.Context, query model.ListSupportGrantsQuery) ([]tenantentity.SupportGrant, string, error) {
	return u.repository.ListByTenant(ctx, query.TenantID, query.PageSize, query.PageToken)
}
