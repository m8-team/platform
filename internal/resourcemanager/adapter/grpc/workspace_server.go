package grpcadapter

import (
	"context"
	"errors"
	"math"
	"strings"

	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/app/command"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WorkspaceServer struct {
	resourcemanagerpb.UnimplementedWorkspaceServiceServer
	application  *usecase.WorkspaceService
	clock        ports.Clock
	operationIDs OperationIDGenerator
}

func NewWorkspaceServer(application *usecase.WorkspaceService, clock ports.Clock, ids OperationIDGenerator) (*WorkspaceServer, error) {
	if application == nil || clock == nil || ids == nil {
		return nil, errors.New("workspace gRPC dependencies are required")
	}
	return &WorkspaceServer{application: application, clock: clock, operationIDs: ids}, nil
}
func parseWorkspaceID(raw string) (workspace.ID, error) {
	if len(raw) != 36 {
		return workspace.ID{}, organization.ErrInvalidOrganizationID
	}
	id, err := organization.ParseID(raw)
	if err != nil || !strings.EqualFold(id.String(), raw) {
		return workspace.ID{}, organization.ErrInvalidOrganizationID
	}
	return id, nil
}
func (s *WorkspaceServer) GetWorkspace(ctx context.Context, r *resourcemanagerpb.GetWorkspaceRequest) (*resourcemanagerpb.Workspace, error) {
	if r == nil {
		return nil, invalidArgument("request is required")
	}
	id, e := parseWorkspaceID(r.GetId())
	if e != nil {
		return nil, invalidArgument("id must be a canonical non-zero UUID")
	}
	w, e := s.application.Get(ctx, query.GetWorkspace{ID: id})
	if e != nil {
		return nil, mapError(e)
	}
	return workspaceToProto(w)
}
func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, r *resourcemanagerpb.ListWorkspacesRequest) (*resourcemanagerpb.ListWorkspacesResponse, error) {
	if r == nil {
		return nil, invalidArgument("request is required")
	}
	org, e := parseCanonicalOrganizationID(r.GetOrganizationId())
	if e != nil {
		return nil, invalidArgument("organization_id must be a canonical non-zero UUID")
	}
	result, e := s.application.List(ctx, query.ListWorkspaces{OrganizationID: org, PageSize: int(r.GetPageSize()), PageToken: r.GetPageToken(), Filter: r.GetFilter(), OrderBy: r.GetOrderBy(), ShowDeleted: r.GetShowDeleted()})
	if e != nil {
		return nil, mapError(e)
	}
	if result.TotalSize > math.MaxInt32 {
		return nil, status.Error(codes.Internal, "workspace total size exceeds API range")
	}
	values := make([]*resourcemanagerpb.Workspace, 0, len(result.Workspaces))
	for _, w := range result.Workspaces {
		v, e := workspaceToProto(w)
		if e != nil {
			return nil, e
		}
		values = append(values, v)
	}
	return &resourcemanagerpb.ListWorkspacesResponse{Workspaces: values, NextPageToken: result.NextPageToken, TotalSize: int32(result.TotalSize)}, nil
}
func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, r *resourcemanagerpb.CreateWorkspaceRequest) (*longrunningpb.Operation, error) {
	if r == nil || r.GetWorkspace() == nil {
		return nil, invalidArgument("workspace is required")
	}
	v := r.GetWorkspace()
	if v.GetId() != "" || v.GetOrganizationId() != "" || v.GetState() != 0 || v.GetCreateTime() != nil || v.GetUpdateTime() != nil || v.GetDeleteTime() != nil || v.GetPurgeTime() != nil || v.GetVersion() != 0 {
		return nil, invalidArgument("workspace contains server-assigned fields")
	}
	org, e := parseCanonicalOrganizationID(r.GetOrganizationId())
	if e != nil {
		return nil, invalidArgument("organization_id must be a canonical non-zero UUID")
	}
	w, e := s.application.Create(ctx, command.CreateWorkspace{OrganizationID: org, Name: v.GetName(), Description: v.GetDescription(), Labels: v.GetLabels()})
	if e != nil {
		return nil, mapError(e)
	}
	return s.operation(w, false)
}
func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, r *resourcemanagerpb.UpdateWorkspaceRequest) (*longrunningpb.Operation, error) {
	if r == nil || r.GetWorkspace() == nil {
		return nil, invalidArgument("workspace is required")
	}
	v := r.GetWorkspace()
	id, e := parseWorkspaceID(v.GetId())
	if e != nil {
		return nil, invalidArgument("workspace.id must be a canonical non-zero UUID")
	}
	paths, e := mutableUpdatePaths(r.GetUpdateMask())
	if e != nil {
		return nil, e
	}
	cmd := command.UpdateWorkspace{ID: id, ExpectedVersion: types.Version(v.GetVersion())}
	if paths["name"] {
		x := v.GetName()
		cmd.Name = &x
	}
	if paths["description"] {
		x := v.GetDescription()
		cmd.Description = &x
	}
	if paths["labels"] {
		x := cloneStringMap(v.GetLabels())
		cmd.Labels = &x
	}
	w, e := s.application.Update(ctx, cmd)
	if e != nil {
		return nil, mapError(e)
	}
	return s.operation(w, false)
}
func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, r *resourcemanagerpb.DeleteWorkspaceRequest) (*longrunningpb.Operation, error) {
	if r == nil {
		return nil, invalidArgument("request is required")
	}
	id, e := parseWorkspaceID(r.GetId())
	if e != nil {
		return nil, invalidArgument("id must be a canonical non-zero UUID")
	}
	w, e := s.application.Delete(ctx, command.DeleteWorkspace{ID: id, ExpectedVersion: types.Version(r.GetVersion()), AllowMissing: r.GetAllowMissing()})
	if e != nil {
		return nil, mapError(e)
	}
	if w == nil {
		w, _ = workspace.New(workspace.CreateParams{ID: id, OrganizationID: id, Now: s.clock.Now()})
	}
	return s.operation(w, true)
}
func (s *WorkspaceServer) UndeleteWorkspace(ctx context.Context, r *resourcemanagerpb.UndeleteWorkspaceRequest) (*longrunningpb.Operation, error) {
	if r == nil {
		return nil, invalidArgument("request is required")
	}
	id, e := parseWorkspaceID(r.GetId())
	if e != nil {
		return nil, invalidArgument("id must be a canonical non-zero UUID")
	}
	w, e := s.application.Undelete(ctx, command.UndeleteWorkspace{ID: id})
	if e != nil {
		return nil, mapError(e)
	}
	return s.operation(w, false)
}
func workspaceToProto(w *workspace.Workspace) (*resourcemanagerpb.Workspace, error) {
	if w == nil {
		return nil, status.Error(codes.Internal, "workspace is nil")
	}
	ct, e := timestampFromTime(w.CreateTime())
	if e != nil {
		return nil, e
	}
	ut, e := timestampFromTime(w.UpdateTime())
	if e != nil {
		return nil, e
	}
	dt, e := timestampFromOptionalTime(w.DeleteTime())
	if e != nil {
		return nil, e
	}
	pt, e := timestampFromOptionalTime(w.PurgeTime())
	if e != nil {
		return nil, e
	}
	return &resourcemanagerpb.Workspace{Id: w.ID().String(), OrganizationId: w.OrganizationID().String(), State: resourcemanagerpb.Workspace_State(w.State()), Name: w.Name(), Description: w.Description(), CreateTime: ct, UpdateTime: ut, DeleteTime: dt, PurgeTime: pt, Version: w.Version().Int64(), Labels: w.Labels()}, nil
}

var _ resourcemanagerpb.WorkspaceServiceServer = (*WorkspaceServer)(nil)
