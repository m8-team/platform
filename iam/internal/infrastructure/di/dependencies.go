package di

import (
	"context"

	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	foundationgrpc "github.com/m8platform/platform/iam/internal/foundation/grpcserver"
	foundationmetrics "github.com/m8platform/platform/iam/internal/foundation/metrics"
	legacyspicedb "github.com/m8platform/platform/iam/internal/spicedb"
	redisstore "github.com/m8platform/platform/iam/internal/storage/redis"
	ydbstore "github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/temporalx"
	legacytopics "github.com/m8platform/platform/iam/internal/topics"
	"go.uber.org/zap"
)

type Dependencies struct {
	Logger    *zap.Logger
	Validator foundationgrpc.Validator
	Metrics   *foundationmetrics.Metrics
	Store     *ydbstore.Client
	Cache     *redisstore.Cache
	Publisher *legacytopics.Publisher
	Workflows *temporalx.WorkflowStarter
	SpiceDB   *legacyspicedb.Client
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
