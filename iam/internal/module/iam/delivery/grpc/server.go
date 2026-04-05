package grpc

import (
	legacygrpc "github.com/m8platform/platform/iam/internal/adapter/in/grpc"
	legacyidentity "github.com/m8platform/platform/iam/internal/identity"
	identityuc "github.com/m8platform/platform/iam/internal/usecase/identity"
	"go.uber.org/zap"
)

type Server struct {
	*legacygrpc.IdentityServer
}

func NewServer(
	legacy *legacyidentity.Service,
	logger *zap.Logger,
	createServiceAccount *identityuc.CreateServiceAccountUseCase,
	rotateClientSecret *identityuc.RotateOAuthClientSecretUseCase,
) *Server {
	return &Server{
		IdentityServer: legacygrpc.NewIdentityServer(legacy, logger, createServiceAccount, rotateClientSecret),
	}
}
