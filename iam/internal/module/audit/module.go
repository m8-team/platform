package audit

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	foundationstore "github.com/m8platform/platform/iam/internal/foundation/store"
	deliverygrpc "github.com/m8platform/platform/iam/internal/module/audit/delivery/grpc"
	"google.golang.org/grpc"
)

type Module struct {
	auditServer *deliverygrpc.AuditServer
	opsServer   *deliverygrpc.OperationsServer
}

func New(store foundationstore.DocumentStore) *Module {
	return &Module{
		auditServer: deliverygrpc.NewAuditServer(store),
		opsServer:   deliverygrpc.NewOperationsServer(store),
	}
}

func (m *Module) Name() string {
	return "audit"
}

func (m *Module) RegisterHTTP(reg modulekit.HTTPRegistrar) {}

func (m *Module) RegisterGRPC(reg modulekit.GRPCRegistrar) {
	if m == nil || m.auditServer == nil || m.opsServer == nil {
		return
	}
	reg.RegisterGRPCService(modulekit.GRPCService{
		Name: "audit",
		Register: func(s grpc.ServiceRegistrar) {
			auditv1.RegisterAuditServiceServer(s, m.auditServer)
			opsv1.RegisterOperationsServiceServer(s, m.opsServer)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			if err := auditv1.RegisterAuditServiceHandler(ctx, mux, conn); err != nil {
				return err
			}
			return opsv1.RegisterOperationsServiceHandler(ctx, mux, conn)
		},
	})
}

func (m *Module) RegisterConsumers(reg modulekit.ConsumerRegistrar) {}

func (m *Module) RegisterWorkers(reg modulekit.WorkerRegistrar) {}
