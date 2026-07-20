package workspace_test

import (
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func TestWorkspaceLifecycleAndParent(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	organizationID := organization.NewID()
	value, err := workspace.New(workspace.CreateParams{
		ID:             organization.NewID(),
		OrganizationID: organizationID,
		Name:           "Production",
		Now:            now,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if !value.OrganizationID().Equal(organizationID) {
		t.Fatalf("OrganizationID() = %s, want %s", value.OrganizationID(), organizationID)
	}
	if value.State() != workspace.StateActive || value.Version().Int64() != 1 {
		t.Fatalf("new workspace state/version = %s/%d", value.State(), value.Version().Int64())
	}

	if err := value.Delete(now.Add(time.Minute), now.Add(24*time.Hour), value.Version()); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !value.IsDeleted() || value.DeleteTime() == nil || value.PurgeTime() == nil {
		t.Fatal("Delete() did not retain a tombstone")
	}
	if err := value.Undelete(now.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Undelete() error = %v", err)
	}
	if value.IsDeleted() || value.DeleteTime() != nil || value.PurgeTime() != nil {
		t.Fatal("Undelete() did not restore the workspace")
	}
}

func TestWorkspaceRequiresParentOrganization(t *testing.T) {
	_, err := workspace.New(workspace.CreateParams{ID: organization.NewID(), Now: time.Now()})
	if err == nil {
		t.Fatal("New() error = nil, want invalid parent error")
	}
}
