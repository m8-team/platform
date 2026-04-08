package grpcpresenter

import (
	"context"
	"errors"

	"github.com/m8platform/platform/internal/entity/resourcemanager/hierarchy"
	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	projectentity "github.com/m8platform/platform/internal/entity/resourcemanager/project"
	workspaceentity "github.com/m8platform/platform/internal/entity/resourcemanager/workspace"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PresentError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	case errors.Is(err, organizationentity.ErrNotFound),
		errors.Is(err, workspaceentity.ErrNotFound),
		errors.Is(err, projectentity.ErrNotFound),
		errors.Is(err, hierarchy.ErrParentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecasecommon.ErrInvalidMask),
		errors.Is(err, usecasecommon.ErrInvalidInput),
		errors.Is(err, organizationentity.ErrInvalidID),
		errors.Is(err, organizationentity.ErrImmutableID),
		errors.Is(err, organizationentity.ErrInvalidUpdatePath),
		errors.Is(err, workspaceentity.ErrInvalidID),
		errors.Is(err, workspaceentity.ErrInvalidParentID),
		errors.Is(err, workspaceentity.ErrImmutableID),
		errors.Is(err, workspaceentity.ErrImmutableParent),
		errors.Is(err, workspaceentity.ErrInvalidUpdatePath),
		errors.Is(err, projectentity.ErrInvalidID),
		errors.Is(err, projectentity.ErrInvalidParentID),
		errors.Is(err, projectentity.ErrImmutableID),
		errors.Is(err, projectentity.ErrImmutableParent),
		errors.Is(err, projectentity.ErrInvalidUpdatePath):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, organizationentity.ErrETagMismatch),
		errors.Is(err, workspaceentity.ErrETagMismatch),
		errors.Is(err, projectentity.ErrETagMismatch):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, organizationentity.ErrDeleted),
		errors.Is(err, organizationentity.ErrAlreadyDeleted),
		errors.Is(err, organizationentity.ErrNotDeleted),
		errors.Is(err, workspaceentity.ErrDeleted),
		errors.Is(err, workspaceentity.ErrAlreadyDeleted),
		errors.Is(err, workspaceentity.ErrNotDeleted),
		errors.Is(err, projectentity.ErrDeleted),
		errors.Is(err, projectentity.ErrAlreadyDeleted),
		errors.Is(err, projectentity.ErrNotDeleted),
		errors.Is(err, hierarchy.ErrDeleteBlocked),
		errors.Is(err, hierarchy.ErrParentDeleted),
		errors.Is(err, hierarchy.ErrUndeleteParentInvalid):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, usecasecommon.ErrDuplicateRequest):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
