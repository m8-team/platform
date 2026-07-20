package grpcadapter

import (
	"fmt"
	"strings"
	"time"

	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	commonpb "github.com/m8-team/go-genproto/m8/platform/common/operation/v1"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *OrganizationServer) completedOrganizationOperation(
	id organization.ID,
	response *resourcemanagerpb.OrganizationOperationResponse,
) (*longrunningpb.Operation, error) {
	return completedOperation(s.clock, s.operationIDs, newOrganizationResourceRef(id), response)
}

func (s *OrganizationServer) completedDeleteOperation(id organization.ID) (*longrunningpb.Operation, error) {
	resource := newOrganizationResourceRef(id)
	return completedOperation(s.clock, s.operationIDs, resource, &commonpb.OperationResponse{Resource: resource})
}

func (s *WorkspaceServer) completedWorkspaceOperation(
	prepared preparedOperation,
	value *workspace.Workspace,
) (*longrunningpb.Operation, error) {
	mapped, err := workspaceToProto(value)
	if err != nil {
		return nil, fmt.Errorf("map workspace operation response: %w", err)
	}
	return buildCompletedOperation(
		prepared,
		newWorkspaceResourceRef(value.ID()),
		&resourcemanagerpb.WorkspaceOperationResponse{Workspace: mapped},
	)
}

func (s *WorkspaceServer) completedWorkspaceDeleteOperation(
	prepared preparedOperation,
	id workspace.ID,
) (*longrunningpb.Operation, error) {
	resource := newWorkspaceResourceRef(id)
	return buildCompletedOperation(prepared, resource, &commonpb.OperationResponse{Resource: resource})
}

type preparedOperation struct {
	id  string
	now time.Time
}

func completedOperation(
	clock ports.Clock,
	operationIDs OperationIDGenerator,
	resource *commonpb.ResourceRef,
	response proto.Message,
) (*longrunningpb.Operation, error) {
	prepared, err := prepareCompletedOperation(clock, operationIDs)
	if err != nil {
		return nil, err
	}
	return buildCompletedOperation(prepared, resource, response)
}

func prepareCompletedOperation(clock ports.Clock, operationIDs OperationIDGenerator) (preparedOperation, error) {
	operationID, err := canonicalOperationID(operationIDs.NewOperationID())
	if err != nil {
		return preparedOperation{}, fmt.Errorf("invalid generated operation id: %w", err)
	}
	now := clock.Now().UTC()
	if _, err := timestampFromTime(now); err != nil {
		return preparedOperation{}, fmt.Errorf("operation timestamp: %w", err)
	}
	return preparedOperation{id: operationID, now: now}, nil
}

func buildCompletedOperation(
	prepared preparedOperation,
	resource *commonpb.ResourceRef,
	response proto.Message,
) (*longrunningpb.Operation, error) {
	timestamp, err := timestampFromTime(prepared.now)
	if err != nil {
		return nil, fmt.Errorf("operation timestamp: %w", err)
	}

	metadata, err := anypb.New(&commonpb.OperationMetadata{
		OperationId: prepared.id,
		Resource:    resource,
		State:       commonpb.OperationMetadata_SUCCEEDED,
		CreateTime:  timestamp,
		StartTime:   timestamppb.New(prepared.now),
		UpdateTime:  timestamppb.New(prepared.now),
		EndTime:     timestamppb.New(prepared.now),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal operation metadata: %w", err)
	}
	packedResponse, err := anypb.New(response)
	if err != nil {
		return nil, fmt.Errorf("marshal operation response: %w", err)
	}

	return &longrunningpb.Operation{
		Name:     "operations/" + prepared.id,
		Metadata: metadata,
		Done:     true,
		Result: &longrunningpb.Operation_Response{
			Response: packedResponse,
		},
	}, nil
}

func canonicalOperationID(raw string) (string, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil || parsed == uuid.Nil || !strings.EqualFold(parsed.String(), raw) {
		return "", fmt.Errorf("operation id must be a canonical non-zero UUID")
	}
	return parsed.String(), nil
}

func newOrganizationResourceRef(id organization.ID) *commonpb.ResourceRef {
	return &commonpb.ResourceRef{
		Type: organization.ResourceType,
		Id:   id.String(),
		Name: "organizations/" + id.String(),
	}
}

func newWorkspaceResourceRef(id workspace.ID) *commonpb.ResourceRef {
	return &commonpb.ResourceRef{
		Type: workspace.ResourceType,
		Id:   id.String(),
		Name: "workspaces/" + id.String(),
	}
}

func timestampFromTime(value time.Time) (*timestamppb.Timestamp, error) {
	if value.IsZero() {
		return nil, fmt.Errorf("timestamp is zero")
	}
	timestamp := timestamppb.New(value.UTC())
	if err := timestamp.CheckValid(); err != nil {
		return nil, err
	}
	return timestamp, nil
}

func timestampFromOptionalTime(value *time.Time) (*timestamppb.Timestamp, error) {
	if value == nil {
		return nil, nil
	}
	return timestampFromTime(*value)
}
