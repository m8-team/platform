package organization

import (
	"errors"
	"testing"
	"time"
)

func TestOrganizationUpdateRejectsImmutableID(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	aggregate, err := New(CreateParams{
		ID:   "11111111-1111-4111-8111-111111111111",
		Now:  now,
		ETag: "etag-1",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = aggregate.Update([]string{"name"}, UpdateFields{
		ID:   "22222222-2222-4222-8222-222222222222",
		Name: stringPtr("renamed"),
	}, now.Add(time.Minute), "etag-2")
	if !errors.Is(err, ErrImmutableID) {
		t.Fatalf("expected ErrImmutableID, got %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
