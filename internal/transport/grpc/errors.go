package grpctransport

import (
	"errors"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/hierarchy"
	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, appcommon.ErrDuplicateRequest):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, appcommon.ErrInvalidMask):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, organization.ErrNotFound),
		errors.Is(err, workspace.ErrNotFound),
		errors.Is(err, project.ErrNotFound),
		errors.Is(err, hierarchy.ErrParentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, organization.ErrETagMismatch),
		errors.Is(err, workspace.ErrETagMismatch),
		errors.Is(err, project.ErrETagMismatch):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, hierarchy.ErrDeleteBlocked),
		errors.Is(err, hierarchy.ErrParentDeleted),
		errors.Is(err, hierarchy.ErrUndeleteParentInvalid),
		errors.Is(err, organization.ErrAlreadyDeleted),
		errors.Is(err, organization.ErrNotDeleted),
		errors.Is(err, workspace.ErrAlreadyDeleted),
		errors.Is(err, workspace.ErrNotDeleted),
		errors.Is(err, project.ErrAlreadyDeleted),
		errors.Is(err, project.ErrNotDeleted):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, organization.ErrInvalidID),
		errors.Is(err, organization.ErrInvalidUpdatePath),
		errors.Is(err, organization.ErrImmutableID),
		errors.Is(err, organization.ErrDeleted),
		errors.Is(err, workspace.ErrInvalidID),
		errors.Is(err, workspace.ErrInvalidParentID),
		errors.Is(err, workspace.ErrInvalidUpdatePath),
		errors.Is(err, workspace.ErrImmutableID),
		errors.Is(err, workspace.ErrImmutableParent),
		errors.Is(err, workspace.ErrDeleted),
		errors.Is(err, project.ErrInvalidID),
		errors.Is(err, project.ErrInvalidParentID),
		errors.Is(err, project.ErrInvalidUpdatePath),
		errors.Is(err, project.ErrImmutableID),
		errors.Is(err, project.ErrImmutableParent),
		errors.Is(err, project.ErrDeleted):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ports.ErrNotImplemented):
		return status.Error(codes.Unimplemented, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
