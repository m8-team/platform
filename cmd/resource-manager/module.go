package main

import (
	"github.com/m8-team/platform/internal/platform/telemetry"
	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/authz"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/system"
	organizationapp "github.com/m8-team/platform/internal/resourcemanager/app/organization"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	workspaceapp "github.com/m8-team/platform/internal/resourcemanager/app/workspace"
	"go.uber.org/fx"
)

// This executable chooses local adapters. The business module has no dependency
// on Fx, memory repositories or a particular authorization implementation.
func resourceManagerModule(cfg resourcemanager.Config, allowUnauthenticated bool) fx.Option {
	return fx.Module("resourcemanager",
		fx.Provide(func() ports.Clock { return system.NewClock() }),
		fx.Provide(func(clock ports.Clock, observer *telemetry.Observer) (*resourcemanager.Services, error) {
			authorizer := authz.DenyAll()
			if allowUnauthenticated {
				authorizer = authz.AllowAll()
			}
			ids := system.NewIDGenerator()
			return resourcemanager.New(cfg, resourcemanager.Dependencies{
				Observer:      observer,
				Organizations: memory.NewOrganizationRepository(), Workspaces: memory.NewWorkspaceRepository(),
				Authorizer: authorizer, Clock: clock, OrganizationIDs: ids, WorkspaceIDs: ids,
			})
		}),
		fx.Provide(func(s *resourcemanager.Services) *organizationapp.OrganizationService { return s.Organizations }),
		fx.Provide(func(s *resourcemanager.Services) *workspaceapp.WorkspaceService { return s.Workspaces }),
	)
}
