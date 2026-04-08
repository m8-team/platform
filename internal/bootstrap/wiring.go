package bootstrap

import (
	"context"

	grpcadapter "github.com/m8platform/platform/internal/adapters/inbound/grpc/resourcemanager"
	httpadapter "github.com/m8platform/platform/internal/adapters/inbound/http"
	eventadapter "github.com/m8platform/platform/internal/adapters/outbound/events"
	"github.com/m8platform/platform/internal/adapters/outbound/filtering"
	"github.com/m8platform/platform/internal/adapters/outbound/idempotency"
	"github.com/m8platform/platform/internal/adapters/outbound/ordering"
	"github.com/m8platform/platform/internal/adapters/outbound/outbox"
	"github.com/m8platform/platform/internal/adapters/outbound/postgres/resourcemanager"
	grpcpresenter "github.com/m8platform/platform/internal/adapters/presenters/grpc/resourcemanager"
	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/frameworks/broker"
	"github.com/m8platform/platform/internal/frameworks/config"
	"github.com/m8platform/platform/internal/frameworks/database"
	"github.com/m8platform/platform/internal/frameworks/grpcserver"
	"github.com/m8platform/platform/internal/frameworks/httpserver"
	"github.com/m8platform/platform/internal/frameworks/telemetry"
	"github.com/m8platform/platform/internal/platform"
	organizationcmd "github.com/m8platform/platform/internal/usecase/resourcemanager/commands/organization"
	projectcmd "github.com/m8platform/platform/internal/usecase/resourcemanager/commands/project"
	workspacecmd "github.com/m8platform/platform/internal/usecase/resourcemanager/commands/workspace"
	organizationqry "github.com/m8platform/platform/internal/usecase/resourcemanager/queries/organization"
	projectqry "github.com/m8platform/platform/internal/usecase/resourcemanager/queries/project"
	workspaceqry "github.com/m8platform/platform/internal/usecase/resourcemanager/queries/workspace"
)

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	_ = telemetry.NewLogger()
	_ = telemetry.NewMetrics()
	_ = telemetry.NewTracer()

	db := database.NewPostgres(database.Config{DSN: cfg.PostgresDSN})
	store := postgres.NewStore()
	txManager := postgres.TxManager{}
	clock := platform.SystemClock{}
	uuidGenerator := platform.UUIDGenerator{}
	filterValidator := filtering.AIP160Validator{}
	orderValidator := ordering.AIP132Validator{}
	idempotencyStore := idempotency.NewStore(clock)
	outboxWriter := outbox.NewWriter()
	publisher := &eventadapter.Publisher{Client: broker.NopClient{}}
	dispatcher := outbox.Dispatcher{Writer: outboxWriter, Publisher: publisher}

	orgRepository := postgres.OrganizationRepository{Store: store}
	workspaceRepository := postgres.WorkspaceRepository{Store: store}
	projectRepository := postgres.ProjectRepository{Store: store}
	hierarchyReader := postgres.HierarchyReader{Store: store}

	deletePolicy := domainservices.DeletePolicy{}
	hierarchyPolicy := domainservices.HierarchyPolicy{}
	undeletePolicy := domainservices.UndeletePolicy{}

	orgCommands := organizationcmd.CommandService{
		CreateHandler: organizationcmd.CreateInteractor{
			TxManager:        txManager,
			Repository:       orgRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UpdateHandler: organizationcmd.UpdateInteractor{
			TxManager:        txManager,
			Repository:       orgRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		DeleteHandler: organizationcmd.DeleteInteractor{
			TxManager:        txManager,
			Repository:       orgRepository,
			HierarchyReader:  hierarchyReader,
			DeletePolicy:     deletePolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UndeleteHandler: organizationcmd.UndeleteInteractor{
			TxManager:        txManager,
			Repository:       orgRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
	}
	orgQueries := organizationqry.QueryService{
		GetHandler: organizationqry.GetInteractor{Repository: orgRepository},
		ListHandler: organizationqry.ListInteractor{
			Repository:      orgRepository,
			FilterValidator: filterValidator,
			OrderValidator:  orderValidator,
		},
	}

	workspaceCommands := workspacecmd.CommandService{
		CreateHandler: workspacecmd.CreateInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			HierarchyReader:  hierarchyReader,
			HierarchyPolicy:  hierarchyPolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UpdateHandler: workspacecmd.UpdateInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		DeleteHandler: workspacecmd.DeleteInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			HierarchyReader:  hierarchyReader,
			DeletePolicy:     deletePolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UndeleteHandler: workspacecmd.UndeleteInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			HierarchyReader:  hierarchyReader,
			UndeletePolicy:   undeletePolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
	}
	workspaceQueries := workspaceqry.QueryService{
		GetHandler: workspaceqry.GetInteractor{Repository: workspaceRepository},
		ListHandler: workspaceqry.ListInteractor{
			Repository:      workspaceRepository,
			FilterValidator: filterValidator,
			OrderValidator:  orderValidator,
		},
	}

	projectCommands := projectcmd.CommandService{
		CreateHandler: projectcmd.CreateInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			HierarchyReader:  hierarchyReader,
			HierarchyPolicy:  hierarchyPolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UpdateHandler: projectcmd.UpdateInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		DeleteHandler: projectcmd.DeleteInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UndeleteHandler: projectcmd.UndeleteInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			HierarchyReader:  hierarchyReader,
			UndeletePolicy:   undeletePolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
	}
	projectQueries := projectqry.QueryService{
		GetHandler: projectqry.GetInteractor{Repository: projectRepository},
		ListHandler: projectqry.ListInteractor{
			Repository:      projectRepository,
			FilterValidator: filterValidator,
			OrderValidator:  orderValidator,
		},
	}

	organizationServer := grpcadapter.OrganizationServiceServer{
		Commands:  orgCommands,
		Queries:   orgQueries,
		Presenter: grpcpresenter.OrganizationPresenter{},
	}
	workspaceServer := grpcadapter.WorkspaceServiceServer{
		Commands:  workspaceCommands,
		Queries:   workspaceQueries,
		Presenter: grpcpresenter.WorkspacePresenter{},
	}
	projectServer := grpcadapter.ProjectServiceServer{
		Commands:  projectCommands,
		Queries:   projectQueries,
		Presenter: grpcpresenter.ProjectPresenter{},
	}

	grpcSrv := grpcserver.New(organizationServer, workspaceServer, projectServer)
	gatewayMux, err := httpadapter.NewGateway(ctx, organizationServer, workspaceServer, projectServer)
	if err != nil {
		return nil, err
	}
	httpSrv := httpserver.New(cfg.HTTPAddress, gatewayMux)

	return &App{
		Config:             cfg,
		Database:           db,
		GRPCServer:         grpcSrv,
		HTTPServer:         httpSrv,
		OutboxDispatcher:   dispatcher,
		OrganizationServer: organizationServer,
		WorkspaceServer:    workspaceServer,
		ProjectServer:      projectServer,
	}, nil
}
