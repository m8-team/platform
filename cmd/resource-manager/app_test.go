package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	resourcemanagerpb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestNewAppBuilds(t *testing.T) {
	app := NewApp(Config{
		HTTP:       HTTPConfig{Address: "127.0.0.1:0"},
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:       grpcserver.Config{Address: "127.0.0.1:0"},
	})
	if err := app.Err(); err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
}

func TestApplicationStartsAndJoinsServersOnShutdown(t *testing.T) {
	app := NewApp(Config{
		HTTP:       HTTPConfig{Address: "127.0.0.1:0"},
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:       grpcserver.Config{Address: "127.0.0.1:0"},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestAppRegistersOrganizationServiceBeforeStart(t *testing.T) {
	cfg := Config{
		HTTP:       HTTPConfig{Address: "127.0.0.1:0"},
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:       grpcserver.Config{Address: "127.0.0.1:0"},
	}
	var server *grpc.Server
	options := append(appOptions(cfg), fx.Populate(&server))
	app := fx.New(options...)
	if err := app.Err(); err != nil {
		t.Fatalf("build app: %v", err)
	}

	serviceName := resourcemanagerpb.OrganizationService_ServiceDesc.ServiceName
	if _, exists := server.GetServiceInfo()[serviceName]; !exists {
		t.Fatalf("gRPC service %q is not registered", serviceName)
	}
	workspaceServiceName := resourcemanagerpb.WorkspaceService_ServiceDesc.ServiceName
	if _, exists := server.GetServiceInfo()[workspaceServiceName]; !exists {
		t.Fatalf("gRPC service %q is not registered", workspaceServiceName)
	}
}

func TestOrganizationRESTGatewayCreatesOrganization(t *testing.T) {
	handler := buildHTTPHandler(t, true)
	request := httptest.NewRequest(
		http.MethodPost,
		"/resource-manager/v1/organizations",
		strings.NewReader(`{
  "name": "postman-test",
  "description": "Created through the REST gateway",
  "labels": {"m8.io/source": "test"}
}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("POST organization status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body)
	}

	operation := &longrunningpb.Operation{}
	if err := protojson.Unmarshal(response.Body.Bytes(), operation); err != nil {
		t.Fatalf("decode operation response: %v; body = %s", err, response.Body)
	}
	if !operation.GetDone() {
		t.Fatal("operation done = false, want true")
	}

	result := &resourcemanagerpb.OrganizationOperationResponse{}
	if err := operation.GetResponse().UnmarshalTo(result); err != nil {
		t.Fatalf("decode organization operation response: %v", err)
	}
	if result.GetOrganization().GetId() == "" {
		t.Fatal("created organization id is empty")
	}
	if result.GetOrganization().GetName() != "postman-test" {
		t.Fatalf("created organization name = %q, want postman-test", result.GetOrganization().GetName())
	}
}

func TestOrganizationRESTGatewayDeniesMutationsByDefault(t *testing.T) {
	handler := buildHTTPHandler(t, false)
	request := httptest.NewRequest(
		http.MethodPost,
		"/resource-manager/v1/organizations",
		strings.NewReader(`{"name":"denied"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("POST organization status = %d, want %d; body = %s", response.Code, http.StatusForbidden, response.Body)
	}
}

func TestWorkspaceRESTGatewayCreatesAndListsWorkspace(t *testing.T) {
	handler := buildHTTPHandler(t, true)
	organizationID := createOrganizationThroughREST(t, handler)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/resource-manager/v1/workspaces?organizationId="+organizationID,
		strings.NewReader(`{"name":"production","labels":{"environment":"prod"}}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("POST workspace status = %d, want %d; body = %s", createResponse.Code, http.StatusOK, createResponse.Body)
	}
	operation := &longrunningpb.Operation{}
	if err := protojson.Unmarshal(createResponse.Body.Bytes(), operation); err != nil {
		t.Fatalf("decode workspace operation: %v", err)
	}
	created := &resourcemanagerpb.WorkspaceOperationResponse{}
	if err := operation.GetResponse().UnmarshalTo(created); err != nil {
		t.Fatalf("decode workspace operation response: %v", err)
	}
	if created.GetWorkspace().GetOrganizationId() != organizationID || created.GetWorkspace().GetName() != "production" {
		t.Fatalf("created workspace = %+v", created.GetWorkspace())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/resource-manager/v1/workspaces?organizationId="+organizationID, nil)
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("GET workspaces status = %d, want %d; body = %s", listResponse.Code, http.StatusOK, listResponse.Body)
	}
	list := &resourcemanagerpb.ListWorkspacesResponse{}
	if err := protojson.Unmarshal(listResponse.Body.Bytes(), list); err != nil {
		t.Fatalf("decode workspace list: %v", err)
	}
	if len(list.GetWorkspaces()) != 1 || list.GetWorkspaces()[0].GetId() != created.GetWorkspace().GetId() {
		t.Fatalf("workspace list = %+v", list.GetWorkspaces())
	}
}

func createOrganizationThroughREST(t *testing.T, handler http.Handler) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/resource-manager/v1/organizations", strings.NewReader(`{"name":"parent"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("POST parent organization status = %d; body = %s", response.Code, response.Body)
	}
	operation := &longrunningpb.Operation{}
	if err := protojson.Unmarshal(response.Body.Bytes(), operation); err != nil {
		t.Fatal(err)
	}
	created := &resourcemanagerpb.OrganizationOperationResponse{}
	if err := operation.GetResponse().UnmarshalTo(created); err != nil {
		t.Fatal(err)
	}
	return created.GetOrganization().GetId()
}

func TestResourceManagerHTTPHandlerDoesNotExposeHealthRoutes(t *testing.T) {
	handler := buildHTTPHandler(t, false)
	request := httptest.NewRequest(http.MethodGet, "/livez", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /livez status = %d, want %d; body = %s", response.Code, http.StatusNotFound, response.Body)
	}
}

func TestHealthHTTPHandlerKeepsHealthRoutes(t *testing.T) {
	handler := buildHealthHTTPHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/livez", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /livez status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body)
	}
}

func buildHTTPHandler(t *testing.T, allowUnauthenticated bool) http.Handler {
	t.Helper()

	cfg := Config{
		HTTP:                 HTTPConfig{Address: "127.0.0.1:0"},
		HealthHTTP:           HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:                 grpcserver.Config{Address: "127.0.0.1:0"},
		AllowUnauthenticated: allowUnauthenticated,
	}
	var wrapped resourceManagerHTTPHandler
	options := append(appOptions(cfg), fx.Populate(&wrapped))
	app := fx.New(options...)
	if err := app.Err(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	if wrapped.Handler == nil {
		t.Fatal("HTTP handler is nil")
	}

	return wrapped.Handler
}

func buildHealthHTTPHandler(t *testing.T) http.Handler {
	t.Helper()

	cfg := Config{
		HTTP:       HTTPConfig{Address: "127.0.0.1:0"},
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
		GRPC:       grpcserver.Config{Address: "127.0.0.1:0"},
	}
	var wrapped healthHTTPHandler
	options := append(appOptions(cfg), fx.Populate(&wrapped))
	app := fx.New(options...)
	if err := app.Err(); err != nil {
		t.Fatalf("build app: %v", err)
	}
	if wrapped.Handler == nil {
		t.Fatal("health HTTP handler is nil")
	}

	return wrapped.Handler
}
