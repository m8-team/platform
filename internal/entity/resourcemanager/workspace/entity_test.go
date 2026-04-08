package workspace

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/m8platform/platform/internal/testutil"
)

func TestEntityUpdateRejectsParentChange(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)
	entity, err := New(CreateParams{
		ID:             uuid.NewString(),
		OrganizationID: uuid.NewString(),
		Name:           "Workspace",
		Now:            now,
		ETag:           "etag-1",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	err = entity.Update([]string{"name"}, UpdateParams{
		ID:             entity.ID,
		OrganizationID: uuid.NewString(),
		Name:           testutil.StringPointer("Renamed"),
		ETag:           "etag-1",
	}, now.Add(time.Minute), "etag-2")
	if !errors.Is(err, ErrImmutableParent) {
		t.Fatalf("expected ErrImmutableParent, got %v", err)
	}
}
