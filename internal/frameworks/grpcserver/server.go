package grpcserver

import (
	grpcadapter "github.com/m8platform/platform/internal/adapters/inbound/grpc/resourcemanager"
	"google.golang.org/grpc"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

func New(
	organizationServer grpcadapter.OrganizationServiceServer,
	workspaceServer grpcadapter.WorkspaceServiceServer,
	projectServer grpcadapter.ProjectServiceServer,
) *grpc.Server {
	server := grpc.NewServer()
	resourcemanagerv1.RegisterOrganizationServiceServer(server, organizationServer)
	resourcemanagerv1.RegisterWorkspaceServiceServer(server, workspaceServer)
	resourcemanagerv1.RegisterProjectServiceServer(server, projectServer)
	return server
}
