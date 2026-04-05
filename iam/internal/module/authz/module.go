package authz

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	legacyauthz "github.com/m8platform/platform/iam/internal/authz"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	"github.com/m8platform/platform/iam/internal/module/authz/api"
	deliverygrpc "github.com/m8platform/platform/iam/internal/module/authz/delivery/grpc"
	authzuc "github.com/m8platform/platform/iam/internal/usecase/authz"
	"github.com/m8platform/platform/iam/internal/usecase/port"
	"google.golang.org/grpc"
)

type Dependencies struct {
	LegacyService *legacyauthz.Service
	CheckAccess   *authzuc.CheckAccessUseCase
	Bindings      port.AccessBindingRepository
	Roles         port.RolePermissionResolver
}

type Module struct {
	server *deliverygrpc.Server
	facade api.Facade
}

func New(deps Dependencies) *Module {
	return &Module{
		server: deliverygrpc.NewServer(deps.LegacyService, deps.CheckAccess, deps.Bindings, deps.Roles),
		facade: api.New(deps.CheckAccess),
	}
}

func (m *Module) Name() string {
	return "authz"
}

func (m *Module) Facade() api.Facade {
	if m == nil {
		return nil
	}
	return m.facade
}

func (m *Module) RegisterHTTP(reg modulekit.HTTPRegistrar) {}

func (m *Module) RegisterGRPC(reg modulekit.GRPCRegistrar) {
	if m == nil || m.server == nil {
		return
	}
	reg.RegisterGRPCService(modulekit.GRPCService{
		Name: "authz.authorization",
		Register: func(s grpc.ServiceRegistrar) {
			authzv1.RegisterAuthorizationFacadeServiceServer(s, m.server)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return authzv1.RegisterAuthorizationFacadeServiceHandler(ctx, mux, conn)
		},
	})
}

func (m *Module) RegisterConsumers(reg modulekit.ConsumerRegistrar) {}

func (m *Module) RegisterWorkers(reg modulekit.WorkerRegistrar) {}
