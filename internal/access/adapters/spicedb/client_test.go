package spicedb

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

const tenantValue = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
const otherTenantValue = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	permission, _ := domain.NewPermission("m8.workspace.read")
	definition, err := domain.NewPermissionDefinition(permission, "workspace", "view")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewStaticPermissionCatalog([]domain.PermissionDefinition{definition})
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(Config{Endpoint: server.URL, BearerToken: "test-only", ModelRevision: "model-v1", AllowInsecureLoopback: true}, catalog, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func testRequest(t *testing.T, tenantString string, token string) domain.CheckPermissionRequest {
	t.Helper()
	tenant, err := domain.NewTenantID(tenantString)
	if err != nil {
		t.Fatal(err)
	}
	subject, err := domain.NewSubject("user", "same-id")
	if err != nil {
		t.Fatal(err)
	}
	resource, err := domain.NewResource("workspace", "same-id")
	if err != nil {
		t.Fatal(err)
	}
	permission, err := domain.NewPermission("m8.workspace.read")
	if err != nil {
		t.Fatal(err)
	}
	revision, err := domain.NewModelRevision("model-v1")
	if err != nil {
		t.Fatal(err)
	}
	req, err := domain.NewCheckPermissionRequest(domain.CheckPermissionInput{TenantID: tenant, Subject: subject, Resource: resource, Permission: permission, ModelRevision: revision, MinZedToken: token, Critical: true})
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestCheckPermissionTenantIsolationAndStrongConsistency(t *testing.T) {
	var ids []string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/permissions/check" || r.Header.Get("Authorization") != "Bearer test-only" {
			t.Errorf("incorrect method/auth: %s", r.URL.Path)
		}
		var body struct {
			Consistency map[string]any   `json:"consistency"`
			Resource    objectReference  `json:"resource"`
			Subject     subjectReference `json:"subject"`
			Permission  string           `json:"permission"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Consistency["fullyConsistent"] != true {
			t.Error("check must use strong consistency by default")
		}
		if body.Resource.ObjectID != body.Subject.Object.ObjectID {
			t.Error("subject and resource should match when IDs equal")
		}
		if body.Permission != "view" {
			t.Error("missing registry mapping")
		}
		ids = append(ids, body.Resource.ObjectID)
		_, _ = w.Write([]byte(`{"checkedAt":{"token":"zed-v1"},"permissionship":"PERMISSIONSHIP_HAS_PERMISSION"}`))
	})
	a, err := client.CheckPermission(context.Background(), testRequest(t, tenantValue, ""))
	if err != nil || !a.Allowed() || a.ZedToken() != "zed-v1" {
		t.Fatalf("allow: %v %#v", err, a)
	}
	b, err := client.CheckPermission(context.Background(), testRequest(t, otherTenantValue, ""))
	if err != nil || !b.Allowed() {
		t.Fatalf("second tenant: %v", err)
	}
	if len(ids) != 2 || ids[0] == ids[1] || !strings.HasPrefix(ids[0], tenantValue+"/") || !strings.HasPrefix(ids[1], otherTenantValue+"/") {
		t.Fatalf("tenant identities leaked: %v", ids)
	}
}

func TestCheckPermissionZedTokenAndConditionalDenial(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Consistency map[string]json.RawMessage `json:"consistency"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		var fresh struct {
			Token string `json:"token"`
		}
		_ = json.Unmarshal(body.Consistency["atLeastAsFresh"], &fresh)
		if fresh.Token != "must-see-revoke" {
			t.Errorf("wrong token: %v", body.Consistency)
		}
		_, _ = w.Write([]byte(`{"checkedAt":{"token":"newer"},"permissionship":"PERMISSIONSHIP_CONDITIONAL_PERMISSION","partialCaveatInfo":{"missingRequiredContext":["mfa"]}}`))
	})
	decision, err := client.CheckPermission(context.Background(), testRequest(t, tenantValue, "must-see-revoke"))
	if err != nil || decision.Allowed() || decision.Reason() != "spicedb_missing_caveat_context" {
		t.Fatalf("conditional must deny: %v %#v", err, decision)
	}
}

func TestCheckPermissionFailClosedOnProtocolAndTransport(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"bad result", 200, `{"checkedAt":{"token":"x"},"permissionship":"PERMISSIONSHIP_UNSPECIFIED"}`, ErrInvalidResponse},
		{"missing revision", 200, `{"permissionship":"PERMISSIONSHIP_HAS_PERMISSION"}`, ErrInvalidResponse},
		{"invalid json", 200, `{`, ErrInvalidResponse},
		{"service unavailable", 503, `{"error":"secret"}`, ports.ErrPermissionEngineUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := c.CheckPermission(context.Background(), testRequest(t, tenantValue, ""))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v; want %v", err, tc.want)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("SpiceDB response leaked")
			}
		})
	}
}

