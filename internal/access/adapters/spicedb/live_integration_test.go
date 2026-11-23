//go:build integration

package spicedb

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/google/uuid"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

func liveSpiceDB(t *testing.T) (*Client, context.Context) {
	t.Helper()
	endpoint := os.Getenv("M8_ACCESS_TEST_SPICEDB_ENDPOINT")
	token := os.Getenv("M8_ACCESS_TEST_SPICEDB_TOKEN")
	if endpoint == "" || token == "" {
		t.Skip("requires disposable SpiceDB gRPC test service")
	}
	p, _ := domain.NewPermission("m8.workspace.read")
	def, _ := domain.NewPermissionDefinition(p, "workspace", "view")
	catalog, err := NewStaticPermissionCatalog([]domain.PermissionDefinition{def})
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(Config{Endpoint: endpoint, BearerToken: token, ModelRevision: "test-model-v1", AllowInsecureLoopback: true}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	schema, err := os.ReadFile("../../schema/m8.zed")
	if err != nil {
		t.Fatal(err)
	}
	var wrote bool
	for i := 0; i < 25; i++ {
		_, err = client.sdk.WriteSchema(ctx, &pb.WriteSchemaRequest{Schema: string(schema)})
		if err == nil {
			wrote = true
			break
		}
		time.Sleep(400 * time.Millisecond)
	}
	if !wrote {
		t.Fatalf("real SpiceDB schema write: %v", err)
	}
	return client, ctx
}

// Exercises the real SpiceDB datastore and gRPC preconditions, not mocks.
func TestLiveSchemaFencingIndependentGrants(t *testing.T) {
	client, ctx := liveSpiceDB(t)
	p, _ := domain.NewPermission("m8.workspace.read")
	tenant, _ := domain.NewTenantID(tenantValue)
	workspace, _ := domain.NewResource("workspace", "same-id")
	subject, _ := domain.NewSubject("user", "same-id")
	mkGrant := func(id string) (ports.GraphFence, domain.Relationship, domain.Relationship) {
		fence := ports.GraphFence{TenantID: tenant, Kind: "ACCESS_GRANT", AggregateID: id, Epoch: 1}
		node, _ := domain.NewResource("access_grant", id)
		holder, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: tenant, Resource: node, Relation: "holder", Subject: subject, ExpiresAt: time.Now().UTC().Add(time.Hour)})
		indirect, _ := domain.NewSubject("access_grant", id)
		viewer, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: tenant, Resource: workspace, Relation: "viewer", Subject: indirect, SubjectRelation: "holder"})
		return fence, holder, viewer
	}
	g1, a1, b1 := mkGrant("11111111-1111-4111-8111-111111111111")
	g2, a2, b2 := mkGrant("22222222-2222-4222-8222-222222222222")
	for _, item := range []struct {
		f              ports.GraphFence
		holder, viewer domain.Relationship
	}{{g1, a1, b1}, {g2, a2, b2}} {
		if _, err := client.EstablishFence(ctx, item.f); err != nil {
			t.Fatalf("establish graph epoch: %v", err)
		}
		guarded, err := ports.WithGraphFence(ctx, item.f)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.WriteRelationships(guarded, tenant, []ports.RelationshipChange{
			{Operation: ports.RelationshipTouch, Relationship: item.holder},
			{Operation: ports.RelationshipTouch, Relationship: item.viewer},
		})
		if err != nil {
			t.Fatalf("write grant tuple: %v", err)
		}
	}
	evaluate := func(checkTenant domain.TenantID, allowed bool) {
		t.Helper()
		rev, _ := domain.NewModelRevision("test-model-v1")
		req, err := domain.NewCheckPermissionRequest(domain.CheckPermissionInput{TenantID: checkTenant, Subject: subject,
			Permission: p, Resource: workspace, ModelRevision: rev, Critical: true})
		if err != nil {
			t.Fatal(err)
		}
		result, err := client.CheckPermission(ctx, req)
		if err != nil || result.Allowed() != allowed {
			t.Fatalf("tenant %s: permission should be %v: %+v %v", checkTenant, allowed, result, err)
		}
	}
	evaluate(tenant, true)
	evaluate(otherTenantValue, false)
	c1, _ := ports.WithGraphFence(ctx, g1)
	if _, err := client.WriteRelationships(c1, tenant, []ports.RelationshipChange{
		{Operation: ports.RelationshipDelete, Relationship: a1}, {Operation: ports.RelationshipDelete, Relationship: b1},
	}); err != nil {
		t.Fatalf("revoke first grant: %v", err)
	}
	evaluate(tenant, true) // independent grant #2 remains effective
	newest := g1
	newest.Epoch = 2
	if _, err := client.EstablishFence(ctx, newest); err != nil {
		t.Fatalf("take over graph fence: %v", err)
	}
	if _, err := client.EstablishFence(ctx, g1); !errors.Is(err, ErrGraphFenceConflict) {
		t.Fatalf("old graph epoch reinstalled: %v", err)
	}
	if _, err := client.WriteRelationships(c1, tenant, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: a1}}); !errors.Is(err, ErrGraphFenceConflict) {
		t.Fatalf("stale graph writer accepted: %v", err)
	}
	c2, _ := ports.WithGraphFence(ctx, g2)
	if _, err := client.WriteRelationships(c2, tenant, []ports.RelationshipChange{
		{Operation: ports.RelationshipDelete, Relationship: a2}, {Operation: ports.RelationshipDelete, Relationship: b2},
	}); err != nil {
		t.Fatalf("revoke second grant: %v", err)
	}
	evaluate(tenant, false)
}

