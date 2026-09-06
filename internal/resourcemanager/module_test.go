package resourcemanager_test

import (
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/authz"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/system"
	organizationapp "github.com/m8-team/platform/internal/resourcemanager/app/organization"
)

func TestComposition(t *testing.T) {
	ids := system.NewIDGenerator()
	deps := resourcemanager.Dependencies{Organizations: memory.NewOrganizationRepository(),
		Workspaces: memory.NewWorkspaceRepository(), Authorizer: authz.DenyAll(),
		Clock: system.NewClock(), OrganizationIDs: ids, WorkspaceIDs: ids}
	for _, tt := range []struct {
		name string
		cfg  resourcemanager.Config
		want error
	}{
		{"valid", resourcemanager.Config{ServiceName: "resource-manager"}, nil},
		{"missing name", resourcemanager.Config{}, resourcemanager.ErrEmptyServiceName},
		{"negative retention", resourcemanager.Config{ServiceName: "resource-manager", SoftDeleteRetention: -1}, organizationapp.ErrInvalidSoftDeleteRetention},
		{"short key", resourcemanager.Config{ServiceName: "resource-manager", PageTokenKey: []byte("short")}, organizationapp.ErrInvalidPageTokenKey},
	} {
		t.Run(tt.name, func(t *testing.T) {
			services, err := resourcemanager.New(tt.cfg, deps)
			if !errors.Is(err, tt.want) {
				t.Fatalf("New() = %v, want %v", err, tt.want)
			}
			if err == nil && (services.Organizations == nil || services.Workspaces == nil) {
				t.Fatal("incomplete composition")
			}
		})
	}
	deps.Authorizer = nil
	if _, err := resourcemanager.New(resourcemanager.Config{ServiceName: "resource-manager"}, deps); !errors.Is(err, organizationapp.ErrOrganizationAuthorizerRequired) {
		t.Fatalf("missing explicit authorizer: %v", err)
	}
}
