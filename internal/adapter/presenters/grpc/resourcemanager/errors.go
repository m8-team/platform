package grpcpresenter

import (
	"context"
	"errors"

	"github.com/m8platform/platform/internal/entity/resourcemanager/hierarchy"
	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
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
		errors.Is(err, hierarchy.ErrParentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecasecommon.ErrInvalidMask),
		errors.Is(err, usecasecommon.ErrInvalidInput),
		errors.Is(err, organizationentity.ErrInvalidID),
		errors.Is(err, organizationentity.ErrImmutableID),
		errors.Is(err, organizationentity.ErrInvalidUpdatePath):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, organizationentity.ErrETagMismatch):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, organizationentity.ErrDeleted),
		errors.Is(err, organizationentity.ErrAlreadyDeleted),
		errors.Is(err, organizationentity.ErrNotDeleted),
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
