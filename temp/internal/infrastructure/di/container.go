package di

import (
	"context"

	"buf.build/go/protovalidate"
	keycloakadapter "github.com/m8platform/platform/iam/internal/adapter/out/keycloak"
	redisadapter "github.com/m8platform/platform/iam/internal/adapter/out/redis"
	spicedbadapter "github.com/m8platform/platform/iam/internal/adapter/out/spicedb"
	temporaladapter "github.com/m8platform/platform/iam/internal/adapter/out/temporalclient"
	topicsadapter "github.com/m8platform/platform/iam/internal/adapter/out/topics"
	ydbadapter "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	foundationgrpc "github.com/m8platform/platform/iam/internal/foundation/grpcserver"
	foundationlogging "github.com/m8platform/platform/iam/internal/foundation/logging"
	foundationmetrics "github.com/m8platform/platform/iam/internal/foundation/metrics"
	"github.com/m8platform/platform/iam/internal/foundation/modulekit"
	modulaudit "github.com/m8platform/platform/iam/internal/module/audit"
	modauthz "github.com/m8platform/platform/iam/internal/module/authz"
	authzport "github.com/m8platform/platform/iam/internal/module/authz/port"
	authzuc "github.com/m8platform/platform/iam/internal/module/authz/usecase"
	modiam "github.com/m8platform/platform/iam/internal/module/iam"
	identityuc "github.com/m8platform/platform/iam/internal/module/iam/usecase"
	modtenant "github.com/m8platform/platform/iam/internal/module/tenant"
	tenantuc "github.com/m8platform/platform/iam/internal/module/tenant/usecase"
	"github.com/m8platform/platform/iam/internal/shared/clock"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Container struct {
	Config    foundationconfig.Config
	Logger    *zap.Logger
	Validator foundationgrpc.Validator
	Metrics   *foundationmetrics.Metrics
	Store     *ydbadapter.Client
	Cache     *redisadapter.Cache
	Publisher *topicsadapter.Publisher
	Workflows *temporaladapter.WorkflowStarter
	SpiceDB   *spicedbadapter.Client
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

	store, err := ydbadapter.Open(ctx, cfg.YDB)
	if err != nil {
		return nil, err
	}
	cache := redisadapter.NewCache(cfg.Redis)
	publisher := topicsadapter.NewPublisher(logger)
	keycloakClient := keycloakadapter.NewClient(cfg.Keycloak)
	spicedbClient := spicedbadapter.NewClient(cfg.SpiceDB)
	workflowStarter, err := temporaladapter.NewWorkflowStarter(cfg.Temporal)
	if err != nil {
		return nil, err
	}

	systemClock := clock.System{}
	identityWorkflowStarter := temporaladapter.NewIdentityWorkflowStarter(workflowStarter)
	serviceAccountRepository := ydbadapter.NewServiceAccountRepository(store)
	serviceAccountEvents := topicsadapter.NewServiceAccountEventPublisher(publisher, cfg.Topics.ServiceAccounts)

	createServiceAccount := identityuc.NewCreateServiceAccountUseCase(
		systemClock,
		serviceAccountRepository,
		keycloakClient,
		identityWorkflowStarter,
		serviceAccountEvents,
	)
	rotateClientSecret := identityuc.NewRotateOAuthClientSecretUseCase(
		systemClock,
		keycloakClient,
		identityWorkflowStarter,
	)

	accessBindings := ydbadapter.NewAccessBindingRepository(store)
	roleResolver := spicedbadapter.RolePermissionResolver{}
	accessCache := redisadapter.NewAccessDecisionCache(cache, cfg.Redis.PolicyVersion)
	var runtimeChecker authzport.AuthorizationChecker
	if cfg.SpiceDB.Endpoint != "" {
		runtimeChecker = spicedbadapter.NewAuthorizationChecker(spicedbClient)
	}
	checkAccess := authzuc.NewCheckAccessUseCase(accessBindings, runtimeChecker, accessCache, roleResolver)
	supportGrantRepository := ydbadapter.NewSupportGrantRepository(store)
	supportGrantEvents := topicsadapter.NewSupportGrantEventPublisher(publisher, cfg.Topics.SupportGrants)
	supportGrantWorkflows := temporaladapter.NewSupportGrantWorkflowStarter(workflowStarter)
	supportAccess := tenantuc.NewSupportAccessUseCase(systemClock, supportGrantRepository, supportGrantEvents, supportGrantWorkflows)

	registry := modulekit.NewRegistry(
		modiam.New(modiam.Dependencies{
			Store:                   store,
			Publisher:               publisher,
			Workflows:               workflowStarter,
			Runtime:                 spicedbClient,
			Keycloak:                keycloakClient,
			Logger:                  logger,
			Topics:                  cfg.Topics,
			CreateServiceAccount:    createServiceAccount,
			RotateOAuthClientSecret: rotateClientSecret,
		}),
		modauthz.New(modauthz.Dependencies{
			Store:         store,
			Cache:         cache,
			Publisher:     publisher,
			Runtime:       spicedbClient,
			Logger:        logger,
			PolicyVersion: cfg.Redis.PolicyVersion,
			Topics:        cfg.Topics,
			CheckAccess:   checkAccess,
			Bindings:      accessBindings,
			Roles:         roleResolver,
		}),
		modtenant.New(modtenant.Dependencies{
			Logger:        logger,
			SupportAccess: supportAccess,
		}),
		modulaudit.New(store),
	)

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
