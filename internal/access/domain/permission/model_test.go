package permission

import (
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

const testTenant = "11111111-1111-4111-8111-111111111111"
const testID = "22222222-2222-4222-8222-222222222222"

func TestModelPublishAndImmutability(t *testing.T) {
	tenant, _ := shared.NewTenantID(testTenant)
	p, _ := NewPermission("workspace.view")
	d, _ := NewDefinition(p, "workspace", "view")
	m, err := NewAuthorizationModel(testID, tenant, []Definition{d})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.WithDefinitions(0, nil); !errors.Is(err, shared.ErrRevisionConflict) {
		t.Fatalf("expected CAS rejection: %v", err)
	}
	revision, _ := shared.NewModelRevision("v1")
	published, err := m.Publish(1, revision)
	if err != nil || published.State() != ModelPublished || published.Revision() != 2 {
		t.Fatalf("publish failed: %v", err)
	}
	if _, err = published.WithDefinitions(2, nil); !errors.Is(err, ErrModelPublished) {
		t.Fatalf("published mutation: %v", err)
	}
	if m.State() != ModelDraft {
		t.Fatal("original mutated")
	}
	if _, err = NewAuthorizationModel(testID, tenant, []Definition{d, d}); !errors.Is(err, ErrDuplicateDefinition) {
		t.Fatalf("duplicate accepted: %v", err)
	}
}

func TestPermissionValidation(t *testing.T) {
	p, _ := NewPermission("organization.edit")
	if err := ValidateSet([]Permission{p, p}, true); !errors.Is(err, ErrDuplicatePermission) {
		t.Fatal(err)
	}
	if _, err := NewPermission("Invalid Name"); err == nil {
		t.Fatal("invalid action accepted")
	}
}
