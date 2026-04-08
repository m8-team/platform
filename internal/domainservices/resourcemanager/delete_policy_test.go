package domainservices

import (
	"errors"
	"testing"

	"github.com/m8platform/platform/internal/entity/resourcemanager/hierarchy"
)

func TestDeletePolicyEnsureAllowed(t *testing.T) {
	t.Parallel()

	err := DeletePolicy{}.EnsureAllowed(true)
	if !errors.Is(err, hierarchy.ErrDeleteBlocked) {
		t.Fatalf("expected ErrDeleteBlocked, got %v", err)
	}
}
