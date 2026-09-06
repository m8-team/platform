package grpcserver

import (
	"context"

	"github.com/m8-team/platform/internal/platform/bootstrap"

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

type lifecycleParams struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Server     *Server
	Supervisor *bootstrap.Supervisor `optional:"true"`
}

func registerLifecycle(params lifecycleParams) {
	lifecycle, server := params.Lifecycle, params.Server
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := server.Start(ctx); err != nil {
				return err
			}
			if params.Supervisor != nil {
				params.Supervisor.Go("grpc", func(context.Context) error { return server.Wait() })
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Stop(ctx)
		},
	})
}
