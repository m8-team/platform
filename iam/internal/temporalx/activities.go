package temporalx

import (
	"context"

	"go.uber.org/zap"
)

type Activities struct {
	Logger *zap.Logger
}

func (a *Activities) CreateServiceAccountMetadata(_ context.Context, input CreateServiceAccountInput) error {
	a.Logger.Info("create service account metadata", zap.String("service_account_id", input.ServiceAccountID))
	return nil
}

func (a *Activities) CreateKeycloakServiceAccount(_ context.Context, input CreateServiceAccountInput) error {
	a.Logger.Info("create keycloak service account", zap.String("service_account_id", input.ServiceAccountID))
	return nil
}

func (a *Activities) SyncServiceAccountBindings(_ context.Context, input CreateServiceAccountInput) error {
	a.Logger.Info("sync service account bindings", zap.String("service_account_id", input.ServiceAccountID))
	return nil
}

func (a *Activities) WriteServiceAccountAudit(_ context.Context, input CreateServiceAccountInput) error {
	a.Logger.Info("write service account audit", zap.String("service_account_id", input.ServiceAccountID))
	return nil
}

func (a *Activities) RotateKeycloakClientSecret(_ context.Context, input RotateClientSecretInput) (string, error) {
	a.Logger.Info("rotate client secret", zap.String("oauth_client_id", input.OAuthClientID))
	return "vault://keycloak/rotated", nil
}

func (a *Activities) WriteSecretRotationAudit(_ context.Context, input RotateClientSecretInput) error {
	a.Logger.Info("write client secret audit", zap.String("oauth_client_id", input.OAuthClientID))
	return nil
}

func (a *Activities) ActivateSupportGrant(_ context.Context, input GrantTemporarySupportAccessInput) error {
	a.Logger.Info("activate support grant", zap.String("support_grant_id", input.SupportGrantID))
	return nil
}

func (a *Activities) ExpireSupportGrant(_ context.Context, input GrantTemporarySupportAccessInput) error {
	a.Logger.Info("expire support grant", zap.String("support_grant_id", input.SupportGrantID))
	return nil
}

func (a *Activities) SyncRelationshipBatch(_ context.Context, input SyncRelationshipsInput) error {
	a.Logger.Info("sync relationship batch", zap.Int("batch_size", input.BatchSize))
	return nil
}

func (a *Activities) RebuildProjection(_ context.Context, input RebuildAccessReadModelsInput) error {
	a.Logger.Info("rebuild projection", zap.String("projection", input.Projection))
	return nil
}

func (a *Activities) ImportFederatedUser(_ context.Context, input ImportFederatedUserInput) error {
	a.Logger.Info("import federated user", zap.String("tenant_id", input.TenantID), zap.String("provider", input.Provider))
	return nil
}
