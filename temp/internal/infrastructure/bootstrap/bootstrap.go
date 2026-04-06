package bootstrap

import (
	"context"
	"fmt"

	foundationapp "github.com/m8platform/platform/iam/internal/foundation/app"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/infrastructure/di"
)

func NewApplication(ctx context.Context, cfg foundationconfig.Config) (*foundationapp.Application, error) {
	container, err := di.NewContainer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("build container: %w", err)
	}

	return foundationapp.New(
		container.GRPC,
		func(ctx context.Context) error { return container.GRPC.Shutdown(ctx) },
		func(context.Context) error { return container.Workflows.Close() },
		func(context.Context) error { return container.SpiceDB.Close() },
		func(ctx context.Context) error { return container.Store.Close(ctx) },
		func(context.Context) error { return container.Logger.Sync() },
	), nil
}
