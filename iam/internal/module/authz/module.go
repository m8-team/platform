package authz

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	graphv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/graph/v1"
	"github.com/m8platform/platform/iam/internal/core"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	"github.com/m8platform/platform/iam/internal/module/authz/api"
	deliverygrpc "github.com/m8platform/platform/iam/internal/module/authz/delivery/grpc"
	"github.com/m8platform/platform/iam/internal/module/authz/port"
	authzuc "github.com/m8platform/platform/iam/internal/module/authz/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Dependencies struct {
	Store         core.DocumentStore
	Cache         core.Cache
	Publisher     core.EventPublisher
	Runtime       core.AuthorizationRuntime
	Logger        *zap.Logger
	PolicyVersion string
	Topics        foundationconfig.TopicsConfig
	CheckAccess   *authzuc.CheckAccessUseCase
	Bindings      port.AccessBindingRepository
	Roles         port.RolePermissionResolver
}

type Module struct {
	server      *deliverygrpc.Server
	graphServer *deliverygrpc.GraphServer
	facade      api.Facade
}

func New(deps Dependencies) *Module {
	return &Module{
		server: deliverygrpc.NewServer(
			deps.Store,
			deps.Cache,
			deps.Publisher,
			deps.Runtime,
			deps.Logger,
			deps.PolicyVersion,
			deps.Topics,
			deps.CheckAccess,
			deps.Bindings,
			deps.Roles,
		),
		graphServer: deliverygrpc.NewGraphServer(deps.Store),
		facade:      api.New(deps.CheckAccess),
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
			graphv1.RegisterGraphServiceServer(s, m.graphServer)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			if err := authzv1.RegisterAuthorizationFacadeServiceHandler(ctx, mux, conn); err != nil {
				return err
			}
			return graphv1.RegisterGraphServiceHandler(ctx, mux, conn)
		},
	})
}

func (m *Module) RegisterConsumers(reg modulekit.ConsumerRegistrar) {}

func (m *Module) RegisterWorkers(reg modulekit.WorkerRegistrar) {}
