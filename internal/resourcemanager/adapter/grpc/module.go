package grpcadapter

import (
	"context"
	"database/sql"
	"fmt"
	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	accessusecase "github.com/m8-team/platform/internal/access/app/usecase"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/ydb"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/orgaccess"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	_ "github.com/ydb-platform/ydb-go-sdk/v3"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"time"
)

// Register the new canonical v1 service, not the retired OrganizationService.
func Module() fx.Option {
	return fx.Module("resource-manager-v1-grpc",
		fx.Provide(v1engine.NewStore),
		fx.Provide(provideOrgStructureStore),
		fx.Provide(newConfiguredOrgStructureServer),
		fx.Provide(fx.Annotate(newOrgStructureRegistration, fx.ResultTags(grpcserver.RegistrationResultTag))),
		fx.Provide(newConfiguredResourceServer),
		fx.Provide(asV1ServiceServer),
		fx.Provide(fx.Annotate(newRegistration, fx.ResultTags(grpcserver.RegistrationResultTag))),
	)
}
func asV1ServiceServer(server *ResourceServer) pb.ResourceManagerServiceServer { return server }
func newRegistration(server *ResourceServer) grpcserver.Registration {
	return func(registrar grpc.ServiceRegistrar) error {
		pb.RegisterResourceManagerServiceServer(registrar, server)
		return nil
	}
}

// newOrgStructureRegistration registers the typed v1 organization graph APIs.
// The production security gate is deny-all until M8 Access is connected.
func newOrgStructureRegistration(s *OrgStructureServer) grpcserver.Registration {
	return func(registrar grpc.ServiceRegistrar) error {
		pb.RegisterOrgStructureServiceServer(registrar, s)
		pb.RegisterOrgUnitChangeSetServiceServer(registrar, s)
		return nil
	}
}

type orgStructureGRPCParams struct {
	fx.In
	Store  orgstructure.Store
	Config resourcemanager.Config
	Access *accessusecase.CheckPermissionHandler `optional:"true"`
	Impact orgaccess.ImpactReviewer `optional:"true"`
}

func newConfiguredOrgStructureServer(p orgStructureGRPCParams) (*OrgStructureServer, error) {
	return newOrgStructureServerWithImpact(p.Store, p.Config, p.Access, p.Impact)
}

// Selects YDB when configured; the memory store is never used as a silent
// substitute for a failed YDB connection. Production must configure YDB.
func provideOrgStructureStore(cfg resourcemanager.Config, lifecycle fx.Lifecycle) (orgstructure.Store, error) {
	if cfg.YDBDSN == "" {
		// With no DSN production authorization still fails closed.
		return memory.NewOrgStructureStore(), nil
	}
	if cfg.AllowUnauthenticated {
		return nil, fmt.Errorf("YDB mode cannot enable unauthenticated development access")
	}
	db, err := sql.Open("ydb", cfg.YDBDSN)
	if err != nil {
		return nil, fmt.Errorf("open YDB: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect YDB: %w", err)
	}
	store, err := ydb.NewStore(db, cfg.YDBTablePrefix)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { return db.Close() }})
	return store, nil
}

func newConfiguredResourceServer(store *v1engine.Store, cfg resourcemanager.Config, orgStore orgstructure.Store) *ResourceServer {
	bootstrap, ok := orgStore.(orgstructure.BootstrapRepository)
	if !ok {
		return NewResourceServer(store, cfg)
	}
	return NewResourceServer(store, cfg, bootstrap)
}
