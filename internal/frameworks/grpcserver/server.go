package grpcserver

import (
	grpcadapter "github.com/m8platform/platform/internal/adapter/inbound/grpc/resourcemanager"
	"google.golang.org/grpc"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

func New(
	organizationServer grpcadapter.OrganizationServiceServer,
) *grpc.Server {
	server := grpc.NewServer()
	resourcemanagerv1.RegisterOrganizationServiceServer(server, organizationServer)
	return server
}
