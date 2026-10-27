package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func requestGateway(t *testing.T, handler http.Handler, method, path, body string, expected int) string {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("%s %s: status=%d, want=%d, body=%s", method, path, response.Code, expected, response.Body.String())
	}
	return response.Body.String()
}

func createHTTPTestOrganization(t *testing.T, handler http.Handler) (string, string) {
	t.Helper()
	tenantBody := requestGateway(t, handler, http.MethodPost, "/resource-manager/v1/resources",
		`{"resource":{"tenant":{"metadata":{"name":"Demo Tenant"}}},"idempotencyKey":"org-http-tenant-1"}`, http.StatusOK)
	var tenant pb.ResourceRecord
	if err := protojson.Unmarshal([]byte(tenantBody), &tenant); err != nil {
		t.Fatal(err)
	}
	tenantID := tenant.GetTenant().GetMetadata().GetId()
	if tenantID == "" {
		t.Fatal("Tenant ID missing from response")
	}
	body := fmt.Sprintf(`{"resource":{"organization":{"tenantId":%q,"metadata":{"name":"Alpha"}}},"idempotencyKey":"org-http-create-1"}`, tenantID)
	orgBody := requestGateway(t, handler, http.MethodPost, "/resource-manager/v1/resources", body, http.StatusOK)
	var organization pb.ResourceRecord
	if err := protojson.Unmarshal([]byte(orgBody), &organization); err != nil {
		t.Fatal(err)
	}
	orgID := organization.GetOrganization().GetMetadata().GetId()
	if orgID == "" {
		t.Fatal("Organization ID missing from response")
	}
	return tenantID, orgID
}

func TestOrgStructureRESTGatewayInitialEmptyTree(t *testing.T) {
	handler := buildHTTPHandler(t, true)
	_, orgID := createHTTPTestOrganization(t, handler)
	path := fmt.Sprintf("/resource-manager/v1/organizations/%s/org-units?pageSize=200", orgID)
	body := requestGateway(t, handler, http.MethodGet, path, "", http.StatusOK)
	var result pb.ListOrgUnitsResponse
	if err := protojson.Unmarshal([]byte(body), &result); err != nil {
		t.Fatal(err)
	}
	if result.GetStructureRevision() != 1 || len(result.GetOrgUnits()) != 0 {
		t.Fatalf("unexpected initial structure: %s", body)
	}
	// Regression: protojson emits int64 revision as a quoted decimal string and
	// by default omits the empty repeated orgUnits field.
	if !strings.Contains(body, `"structureRevision":"1"`) {
		t.Fatalf("expected protojson int64 string; got %s", body)
	}
}

func TestOrgStructureRESTGatewayValidateApplyAndList(t *testing.T) {
	handler := buildHTTPHandler(t, true)
	_, orgID := createHTTPTestOrganization(t, handler)
	change := fmt.Sprintf(`{"changeSet":{"id":"org-http-change-1","organizationId":%q,"expectedStructureRevision":"1","mutations":[{"kind":"ORG_UNIT_MUTATION_KIND_CREATE","orgUnitId":"engineering","displayName":"Engineering"}]}}`, orgID)
	validationBody := requestGateway(t, handler, http.MethodPost,
		"/resource-manager/v1/org-unit-change-sets:validate", change, http.StatusOK)
	var validation pb.ValidateOrgUnitChangeSetResponse
	if err := protojson.Unmarshal([]byte(validationBody), &validation); err != nil {
		t.Fatal(err)
	}
	if !validation.GetValid() || validation.GetProposedStructureRevision() != 2 {
		t.Fatalf("unexpected validation result: %s", validationBody)
	}
	apply := fmt.Sprintf(`{"changeSet":{"id":"org-http-change-1","organizationId":%q,"expectedStructureRevision":"1","mutations":[{"kind":"ORG_UNIT_MUTATION_KIND_CREATE","orgUnitId":"engineering","displayName":"Engineering"}]},"idempotencyKey":"org-http-change-1"}`, orgID)
	appliedBody := requestGateway(t, handler, http.MethodPost,
		"/resource-manager/v1/org-unit-change-sets:apply", apply, http.StatusOK)
	var applied pb.ApplyOrgUnitChangeSetResponse
	if err := protojson.Unmarshal([]byte(appliedBody), &applied); err != nil {
		t.Fatal(err)
	}
	if applied.GetStructureRevision() != 2 {
		t.Fatalf("unexpected apply result: %s", appliedBody)
	}
	body := requestGateway(t, handler, http.MethodGet,
		fmt.Sprintf("/resource-manager/v1/organizations/%s/org-units?pageSize=200", orgID),
		"", http.StatusOK)
	var result pb.ListOrgUnitsResponse
	if err := protojson.Unmarshal([]byte(body), &result); err != nil {
		t.Fatal(err)
	}
	if result.GetStructureRevision() != 2 || len(result.GetOrgUnits()) != 1 ||
		result.GetOrgUnits()[0].GetMetadata().GetName() != "Engineering" {
		t.Fatalf("unexpected structure after apply: %s", body)
	}
}

func TestOrgStructureRESTGatewayMissingOrganization(t *testing.T) {
	handler := buildHTTPHandler(t, true)
	body := requestGateway(t, handler, http.MethodGet,
		"/resource-manager/v1/organizations/does-not-exist/org-units", "", http.StatusNotFound)
	// A missing Organization is a real backend NOT_FOUND, not an unregistered
	// gateway path returning the generic message "Not Found".
	if strings.Contains(body, `"message":"Not Found"`) {
		t.Fatalf("unregistered route instead of backend error: %s", body)
	}
}

func TestOrgStructureRESTGatewayDefaultDenied(t *testing.T) {
	handler := buildHTTPHandler(t, false)
	requestGateway(t, handler, http.MethodGet,
		"/resource-manager/v1/organizations/a-real-looking-id/org-units", "", http.StatusUnauthorized)
	requestGateway(t, handler, http.MethodPost,
		"/resource-manager/v1/org-unit-change-sets:validate",
		`{"changeSet":{"id":"x","organizationId":"a-real-looking-id","expectedStructureRevision":"1"}}`,
		http.StatusUnauthorized)
}
