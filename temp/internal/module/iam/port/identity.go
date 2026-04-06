package port

import (
	"context"

	"github.com/m8platform/platform/iam/internal/module/iam/entity"
	"github.com/m8platform/platform/iam/internal/module/iam/model"
)

type ServiceAccountRepository interface {
	Save(ctx context.Context, account entity.ServiceAccount) error
}

type ServiceAccountProvisioner interface {
	CreateConfidentialClient(ctx context.Context, tenantID string, clientID string, displayName string, serviceAccountsEnabled bool) (string, error)
}

type ServiceAccountWorkflowStarter interface {
	StartCreateServiceAccount(ctx context.Context, workflow model.CreateServiceAccountWorkflow) error
}

type ServiceAccountEventPublisher interface {
	PublishServiceAccountCreated(ctx context.Context, event model.ServiceAccountCreatedEvent) error
}

type OAuthClientSecretRotator interface {
	RotateOAuthClientSecret(ctx context.Context, oauthClientID string) (string, error)
}

type OAuthClientSecretWorkflowStarter interface {
	StartRotateOAuthClientSecret(ctx context.Context, workflow model.RotateOAuthClientSecretWorkflow) error
}
