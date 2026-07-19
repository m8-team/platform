package grpcserver

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// Module creates one process-level gRPC server. Service modules contribute
// Registration callbacks through RegistrationGroup without making the
// platform foundation depend on concrete business modules.
func Module(config Config, options ...grpc.ServerOption) fx.Option {
	config = config.normalized()

	return fx.Module(
		"platform-grpc-server",
		fx.Supply(config),
		fx.Provide(func(params registrationParams) (*grpc.Server, error) {
			return newGRPCServer(params, options...)
		}),
		fx.Provide(NewServer),
		fx.Invoke(registerLifecycle),
	)
}

func registerLifecycle(lifecycle fx.Lifecycle, server *Server) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return server.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return server.Stop(ctx)
		},
	})
}
