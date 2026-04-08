package workspacecmd

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/m8platform/platform/internal/adapters/outbound/idempotency"
	"github.com/m8platform/platform/internal/adapters/outbound/outbox"
	"github.com/m8platform/platform/internal/adapters/outbound/postgres/resourcemanager"
	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/entities/resourcemanager/hierarchy"
	organizationentity "github.com/m8platform/platform/internal/entities/resourcemanager/organization"
	"github.com/m8platform/platform/internal/testutil"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

func TestCreateWorkspaceRequiresActiveParent(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)
	store := postgres.NewStore()
	orgRepository := postgres.OrganizationRepository{Store: store}
	orgID := uuid.NewString()
	org, err := organizationentity.New(organizationentity.CreateParams{
		ID:   orgID,
		Name: "Acme",
		Now:  now,
		ETag: "etag-1",
	})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	if err := org.SoftDelete(now.Add(time.Minute), now.Add(24*time.Hour), "etag-1", "etag-2"); err != nil {
		t.Fatalf("soft delete organization: %v", err)
	}
	if err := orgRepository.Create(t.Context(), org); err != nil {
		t.Fatalf("seed organization: %v", err)
	}

	clock := testutil.FakeClock{Current: now.Add(2 * time.Minute)}
	uuidGen := &testutil.SequenceUUIDGenerator{Values: []string{uuid.NewString(), "etag-3"}}
	interactor := CreateInteractor{
		TxManager:        postgres.TxManager{},
		Repository:       postgres.WorkspaceRepository{Store: store},
		HierarchyReader:  postgres.HierarchyReader{Store: store},
		HierarchyPolicy:  domainservices.HierarchyPolicy{},
		IdempotencyStore: idempotency.NewStore(clock),
		OutboxWriter:     outbox.NewWriter(),
		Clock:            clock,
		UUIDGenerator:    uuidGen,
	}

	_, err = interactor.Execute(t.Context(), boundaries.CreateWorkspaceInput{
		OrganizationID: orgID,
		Name:           "Workspace",
	})
	if !errors.Is(err, hierarchy.ErrParentDeleted) {
		t.Fatalf("expected ErrParentDeleted, got %v", err)
	}
}
