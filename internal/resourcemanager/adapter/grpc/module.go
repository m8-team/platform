package grpcadapter

import (
	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// Register the new canonical v1 service, not the retired OrganizationService.
func Module() fx.Option {
	return fx.Module("resource-manager-v1-grpc",
		fx.Provide(v1engine.NewStore),
		fx.Provide(memory.NewOrgStructureStore),
		fx.Provide(NewOrgStructureServer),
		fx.Provide(fx.Annotate(newOrgStructureRegistration, fx.ResultTags(grpcserver.RegistrationResultTag))),
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

 // newOrgStructureRegistration registers the typed v1 organization graph APIs.
 // The production security gate is deny-all until M8 Access is connected.
 func newOrgStructureRegistration(s *OrgStructureServer) grpcserver.Registration {
   return func(registrar grpc.ServiceRegistrar) error {
     pb.RegisterOrgStructureServiceServer(registrar,s)
     pb.RegisterOrgUnitChangeSetServiceServer(registrar,s)
     return nil
   }
 }
