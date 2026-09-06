package grpcadapter

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	workspaceapp "github.com/m8-team/platform/internal/resourcemanager/app/workspace"

	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WorkspaceServer struct {
	resourcemanagerpb.UnimplementedWorkspaceServiceServer

	application  *workspaceapp.WorkspaceService
	clock        ports.Clock
	operationIDs OperationIDGenerator
}

func NewWorkspaceServer(
	application *workspaceapp.WorkspaceService,
	clock ports.Clock,
	operationIDs OperationIDGenerator,
) (*WorkspaceServer, error) {
	if application == nil {
		return nil, errors.New("workspace application service is required")
	}
	if clock == nil {
		return nil, errors.New("workspace gRPC clock is required")
	}
	if operationIDs == nil {
		return nil, errors.New("operation id generator is required")
	}

	return &WorkspaceServer{
		application:  application,
		clock:        clock,
		operationIDs: operationIDs,
	}, nil
}

func (s *WorkspaceServer) GetWorkspace(
	ctx context.Context,
	request *resourcemanagerpb.GetWorkspaceRequest,
) (*resourcemanagerpb.Workspace, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	id, err := parseCanonicalWorkspaceID(request.GetId())
	if err != nil {
		return nil, invalidArgument("id must be a canonical non-zero UUID")
	}

	value, err := s.application.Get(ctx, workspaceapp.GetWorkspace{ID: id})
	if err != nil {
		return nil, mapError(err)
	}
	response, err := workspaceToProto(value)
	if err != nil {
		return nil, status.Error(codes.Internal, "map workspace response")
	}
	return response, nil
}

