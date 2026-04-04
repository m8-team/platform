package grpc

import (
	"context"

	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	legacyidentity "github.com/m8platform/platform/iam/internal/identity"
	identityuc "github.com/m8platform/platform/iam/internal/usecase/identity"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IdentityServer struct {
	*legacyidentity.Service

	logger               *zap.Logger
	createServiceAccount *identityuc.CreateServiceAccountUseCase
	rotateClientSecret   *identityuc.RotateOAuthClientSecretUseCase
}

func NewIdentityServer(
	legacy *legacyidentity.Service,
	logger *zap.Logger,
	createServiceAccount *identityuc.CreateServiceAccountUseCase,
	rotateClientSecret *identityuc.RotateOAuthClientSecretUseCase,
) *IdentityServer {
	return &IdentityServer{
		Service:              legacy,
		logger:               logger,
		createServiceAccount: createServiceAccount,
		rotateClientSecret:   rotateClientSecret,
	}
}

func (s *IdentityServer) CreateServiceAccount(ctx context.Context, req *identityv1.CreateServiceAccountRequest) (*identityv1.ServiceAccount, error) {
	if s.createServiceAccount == nil {
		return s.Service.CreateServiceAccount(ctx, req)
	}

	result, err := s.createServiceAccount.Execute(ctx, model.CreateServiceAccountCommand{
		ServiceAccountID: req.GetServiceAccountId(),
		TenantID:         req.GetTenantId(),
		DisplayName:      req.GetDisplayName(),
		Description:      req.GetDescription(),
		PerformedBy:      req.GetPerformedBy(),
	})
	if err != nil {
		return nil, err
	}

	s.logWarnings("create service account", result.Warnings)
	return &identityv1.ServiceAccount{
		ServiceAccountId: result.Account.ID,
		TenantId:         result.Account.TenantID,
		DisplayName:      result.Account.DisplayName,
		Description:      result.Account.Description,
		Disabled:         result.Account.Disabled,
		KeycloakClientId: result.Account.KeycloakClientID,
		OperationId:      result.Account.OperationID,
		CreatedAt:        timestamppb.New(result.Account.CreatedAt.UTC()),
		UpdatedAt:        timestamppb.New(result.Account.UpdatedAt.UTC()),
	}, nil
}

func (s *IdentityServer) RotateClientSecret(ctx context.Context, req *identityv1.RotateClientSecretRequest) (*identityv1.RotateClientSecretResponse, error) {
	if s.rotateClientSecret == nil {
		return s.Service.RotateClientSecret(ctx, req)
	}

	result, err := s.rotateClientSecret.Execute(ctx, model.RotateOAuthClientSecretCommand{
		OAuthClientID: req.GetOauthClientId(),
		PerformedBy:   req.GetPerformedBy(),
		Reason:        req.GetReason(),
	})
	if err != nil {
		return nil, err
	}

	s.logWarnings("rotate client secret", result.Warnings)
	return &identityv1.RotateClientSecretResponse{
		OperationId: result.OperationID,
		SecretRef:   result.SecretRef,
	}, nil
}

func (s *IdentityServer) logWarnings(operation string, warnings []error) {
	if s == nil || s.logger == nil {
		return
	}
	for _, warning := range warnings {
		if warning == nil {
			continue
		}
		s.logger.Warn(operation+" degraded", zap.Error(warning))
	}
}
