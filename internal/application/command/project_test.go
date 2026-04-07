package command

import (
	"context"
	"errors"
	"testing"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/ports"
	"github.com/m8platform/platform/internal/testutil"
)

func TestCreateProjectRejectsDuplicateIdempotencyKey(t *testing.T) {
	repo := testutil.NewProjectRepository()
	hierarchyRepo := testutil.NewHierarchyRepository()
	idempotency := testutil.NewIdempotencyStore()
	outbox := &testutil.OutboxWriter{}
	uuids := &testutil.UUIDGenerator{Values: []string{
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		"33333333-3333-4333-8333-333333333333",
	}}
	clock := testutil.Clock{}

	hierarchyRepo.Workspaces["aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"] = ports.HierarchyNode{
		ID:     "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Exists: true,
	}

	handler := CreateProjectHandler{
		TxManager:   testutil.TxManager{},
		Repository:  repo,
		Hierarchy:   hierarchyRepo,
		Idempotency: idempotency,
		Outbox:      outbox,
		Clock:       clock,
		UUIDs:       uuids,
	}

	command := CreateProject{
		Metadata:    appcommon.Metadata{IdempotencyKey: "create-1"},
		WorkspaceID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		Name:        "Project A",
	}

	if _, err := handler.Handle(context.Background(), command); err != nil {
		t.Fatalf("first Handle() error = %v", err)
	}

	_, err := handler.Handle(context.Background(), command)
	if !errors.Is(err, appcommon.ErrDuplicateRequest) {
		t.Fatalf("expected ErrDuplicateRequest, got %v", err)
	}
}
