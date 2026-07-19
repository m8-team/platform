package grpcadapter

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func invalidArgument(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	}
	if current, ok := status.FromError(err); ok && current.Code() != codes.Unknown {
		return err
	}

	switch {
	case errors.Is(err, ports.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "authentication is required")
	case errors.Is(err, ports.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, ports.ErrOrganizationNotFound),
		errors.Is(err, organization.ErrOrganizationAlreadyDeleted):
		return status.Error(codes.NotFound, "organization not found")
	case errors.Is(err, ports.ErrOrganizationAlreadyExists):
		return status.Error(codes.AlreadyExists, "organization already exists")
	case errors.Is(err, ports.ErrOrganizationVersionConflict),
		errors.Is(err, organization.ErrVersionMismatch):
		return status.Error(codes.Aborted, "organization version conflict")
	case errors.Is(err, ports.ErrOrganizationRepositoryUnavailable):
		return status.Error(codes.Unavailable, "organization repository is unavailable")
	case errors.Is(err, usecase.ErrInvalidOrganizationPageSize),
		errors.Is(err, usecase.ErrInvalidOrganizationPageToken),
		errors.Is(err, usecase.ErrInvalidOrganizationFilter),
		errors.Is(err, usecase.ErrInvalidOrganizationOrderBy),
		errors.Is(err, ports.ErrInvalidListOrganizationsOptions),
		errors.Is(err, organization.ErrInvalidOrganizationName),
		errors.Is(err, organization.ErrOrganizationNameTooLong),
		errors.Is(err, organization.ErrInvalidOrganizationDescription),
		errors.Is(err, organization.ErrOrganizationDescriptionTooLong),
		errors.Is(err, organization.ErrInvalidOrganizationLabel),
		errors.Is(err, organization.ErrNoOrganizationUpdates):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrOrganizationHasWorkspaces),
		errors.Is(err, organization.ErrOrganizationDeleted),
		errors.Is(err, organization.ErrOrganizationNotDeleted),
		errors.Is(err, organization.ErrInvalidStateTransition),
		errors.Is(err, organization.ErrPurgeTimePassed),
		errors.Is(err, organization.ErrVersionOverflow):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal organization service error")
	}
}
