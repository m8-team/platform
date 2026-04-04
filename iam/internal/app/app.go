package app

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"github.com/m8platform/platform/iam/internal/audit"
	"github.com/m8platform/platform/iam/internal/authz"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/graph"
	"github.com/m8platform/platform/iam/internal/identity"
	"github.com/m8platform/platform/iam/internal/keycloak"
	"github.com/m8platform/platform/iam/internal/observability"
	"github.com/m8platform/platform/iam/internal/ops"
	"github.com/m8platform/platform/iam/internal/spicedb"
	redisstore "github.com/m8platform/platform/iam/internal/storage/redis"
	ydbstore "github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/support"
	"github.com/m8platform/platform/iam/internal/temporalx"
	"github.com/m8platform/platform/iam/internal/topics"
	grpcserver "github.com/m8platform/platform/iam/internal/transport/grpc"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
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

type validatorAdapter struct {
	inner protovalidate.Validator
}

func (v validatorAdapter) Validate(message proto.Message) error {
	return v.inner.Validate(message)
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	logger, err := observability.NewLogger(cfg.Development)
	if err != nil {
		return nil, err
	}

	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	validation := validatorAdapter{inner: validator}

	store, err := ydbstore.Open(ctx, cfg.YDB)
	if err != nil {
		return nil, err
	}
	cache := redisstore.NewCache(cfg.Redis)
	publisher := topics.NewPublisher(logger)
	keycloakClient := keycloak.NewClient(cfg.Keycloak)
	spicedbClient := spicedb.NewClient(cfg.SpiceDB)
	workflowStarter, err := temporalx.NewWorkflowStarter(cfg.Temporal)
	if err != nil {
		return nil, err
	}

	identityService := identity.NewService(store, publisher, workflowStarter, spicedbClient, keycloakClient, logger, cfg)
	authzService := authz.NewService(store, cache, publisher, spicedbClient, logger, cfg)
	graphService := graph.NewService(store)
	supportService := support.NewService(store, publisher, workflowStarter, logger, cfg)
	auditService := audit.NewService(store)
	opsService := ops.NewService(store)

	grpcSrv, err := grpcserver.New(cfg.GRPC, cfg.HTTP, logger, validation, grpcserver.Services{
		Identity: identityService,
		OAuth:    identityService,
		Authz:    authzService,
		Graph:    graphService,
		Support:  supportService,
		Audit:    auditService,
		Ops:      opsService,
	})
	if err != nil {
		return nil, err
	}

	return &Application{
		Config:    cfg,
		Logger:    logger,
		Validator: validation,
		Metrics:   observability.NewMetrics(prometheus.DefaultRegisterer),
		Store:     store,
		Cache:     cache,
		Publisher: publisher,
		Workflows: workflowStarter,
		SpiceDB:   spicedbClient,
		GRPC:      grpcSrv,
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
