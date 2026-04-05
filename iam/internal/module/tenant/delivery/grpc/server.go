package grpc

import (
	legacygrpc "github.com/m8platform/platform/iam/internal/adapter/in/grpc"
	legacysupport "github.com/m8platform/platform/iam/internal/support"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
	"go.uber.org/zap"
)

type Server struct {
	*legacygrpc.SupportServer
}

func NewServer(legacy *legacysupport.Service, logger *zap.Logger, useCase *tenantuc.SupportAccessUseCase) *Server {
	return &Server{
		SupportServer: legacygrpc.NewSupportServer(legacy, logger, useCase),
	}
}
