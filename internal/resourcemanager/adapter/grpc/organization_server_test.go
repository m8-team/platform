package grpcadapter

import (
	"context"
	"strings"
	"testing"
	"time"

	commonpb "github.com/m8-team/go-genproto/m8/platform/common/operation/v1"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/authz"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testOrganizationID = "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"
	testOperationID    = "028f3f16-9950-7a48-9d12-9fb6d8f4c8f2"
)

func TestOrganizationServerLifecycleAndCompletedOperations(t *testing.T) {
	server := newTestOrganizationServer(t, authz.AllowAll(), &stubWorkspaceChildren{})
	ctx := context.Background()

	createdOperation, err := server.CreateOrganization(ctx, &resourcemanagerpb.CreateOrganizationRequest{
		Organization: &resourcemanagerpb.Organization{
			Name:        "Acme",
			Description: "first",
			Labels:      map[string]string{"tier": "one"},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrganization() error = %v", err)
	}
	created := unpackOrganizationOperation(t, createdOperation)
	if created.GetId() != testOrganizationID || created.GetState() != resourcemanagerpb.Organization_ACTIVE {
		t.Fatalf("created organization = %+v", created)
	}
	if created.GetVersion() != 1 || created.GetCreateTime() == nil || created.GetUpdateTime() == nil {
		t.Fatalf("created output fields = %+v", created)
	}
	assertCompletedMetadata(t, createdOperation, testOrganizationID)

	got, err := server.GetOrganization(ctx, &resourcemanagerpb.GetOrganizationRequest{
		OrganizationId: testOrganizationID,
	})
	if err != nil {
		t.Fatalf("GetOrganization() error = %v", err)
	}
	if got.GetName() != "Acme" || got.GetDescription() != "first" {
		t.Fatalf("GetOrganization() = %+v", got)
	}

	updatedOperation, err := server.UpdateOrganization(ctx, &resourcemanagerpb.UpdateOrganizationRequest{
		Organization: &resourcemanagerpb.Organization{
			Id:          testOrganizationID,
			Name:        "Acme 2",
			Description: "ignored",
			Version:     1,
			State:       resourcemanagerpb.Organization_FAILED,
			CreateTime:  timestamppb.New(time.Unix(1, 0)),
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "labels", "name"}},
	})
	if err != nil {
		t.Fatalf("UpdateOrganization() error = %v", err)
	}
	updated := unpackOrganizationOperation(t, updatedOperation)
	if updated.GetName() != "Acme 2" || updated.GetDescription() != "first" {
		t.Fatalf("updated organization = %+v", updated)
	}
	if len(updated.GetLabels()) != 0 {
		t.Fatalf("updated labels = %#v, want cleared map", updated.GetLabels())
	}
	if updated.GetState() != resourcemanagerpb.Organization_ACTIVE || updated.GetVersion() != 2 {
		t.Fatalf("updated state/version = %s/%d", updated.GetState(), updated.GetVersion())
	}

	deleteOperation, err := server.DeleteOrganization(ctx, &resourcemanagerpb.DeleteOrganizationRequest{
		OrganizationId: testOrganizationID,
		Version:        2,
	})
	if err != nil {
		t.Fatalf("DeleteOrganization() error = %v", err)
	}
	if !deleteOperation.GetDone() {
		t.Fatal("delete operation is not done")
	}
	deleteResponse := &commonpb.OperationResponse{}
	if got, want := deleteOperation.GetResponse().GetTypeUrl(), "type.googleapis.com/m8.platform.common.operation.v1.OperationResponse"; got != want {
		t.Fatalf("delete response type URL = %q, want %q", got, want)
	}
	if err := anypb.UnmarshalTo(deleteOperation.GetResponse(), deleteResponse, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unpack delete response: %v", err)
	}
	if deleteResponse.GetResource().GetId() != testOrganizationID {
		t.Fatalf("delete resource = %+v", deleteResponse.GetResource())
	}

	deleted, err := server.GetOrganization(ctx, &resourcemanagerpb.GetOrganizationRequest{
		OrganizationId: testOrganizationID,
	})
	if err != nil {
		t.Fatalf("GetOrganization(deleted) error = %v", err)
	}
	if deleted.GetState() != resourcemanagerpb.Organization_DELETED || deleted.GetDeleteTime() == nil || deleted.GetPurgeTime() == nil {
		t.Fatalf("deleted organization = %+v", deleted)
	}
	list, err := server.ListOrganizations(ctx, &resourcemanagerpb.ListOrganizationsRequest{})
	if err != nil {
		t.Fatalf("ListOrganizations() error = %v", err)
	}
	if len(list.GetOrganizations()) != 0 || list.GetTotalSize() != 0 {
		t.Fatalf("default list = %+v, want deleted excluded", list)
	}
	list, err = server.ListOrganizations(ctx, &resourcemanagerpb.ListOrganizationsRequest{ShowDeleted: true})
	if err != nil || len(list.GetOrganizations()) != 1 {
		t.Fatalf("ListOrganizations(show_deleted) = %+v, %v", list, err)
	}

	undeleteOperation, err := server.UndeleteOrganization(ctx, &resourcemanagerpb.UndeleteOrganizationRequest{
		OrganizationId: testOrganizationID,
	})
	if err != nil {
		t.Fatalf("UndeleteOrganization() error = %v", err)
	}
	restored := unpackOrganizationOperation(t, undeleteOperation)
	if restored.GetState() != resourcemanagerpb.Organization_ACTIVE || restored.GetVersion() != 4 {
		t.Fatalf("restored state/version = %s/%d", restored.GetState(), restored.GetVersion())
	}
	if restored.GetDeleteTime() != nil || restored.GetPurgeTime() != nil {
		t.Fatalf("restored deletion timestamps = %v/%v", restored.GetDeleteTime(), restored.GetPurgeTime())
	}
}

func TestOrganizationServerRequestValidation(t *testing.T) {
	server := newTestOrganizationServer(t, authz.AllowAll(), &stubWorkspaceChildren{})
	validOrganization := func() *resourcemanagerpb.Organization {
		return &resourcemanagerpb.Organization{Id: testOrganizationID}
	}

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "nil get", run: func() error { _, err := server.GetOrganization(context.Background(), nil); return err }},
		{name: "noncanonical uuid", run: func() error {
			_, err := server.GetOrganization(context.Background(), &resourcemanagerpb.GetOrganizationRequest{OrganizationId: strings.ReplaceAll(testOrganizationID, "-", "")})
			return err
		}},
		{name: "nil create payload", run: func() error {
			_, err := server.CreateOrganization(context.Background(), &resourcemanagerpb.CreateOrganizationRequest{})
			return err
		}},
		{name: "create id", run: func() error {
			_, err := server.CreateOrganization(context.Background(), &resourcemanagerpb.CreateOrganizationRequest{Organization: validOrganization()})
			return err
		}},
		{name: "create state", run: func() error {
			_, err := server.CreateOrganization(context.Background(), &resourcemanagerpb.CreateOrganizationRequest{Organization: &resourcemanagerpb.Organization{State: resourcemanagerpb.Organization_ACTIVE}})
			return err
		}},
		{name: "create timestamp", run: func() error {
			_, err := server.CreateOrganization(context.Background(), &resourcemanagerpb.CreateOrganizationRequest{Organization: &resourcemanagerpb.Organization{CreateTime: &timestamppb.Timestamp{}}})
			return err
		}},
		{name: "create version", run: func() error {
			_, err := server.CreateOrganization(context.Background(), &resourcemanagerpb.CreateOrganizationRequest{Organization: &resourcemanagerpb.Organization{Version: 1}})
			return err
		}},
		{name: "name unicode rune limit", run: func() error {
			_, err := server.CreateOrganization(context.Background(), &resourcemanagerpb.CreateOrganizationRequest{Organization: &resourcemanagerpb.Organization{Name: strings.Repeat("界", 257)}})
			return err
		}},
		{name: "nil update mask", run: func() error {
			_, err := server.UpdateOrganization(context.Background(), &resourcemanagerpb.UpdateOrganizationRequest{Organization: validOrganization()})
			return err
		}},
		{name: "immutable update path", run: func() error {
			_, err := server.UpdateOrganization(context.Background(), &resourcemanagerpb.UpdateOrganizationRequest{Organization: validOrganization(), UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"state"}}})
			return err
		}},
		{name: "map subpath", run: func() error {
			_, err := server.UpdateOrganization(context.Background(), &resourcemanagerpb.UpdateOrganizationRequest{Organization: validOrganization(), UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"labels.foo"}}})
			return err
		}},
		{name: "wildcard mixed", run: func() error {
			_, err := server.UpdateOrganization(context.Background(), &resourcemanagerpb.UpdateOrganizationRequest{Organization: validOrganization(), UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*", "name"}}})
			return err
		}},
		{name: "negative delete version", run: func() error {
			_, err := server.DeleteOrganization(context.Background(), &resourcemanagerpb.DeleteOrganizationRequest{OrganizationId: testOrganizationID, Version: -1})
			return err
		}},
		{name: "negative page size", run: func() error {
			_, err := server.ListOrganizations(context.Background(), &resourcemanagerpb.ListOrganizationsRequest{PageSize: -1})
			return err
		}},
		{name: "oversized filter", run: func() error {
			_, err := server.ListOrganizations(context.Background(), &resourcemanagerpb.ListOrganizationsRequest{Filter: strings.Repeat("界", 1025)})
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if code := status.Code(test.run()); code != codes.InvalidArgument {
				t.Fatalf("code = %s, want %s", code, codes.InvalidArgument)
			}
		})
	}

	paths, err := mutableUpdatePaths(&fieldmaskpb.FieldMask{Paths: []string{"*"}})
	if err != nil || !paths["name"] || !paths["description"] || !paths["labels"] {
		t.Fatalf("mutableUpdatePaths(*) = %#v, %v", paths, err)
	}
}

