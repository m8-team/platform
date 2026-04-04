package app

import (
	"context"
	"errors"

	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/infrastructure/di"
	"github.com/m8platform/platform/iam/internal/observability"
	"github.com/m8platform/platform/iam/internal/spicedb"
	redisstore "github.com/m8platform/platform/iam/internal/storage/redis"
	ydbstore "github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/temporalx"
	"github.com/m8platform/platform/iam/internal/topics"
	grpcserver "github.com/m8platform/platform/iam/internal/transport/grpc"
)

type Application struct {
	Config    config.Config
	Logger    Logger
	Validator core.Validator
	Metrics   *observability.Metrics
	Store     *ydbstore.Client
	Cache     *redisstore.Cache
	Publisher *topics.Publisher
	Workflows *temporalx.WorkflowStarter
	SpiceDB   *spicedb.Client
	GRPC      *grpcserver.Server
}

type Logger interface {
	Sync() error
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	deps, err := di.Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Application{
		Config:    cfg,
		Logger:    deps.Logger,
		Validator: deps.Validator,
		Metrics:   deps.Metrics,
		Store:     deps.Store,
		Cache:     deps.Cache,
		Publisher: deps.Publisher,
		Workflows: deps.Workflows,
		SpiceDB:   deps.SpiceDB,
		GRPC:      deps.GRPC,
	}, nil
}

func (a *Application) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = a.GRPC.Shutdown(context.Background())
	}()
	err := a.GRPC.Serve()
	if err != nil && errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return err
}

func (a *Application) Close(ctx context.Context) error {
	if a == nil {
		return nil
	}
	if a.GRPC != nil {
		_ = a.GRPC.Shutdown(ctx)
	}
	if a.Workflows != nil {
		_ = a.Workflows.Close()
	}
	if a.SpiceDB != nil {
		_ = a.SpiceDB.Close()
	}
	if a.Store != nil {
		_ = a.Store.Close(ctx)
	}
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
	return nil
}
