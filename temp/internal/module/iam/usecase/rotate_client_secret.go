package usecase

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/iam/internal/module/iam/model"
	"github.com/m8platform/platform/iam/internal/module/iam/port"
	sharedclock "github.com/m8platform/platform/iam/internal/shared/clock"
)

type RotateOAuthClientSecretUseCase struct {
	clock     sharedclock.Clock
	rotator   port.OAuthClientSecretRotator
	workflows port.OAuthClientSecretWorkflowStarter
}

func NewRotateOAuthClientSecretUseCase(
	clock sharedclock.Clock,
	rotator port.OAuthClientSecretRotator,
	workflows port.OAuthClientSecretWorkflowStarter,
) *RotateOAuthClientSecretUseCase {
	return &RotateOAuthClientSecretUseCase{
		clock:     clock,
		rotator:   rotator,
		workflows: workflows,
	}
}

func (u *RotateOAuthClientSecretUseCase) Execute(ctx context.Context, cmd model.RotateOAuthClientSecretCommand) (model.RotateOAuthClientSecretResult, error) {
	now := u.clock.Now().UTC()
	operationID := fmt.Sprintf("rotate-%d", now.UnixNano())
	secretRef := fmt.Sprintf("vault://oauth/%s/%s", "clients", cmd.OAuthClientID)
	warnings := make([]error, 0, 2)

	if u.rotator != nil {
		ref, err := u.rotator.RotateOAuthClientSecret(ctx, cmd.OAuthClientID)
		if err != nil {
			warnings = append(warnings, err)
		} else if ref != "" {
			secretRef = ref
		}
	}

	if u.workflows != nil {
		if err := u.workflows.StartRotateOAuthClientSecret(ctx, model.RotateOAuthClientSecretWorkflow{
			OperationID:   operationID,
			OAuthClientID: cmd.OAuthClientID,
			RequestedBy:   cmd.PerformedBy,
			Reason:        cmd.Reason,
		}); err != nil {
			warnings = append(warnings, err)
		}
	}

	return model.RotateOAuthClientSecretResult{
		OperationID: operationID,
		SecretRef:   secretRef,
		Warnings:    warnings,
	}, nil
}
