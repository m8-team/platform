package identity

import (
	"context"

	identityentity "github.com/m8platform/platform/iam/internal/entity/identity"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	"github.com/m8platform/platform/iam/internal/usecase/port"
)

type CreateServiceAccountUseCase struct {
	clock        port.Clock
	repository   port.ServiceAccountRepository
	provisioner  port.ServiceAccountProvisioner
	workflows    port.ServiceAccountWorkflowStarter
	eventPublish port.ServiceAccountEventPublisher
}

func NewCreateServiceAccountUseCase(
	clock port.Clock,
	repository port.ServiceAccountRepository,
	provisioner port.ServiceAccountProvisioner,
	workflows port.ServiceAccountWorkflowStarter,
	eventPublish port.ServiceAccountEventPublisher,
) *CreateServiceAccountUseCase {
	return &CreateServiceAccountUseCase{
		clock:        clock,
		repository:   repository,
		provisioner:  provisioner,
		workflows:    workflows,
		eventPublish: eventPublish,
	}
}

func (u *CreateServiceAccountUseCase) Execute(ctx context.Context, cmd model.CreateServiceAccountCommand) (model.CreateServiceAccountResult, error) {
	now := u.clock.Now().UTC()
	account, err := identityentity.NewServiceAccount(identityentity.NewServiceAccountParams{
		ID:          cmd.ServiceAccountID,
		TenantID:    cmd.TenantID,
		DisplayName: cmd.DisplayName,
		Description: cmd.Description,
		Now:         now,
	})
	if err != nil {
		return model.CreateServiceAccountResult{}, err
	}

	warnings := make([]error, 0, 2)
	if u.provisioner != nil {
		keycloakClientID, provisionErr := u.provisioner.CreateConfidentialClient(ctx, account.TenantID, account.ID, account.DisplayName, true)
		if provisionErr != nil {
			warnings = append(warnings, provisionErr)
		} else {
			account = account.WithKeycloakClientID(keycloakClientID)
		}
	}

	if err := u.repository.Save(ctx, account); err != nil {
		return model.CreateServiceAccountResult{}, err
	}

	if u.workflows != nil {
		if workflowErr := u.workflows.StartCreateServiceAccount(ctx, model.CreateServiceAccountWorkflow{
			OperationID:      account.OperationID,
			ServiceAccountID: account.ID,
			TenantID:         account.TenantID,
			DisplayName:      account.DisplayName,
			Description:      account.Description,
			RequestedBy:      cmd.PerformedBy,
		}); workflowErr != nil {
			warnings = append(warnings, workflowErr)
		}
	}

	if err := u.eventPublish.PublishServiceAccountCreated(ctx, model.ServiceAccountCreatedEvent{
		OperationID: account.OperationID,
		OccurredAt:  now,
		PerformedBy: cmd.PerformedBy,
		Account:     account,
	}); err != nil {
		return model.CreateServiceAccountResult{}, err
	}

	return model.CreateServiceAccountResult{
		Account:  account,
		Warnings: warnings,
	}, nil
}
