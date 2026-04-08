package organizationcommand

import (
	"errors"
	"testing"

	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

func TestUpdateMaskValidatorRejectsDuplicates(t *testing.T) {
	t.Parallel()

	name := "Acme"
	err := UpdateMaskValidator{}.Validate(organizationboundary.UpdateOrganizationInput{
		UpdateMask: []string{"name", "name"},
		Name:       &name,
	})
	if !errors.Is(err, usecasecommon.ErrInvalidMask) {
		t.Fatalf("expected ErrInvalidMask, got %v", err)
	}
}

func TestUpdateMaskValidatorRejectsMaskPayloadMismatch(t *testing.T) {
	t.Parallel()

	name := "Acme"
	err := UpdateMaskValidator{}.Validate(organizationboundary.UpdateOrganizationInput{
		UpdateMask: []string{"description"},
		Name:       &name,
	})
	if !errors.Is(err, usecasecommon.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
