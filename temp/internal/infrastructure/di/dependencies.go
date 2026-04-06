package di

import (
	"context"

	redisadapter "github.com/m8platform/platform/iam/internal/adapter/out/redis"
	spicedbadapter "github.com/m8platform/platform/iam/internal/adapter/out/spicedb"
	temporaladapter "github.com/m8platform/platform/iam/internal/adapter/out/temporalclient"
	topicsadapter "github.com/m8platform/platform/iam/internal/adapter/out/topics"
	ydbadapter "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	foundationgrpc "github.com/m8platform/platform/iam/internal/foundation/grpcserver"
	foundationmetrics "github.com/m8platform/platform/iam/internal/foundation/metrics"
	"go.uber.org/zap"
)

type Dependencies struct {
	Logger    *zap.Logger
	Validator foundationgrpc.Validator
	Metrics   *foundationmetrics.Metrics
	Store     *ydbadapter.Client
	Cache     *redisadapter.Cache
	Publisher *topicsadapter.Publisher
	Workflows *temporaladapter.WorkflowStarter
	SpiceDB   *spicedbadapter.Client
	GRPC      *foundationgrpc.Server
}

func Build(ctx context.Context, cfg foundationconfig.Config) (*Dependencies, error) {
	container, err := NewContainer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Dependencies{
		Logger:    container.Logger,
		Validator: container.Validator,
		Metrics:   container.Metrics,
		Store:     container.Store,
		Cache:     container.Cache,
		Publisher: container.Publisher,
		Workflows: container.Workflows,
		SpiceDB:   container.SpiceDB,
		GRPC:      container.GRPC,
	}, nil
}
