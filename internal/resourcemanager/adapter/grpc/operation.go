package grpcadapter

import (
	"fmt"
	"time"

	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	commonpb "github.com/m8-team/go-genproto/m8/platform/common/operation/v1"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *OrganizationServer) completedOrganizationOperation(
	id organization.ID,
	response *resourcemanagerpb.OrganizationOperationResponse,
) (*longrunningpb.Operation, error) {
	return s.completedOperation(id, response)
}

func (s *OrganizationServer) completedDeleteOperation(id organization.ID) (*longrunningpb.Operation, error) {
	return s.completedOperation(id, &commonpb.OperationResponse{Resource: newResourceRef(id)})
}

func (s *OrganizationServer) completedOperation(
	id organization.ID,
	response proto.Message,
) (*longrunningpb.Operation, error) {
	operationID := s.operationIDs.NewOperationID()
	parsedOperationID, err := parseCanonicalOrganizationID(operationID)
	if err != nil {
		return nil, fmt.Errorf("invalid generated operation id: %w", err)
	}
	operationID = parsedOperationID.String()
	now := s.clock.Now().UTC()
	timestamp, err := timestampFromTime(now)
	if err != nil {
		return nil, fmt.Errorf("operation timestamp: %w", err)
	}

	metadata, err := anypb.New(&commonpb.OperationMetadata{
		OperationId: operationID,
		Resource:    newResourceRef(id),
		State:       commonpb.OperationMetadata_SUCCEEDED,
		CreateTime:  timestamp,
		StartTime:   timestamppb.New(now),
		UpdateTime:  timestamppb.New(now),
		EndTime:     timestamppb.New(now),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal operation metadata: %w", err)
	}
	packedResponse, err := anypb.New(response)
	if err != nil {
		return nil, fmt.Errorf("marshal operation response: %w", err)
	}

	return &longrunningpb.Operation{
		Name:     "operations/" + operationID,
		Metadata: metadata,
		Done:     true,
		Result: &longrunningpb.Operation_Response{
			Response: packedResponse,
		},
	}, nil
}

func newResourceRef(id organization.ID) *commonpb.ResourceRef {
	return &commonpb.ResourceRef{
		Type: organization.ResourceType,
		Id:   id.String(),
		Name: "organizations/" + id.String(),
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
