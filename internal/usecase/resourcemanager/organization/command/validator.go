package organizationcommand

import (
	"fmt"
	"slices"
	"strings"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/requestmeta"
)

type MetadataValidator interface {
	Validate(requestmeta.RequestMetadata) error
}

type RequiredMetadataValidator struct{}

func (RequiredMetadataValidator) Validate(metadata requestmeta.RequestMetadata) error {
	if strings.TrimSpace(metadata.Actor) == "" {
		return fmt.Errorf("%w: actor is required", usecasecommon.ErrInvalidInput)
	}
	if strings.TrimSpace(metadata.CorrelationID) == "" {
		return fmt.Errorf("%w: correlation_id is required", usecasecommon.ErrInvalidInput)
	}
	return nil
}

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
