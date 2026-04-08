package organizationcommand

import (
	"fmt"
	"slices"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

type UpdateInputValidator interface {
	Validate(organizationboundary.UpdateOrganizationInput) error
}

type UpdateMaskValidator struct{}

func (UpdateMaskValidator) Validate(input organizationboundary.UpdateOrganizationInput) error {
	if len(input.UpdateMask) == 0 {
		return usecasecommon.ErrInvalidMask
	}

	seen := make(map[string]struct{}, len(input.UpdateMask))
	for _, path := range input.UpdateMask {
		if _, duplicate := seen[path]; duplicate {
			return fmt.Errorf("%w: duplicate path %s", usecasecommon.ErrInvalidMask, path)
		}
		seen[path] = struct{}{}
		if !slices.Contains(organizationentity.AllowedUpdatePaths, path) {
			return fmt.Errorf("%w: %s", usecasecommon.ErrInvalidMask, path)
		}
	}

	if _, ok := seen["name"]; ok && input.Name == nil {
		return fmt.Errorf("%w: missing payload for name", usecasecommon.ErrInvalidInput)
	}
	if _, ok := seen["description"]; ok && input.Description == nil {
		return fmt.Errorf("%w: missing payload for description", usecasecommon.ErrInvalidInput)
	}
	if _, ok := seen["annotations"]; ok && input.Annotations == nil {
		return fmt.Errorf("%w: missing payload for annotations", usecasecommon.ErrInvalidInput)
	}
	if input.Name != nil && !contains(seen, "name") {
		return fmt.Errorf("%w: name provided without update_mask", usecasecommon.ErrInvalidInput)
	}
	if input.Description != nil && !contains(seen, "description") {
		return fmt.Errorf("%w: description provided without update_mask", usecasecommon.ErrInvalidInput)
	}
	if input.Annotations != nil && !contains(seen, "annotations") {
		return fmt.Errorf("%w: annotations provided without update_mask", usecasecommon.ErrInvalidInput)
	}
	return nil
}

func contains(values map[string]struct{}, key string) bool {
	_, ok := values[key]
	return ok
}
