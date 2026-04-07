package httptransport

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/m8platform/platform/internal/transport/grpc"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

type Gateway struct {
	Servers grpctransport.ServerSet
}

func NewGateway(servers grpctransport.ServerSet) Gateway {
	return Gateway{Servers: servers}
}

func (g Gateway) Handler(ctx context.Context) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()
	if g.Servers.Organization != nil {
		if err := resourcemanagerv1.RegisterOrganizationServiceHandlerServer(ctx, mux, g.Servers.Organization); err != nil {
			return nil, fmt.Errorf("register organization gateway: %w", err)
		}
	}
	if g.Servers.Workspace != nil {
		if err := resourcemanagerv1.RegisterWorkspaceServiceHandlerServer(ctx, mux, g.Servers.Workspace); err != nil {
			return nil, fmt.Errorf("register workspace gateway: %w", err)
		}
	}
	if g.Servers.Project != nil {
		if err := resourcemanagerv1.RegisterProjectServiceHandlerServer(ctx, mux, g.Servers.Project); err != nil {
			return nil, fmt.Errorf("register project gateway: %w", err)
		}
	}
	return mux, nil
}
