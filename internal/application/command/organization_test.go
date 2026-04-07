package command

import (
	"context"
	"errors"
	"testing"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/testutil"
)

func TestUpdateOrganizationRejectsETagConflict(t *testing.T) {
	repo := testutil.NewOrganizationRepository()
	idempotency := testutil.NewIdempotencyStore()
	outbox := &testutil.OutboxWriter{}
	uuids := &testutil.UUIDGenerator{Values: []string{"etag-2"}}
	clock := testutil.Clock{}

	item, err := organization.New(organization.CreateParams{
		ID:   "11111111-1111-4111-8111-111111111111",
		Now:  clock.Now(),
		ETag: "etag-1",
	})
	if err != nil {
		t.Fatalf("organization.New() error = %v", err)
	}
	repo.Items[item.ID] = item

	handler := UpdateOrganizationHandler{
		TxManager:   testutil.TxManager{},
		Repository:  repo,
		Idempotency: idempotency,
		Outbox:      outbox,
		Clock:       clock,
		UUIDs:       uuids,
	}

	_, err = handler.Handle(context.Background(), UpdateOrganization{
		Metadata:   appcommon.Metadata{IdempotencyKey: "update-1"},
		ID:         item.ID,
		ETag:       "stale-etag",
		UpdateMask: []string{"name"},
		Name:       stringPtr("changed"),
	})
	if !errors.Is(err, organization.ErrETagMismatch) {
		t.Fatalf("expected ErrETagMismatch, got %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
