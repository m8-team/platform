package httpadapter

import (
	"context"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcadapter "github.com/m8platform/platform/internal/adapter/inbound/grpc/resourcemanager"
)

func NewGateway(
	ctx context.Context,
	organizationServer grpcadapter.OrganizationServiceServer,
	workspaceServer grpcadapter.WorkspaceServiceServer,
	projectServer grpcadapter.ProjectServiceServer,
) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()
	if err := resourcemanagerv1.RegisterOrganizationServiceHandlerServer(ctx, mux, organizationServer); err != nil {
		return nil, err
	}
	if err := resourcemanagerv1.RegisterWorkspaceServiceHandlerServer(ctx, mux, workspaceServer); err != nil {
		return nil, err
	}
	if err := resourcemanagerv1.RegisterProjectServiceHandlerServer(ctx, mux, projectServer); err != nil {
		return nil, err
	}
	return mux, nil
}
