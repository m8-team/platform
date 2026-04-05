package grpc

import (
	legacygrpc "github.com/m8platform/platform/iam/internal/adapter/in/grpc"
	legacyauthz "github.com/m8platform/platform/iam/internal/authz"
	authzuc "github.com/m8platform/platform/iam/internal/usecase/authz"
	"github.com/m8platform/platform/iam/internal/usecase/port"
)

type Server struct {
	*legacygrpc.AuthorizationServer
}

func NewServer(
	legacy *legacyauthz.Service,
	checkAccess *authzuc.CheckAccessUseCase,
	bindings port.AccessBindingRepository,
	roles port.RolePermissionResolver,
) *Server {
	return &Server{
		AuthorizationServer: legacygrpc.NewAuthorizationServer(legacy, checkAccess, bindings, roles),
	}
}
