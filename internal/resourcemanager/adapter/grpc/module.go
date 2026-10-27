package grpcadapter

import (
	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// Register the new canonical v1 service, not the retired OrganizationService.
func Module() fx.Option {
	return fx.Module("resource-manager-v1-grpc",
		fx.Provide(v1engine.NewStore),
		fx.Provide(NewResourceServer),
		fx.Provide(asV1ServiceServer),
		fx.Provide(fx.Annotate(newRegistration, fx.ResultTags(grpcserver.RegistrationResultTag))),
	)
}
func asV1ServiceServer(server *ResourceServer) pb.ResourceManagerServiceServer { return server }
func newRegistration(server *ResourceServer) grpcserver.Registration {
	return func(registrar grpc.ServiceRegistrar) error {
		pb.RegisterResourceManagerServiceServer(registrar, server)
		return nil
	}
}
