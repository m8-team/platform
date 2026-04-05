package di

import (
	"context"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	graphv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/graph/v1"
	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	keycloakadapter "github.com/m8platform/platform/iam/internal/adapter/out/keycloak"
	redisadapter "github.com/m8platform/platform/iam/internal/adapter/out/redis"
	spicedbadapter "github.com/m8platform/platform/iam/internal/adapter/out/spicedb"
	temporaladapter "github.com/m8platform/platform/iam/internal/adapter/out/temporalclient"
	topicsadapter "github.com/m8platform/platform/iam/internal/adapter/out/topics"
	ydbadapter "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	legacyaudit "github.com/m8platform/platform/iam/internal/audit"
	legacyauthz "github.com/m8platform/platform/iam/internal/authz"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	foundationgrpc "github.com/m8platform/platform/iam/internal/foundation/grpcserver"
	foundationlogging "github.com/m8platform/platform/iam/internal/foundation/logging"
	foundationmetrics "github.com/m8platform/platform/iam/internal/foundation/metrics"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	"github.com/m8platform/platform/iam/internal/graph"
	legacyidentity "github.com/m8platform/platform/iam/internal/identity"
	legacykeycloak "github.com/m8platform/platform/iam/internal/keycloak"
	modulaudit "github.com/m8platform/platform/iam/internal/module/audit"
	modauthz "github.com/m8platform/platform/iam/internal/module/authz"
	modiam "github.com/m8platform/platform/iam/internal/module/iam"
	modtenant "github.com/m8platform/platform/iam/internal/module/tenant"
	"github.com/m8platform/platform/iam/internal/ops"
	"github.com/m8platform/platform/iam/internal/shared/clock"
	legacyspicedb "github.com/m8platform/platform/iam/internal/spicedb"
	redisstore "github.com/m8platform/platform/iam/internal/storage/redis"
	ydbstore "github.com/m8platform/platform/iam/internal/storage/ydb"
	"github.com/m8platform/platform/iam/internal/support"
	"github.com/m8platform/platform/iam/internal/temporalx"
	legacytopics "github.com/m8platform/platform/iam/internal/topics"
	authzuc "github.com/m8platform/platform/iam/internal/usecase/authz"
	identityuc "github.com/m8platform/platform/iam/internal/usecase/identity"
	usecaseport "github.com/m8platform/platform/iam/internal/usecase/port"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Container struct {
	Config    foundationconfig.Config
	Logger    *zap.Logger
	Validator foundationgrpc.Validator
	Metrics   *foundationmetrics.Metrics
	Store     *ydbstore.Client
	Cache     *redisstore.Cache
	Publisher *legacytopics.Publisher
	Workflows *temporalx.WorkflowStarter
	SpiceDB   *legacyspicedb.Client
	Modules   *modulekit.Registry
	GRPC      *foundationgrpc.Server
}

type validatorAdapter struct {
	inner protovalidate.Validator
}

func (v validatorAdapter) Validate(message proto.Message) error {
	return v.inner.Validate(message)
}

func NewContainer(ctx context.Context, cfg foundationconfig.Config) (*Container, error) {
	logger, err := foundationlogging.New(cfg.Development)
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

	systemClock := clock.System{}
	keycloakAdapter := keycloakadapter.NewClient(keycloakClient)
	identityWorkflowStarter := temporaladapter.NewIdentityWorkflowStarter(workflowStarter)
	serviceAccountRepository := ydbadapter.NewServiceAccountRepository(store)
	serviceAccountEvents := topicsadapter.NewServiceAccountEventPublisher(publisher, cfg.Topics.ServiceAccounts)

	createServiceAccount := identityuc.NewCreateServiceAccountUseCase(
		systemClock,
		serviceAccountRepository,
		keycloakAdapter,
		identityWorkflowStarter,
		serviceAccountEvents,
	)
	rotateClientSecret := identityuc.NewRotateOAuthClientSecretUseCase(
		systemClock,
		keycloakAdapter,
		identityWorkflowStarter,
	)

	legacyIdentityService := legacyidentity.NewService(store, publisher, workflowStarter, spicedbClient, keycloakClient, logger, cfg)

	legacyAuthzService := legacyauthz.NewService(store, cache, publisher, spicedbClient, logger, cfg)
	accessBindings := ydbadapter.NewAccessBindingRepository(store)
	roleResolver := spicedbadapter.RolePermissionResolver{}
	accessCache := redisadapter.NewAccessDecisionCache(cache, cfg.Redis.PolicyVersion)
	var runtimeChecker usecaseport.AuthorizationChecker
	if cfg.SpiceDB.Endpoint != "" {
		runtimeChecker = spicedbadapter.NewAuthorizationChecker(spicedbClient)
	}
	checkAccess := authzuc.NewCheckAccessUseCase(accessBindings, runtimeChecker, accessCache, roleResolver)

	legacySupportService := support.NewService(store, publisher, workflowStarter, logger, cfg)
	supportGrantRepository := ydbadapter.NewSupportGrantRepository(store)
	supportGrantEvents := topicsadapter.NewSupportGrantEventPublisher(publisher, cfg.Topics.SupportGrants)
	supportGrantWorkflows := temporaladapter.NewSupportGrantWorkflowStarter(workflowStarter)
	supportAccess := tenantuc.NewSupportAccessUseCase(systemClock, supportGrantRepository, supportGrantEvents, supportGrantWorkflows)

	auditService := legacyaudit.NewService(store)

	registry := modulekit.NewRegistry(
		modiam.New(modiam.Dependencies{
			LegacyService:           legacyIdentityService,
			Logger:                  logger,
			CreateServiceAccount:    createServiceAccount,
			RotateOAuthClientSecret: rotateClientSecret,
		}),
		modauthz.New(modauthz.Dependencies{
			LegacyService: legacyAuthzService,
			CheckAccess:   checkAccess,
			Bindings:      accessBindings,
			Roles:         roleResolver,
		}),
		modtenant.New(modtenant.Dependencies{
			LegacyService: legacySupportService,
			Logger:        logger,
			SupportAccess: supportAccess,
		}),
		modulaudit.New(auditService),
	)

	graphService := graph.NewService(store)
	registry.GRPC().RegisterGRPCService(modulekit.GRPCService{
		Name: "graph",
		Register: func(s grpc.ServiceRegistrar) {
			graphv1.RegisterGraphServiceServer(s, graphService)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return graphv1.RegisterGraphServiceHandler(ctx, mux, conn)
		},
	})

	opsService := ops.NewService(store)
	registry.GRPC().RegisterGRPCService(modulekit.GRPCService{
		Name: "operations",
		Register: func(s grpc.ServiceRegistrar) {
			opsv1.RegisterOperationsServiceServer(s, opsService)
		},
		RegisterGateway: func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return opsv1.RegisterOperationsServiceHandler(ctx, mux, conn)
		},
	})

	grpcServer, err := foundationgrpc.New(cfg.GRPC, cfg.HTTP, logger, validation, registry.GRPC().Services())
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:    cfg,
		Logger:    logger,
		Validator: validation,
		Metrics:   foundationmetrics.New(prometheus.DefaultRegisterer),
		Store:     store,
		Cache:     cache,
		Publisher: publisher,
		Workflows: workflowStarter,
		SpiceDB:   spicedbClient,
		Modules:   registry,
		GRPC:      grpcServer,
	}, nil
}
