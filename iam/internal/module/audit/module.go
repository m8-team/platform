package audit

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	legacyaudit "github.com/m8platform/platform/iam/internal/audit"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	deliverygrpc "github.com/m8platform/platform/iam/internal/module/audit/delivery/grpc"
	"google.golang.org/grpc"
)

type Module struct {
	server *deliverygrpc.Server
}

func New(service *legacyaudit.Service) *Module {
	return &Module{server: deliverygrpc.NewServer(service)}
}

func (m *Module) Name() string {
	return "audit"
}

func (m *Module) RegisterHTTP(reg modulekit.HTTPRegistrar) {}

func (m *Module) RegisterGRPC(reg modulekit.GRPCRegistrar) {
	if m == nil || m.server == nil {
		return
	}
	reg.RegisterGRPCService(modulekit.GRPCService{
		Name: "audit",
		Register: func(s grpc.ServiceRegistrar) {
			auditv1.RegisterAuditServiceServer(s, m.server)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return auditv1.RegisterAuditServiceHandler(ctx, mux, conn)
		},
	})
}

func (m *Module) RegisterConsumers(reg modulekit.ConsumerRegistrar) {}

func (m *Module) RegisterWorkers(reg modulekit.WorkerRegistrar) {}
