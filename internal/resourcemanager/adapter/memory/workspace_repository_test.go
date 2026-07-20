package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func TestWorkspaceRepositoryListsOnlyRequestedParent(t *testing.T) {
	repository := memory.NewWorkspaceRepository()
	firstParent, secondParent := organization.NewID(), organization.NewID()
	for _, parent := range []organization.ID{firstParent, secondParent} {
		value, err := workspace.New(workspace.CreateParams{ID: workspace.NewID(), OrganizationID: parent, Now: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Create(context.Background(), value); err != nil {
			t.Fatal(err)
		}
	}

	result, err := repository.List(context.Background(), ports.ListWorkspacesOptions{OrganizationID: firstParent})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Workspaces) != 1 || !result.Workspaces[0].OrganizationID().Equal(firstParent) {
		t.Fatalf("List() returned %d values for wrong parent", len(result.Workspaces))
	}

	all, err := repository.List(context.Background(), ports.ListWorkspacesOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Workspaces) != 2 {
		t.Fatalf("List() without parent returned %d values, want 2", len(all.Workspaces))
	}
}

func TestWorkspaceRepositoryCompareAndSwapAndHierarchyProjection(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewWorkspaceRepository()
	parent := organization.NewID()
	value, err := workspace.New(workspace.CreateParams{
		ID: workspace.NewID(), OrganizationID: parent, Name: "Initial", Now: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, value); err != nil {
		t.Fatal(err)
	}
	hasChildren, err := repository.HasNonDeleted(ctx, parent)
	if err != nil || !hasChildren {
		t.Fatalf("HasNonDeleted() = %t, %v", hasChildren, err)
	}

	first, _ := repository.Get(ctx, value.ID())
	stale, _ := repository.Get(ctx, value.ID())
	name := "Winner"
	if err := first.Update(workspace.UpdateParams{Name: &name, Now: first.UpdateTime().Add(time.Second), ExpectedVersion: first.Version()}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Update(ctx, first, value.Version()); err != nil {
		t.Fatal(err)
	}
	staleName := "Stale"
	if err := stale.Update(workspace.UpdateParams{Name: &staleName, Now: stale.UpdateTime().Add(time.Second), ExpectedVersion: stale.Version()}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Update(ctx, stale, value.Version()); !errors.Is(err, ports.ErrWorkspaceVersionConflict) {
		t.Fatalf("stale Update() error = %v, want %v", err, ports.ErrWorkspaceVersionConflict)
	}

	current, _ := repository.Get(ctx, value.ID())
	if err := repository.Update(ctx, current, current.Version()); !errors.Is(err, ports.ErrInvalidWorkspaceVersion) {
		t.Fatalf("unchanged Update() error = %v, want %v", err, ports.ErrInvalidWorkspaceVersion)
	}
}