func (s *WorkspaceServer) ListWorkspaces(
	ctx context.Context,
	request *resourcemanagerpb.ListWorkspacesRequest,
) (*resourcemanagerpb.ListWorkspacesResponse, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	if request.GetPageSize() < 0 || request.GetPageSize() > ports.MaxWorkspacePageSize {
		return nil, invalidArgument("page_size must be between 0 and 1000")
	}
	if utf8.RuneCountInString(request.GetPageToken()) > maxPageTokenRunes {
		return nil, invalidArgument("page_token exceeds 1024 characters")
	}
	if utf8.RuneCountInString(request.GetFilter()) > maxFilterRunes {
		return nil, invalidArgument("filter exceeds 1024 characters")
	}
	if utf8.RuneCountInString(request.GetOrderBy()) > maxOrderByRunes {
		return nil, invalidArgument("order_by exceeds 128 characters")
	}
	var organizationID organization.ID
	if request.GetOrganizationId() != "" {
		var err error
		organizationID, err = parseCanonicalOrganizationID(request.GetOrganizationId())
		if err != nil {
			return nil, invalidArgument("organization_id must be a canonical non-zero UUID")
		}
	}

	result, err := s.application.List(ctx, workspaceapp.ListWorkspaces{
		OrganizationID: organizationID,
		PageSize:       int(request.GetPageSize()),
		PageToken:      request.GetPageToken(),
		Filter:         request.GetFilter(),
		OrderBy:        request.GetOrderBy(),
		ShowDeleted:    request.GetShowDeleted(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	if result.TotalSize < 0 || result.TotalSize > math.MaxInt32 {
		return nil, status.Error(codes.Internal, "workspace total size exceeds API range")
	}

	workspaces := make([]*resourcemanagerpb.Workspace, 0, len(result.Workspaces))
	for _, value := range result.Workspaces {
		mapped, err := workspaceToProto(value)
		if err != nil {
			return nil, status.Error(codes.Internal, "map workspace list response")
		}
		workspaces = append(workspaces, mapped)
	}

	return &resourcemanagerpb.ListWorkspacesResponse{
		Workspaces:    workspaces,
		NextPageToken: result.NextPageToken,
		TotalSize:     int32(result.TotalSize),
	}, nil
}

func (s *WorkspaceServer) CreateWorkspace(
	ctx context.Context,
	request *resourcemanagerpb.CreateWorkspaceRequest,
) (*longrunningpb.Operation, error) {
	organizationID, input, err := validateCreateWorkspaceRequest(request)
	if err != nil {
		return nil, err
	}
	prepared, err := prepareCompletedOperation(s.clock, s.operationIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "prepare workspace operation")
	}

	value, err := s.application.Create(ctx, workspaceapp.CreateWorkspace{
		OrganizationID: organizationID,
		Name:           input.GetName(),
		Description:    input.GetDescription(),
		Labels:         input.GetLabels(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	operation, err := s.completedWorkspaceOperation(prepared, value)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed workspace operation")
	}
	return operation, nil
}

func (s *WorkspaceServer) UpdateWorkspace(
	ctx context.Context,
	request *resourcemanagerpb.UpdateWorkspaceRequest,
) (*longrunningpb.Operation, error) {
	cmd, err := workspaceUpdateCommand(request)
	if err != nil {
		return nil, err
	}
	prepared, err := prepareCompletedOperation(s.clock, s.operationIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "prepare workspace operation")
	}

	value, err := s.application.Update(ctx, cmd)
	if err != nil {
		return nil, mapError(err)
	}
	operation, err := s.completedWorkspaceOperation(prepared, value)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed workspace operation")
	}
	return operation, nil
}

func (s *WorkspaceServer) DeleteWorkspace(
	ctx context.Context,
	request *resourcemanagerpb.DeleteWorkspaceRequest,
) (*longrunningpb.Operation, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	id, err := parseCanonicalWorkspaceID(request.GetId())
	if err != nil {
		return nil, invalidArgument("id must be a canonical non-zero UUID")
	}
	if request.GetVersion() < 0 {
		return nil, invalidArgument("version must be non-negative")
	}
	prepared, err := prepareCompletedOperation(s.clock, s.operationIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "prepare workspace delete operation")
	}

	_, err = s.application.Delete(ctx, workspaceapp.DeleteWorkspace{
		ID:              id,
		ExpectedVersion: types.Version(request.GetVersion()),
		AllowMissing:    request.GetAllowMissing(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	operation, err := s.completedWorkspaceDeleteOperation(prepared, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed workspace delete operation")
	}
	return operation, nil
}

func (s *WorkspaceServer) UndeleteWorkspace(
	ctx context.Context,
	request *resourcemanagerpb.UndeleteWorkspaceRequest,
) (*longrunningpb.Operation, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	id, err := parseCanonicalWorkspaceID(request.GetId())
	if err != nil {
		return nil, invalidArgument("id must be a canonical non-zero UUID")
	}
	prepared, err := prepareCompletedOperation(s.clock, s.operationIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "prepare workspace operation")
	}

	if request.GetVersion() < 0 {
		return nil, invalidArgument("version must be non-negative")
	}
	value, err := s.application.Undelete(ctx, workspaceapp.UndeleteWorkspace{ID: id, ExpectedVersion: types.Version(request.GetVersion())})
	if err != nil {
		return nil, mapError(err)
	}
	operation, err := s.completedWorkspaceOperation(prepared, value)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed workspace operation")
	}
	return operation, nil
}

func validateCreateWorkspaceRequest(
	request *resourcemanagerpb.CreateWorkspaceRequest,
) (organization.ID, *resourcemanagerpb.WorkspaceInput, error) {
	if request == nil {
		return organization.ID{}, nil, invalidArgument("request is required")
	}
	value := request.GetWorkspace()
	if value == nil {
		return organization.ID{}, nil, invalidArgument("workspace is required")
	}
	organizationID, err := parseCanonicalOrganizationID(request.GetOrganizationId())
	if err != nil {
		return organization.ID{}, nil, invalidArgument("organization_id must be a canonical non-zero UUID")
	}
	return organizationID, value, nil
}

func workspaceUpdateCommand(
	request *resourcemanagerpb.UpdateWorkspaceRequest,
) (workspaceapp.UpdateWorkspace, error) {
	if request == nil {
		return workspaceapp.UpdateWorkspace{}, invalidArgument("request is required")
	}
	value := request.GetWorkspace()
	if value == nil {
		return workspaceapp.UpdateWorkspace{}, invalidArgument("workspace is required")
	}
	id, err := parseCanonicalWorkspaceID(value.GetId())
	if err != nil {
		return workspaceapp.UpdateWorkspace{}, invalidArgument("workspace.id must be a canonical non-zero UUID")
	}
	if value.GetVersion() < 0 {
		return workspaceapp.UpdateWorkspace{}, invalidArgument("workspace.version must be non-negative")
	}
	paths, err := mutableUpdatePaths(request.GetUpdateMask())
	if err != nil {
		return workspaceapp.UpdateWorkspace{}, err
	}

	cmd := workspaceapp.UpdateWorkspace{ID: id, ExpectedVersion: types.Version(value.GetVersion())}
	if paths["name"] {
		name := value.GetName()
		cmd.Name = &name
	}
	if paths["description"] {
		description := value.GetDescription()
		cmd.Description = &description
	}
	if paths["labels"] {
		labels := cloneStringMap(value.GetLabels())
		cmd.Labels = &labels
	}
	return cmd, nil
}

func parseCanonicalWorkspaceID(raw string) (workspace.ID, error) {
	if len(raw) != 36 {
		return workspace.ID{}, workspace.ErrInvalidWorkspaceID
	}
	id, err := workspace.ParseID(raw)
	if err != nil || !strings.EqualFold(id.String(), raw) {
		return workspace.ID{}, workspace.ErrInvalidWorkspaceID
	}
	return id, nil
}

func workspaceToProto(value *workspace.Workspace) (*resourcemanagerpb.Workspace, error) {
	if value == nil {
		return nil, errors.New("workspace is nil")
	}
	createTime, err := timestampFromTime(value.CreateTime())
	if err != nil {
		return nil, err
	}
	updateTime, err := timestampFromTime(value.UpdateTime())
	if err != nil {
		return nil, err
	}
	deleteTime, err := timestampFromOptionalTime(value.DeleteTime())
	if err != nil {
		return nil, err
	}
	purgeTime, err := timestampFromOptionalTime(value.PurgeTime())
	if err != nil {
		return nil, err
	}

	return &resourcemanagerpb.Workspace{
		Id:             value.ID().String(),
		OrganizationId: value.OrganizationID().String(),
		State:          workspaceStateToProto(value.State()),
		Name:           value.Name(),
		Description:    value.Description(),
		CreateTime:     createTime,
		UpdateTime:     updateTime,
		DeleteTime:     deleteTime,
		PurgeTime:      purgeTime,
		Version:        value.Version().Int64(),
		Labels:         value.Labels(),
	}, nil
}

func workspaceStateToProto(state workspace.State) resourcemanagerpb.Workspace_State {
	switch state {
	case workspace.StateCreating:
		return resourcemanagerpb.Workspace_CREATING
	case workspace.StateActive:
		return resourcemanagerpb.Workspace_ACTIVE
	case workspace.StateSuspended:
		return resourcemanagerpb.Workspace_SUSPENDED
	case workspace.StateDeleting:
		return resourcemanagerpb.Workspace_DELETING
	case workspace.StateDeleted:
		return resourcemanagerpb.Workspace_DELETED
	case workspace.StateFailed:
		return resourcemanagerpb.Workspace_FAILED
	default:
		return resourcemanagerpb.Workspace_STATE_UNSPECIFIED
	}
}

var _ resourcemanagerpb.WorkspaceServiceServer = (*WorkspaceServer)(nil)
