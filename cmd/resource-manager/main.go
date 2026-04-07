package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/m8platform/platform/internal/application/command"
	"github.com/m8platform/platform/internal/application/query"
	clockinfra "github.com/m8platform/platform/internal/infra/clock"
	filterinfra "github.com/m8platform/platform/internal/infra/filters"
	idempotencyinfra "github.com/m8platform/platform/internal/infra/idempotency"
	orderinfra "github.com/m8platform/platform/internal/infra/ordering"
	outboxinfra "github.com/m8platform/platform/internal/infra/outbox"
	postgresinfra "github.com/m8platform/platform/internal/infra/postgres"
	uuidinfra "github.com/m8platform/platform/internal/infra/uuid"
	grpctransport "github.com/m8platform/platform/internal/transport/grpc"
	httptransport "github.com/m8platform/platform/internal/transport/http"
	"github.com/m8platform/platform/internal/transport/middleware"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Printf("resource-manager: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	clock := clockinfra.SystemClock{}
	uuids := uuidinfra.Generator{}
	idempotency := idempotencyinfra.NewStore()
	outbox := outboxinfra.NewStore()

	orgRepo := postgresinfra.NewOrganizationRepository(nil)
	workspaceRepo := postgresinfra.NewWorkspaceRepository(nil)
	projectRepo := postgresinfra.NewProjectRepository(nil)
	hierarchyRepo := postgresinfra.NewHierarchyRepository(nil)
	txManager := postgresinfra.NewTxManager(nil)

	filterParser := filterinfra.Parser{}
	orderParser := orderinfra.Parser{}
	actorResolver := middleware.NewMetadataActorResolver()

	servers := grpctransport.ServerSet{
		Organization: &grpctransport.OrganizationServer{
			Get: query.GetOrganizationHandler{Repository: orgRepo},
			List: query.ListOrganizationsHandler{
				Repository:   orgRepo,
				FilterParser: filterParser,
				OrderParser:  orderParser,
			},
			Create: command.CreateOrganizationHandler{
				TxManager:   txManager,
				Repository:  orgRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Update: command.UpdateOrganizationHandler{
				TxManager:   txManager,
				Repository:  orgRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Delete: command.DeleteOrganizationHandler{
				TxManager:   txManager,
				Repository:  orgRepo,
				Hierarchy:   hierarchyRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Undelete: command.UndeleteOrganizationHandler{
				TxManager:   txManager,
				Repository:  orgRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			ActorResolver: actorResolver,
		},
		Workspace: &grpctransport.WorkspaceServer{
			Get: query.GetWorkspaceHandler{Repository: workspaceRepo},
			List: query.ListWorkspacesHandler{
				Repository:   workspaceRepo,
				FilterParser: filterParser,
				OrderParser:  orderParser,
			},
			Create: command.CreateWorkspaceHandler{
				TxManager:   txManager,
				Repository:  workspaceRepo,
				Hierarchy:   hierarchyRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Update: command.UpdateWorkspaceHandler{
				TxManager:   txManager,
				Repository:  workspaceRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Delete: command.DeleteWorkspaceHandler{
				TxManager:   txManager,
				Repository:  workspaceRepo,
				Hierarchy:   hierarchyRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Undelete: command.UndeleteWorkspaceHandler{
				TxManager:   txManager,
				Repository:  workspaceRepo,
				Hierarchy:   hierarchyRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			ActorResolver: actorResolver,
		},
		Project: &grpctransport.ProjectServer{
			Get: query.GetProjectHandler{Repository: projectRepo},
			List: query.ListProjectsHandler{
				Repository:   projectRepo,
				FilterParser: filterParser,
				OrderParser:  orderParser,
			},
			Create: command.CreateProjectHandler{
				TxManager:   txManager,
				Repository:  projectRepo,
				Hierarchy:   hierarchyRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Update: command.UpdateProjectHandler{
				TxManager:   txManager,
				Repository:  projectRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Delete: command.DeleteProjectHandler{
				TxManager:   txManager,
				Repository:  projectRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			Undelete: command.UndeleteProjectHandler{
				TxManager:   txManager,
				Repository:  projectRepo,
				Hierarchy:   hierarchyRepo,
				Idempotency: idempotency,
				Outbox:      outbox,
				Clock:       clock,
				UUIDs:       uuids,
			},
			ActorResolver: actorResolver,
		},
	}

	grpcListener, err := net.Listen("tcp", envOrDefault("RESOURCE_MANAGER_GRPC_ADDR", ":8080"))
	if err != nil {
		return err
	}
	defer grpcListener.Close()

	grpcServer := grpc.NewServer()
	if err := servers.Register(grpcServer); err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go func() {
		errCh <- grpcServer.Serve(grpcListener)
	}()

	httpAddr := envOrDefault("RESOURCE_MANAGER_HTTP_ADDR", "")
	var httpServer *http.Server
	if httpAddr != "" {
		gateway, err := httptransport.NewGateway(servers).Handler(ctx)
		if err != nil {
			return err
		}
		httpServer = &http.Server{Addr: httpAddr, Handler: gateway}
		go func() {
			errCh <- httpServer.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		if httpServer != nil {
			_ = httpServer.Shutdown(context.Background())
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
