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
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const testWorkspaceID = "118f3f16-9950-7a48-9d12-9fb6d8f4c8f2"

func TestWorkspaceServerLifecycleAndCompletedOperations(t *testing.T) {
	server := newTestWorkspaceServer(t, authz.AllowAll(), &sequenceOperationIDGenerator{})
	ctx := context.Background()

	createdOperation, err := server.CreateWorkspace(ctx, &resourcemanagerpb.CreateWorkspaceRequest{
		OrganizationId: testOrganizationID,
		Workspace: &resourcemanagerpb.Workspace{
			Name:        "Production",
			Description: "Primary workspace",
			Labels:      map[string]string{"tier": "one"},
		},
	})
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	created := unpackWorkspaceOperation(t, createdOperation)
	if created.GetId() != testWorkspaceID || created.GetOrganizationId() != testOrganizationID {
		t.Fatalf("created workspace IDs = %s/%s", created.GetId(), created.GetOrganizationId())
	}
	if created.GetState() != resourcemanagerpb.Workspace_ACTIVE || created.GetVersion() != 1 {
		t.Fatalf("created workspace state/version = %s/%d", created.GetState(), created.GetVersion())
	}
	assertWorkspaceOperationMetadata(t, createdOperation, testWorkspaceID)

	updatedOperation, err := server.UpdateWorkspace(ctx, &resourcemanagerpb.UpdateWorkspaceRequest{
		Workspace:  &resourcemanagerpb.Workspace{Id: testWorkspaceID, Name: "Primary", Version: 1},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
	})
	if err != nil {
		t.Fatalf("UpdateWorkspace() error = %v", err)
	}
	updated := unpackWorkspaceOperation(t, updatedOperation)
	if updated.GetName() != "Primary" || updated.GetVersion() != 2 {
		t.Fatalf("updated workspace = %+v", updated)
	}

	deleteOperation, err := server.DeleteWorkspace(ctx, &resourcemanagerpb.DeleteWorkspaceRequest{Id: testWorkspaceID, Version: 2})
	if err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}
	deleteResponse := &commonpb.OperationResponse{}
	if err := anypb.UnmarshalTo(deleteOperation.GetResponse(), deleteResponse, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unpack delete response: %v", err)
	}
	if deleteResponse.GetResource().GetId() != testWorkspaceID {
		t.Fatalf("delete resource = %+v", deleteResponse.GetResource())
	}

	undeleteOperation, err := server.UndeleteWorkspace(ctx, &resourcemanagerpb.UndeleteWorkspaceRequest{Id: testWorkspaceID})
	if err != nil {
		t.Fatalf("UndeleteWorkspace() error = %v", err)
	}
	if restored := unpackWorkspaceOperation(t, undeleteOperation); restored.GetState() != resourcemanagerpb.Workspace_ACTIVE || restored.GetVersion() != 4 {
		t.Fatalf("restored workspace = %+v", restored)
	}
}

func TestWorkspaceServerRequestValidation(t *testing.T) {
	server := newTestWorkspaceServer(t, authz.AllowAll(), &sequenceOperationIDGenerator{})
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "nil get", run: func() error { _, err := server.GetWorkspace(context.Background(), nil); return err }},
		{name: "invalid get id", run: func() error {
			_, err := server.GetWorkspace(context.Background(), &resourcemanagerpb.GetWorkspaceRequest{Id: "invalid"})
			return err
		}},
		{name: "negative page size", run: func() error {
			_, err := server.ListWorkspaces(context.Background(), &resourcemanagerpb.ListWorkspacesRequest{OrganizationId: testOrganizationID, PageSize: -1})
			return err
		}},
		{name: "oversized page token", run: func() error {
			_, err := server.ListWorkspaces(context.Background(), &resourcemanagerpb.ListWorkspacesRequest{OrganizationId: testOrganizationID, PageToken: strings.Repeat("界", 1025)})
			return err
		}},
		{name: "oversized filter", run: func() error {
			_, err := server.ListWorkspaces(context.Background(), &resourcemanagerpb.ListWorkspacesRequest{OrganizationId: testOrganizationID, Filter: strings.Repeat("界", 1025)})
			return err
		}},
		{name: "oversized order", run: func() error {
			_, err := server.ListWorkspaces(context.Background(), &resourcemanagerpb.ListWorkspacesRequest{OrganizationId: testOrganizationID, OrderBy: strings.Repeat("界", 129)})
			return err
		}},
		{name: "negative update version", run: func() error {
			_, err := server.UpdateWorkspace(context.Background(), &resourcemanagerpb.UpdateWorkspaceRequest{Workspace: &resourcemanagerpb.Workspace{Id: testWorkspaceID, Version: -1}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}}})
			return err
		}},
		{name: "negative delete version", run: func() error {
			_, err := server.DeleteWorkspace(context.Background(), &resourcemanagerpb.DeleteWorkspaceRequest{Id: testWorkspaceID, Version: -1})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if code := status.Code(test.run()); code != codes.InvalidArgument {
				t.Fatalf("status code = %s, want %s", code, codes.InvalidArgument)
			}
		})
	}
}