func TestOrganizationServerDeleteAllowMissingAndStatusMapping(t *testing.T) {
	server := newTestOrganizationServer(t, authz.AllowAll(), &stubWorkspaceChildren{})

	operation, err := server.DeleteOrganization(context.Background(), &resourcemanagerpb.DeleteOrganizationRequest{
		OrganizationId: testOrganizationID,
		AllowMissing:   true,
	})
	if err != nil || operation == nil || !operation.GetDone() {
		t.Fatalf("DeleteOrganization(allow_missing) = %+v, %v", operation, err)
	}
	_, err = server.DeleteOrganization(context.Background(), &resourcemanagerpb.DeleteOrganizationRequest{
		OrganizationId: testOrganizationID,
	})
	if code := status.Code(err); code != codes.NotFound {
		t.Fatalf("DeleteOrganization(missing) code = %s, want %s", code, codes.NotFound)
	}

	mappings := []struct {
		err  error
		code codes.Code
	}{
		{err: ports.ErrUnauthenticated, code: codes.Unauthenticated},
		{err: ports.ErrPermissionDenied, code: codes.PermissionDenied},
		{err: ports.ErrOrganizationNotFound, code: codes.NotFound},
		{err: ports.ErrOrganizationAlreadyExists, code: codes.AlreadyExists},
		{err: organization.ErrVersionMismatch, code: codes.Aborted},
		{err: ports.ErrOrganizationRepositoryUnavailable, code: codes.Unavailable},
		{err: usecase.ErrOrganizationHasWorkspaces, code: codes.FailedPrecondition},
		{err: usecase.ErrInvalidOrganizationFilter, code: codes.InvalidArgument},
		{err: usecase.ErrGeneratedOrganizationID, code: codes.Internal},
		{err: context.Canceled, code: codes.Canceled},
		{err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
	}
	for _, mapping := range mappings {
		if got := status.Code(mapError(mapping.err)); got != mapping.code {
			t.Errorf("mapError(%v) = %s, want %s", mapping.err, got, mapping.code)
		}
	}
}

func unpackOrganizationOperation(
	t *testing.T,
	operation interface {
		GetDone() bool
		GetResponse() *anypb.Any
	},
) *resourcemanagerpb.Organization {
	t.Helper()
	if !operation.GetDone() {
		t.Fatal("operation is not completed")
	}
	response := &resourcemanagerpb.OrganizationOperationResponse{}
	if got, want := operation.GetResponse().GetTypeUrl(), "type.googleapis.com/m8.platform.resourcemanager.v1.OrganizationOperationResponse"; got != want {
		t.Fatalf("organization response type URL = %q, want %q", got, want)
	}
	if err := anypb.UnmarshalTo(operation.GetResponse(), response, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unpack organization operation: %v", err)
	}
	if response.GetOrganization() == nil {
		t.Fatal("operation organization is nil")
	}
	return response.GetOrganization()
}

func assertCompletedMetadata(t *testing.T, operation interface {
	GetName() string
	GetDone() bool
	GetMetadata() *anypb.Any
}, organizationID string) {
	t.Helper()
	if !operation.GetDone() || !strings.HasPrefix(operation.GetName(), "operations/") {
		t.Fatalf("operation name/done = %q/%v", operation.GetName(), operation.GetDone())
	}
	metadata := &commonpb.OperationMetadata{}
	if got, want := operation.GetMetadata().GetTypeUrl(), "type.googleapis.com/m8.platform.common.operation.v1.OperationMetadata"; got != want {
		t.Fatalf("metadata type URL = %q, want %q", got, want)
	}
	if err := anypb.UnmarshalTo(operation.GetMetadata(), metadata, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unpack metadata: %v", err)
	}
	if operation.GetName() != "operations/"+metadata.GetOperationId() {
		t.Fatalf("operation name = %q, metadata id = %q", operation.GetName(), metadata.GetOperationId())
	}
	if metadata.GetState() != commonpb.OperationMetadata_SUCCEEDED {
		t.Fatalf("metadata state = %s", metadata.GetState())
	}
	resource := metadata.GetResource()
	if resource.GetType() != organization.ResourceType || resource.GetId() != organizationID ||
		resource.GetName() != "organizations/"+organizationID {
		t.Fatalf("metadata resource = %+v", resource)
	}
	for name, timestamp := range map[string]*timestamppb.Timestamp{
		"create": metadata.GetCreateTime(),
		"start":  metadata.GetStartTime(),
		"update": metadata.GetUpdateTime(),
		"end":    metadata.GetEndTime(),
	} {
		if timestamp == nil || timestamp.CheckValid() != nil {
			t.Errorf("%s timestamp = %v", name, timestamp)
		}
	}
}

func newTestOrganizationServer(
	t *testing.T,
	authorizer ports.Authorizer,
	children ports.WorkspaceChildren,
) *OrganizationServer {
	t.Helper()
	clock := &fixedClock{now: time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)}
	application, err := usecase.NewOrganizationService(
		memory.NewOrganizationRepository(),
		authorizer,
		clock,
		fixedOrganizationIDGenerator{id: organization.MustParseID(testOrganizationID)},
		children,
		usecase.OrganizationServiceConfig{
			SoftDeleteRetention: 24 * time.Hour,
			PageTokenKey:        []byte("01234567890123456789012345678901"),
		},
	)
	if err != nil {
		t.Fatalf("NewOrganizationService() error = %v", err)
	}
	server, err := NewOrganizationServer(application, clock, &sequenceOperationIDGenerator{})
	if err != nil {
		t.Fatalf("NewOrganizationServer() error = %v", err)
	}
	return server
}

type fixedClock struct {
	now time.Time
}

func (c *fixedClock) Now() time.Time { return c.now }

type fixedOrganizationIDGenerator struct {
	id organization.ID
}

func (g fixedOrganizationIDGenerator) NewID() organization.ID { return g.id }

type sequenceOperationIDGenerator struct{}

func (g *sequenceOperationIDGenerator) NewOperationID() string {
	return testOperationID
}

type stubWorkspaceChildren struct {
	hasChildren bool
	err         error
}

func (s *stubWorkspaceChildren) HasNonDeleted(context.Context, organization.ID) (bool, error) {
	return s.hasChildren, s.err
}