func TestCheckPermissionModelRevisionMismatchNeverCallsSpiceDB(t *testing.T) {
	called := false
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { called = true })
	request := testRequest(t, tenantValue, "")
	other, _ := domain.NewModelRevision("model-v2")
	// An unregistered or unverified schema revision must not be silently used.
	client.revision = other
	_, err := client.CheckPermission(context.Background(), request)
	if !errors.Is(err, ErrSchemaRevisionMismatch) || called {
		t.Fatalf("mismatch: %v, called=%v", err, called)
	}
}

func TestWriteRelationshipsExactTenantAndToken(t *testing.T) {
	var payload struct {
		Updates []struct {
			Operation    string            `json:"operation"`
			Relationship graphRelationship `json:"relationship"`
		} `json:"updates"`
	}
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/relationships/write" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		_, _ = w.Write([]byte(`{"writtenAt":{"token":"write-revision"}}`))
	})
	tenant, _ := domain.NewTenantID(tenantValue)
	res, _ := domain.NewResource("workspace", "w-1")
	subject, _ := domain.NewSubject("user", "u-1")
	relationship, err := domain.NewRelationship(domain.RelationshipInput{TenantID: tenant, Resource: res, Relation: "viewer", Subject: subject, ConditionRef: "mfa_verified"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.WriteRelationships(context.Background(), tenant, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: relationship}})
	if err != nil || result.ZedToken != "write-revision" {
		t.Fatalf("write failed: %v", err)
	}
	if len(payload.Updates) != 1 || payload.Updates[0].Operation != "OPERATION_TOUCH" || payload.Updates[0].Relationship.OptionalCaveat.CaveatName != "mfa_verified" {
		t.Fatalf("wrong graph tuple: %+v", payload)
	}
}

func TestWriteRelationshipsRejectsCrossTenantAndDuplicates(t *testing.T) {
	called := false
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { called = true })
	tenant, _ := domain.NewTenantID(tenantValue)
	other, _ := domain.NewTenantID(otherTenantValue)
	res, _ := domain.NewResource("workspace", "w-1")
	subject, _ := domain.NewSubject("user", "u-1")
	rel, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: other, Resource: res, Relation: "viewer", Subject: subject})
	_, err := client.WriteRelationships(context.Background(), tenant, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: rel}})
	if !errors.Is(err, ports.ErrInvalidRelationshipChange) {
		t.Fatalf("wrong tenant: %v", err)
	}
	rel, _ = domain.NewRelationship(domain.RelationshipInput{TenantID: tenant, Resource: res, Relation: "viewer", Subject: subject})
	changes := []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: rel}, {Operation: ports.RelationshipDelete, Relationship: rel}}
	_, err = client.WriteRelationships(context.Background(), tenant, changes)
	if !errors.Is(err, ports.ErrInvalidRelationshipChange) || called {
		t.Fatalf("duplicate must be rejected before call: %v %v", err, called)
	}
}

func TestNewClientRequiresSecureEndpoint(t *testing.T) {
	permission, _ := domain.NewPermission("m8.workspace.read")
	definition, _ := domain.NewPermissionDefinition(permission, "workspace", "view")
	catalog, _ := NewStaticPermissionCatalog([]domain.PermissionDefinition{definition})
	for _, url := range []string{"http://example.com", "http://localhost:8443", "https://user:password@example.com", "https://example.com?q=unsafe"} {
		_, err := NewClient(Config{Endpoint: url, BearerToken: "secret", ModelRevision: "model-v1"}, catalog, nil)
		if !errors.Is(err, ErrInvalidConfig) {
			t.Errorf("%q accepted: %v", url, err)
		}
	}
}

func TestCancelledCheckNeverSendsRequest(t *testing.T) {
	called := false
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { called = true })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.CheckPermission(ctx, testRequest(t, tenantValue, ""))
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("canceled: %v called=%v", err, called)
	}
}

func TestNetworkDeadlineIsClassified(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(25 * time.Millisecond)
		_, _ = w.Write([]byte(`{"checkedAt":{"token":"x"},"permissionship":"PERMISSIONSHIP_HAS_PERMISSION"}`))
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := c.CheckPermission(ctx, testRequest(t, tenantValue, ""))
	if !errors.Is(err, ports.ErrPermissionEngineTimeout) {
		t.Fatalf("timeout not classified: %v", err)
	}
}
