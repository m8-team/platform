package grpcadapter

import (
	"context"
	"slices"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/m8platform/platform/internal/platform/middleware"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/requestmeta"
)

func requestMetadata(ctx context.Context) requestmeta.RequestMetadata {
	meta := middleware.FromGRPCContext(ctx)
	return requestmeta.RequestMetadata{
		Actor:          meta.Actor,
		CorrelationID:  meta.CorrelationID,
		CausationID:    meta.CausationID,
		IdempotencyKey: meta.IdempotencyKey,
	}
}

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

func fieldMaskPathsWorkspace(mask *resourcemanagerv1.UpdateWorkspaceRequest) []string {
	if mask.GetUpdateMask() == nil {
		return nil
	}
	return append([]string(nil), mask.GetUpdateMask().GetPaths()...)
}

func fieldMaskPathsProject(mask *resourcemanagerv1.UpdateProjectRequest) []string {
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