func TestLiveReadProjectionEdgesPageBoundary(t *testing.T) {
	// The extra holder is on the continuation page of the resource filter.
	checkLiveProjectionEdgeLimit(t, 1001)
}

func TestLiveReadProjectionEdgesLimitAcrossFilters(t *testing.T) {
	// 500 holders and 500 workspace viewers fit. The 501st viewer exceeds
	// the aggregate limit while both filters still return a partial page.
	checkLiveProjectionEdgeLimit(t, 500)
}

func checkLiveProjectionEdgeLimit(t *testing.T, holderCount int) {
	t.Helper()
	client, ctx := liveSpiceDB(t)
	const pageSize = 1000
	fence := ports.GraphFence{TenantID: tenantValue, Kind: "ROLE_BINDING", AggregateID: uuid.NewString(), Epoch: 1}
	if _, err := client.EstablishFence(ctx, fence); err != nil {
		t.Fatal(err)
	}
	guarded, err := ports.WithGraphFence(ctx, fence)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := domain.NewResource("access_binding", fence.AggregateID)
	if err != nil {
		t.Fatal(err)
	}

	bindingSubject, err := domain.NewSubject(binding.Type(), binding.ID())
	if err != nil {
		t.Fatal(err)
	}
	changes := make([]ports.RelationshipChange, pageSize+1)
	for i := range changes {
		input := domain.RelationshipInput{TenantID: fence.TenantID}
		if i < holderCount {
			user, err := domain.NewSubject("user", strconv.Itoa(i))
			if err != nil {
				t.Fatal(err)
			}
			input.Resource, input.Relation, input.Subject = binding, "holder", user
		} else {
			workspace, err := domain.NewResource("workspace", strconv.Itoa(i))
			if err != nil {
				t.Fatal(err)
			}
			input.Resource, input.Relation = workspace, "viewer"
			input.Subject, input.SubjectRelation = bindingSubject, "holder"
		}
		rel, err := domain.NewRelationship(input)
		if err != nil {
			t.Fatal(err)
		}
		changes[i] = ports.RelationshipChange{Operation: ports.RelationshipCreate, Relationship: rel}
	}
	if _, err := client.WriteRelationships(guarded, fence.TenantID, changes[:pageSize]); err != nil {
		t.Fatal(err)
	}
	edges, err := client.ReadProjectionEdges(ctx, fence)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != pageSize {
		t.Fatalf("read %d edges, want %d", len(edges), pageSize)
	}
	want := make(map[domain.RelationshipKey]bool, pageSize)
	for _, change := range changes[:pageSize] {
		want[change.Relationship.Key()] = true
	}
	for _, edge := range edges {
		if !want[edge.Key()] {
			t.Fatalf("unexpected or duplicate edge: %+v", edge.Key())
		}
		delete(want, edge.Key())
	}

	if _, err := client.WriteRelationships(guarded, fence.TenantID, changes[pageSize:]); err != nil {
		t.Fatal(err)
	}
	edges, err = client.ReadProjectionEdges(ctx, fence)
	if !errors.Is(err, ErrInvalidResponse) || edges != nil {
		t.Fatalf("overflow returned %d edges and error %v, want nil edges and ErrInvalidResponse", len(edges), err)
	}
}
