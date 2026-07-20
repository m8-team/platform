package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func TestWorkspaceRepositoryListsOnlyRequestedParent(t *testing.T) {
	repository := memory.NewWorkspaceRepository()
	firstParent, secondParent := organization.NewID(), organization.NewID()
	for _, parent := range []organization.ID{firstParent, secondParent} {
		value, err := workspace.New(workspace.CreateParams{ID: organization.NewID(), OrganizationID: parent, Now: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Create(context.Background(), value); err != nil {
			t.Fatal(err)
		}
	}

	values, err := repository.List(context.Background(), firstParent, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || !values[0].OrganizationID().Equal(firstParent) {
		t.Fatalf("List() returned %d values for wrong parent", len(values))
	}
}
