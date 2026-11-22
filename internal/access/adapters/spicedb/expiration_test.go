package spicedb

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

func TestExpirationTransportAndDeletionIdentity(t *testing.T) {
	expiry := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	var seen [][]map[string]any
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Updates []map[string]any `json:"updates"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		seen = append(seen, payload.Updates)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"writtenAt":{"token":"zed-write"}}`))
	})
	tn, e := domain.NewTenantID(tenantValue)
	if e != nil {
		t.Fatal(e)
	}
	obj, e := domain.NewResource("access_grant", "11111111-1111-4111-8111-111111111111")
	if e != nil {
		t.Fatal(e)
	}
	subject, e := domain.NewSubject("user", "alice")
	if e != nil {
		t.Fatal(e)
	}
	rel, e := domain.NewRelationship(domain.RelationshipInput{TenantID: tn, Resource: obj, Subject: subject, Relation: "holder", ExpiresAt: expiry})
	if e != nil {
		t.Fatal(e)
	}
	for _, op := range []ports.RelationshipOperation{ports.RelationshipTouch, ports.RelationshipDelete} {
		_, e = client.WriteRelationships(context.Background(), tn, []ports.RelationshipChange{{Operation: op, Relationship: rel}})
		if e != nil {
			t.Fatal(e)
		}
	}
	if len(seen) != 2 {
		t.Fatalf("requests=%d", len(seen))
	}
	wr := seen[0][0]["relationship"].(map[string]any)
	if wr["optionalExpiresAt"] != "2027-01-02T03:04:05Z" {
		t.Fatalf("expiration absent: %+v", wr)
	}
	deletion := seen[1][0]["relationship"].(map[string]any)
	if _, ok := deletion["optionalExpiresAt"]; ok {
		t.Fatalf("delete must omit expiration: %+v", deletion)
	}
	if seen[1][0]["operation"] != "OPERATION_DELETE" {
		t.Fatalf("unexpected delete operation")
	}
}
