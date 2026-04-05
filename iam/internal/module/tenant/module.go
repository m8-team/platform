package tenant

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	"github.com/m8platform/platform/iam/internal/module/tenant/api"
	deliverygrpc "github.com/m8platform/platform/iam/internal/module/tenant/delivery/grpc"
	deliverytopic "github.com/m8platform/platform/iam/internal/module/tenant/delivery/topic"
	tenantuc "github.com/m8platform/platform/iam/internal/module/tenant/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Dependencies struct {
	Logger        *zap.Logger
	SupportAccess *tenantuc.SupportAccessUseCase
}

type Module struct {
	server   *deliverygrpc.Server
	facade   api.Facade
	consumer *deliverytopic.SupportAccessConsumer
}

func New(deps Dependencies) *Module {
	return &Module{
		server:   deliverygrpc.NewServer(deps.Logger, deps.SupportAccess),
		facade:   api.New(deps.SupportAccess),
		consumer: deliverytopic.NewSupportAccessConsumer(deps.SupportAccess),
	}
}

func (m *Module) Name() string {
	return "tenant"
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
		Name: "tenant.support",
		Register: func(s grpc.ServiceRegistrar) {
			supportv1.RegisterSupportAccessServiceServer(s, m.server)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return supportv1.RegisterSupportAccessServiceHandler(ctx, mux, conn)
		},
	})
}

func (m *Module) RegisterConsumers(reg modulekit.ConsumerRegistrar) {
	if m == nil || m.consumer == nil {
		return
	}
	reg.RegisterConsumer(modulekit.ConsumerRegistration{
		Name:    "tenant.support.grant_temporary_access",
		Handler: m.consumer.HandleGrantTemporaryAccess,
	})
}

func (m *Module) RegisterWorkers(reg modulekit.WorkerRegistrar) {}
