package grpcadapter

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/app/command"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	maxPageTokenRunes = 1024
	maxFilterRunes    = 1024
	maxOrderByRunes   = 128
)

type OperationIDGenerator interface {
	NewOperationID() string
}

type OrganizationServer struct {
	resourcemanagerpb.UnimplementedOrganizationServiceServer

	application  *usecase.OrganizationService
	clock        ports.Clock
	operationIDs OperationIDGenerator
}

func NewOrganizationServer(
	application *usecase.OrganizationService,
	clock ports.Clock,
	operationIDs OperationIDGenerator,
) (*OrganizationServer, error) {
	if application == nil {
		return nil, errors.New("organization application service is required")
	}
	if clock == nil {
		return nil, errors.New("organization gRPC clock is required")
	}
	if operationIDs == nil {
		return nil, errors.New("operation id generator is required")
	}

	return &OrganizationServer{
		application:  application,
		clock:        clock,
		operationIDs: operationIDs,
	}, nil
}

func (s *OrganizationServer) GetOrganization(
	ctx context.Context,
	request *resourcemanagerpb.GetOrganizationRequest,
) (*resourcemanagerpb.Organization, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	id, err := parseCanonicalOrganizationID(request.GetOrganizationId())
	if err != nil {
		return nil, invalidArgument("organization_id must be a canonical non-zero UUID")
	}

	value, err := s.application.Get(ctx, query.GetOrganization{ID: id})
	if err != nil {
		return nil, mapError(err)
	}
	response, err := organizationToProto(value)
	if err != nil {
		return nil, status.Error(codes.Internal, "map organization response")
	}
	return response, nil
}

