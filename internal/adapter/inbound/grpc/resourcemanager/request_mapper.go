package grpcadapter

import (
	"slices"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

func cloneMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func fieldMaskPaths(mask *resourcemanagerv1.UpdateOrganizationRequest) []string {
	if mask.GetUpdateMask() == nil {
		return nil
	}
	return append([]string(nil), mask.GetUpdateMask().GetPaths()...)
}

func optionalString(mask []string, path string, value string) *string {
	if !slices.Contains(mask, path) {
		return nil
	}
	v := value
	return &v
}

func optionalMap(mask []string, path string, value map[string]string) *map[string]string {
	if !slices.Contains(mask, path) {
		return nil
	}
	cloned := cloneMap(value)
	return &cloned
}
