package command

import (
	"context"
	"errors"
	"testing"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/hierarchy"
	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/testutil"
)

func TestDeleteWorkspaceBlockedByActiveProjects(t *testing.T) {
	repo := testutil.NewWorkspaceRepository()
	hierarchyRepo := testutil.NewHierarchyRepository()
	idempotency := testutil.NewIdempotencyStore()
	outbox := &testutil.OutboxWriter{}
	uuids := &testutil.UUIDGenerator{Values: []string{"etag-2"}}
	clock := testutil.Clock{}

	item, err := workspace.New(workspace.CreateParams{
		ID:             "11111111-1111-4111-8111-111111111111",
		OrganizationID: "22222222-2222-4222-8222-222222222222",
		Now:            clock.Now(),
		ETag:           "etag-1",
	})
	if err != nil {
		t.Fatalf("workspace.New() error = %v", err)
	}
	repo.Items[item.ID] = item
	hierarchyRepo.ActiveProjects[item.ID] = true

	handler := DeleteWorkspaceHandler{
		TxManager:   testutil.TxManager{},
		Repository:  repo,
		Hierarchy:   hierarchyRepo,
		Idempotency: idempotency,
		Outbox:      outbox,
		Clock:       clock,
		UUIDs:       uuids,
	}

	err = handler.Handle(context.Background(), DeleteWorkspace{
		Metadata: appcommon.Metadata{IdempotencyKey: "delete-1"},
		ID:       item.ID,
		ETag:     "etag-1",
	})
	if !errors.Is(err, hierarchy.ErrDeleteBlocked) {
		t.Fatalf("expected ErrDeleteBlocked, got %v", err)
	}
}
