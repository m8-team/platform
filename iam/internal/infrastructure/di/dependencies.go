package di

import (
	"context"

	"buf.build/go/protovalidate"
	grpcadapter "github.com/m8platform/platform/iam/internal/adapter/in/grpc"
	keycloakadapter "github.com/m8platform/platform/iam/internal/adapter/out/keycloak"
	redisadapter "github.com/m8platform/platform/iam/internal/adapter/out/redis"
	spicedbadapter "github.com/m8platform/platform/iam/internal/adapter/out/spicedb"
	temporaladapter "github.com/m8platform/platform/iam/internal/adapter/out/temporalclient"
	topicsadapter "github.com/m8platform/platform/iam/internal/adapter/out/topics"
	ydbadapter "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	"github.com/m8platform/platform/iam/internal/audit"
	legacyauthz "github.com/m8platform/platform/iam/internal/authz"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/graph"
	legacyidentity "github.com/m8platform/platform/iam/internal/identity"
	infraClock "github.com/m8platform/platform/iam/internal/infrastructure/clock"
	legacykeycloak "github.com/m8platform/platform/iam/internal/keycloak"
	"github.com/m8platform/platform/iam/internal/observability"
	"github.com/m8platform/platform/iam/internal/ops"
	legacyspicedb "github.com/m8platform/platform/iam/internal/spicedb"
	redisstore "github.com/m8platform/platform/iam/internal/storage/redis"
	ydbstore "github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/support"
	"github.com/m8platform/platform/iam/internal/temporalx"
	legacytopics "github.com/m8platform/platform/iam/internal/topics"
	grpcserver "github.com/m8platform/platform/iam/internal/transport/grpc"
	authzuc "github.com/m8platform/platform/iam/internal/usecase/authz"
	identityuc "github.com/m8platform/platform/iam/internal/usecase/identity"
	usecaseport "github.com/m8platform/platform/iam/internal/usecase/port"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Dependencies struct {
	Logger    *zap.Logger
	Validator core.Validator
	Metrics   *observability.Metrics
	Store     *ydbstore.Client
	Cache     *redisstore.Cache
	Publisher *legacytopics.Publisher
	Workflows *temporalx.WorkflowStarter
	SpiceDB   *legacyspicedb.Client
	GRPC      *grpcserver.Server
}

type validatorAdapter struct {
	inner protovalidate.Validator
}

func (v validatorAdapter) Validate(message proto.Message) error {
	return v.inner.Validate(message)
}

func Build(ctx context.Context, cfg config.Config) (*Dependencies, error) {
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
	publisher := legacytopics.NewPublisher(logger)
	keycloakClient := legacykeycloak.NewClient(cfg.Keycloak)
	spicedbClient := legacyspicedb.NewClient(cfg.SpiceDB)
	workflowStarter, err := temporalx.NewWorkflowStarter(cfg.Temporal)
	if err != nil {
		return nil, err
	}

	clock := infraClock.SystemClock{}
	keycloakAdapter := keycloakadapter.NewClient(keycloakClient)
	identityWorkflowStarter := temporaladapter.NewIdentityWorkflowStarter(workflowStarter)
	serviceAccountRepository := ydbadapter.NewServiceAccountRepository(store)
	serviceAccountEvents := topicsadapter.NewServiceAccountEventPublisher(publisher, cfg.Topics.ServiceAccounts)

	createServiceAccount := identityuc.NewCreateServiceAccountUseCase(
		clock,
		serviceAccountRepository,
		keycloakAdapter,
		identityWorkflowStarter,
		serviceAccountEvents,
	)
	rotateClientSecret := identityuc.NewRotateOAuthClientSecretUseCase(
		clock,
		keycloakAdapter,
		identityWorkflowStarter,
	)

	legacyIdentityService := legacyidentity.NewService(store, publisher, workflowStarter, spicedbClient, keycloakClient, logger, cfg)
	identityServer := grpcadapter.NewIdentityServer(legacyIdentityService, logger, createServiceAccount, rotateClientSecret)

	legacyAuthzService := legacyauthz.NewService(store, cache, publisher, spicedbClient, logger, cfg)
	accessBindings := ydbadapter.NewAccessBindingRepository(store)
	roleResolver := spicedbadapter.RolePermissionResolver{}
	accessCache := redisadapter.NewAccessDecisionCache(cache, cfg.Redis.PolicyVersion)
	var runtimeChecker usecaseport.AuthorizationChecker
	if cfg.SpiceDB.Endpoint != "" {
		runtimeChecker = spicedbadapter.NewAuthorizationChecker(spicedbClient)
	}
	checkAccess := authzuc.NewCheckAccessUseCase(accessBindings, runtimeChecker, accessCache, roleResolver)
	authzServer := grpcadapter.NewAuthorizationServer(legacyAuthzService, checkAccess, accessBindings, roleResolver)

	graphService := graph.NewService(store)
	legacySupportService := support.NewService(store, publisher, workflowStarter, logger, cfg)
	supportGrantRepository := ydbadapter.NewSupportGrantRepository(store)
	supportGrantEvents := topicsadapter.NewSupportGrantEventPublisher(publisher, cfg.Topics.SupportGrants)
	supportGrantWorkflows := temporaladapter.NewSupportGrantWorkflowStarter(workflowStarter)
	supportAccess := tenantuc.NewSupportAccessUseCase(clock, supportGrantRepository, supportGrantEvents, supportGrantWorkflows)
	supportServer := grpcadapter.NewSupportServer(legacySupportService, logger, supportAccess)
	auditService := audit.NewService(store)
	opsService := ops.NewService(store)

	grpcSrv, err := grpcserver.New(cfg.GRPC, cfg.HTTP, logger, validation, grpcserver.Services{
		Identity: identityServer,
		OAuth:    identityServer,
		Authz:    authzServer,
		Graph:    graphService,
		Support:  supportServer,
		Audit:    auditService,
		Ops:      opsService,
	})
	if err != nil {
		return nil, err
	}

	return &Dependencies{
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
