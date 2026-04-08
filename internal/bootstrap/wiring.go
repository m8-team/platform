package bootstrap

import (
	"context"

	grpcadapter "github.com/m8platform/platform/internal/adapter/inbound/grpc/resourcemanager"
	httpadapter "github.com/m8platform/platform/internal/adapter/inbound/http"
	eventadapter "github.com/m8platform/platform/internal/adapter/outbound/events"
	"github.com/m8platform/platform/internal/adapter/outbound/filtering"
	"github.com/m8platform/platform/internal/adapter/outbound/idempotency"
	"github.com/m8platform/platform/internal/adapter/outbound/ordering"
	"github.com/m8platform/platform/internal/adapter/outbound/outbox"
	"github.com/m8platform/platform/internal/adapter/outbound/postgres/resourcemanager"
	grpcpresenter "github.com/m8platform/platform/internal/adapter/presenters/grpc/resourcemanager"
	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/frameworks/broker"
	"github.com/m8platform/platform/internal/frameworks/config"
	"github.com/m8platform/platform/internal/frameworks/database"
	"github.com/m8platform/platform/internal/frameworks/grpcserver"
	"github.com/m8platform/platform/internal/frameworks/httpserver"
	"github.com/m8platform/platform/internal/frameworks/telemetry"
	"github.com/m8platform/platform/internal/platform"
	projectcommand "github.com/m8platform/platform/internal/usecase/resourcemanager/command/project"
	workspacecommand "github.com/m8platform/platform/internal/usecase/resourcemanager/command/workspace"
	organizationcmd "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/command"
	organizationqry "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/query"
	projectquery "github.com/m8platform/platform/internal/usecase/resourcemanager/query/project"
	workspacequery "github.com/m8platform/platform/internal/usecase/resourcemanager/query/workspace"
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
			Executor: organizationcmd.CommandExecutor{
				TxManager:        txManager,
				IdempotencyStore: idempotencyStore,
			},
			Writer:        orgRepository,
			OutboxWriter:  outboxWriter,
			Clock:         clock,
			UUIDGenerator: uuidGenerator,
		},
		UpdateHandler: organizationcmd.UpdateInteractor{
			Executor: organizationcmd.CommandExecutor{
				TxManager:        txManager,
				IdempotencyStore: idempotencyStore,
			},
			Reader:         orgRepository,
			Writer:         orgRepository,
			OutboxWriter:   outboxWriter,
			Clock:          clock,
			UUIDGenerator:  uuidGenerator,
			InputValidator: organizationcmd.UpdateMaskValidator{},
		},
		DeleteHandler: organizationcmd.DeleteInteractor{
			Executor: organizationcmd.CommandExecutor{
				TxManager:        txManager,
				IdempotencyStore: idempotencyStore,
			},
			Reader:          orgRepository,
			Writer:          orgRepository,
			HierarchyReader: hierarchyReader,
			DeletePolicy:    deletePolicy,
			OutboxWriter:    outboxWriter,
			Clock:           clock,
			UUIDGenerator:   uuidGenerator,
		},
		UndeleteHandler: organizationcmd.UndeleteInteractor{
			Executor: organizationcmd.CommandExecutor{
				TxManager:        txManager,
				IdempotencyStore: idempotencyStore,
			},
			Reader:        orgRepository,
			Writer:        orgRepository,
			OutboxWriter:  outboxWriter,
			Clock:         clock,
			UUIDGenerator: uuidGenerator,
		},
		MetadataValidator: organizationcmd.RequiredMetadataValidator{},
	}
	orgQueries := organizationqry.QueryService{
		GetHandler:  organizationqry.GetInteractor{Reader: orgRepository},
		ListHandler: organizationqry.ListInteractor{Reader: orgRepository},
		ListValidator: organizationqry.QueryValidator{
			FilterValidator: filterValidator,
			OrderValidator:  orderValidator,
		},
	}

	workspaceCommands := workspacecommand.CommandService{
		CreateHandler: workspacecommand.CreateInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			HierarchyReader:  hierarchyReader,
			HierarchyPolicy:  hierarchyPolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UpdateHandler: workspacecommand.UpdateInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		DeleteHandler: workspacecommand.DeleteInteractor{
			TxManager:        txManager,
			Repository:       workspaceRepository,
			HierarchyReader:  hierarchyReader,
			DeletePolicy:     deletePolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UndeleteHandler: workspacecommand.UndeleteInteractor{
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
	workspaceQueries := workspacequery.QueryService{
		GetHandler: workspacequery.GetInteractor{Repository: workspaceRepository},
		ListHandler: workspacequery.ListInteractor{
			Repository:      workspaceRepository,
			FilterValidator: filterValidator,
			OrderValidator:  orderValidator,
		},
	}

	projectCommands := projectcommand.CommandService{
		CreateHandler: projectcommand.CreateInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			HierarchyReader:  hierarchyReader,
			HierarchyPolicy:  hierarchyPolicy,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UpdateHandler: projectcommand.UpdateInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		DeleteHandler: projectcommand.DeleteInteractor{
			TxManager:        txManager,
			Repository:       projectRepository,
			IdempotencyStore: idempotencyStore,
			OutboxWriter:     outboxWriter,
			Clock:            clock,
			UUIDGenerator:    uuidGenerator,
		},
		UndeleteHandler: projectcommand.UndeleteInteractor{
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
	projectQueries := projectquery.QueryService{
		GetHandler: projectquery.GetInteractor{Repository: projectRepository},
		ListHandler: projectquery.ListInteractor{
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