func TestWorkspaceServerRejectsInvalidGeneratedOperationID(t *testing.T) {
	server := newTestWorkspaceServer(t, authz.AllowAll(), fixedOperationIDGenerator("invalid"))
	_, err := server.CreateWorkspace(context.Background(), &resourcemanagerpb.CreateWorkspaceRequest{
		OrganizationId: testOrganizationID,
		Workspace:      &resourcemanagerpb.Workspace{Name: "test"},
	})
	if code := status.Code(err); code != codes.Internal {
		t.Fatalf("status code = %s, want %s; error = %v", code, codes.Internal, err)
	}
	if _, err := server.GetWorkspace(context.Background(), &resourcemanagerpb.GetWorkspaceRequest{Id: testWorkspaceID}); status.Code(err) != codes.NotFound {
		t.Fatalf("workspace persisted after operation preparation failure: %v", err)
	}
}

func TestWorkspaceServerAllowMissingDeleteReturnsWorkspaceResource(t *testing.T) {
	server := newTestWorkspaceServer(t, authz.AllowAll(), &sequenceOperationIDGenerator{})
	operation, err := server.DeleteWorkspace(context.Background(), &resourcemanagerpb.DeleteWorkspaceRequest{
		Id:           testWorkspaceID,
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}
	response := &commonpb.OperationResponse{}
	if err := anypb.UnmarshalTo(operation.GetResponse(), response, proto.UnmarshalOptions{}); err != nil {
		t.Fatal(err)
	}
	resource := response.GetResource()
	if resource.GetType() != workspace.ResourceType || resource.GetId() != testWorkspaceID || resource.GetName() != "workspaces/"+testWorkspaceID {
		t.Fatalf("delete resource = %+v", resource)
	}
}

func newTestWorkspaceServer(t *testing.T, authorizer ports.Authorizer, operationIDs OperationIDGenerator) *WorkspaceServer {
	t.Helper()
	clock := &fixedClock{now: time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)}
	organizations := memory.NewOrganizationRepository()
	parent, err := organization.New(organization.CreateParams{ID: organization.MustParseID(testOrganizationID), Now: clock.now})
	if err != nil {
		t.Fatal(err)
	}
	if err := organizations.Create(context.Background(), parent); err != nil {
		t.Fatal(err)
	}
	application, err := usecase.NewWorkspaceService(
		memory.NewWorkspaceRepository(),
		organizations,
		authorizer,
		clock,
		fixedWorkspaceIDGenerator{id: workspace.MustParseID(testWorkspaceID)},
		usecase.WorkspaceServiceConfig{SoftDeleteRetention: 24 * time.Hour, PageTokenKey: []byte("01234567890123456789012345678901")},
	)
	if err != nil {
		t.Fatalf("NewWorkspaceService() error = %v", err)
	}
	server, err := NewWorkspaceServer(application, clock, operationIDs)
	if err != nil {
		t.Fatalf("NewWorkspaceServer() error = %v", err)
	}
	return server
}

func unpackWorkspaceOperation(t *testing.T, operation interface{ GetResponse() *anypb.Any }) *resourcemanagerpb.Workspace {
	t.Helper()
	response := &resourcemanagerpb.WorkspaceOperationResponse{}
	if err := anypb.UnmarshalTo(operation.GetResponse(), response, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unpack workspace operation: %v", err)
	}
	return response.GetWorkspace()
}

func assertWorkspaceOperationMetadata(t *testing.T, operation interface{ GetMetadata() *anypb.Any }, workspaceID string) {
	t.Helper()
	metadata := &commonpb.OperationMetadata{}
	if err := anypb.UnmarshalTo(operation.GetMetadata(), metadata, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("unpack operation metadata: %v", err)
	}
	resource := metadata.GetResource()
	if resource.GetType() != workspace.ResourceType || resource.GetId() != workspaceID || resource.GetName() != "workspaces/"+workspaceID {
		t.Fatalf("operation resource = %+v", resource)
	}
}

type fixedWorkspaceIDGenerator struct{ id workspace.ID }

func (g fixedWorkspaceIDGenerator) NewWorkspaceID() workspace.ID { return g.id }

type fixedOperationIDGenerator string

func (g fixedOperationIDGenerator) NewOperationID() string { return string(g) }