func (s *OrganizationServer) ListOrganizations(
	ctx context.Context,
	request *resourcemanagerpb.ListOrganizationsRequest,
) (*resourcemanagerpb.ListOrganizationsResponse, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	if request.GetPageSize() < 0 || request.GetPageSize() > ports.MaxOrganizationPageSize {
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

	result, err := s.application.List(ctx, query.ListOrganizations{
		PageSize:    int(request.GetPageSize()),
		PageToken:   request.GetPageToken(),
		Filter:      request.GetFilter(),
		OrderBy:     request.GetOrderBy(),
		ShowDeleted: request.GetShowDeleted(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	if result.TotalSize < 0 || result.TotalSize > math.MaxInt32 {
		return nil, status.Error(codes.Internal, "organization total size exceeds API range")
	}

	organizations := make([]*resourcemanagerpb.Organization, 0, len(result.Organizations))
	for _, value := range result.Organizations {
		mapped, err := organizationToProto(value)
		if err != nil {
			return nil, status.Error(codes.Internal, "map organization list response")
		}
		organizations = append(organizations, mapped)
	}

	return &resourcemanagerpb.ListOrganizationsResponse{
		Organizations: organizations,
		NextPageToken: result.NextPageToken,
		TotalSize:     int32(result.TotalSize),
	}, nil
}

func (s *OrganizationServer) CreateOrganization(
	ctx context.Context,
	request *resourcemanagerpb.CreateOrganizationRequest,
) (*longrunningpb.Operation, error) {
	input, err := validateCreateRequest(request)
	if err != nil {
		return nil, err
	}

	value, err := s.application.Create(ctx, command.CreateOrganization{
		Name:        input.GetName(),
		Description: input.GetDescription(),
		Labels:      input.GetLabels(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	response, err := organizationToProto(value)
	if err != nil {
		return nil, status.Error(codes.Internal, "map created organization")
	}
	operation, err := s.completedOrganizationOperation(
		value.ID(),
		&resourcemanagerpb.OrganizationOperationResponse{Organization: response},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed organization operation")
	}
	return operation, nil
}

func (s *OrganizationServer) UpdateOrganization(
	ctx context.Context,
	request *resourcemanagerpb.UpdateOrganizationRequest,
) (*longrunningpb.Operation, error) {
	cmd, err := updateCommand(request)
	if err != nil {
		return nil, err
	}

	value, err := s.application.Update(ctx, cmd)
	if err != nil {
		return nil, mapError(err)
	}
	response, err := organizationToProto(value)
	if err != nil {
		return nil, status.Error(codes.Internal, "map updated organization")
	}
	operation, err := s.completedOrganizationOperation(
		value.ID(),
		&resourcemanagerpb.OrganizationOperationResponse{Organization: response},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed organization operation")
	}
	return operation, nil
}

func (s *OrganizationServer) DeleteOrganization(
	ctx context.Context,
	request *resourcemanagerpb.DeleteOrganizationRequest,
) (*longrunningpb.Operation, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	id, err := parseCanonicalOrganizationID(request.GetOrganizationId())
	if err != nil {
		return nil, invalidArgument("organization_id must be a canonical non-zero UUID")
	}
	if request.GetVersion() < 0 {
		return nil, invalidArgument("version must be non-negative")
	}

	_, err = s.application.Delete(ctx, command.DeleteOrganization{
		ID:              id,
		ExpectedVersion: types.Version(request.GetVersion()),
		AllowMissing:    request.GetAllowMissing(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	operation, err := s.completedDeleteOperation(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed organization delete operation")
	}
	return operation, nil
}

func (s *OrganizationServer) UndeleteOrganization(
	ctx context.Context,
	request *resourcemanagerpb.UndeleteOrganizationRequest,
) (*longrunningpb.Operation, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	id, err := parseCanonicalOrganizationID(request.GetOrganizationId())
	if err != nil {
		return nil, invalidArgument("organization_id must be a canonical non-zero UUID")
	}

	value, err := s.application.Undelete(ctx, command.UndeleteOrganization{ID: id})
	if err != nil {
		return nil, mapError(err)
	}
	response, err := organizationToProto(value)
	if err != nil {
		return nil, status.Error(codes.Internal, "map undeleted organization")
	}
	operation, err := s.completedOrganizationOperation(
		value.ID(),
		&resourcemanagerpb.OrganizationOperationResponse{Organization: response},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "build completed organization operation")
	}
	return operation, nil
}

func validateCreateRequest(
	request *resourcemanagerpb.CreateOrganizationRequest,
) (*resourcemanagerpb.Organization, error) {
	if request == nil {
		return nil, invalidArgument("request is required")
	}
	value := request.GetOrganization()
	if value == nil {
		return nil, invalidArgument("organization is required")
	}
	if value.GetId() != "" || value.GetState() != resourcemanagerpb.Organization_STATE_UNSPECIFIED ||
		value.GetCreateTime() != nil || value.GetUpdateTime() != nil || value.GetDeleteTime() != nil ||
		value.GetPurgeTime() != nil || value.GetVersion() != 0 {
		return nil, invalidArgument("organization contains server-assigned fields")
	}
	return value, nil
}

func updateCommand(request *resourcemanagerpb.UpdateOrganizationRequest) (command.UpdateOrganization, error) {
	if request == nil {
		return command.UpdateOrganization{}, invalidArgument("request is required")
	}
	value := request.GetOrganization()
	if value == nil {
		return command.UpdateOrganization{}, invalidArgument("organization is required")
	}
	id, err := parseCanonicalOrganizationID(value.GetId())
	if err != nil {
		return command.UpdateOrganization{}, invalidArgument("organization.id must be a canonical non-zero UUID")
	}
	if value.GetVersion() < 0 {
		return command.UpdateOrganization{}, invalidArgument("organization.version must be non-negative")
	}
	paths, err := mutableUpdatePaths(request.GetUpdateMask())
	if err != nil {
		return command.UpdateOrganization{}, err
	}

	cmd := command.UpdateOrganization{
		ID:              id,
		ExpectedVersion: types.Version(value.GetVersion()),
	}
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

func mutableUpdatePaths(mask *fieldmaskpb.FieldMask) (map[string]bool, error) {
	if mask == nil || len(mask.GetPaths()) == 0 {
		return nil, invalidArgument("update_mask is required and must not be empty")
	}
	paths := mask.GetPaths()
	if len(paths) == 1 && paths[0] == "*" {
		return map[string]bool{"name": true, "description": true, "labels": true}, nil
	}

	result := make(map[string]bool, len(paths))
	for _, path := range paths {
		switch path {
		case "name", "description", "labels":
			result[path] = true
		default:
			return nil, invalidArgument(fmt.Sprintf("update_mask contains unsupported path %q", path))
		}
	}
	return result, nil
}

func parseCanonicalOrganizationID(raw string) (organization.ID, error) {
	if len(raw) != 36 {
		return organization.ID{}, organization.ErrInvalidOrganizationID
	}
	id, err := organization.ParseID(raw)
	if err != nil || !strings.EqualFold(id.String(), raw) {
		return organization.ID{}, organization.ErrInvalidOrganizationID
	}
	return id, nil
}

func organizationToProto(value *organization.Organization) (*resourcemanagerpb.Organization, error) {
	if value == nil {
		return nil, errors.New("organization is nil")
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

	return &resourcemanagerpb.Organization{
		Id:          value.ID().String(),
		State:       stateToProto(value.State()),
		Name:        value.Name(),
		Description: value.Description(),
		CreateTime:  createTime,
		UpdateTime:  updateTime,
		DeleteTime:  deleteTime,
		PurgeTime:   purgeTime,
		Version:     value.Version().Int64(),
		Labels:      value.Labels(),
	}, nil
}

func stateToProto(state organization.State) resourcemanagerpb.Organization_State {
	switch state {
	case organization.StateCreating:
		return resourcemanagerpb.Organization_CREATING
	case organization.StateActive:
		return resourcemanagerpb.Organization_ACTIVE
	case organization.StateSuspended:
		return resourcemanagerpb.Organization_SUSPENDED
	case organization.StateDeleting:
		return resourcemanagerpb.Organization_DELETING
	case organization.StateDeleted:
		return resourcemanagerpb.Organization_DELETED
	case organization.StateFailed:
		return resourcemanagerpb.Organization_FAILED
	default:
		return resourcemanagerpb.Organization_STATE_UNSPECIFIED
	}
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

var _ resourcemanagerpb.OrganizationServiceServer = (*OrganizationServer)(nil)
