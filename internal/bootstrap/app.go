package bootstrap

import (
	"net/http"

	grpcadapter "github.com/m8platform/platform/internal/adapters/inbound/grpc/resourcemanager"
	"github.com/m8platform/platform/internal/adapters/outbound/outbox"
	"github.com/m8platform/platform/internal/frameworks/config"
	"github.com/m8platform/platform/internal/frameworks/database"
	"google.golang.org/grpc"
)

type App struct {
	Config             config.Config
	Database           *database.Postgres
	GRPCServer         *grpc.Server
	HTTPServer         *http.Server
	OutboxDispatcher   outbox.Dispatcher
	OrganizationServer grpcadapter.OrganizationServiceServer
	WorkspaceServer    grpcadapter.WorkspaceServiceServer
	ProjectServer      grpcadapter.ProjectServiceServer
}
