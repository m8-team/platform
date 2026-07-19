package grpcadapter

import (
	"github.com/google/uuid"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func Module() fx.Option {
	return fx.Module(
		"resource-manager-organization-grpc",
		fx.Provide(newOperationIDGenerator),
		fx.Provide(NewOrganizationServer),
		fx.Provide(fx.Annotate(
			newRegistration,
			fx.ResultTags(grpcserver.RegistrationResultTag),
		)),
	)
}

type uuidOperationIDGenerator struct{}

func newOperationIDGenerator() OperationIDGenerator {
	return uuidOperationIDGenerator{}
}

func (uuidOperationIDGenerator) NewOperationID() string {
	return uuid.NewString()
}

func newRegistration(server *OrganizationServer) grpcserver.Registration {
	return func(registrar grpc.ServiceRegistrar) error {
		resourcemanagerpb.RegisterOrganizationServiceServer(registrar, server)
		return nil
	}
}
