package grpcadapter

import (
	"context"
	"errors"

	organizationapp "github.com/m8-team/platform/internal/resourcemanager/app/organization"
	workspaceapp "github.com/m8-team/platform/internal/resourcemanager/app/workspace"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
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
		errors.Is(err, ports.ErrWorkspaceNotFound),
		errors.Is(err, organization.ErrOrganizationAlreadyDeleted),
		errors.Is(err, workspace.ErrWorkspaceAlreadyDeleted):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, ports.ErrOrganizationAlreadyExists):
		return status.Error(codes.AlreadyExists, "organization already exists")
	case errors.Is(err, ports.ErrWorkspaceAlreadyExists):
		return status.Error(codes.AlreadyExists, "workspace already exists")
	case errors.Is(err, ports.ErrOrganizationVersionConflict),
		errors.Is(err, ports.ErrWorkspaceVersionConflict),
		errors.Is(err, organization.ErrVersionMismatch),
		errors.Is(err, workspace.ErrVersionMismatch):
		return status.Error(codes.Aborted, "resource version conflict")
	case errors.Is(err, ports.ErrOrganizationRepositoryUnavailable),
		errors.Is(err, ports.ErrWorkspaceRepositoryUnavailable):
		return status.Error(codes.Unavailable, "resource repository is unavailable")
	case errors.Is(err, organizationapp.ErrInvalidOrganizationPageSize),
		errors.Is(err, organizationapp.ErrInvalidOrganizationPageToken),
		errors.Is(err, organizationapp.ErrInvalidOrganizationFilter),
		errors.Is(err, organizationapp.ErrInvalidOrganizationOrderBy),
		errors.Is(err, ports.ErrInvalidListOrganizationsOptions),
		errors.Is(err, organization.ErrInvalidOrganizationName),
		errors.Is(err, organization.ErrOrganizationNameTooLong),
		errors.Is(err, organization.ErrInvalidOrganizationDescription),
		errors.Is(err, organization.ErrOrganizationDescriptionTooLong),
		errors.Is(err, organization.ErrInvalidOrganizationLabel),
		errors.Is(err, organization.ErrNoOrganizationUpdates),
		errors.Is(err, workspaceapp.ErrInvalidWorkspacePageSize),
		errors.Is(err, workspaceapp.ErrInvalidWorkspacePageToken),
		errors.Is(err, workspaceapp.ErrInvalidWorkspaceFilter),
		errors.Is(err, workspaceapp.ErrInvalidWorkspaceOrderBy),
		errors.Is(err, ports.ErrInvalidListWorkspacesOptions),
		errors.Is(err, workspace.ErrEmptyWorkspaceID),
		errors.Is(err, workspace.ErrInvalidWorkspaceID),
		errors.Is(err, workspace.ErrInvalidWorkspaceName),
		errors.Is(err, workspace.ErrWorkspaceNameTooLong),
		errors.Is(err, workspace.ErrInvalidWorkspaceDescription),
		errors.Is(err, workspace.ErrWorkspaceDescriptionTooLong),
		errors.Is(err, workspace.ErrInvalidWorkspaceLabel),
		errors.Is(err, workspace.ErrInvalidWorkspaceVersion),
		errors.Is(err, workspace.ErrNoWorkspaceUpdates):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, organizationapp.ErrOrganizationHasWorkspaces),
		errors.Is(err, organization.ErrOrganizationDeleted),
		errors.Is(err, organization.ErrOrganizationNotDeleted),
		errors.Is(err, organization.ErrOrganizationNotActive),
		errors.Is(err, organization.ErrInvalidStateTransition),
		errors.Is(err, organization.ErrPurgeTimePassed),
		errors.Is(err, organization.ErrVersionOverflow),
		errors.Is(err, workspace.ErrWorkspaceDeleted),
		errors.Is(err, workspace.ErrWorkspaceNotDeleted),
		errors.Is(err, workspace.ErrInvalidStateTransition),
		errors.Is(err, workspace.ErrPurgeTimePassed),
		errors.Is(err, workspace.ErrVersionOverflow):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal resource manager service error")
	}
}
