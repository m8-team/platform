package port

import (
	"context"

	identityentity "github.com/m8platform/platform/iam/internal/entity/identity"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type ServiceAccountRepository interface {
	Save(ctx context.Context, account identityentity.ServiceAccount) error
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
