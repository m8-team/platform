package iam

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	foundationcontracts "github.com/m8platform/platform/iam/internal/foundation/contracts"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	foundationstore "github.com/m8platform/platform/iam/internal/foundation/store"
	"github.com/m8platform/platform/iam/internal/module/iam/api"
	deliverygrpc "github.com/m8platform/platform/iam/internal/module/iam/delivery/grpc"
	identityuc "github.com/m8platform/platform/iam/internal/module/iam/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Dependencies struct {
	Store                   foundationstore.DocumentStore
	Publisher               foundationcontracts.EventPublisher
	Workflows               foundationcontracts.WorkflowStarter
	Runtime                 foundationcontracts.AuthorizationRuntime
	Keycloak                foundationcontracts.KeycloakClient
	Logger                  *zap.Logger
	Topics                  foundationconfig.TopicsConfig
	CreateServiceAccount    *identityuc.CreateServiceAccountUseCase
	RotateOAuthClientSecret *identityuc.RotateOAuthClientSecretUseCase
}

type Module struct {
	server *deliverygrpc.Server
	facade api.Facade
}

func New(deps Dependencies) *Module {
	return &Module{
		server: deliverygrpc.NewServer(
			deps.Store,
			deps.Publisher,
			deps.Workflows,
			deps.Runtime,
			deps.Keycloak,
			deps.Logger,
			deps.Topics,
			deps.CreateServiceAccount,
			deps.RotateOAuthClientSecret,
		),
		facade: api.New(deps.CreateServiceAccount, deps.RotateOAuthClientSecret),
	}
}

func (m *Module) Name() string {
	return "iam"
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
		Name: "iam.identity",
		Register: func(s grpc.ServiceRegistrar) {
			identityv1.RegisterIdentityServiceServer(s, m.server)
			identityv1.RegisterOAuthFacadeServiceServer(s, m.server)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			if err := identityv1.RegisterIdentityServiceHandler(ctx, mux, conn); err != nil {
				return err
			}
			return identityv1.RegisterOAuthFacadeServiceHandler(ctx, mux, conn)
		},
	})
}

func (m *Module) RegisterConsumers(reg modulekit.ConsumerRegistrar) {}

func (m *Module) RegisterWorkers(reg modulekit.WorkerRegistrar) {}
