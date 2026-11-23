//go:build integration

package ydbcheckpoint

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/adapters/spicedb"
	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type bindingSnapshots map[projection.Key]projection.Snapshot

func (s bindingSnapshots) LoadCurrent(_ context.Context, key projection.Key) (projection.Snapshot, error) {
	snapshot, ok := s[key]
	if !ok {
		return projection.Snapshot{}, fmt.Errorf("missing binding snapshot for %s", key)
	}
	return snapshot, nil
}

// Observe the real graph commit while the projection worker still holds the
// YDB lease and has not yet received the token to publish its checkpoint.
type observedGraph struct {
	*spicedb.Client
	afterWrite func(context.Context, ports.GraphFence, string)
}

func (g *observedGraph) WriteRelationships(ctx context.Context, tenant domain.TenantID, changes []ports.RelationshipChange) (ports.RelationshipWriteResult, error) {
	result, err := g.Client.WriteRelationships(ctx, tenant, changes)
	if err == nil {
		fence, ok := ports.GraphFenceFromContext(ctx)
		if !ok {
			return result, ports.ErrInvalidGraphFence
		}
		g.afterWrite(ctx, fence, result.ZedToken)
	}
	return result, err
}

func TestLivePolicyJournalWaitsForGraphAndAllCheckpoints(t *testing.T) {
	endpoint, token := os.Getenv("M8_ACCESS_TEST_SPICEDB_ENDPOINT"), os.Getenv("M8_ACCESS_TEST_SPICEDB_TOKEN")
	if endpoint == "" || token == "" {
		t.Skip("requires disposable SpiceDB")
	}
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil || (host != "localhost" && !net.ParseIP(host).IsLoopback()) {
		t.Fatal("requires a loopback SpiceDB endpoint")
	}
	store := livePolicyStore(t)
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	schema, err := os.ReadFile("../../schema/m8.zed")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pb.NewSchemaServiceClient(conn).WriteSchema(ctx, &pb.WriteSchemaRequest{Schema: string(schema)}, grpc.WaitForReady(true)); err != nil {
		t.Fatal(err)
	}

	const tenant domain.TenantID = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	permission, err := domain.NewPermission("m8.workspace.read")
	if err != nil {
		t.Fatal(err)
	}
	def, err := domain.NewPermissionDefinition(permission, "workspace", "view")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := spicedb.NewStaticPermissionCatalog([]domain.PermissionDefinition{def})
	if err != nil {
		t.Fatal(err)
	}
	client, err := spicedb.NewWithPermissionsClient("test-model-v1", catalog, pb.NewPermissionsServiceClient(conn))
	if err != nil {
		t.Fatal(err)
	}
	role, err := domain.NewRole(domain.RoleInput{TenantID: tenant, ID: "33333333-3333-4333-8333-333333333333", Name: "reader", Permissions: []domain.Permission{permission}, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	scope, err := domain.NewResource("workspace", "journal-workspace")
	if err != nil {
		t.Fatal(err)
	}
	subject, err := domain.NewSubject("user", "journal-reader")
	if err != nil {
		t.Fatal(err)
	}
	snapshots := bindingSnapshots{}
	manifest := policyjournal.Mutation{TenantID: tenant, ID: "bootstrap-bindings", Bootstrap: true}
	for _, id := range []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"} {
		key := projection.Key{TenantID: tenant, Kind: projection.KindRoleBinding, AggregateID: id}
		binding, err := domain.NewRoleBinding(domain.RoleBindingInput{TenantID: tenant, ID: id, RoleID: role.ID(), Subject: subject, Scope: scope, Revision: 1})
		if err != nil {
			t.Fatal(err)
		}
		snapshots[key] = projection.Snapshot{Key: key, Revision: 1, RoleBinding: &projection.RoleBindingState{Role: role, Binding: binding}}
		manifest.Jobs = append(manifest.Jobs, policyjournal.Job{Key: key, RequiredRevision: 1})
	}
	revision, err := publishLivePolicy(t, store, manifest)
	if err != nil {
		t.Fatal(err)
	}
	assertApplied := func(want uint64) {
		t.Helper()
		state, err := store.LoadPolicyState(ctx, tenant)
		if err != nil || state.Head != revision || state.Applied != want {
			t.Fatalf("policy state %+v, want head %d applied %d: %v", state, revision, want, err)
		}
	}
	assertApplied(0)
	tokens := map[projection.Key]string{}
	graph := &observedGraph{Client: client}
	graph.afterWrite = func(ctx context.Context, f ports.GraphFence, token string) {
		t.Helper()
		key := projection.Key{TenantID: f.TenantID, Kind: projection.Kind(f.Kind), AggregateID: f.AggregateID}
		edges, err := client.ReadProjectionEdges(ctx, f)
		if err != nil {
			t.Fatal(err)
		}
		want := map[domain.RelationshipKey]bool{
			{TenantID: tenant, ResourceType: "access_binding", ResourceID: key.AggregateID, Relation: "holder", SubjectType: subject.Type(), SubjectID: subject.ID()}:                        true,
			{TenantID: tenant, ResourceType: scope.Type(), ResourceID: scope.ID(), Relation: "viewer", SubjectType: "access_binding", SubjectID: key.AggregateID, SubjectRelation: "holder"}: true,
		}
		for _, edge := range edges {
			if !want[edge.Key()] {
				t.Fatalf("unexpected graph edge %+v", edge.Key())
			}
			delete(want, edge.Key())
		}
		if len(want) != 0 {
			t.Fatalf("missing graph edges %v", want)
		}
		if _, err := store.LoadCheckpoint(ctx, key); !errors.Is(err, ErrCheckpointNotConverged) {
			t.Fatalf("checkpoint before commit: %v", err)
		}
		if err := store.CommitPolicyJob(ctx, revision, key); !errors.Is(err, ErrPolicyCheckpointBehind) {
			t.Fatalf("certification before checkpoint: %v", err)
		}
		assertApplied(0)
		tokens[key] = token
	}
	projector, err := projection.New(catalog, graph, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	worker, err := projection.NewWorker(projector, snapshots, store, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	journal, err := policyjournal.NewWorker(worker, store)
	if err != nil {
		t.Fatal(err)
	}
	for i, job := range manifest.Jobs {
		event := projection.Event{Key: job.Key, Revision: job.RequiredRevision, EventID: "binding-" + job.Key.AggregateID}
		if err := journal.Process(ctx, policyjournal.Event{PolicyRevision: revision, Projection: event}); err != nil {
			t.Fatal(err)
		}
		cp, err := store.LoadCheckpoint(ctx, job.Key)
		if err != nil || cp.EventID != event.EventID || cp.Revision != job.RequiredRevision || cp.ZedToken == "" || cp.ZedToken != tokens[job.Key] {
			t.Fatalf("checkpoint for %s did not certify the graph write: %v", job.Key, err)
		}
		if i+1 < len(manifest.Jobs) {
			assertApplied(0)
		}
	}
	assertApplied(revision)
}
