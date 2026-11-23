//go:build integration

package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/adapters/spicedb"
	"github.com/m8-team/platform/internal/access/adapters/ydbcheckpoint"
	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type liveUniversalOwner struct {
	proof consistency.UniversalOwnerProof
}

func (o liveUniversalOwner) ResolveUniversalResource(_ context.Context, _ domain.TenantID, _, _ string) (consistency.UniversalOwnerProof, error) {
	return o.proof, nil
}

type liveUniversalCatalog struct{}

func (liveUniversalCatalog) HasCatalogPermission(_ context.Context, _ domain.TenantID, kind, action string) (bool, error) {
	return kind == "finance.invoice" && action == "finance.invoice.view", nil
}

func applyPolicySchema(t *testing.T, s *Store, ctx context.Context) {
	t.Helper()
	for _, filename := range []string{"../ydbcheckpoint/schema.sql", "../ydbcheckpoint/schema_policy_v2.sql"} {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		var clean []string
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				clean = append(clean, line)
			}
		}
		for _, fragment := range strings.Split(strings.Join(clean, "\n"), ";") {
			statement := strings.TrimSpace(fragment)
			if statement == "" {
				continue
			}
			_, err = s.db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s;", s.prefix, statement))
			if err != nil {
				t.Fatalf("apply checkpoint migration: %v statement=%s", err, statement)
			}
		}
	}
}
func liveProfileAssignment(t *testing.T) (role.ProfileAssignment, projection.ProfileCapability) {
	t.Helper()
	tenant, _ := shared.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	user, _ := shared.NewSubject("user", "profile-reader")
	scope, _ := authorization.NewScope(authorization.ScopeInput{Kind: authorization.ScopeInstance,
		TenantID: tenant, ResourceType: "finance.invoice", ResourceID: "invoice-123"})
	a, err := role.NewProfileAssignment(role.ProfileAssignmentInput{
		ID: "11111111-1111-4111-8111-111111111111", TenantID: tenant, Subject: user,
		Target: role.AssignmentTarget{Kind: role.AssignmentTemplate, Name: "finance.accountant", Version: "1.0.0"},
		Scopes: []authorization.Scope{scope},
	})
	if err != nil {
		t.Fatal(err)
	}
	return a, projection.ProfileCapability{ResourceType: "finance.invoice", ResourceID: "invoice-123", Action: "finance.invoice.view", OwnerRevision: 7}
}
func TestLiveProfilePolicyJournalGraphActivationAndRevocation(t *testing.T) {
	endpoint, token := os.Getenv("M8_ACCESS_TEST_SPICEDB_ENDPOINT"), os.Getenv("M8_ACCESS_TEST_SPICEDB_TOKEN")
	if endpoint == "" || token == "" {
		t.Skip("requires disposable SpiceDB")
	}
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil || host != "localhost" && !net.ParseIP(host).IsLoopback() {
		t.Fatal("only local SpiceDB allowed")
	}
	s := liveStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()
	applyPolicySchema(t, s, ctx)
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	schema, err := os.ReadFile("../../schema/m8.zed")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pb.NewSchemaServiceClient(conn).WriteSchema(ctx, &pb.WriteSchemaRequest{Schema: string(schema)}, grpc.WaitForReady(true))
	if err != nil {
		t.Fatal(err)
	}
	workspacePermission, _ := domain.NewPermission("m8.workspace.read")
	workspaceDef, _ := domain.NewPermissionDefinition(workspacePermission, "workspace", "view")
	catalog, err := spicedb.NewStaticPermissionCatalog([]domain.PermissionDefinition{workspaceDef})
	if err != nil {
		t.Fatal(err)
	}
	graph, err := spicedb.NewWithPermissionsClient("test-model-v1", catalog, pb.NewPermissionsServiceClient(conn))
	if err != nil {
		t.Fatal(err)
	}
	checkpoints, err := ydbcheckpoint.New(s.db, s.prefix, "profile-worker-v3", 4*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	projector, err := projection.New(catalog, graph, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	worker, err := projection.NewWorker(projector, s, checkpoints, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	journalWorker, err := policyjournal.NewWorker(worker, checkpoints)
	if err != nil {
		t.Fatal(err)
	}
	a, cap := liveProfileAssignment(t)
	if _, err := s.CreateAssignment(ctx, a, "stage"); err != nil {
		t.Fatal(err)
	}
	key := projection.Key{TenantID: domain.TenantID(a.TenantID()), Kind: projection.KindProfileAssignment, AggregateID: a.ID()}
	boot := policyjournal.Mutation{TenantID: key.TenantID, ExpectedHead: 0, Bootstrap: true, ID: "profile-bootstrap", Jobs: []policyjournal.Job{{Key: key, RequiredRevision: 1}}}
	var bootHead uint64
	err = s.tx(ctx, func(tx *sql.Tx) error {
		var err error
		bootHead, err = checkpoints.PublishPolicyMutationTx(ctx, tx, boot)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if bootHead != 1 {
		t.Fatalf("bootstrap head: %d", bootHead)
	}
	if err := journalWorker.Process(ctx, policyjournal.Event{PolicyRevision: bootHead, Projection: projection.Event{
		Key: key, EventID: "profile-boot", Revision: 1}}); err != nil {
		t.Fatal(err)
	}
	state, err := checkpoints.LoadPolicyState(ctx, key.TenantID)
	if err != nil || state.Head != 1 || state.Applied != 1 {
		t.Fatalf("bootstrap not certified %+v %v", state, err)
	}

	proof := consistency.UniversalOwnerProof{TenantID: key.TenantID, ResourceType: cap.ResourceType, ResourceID: cap.ResourceID, Revision: 7, Live: true}
	guard, err := consistency.NewUniversalBarrier(checkpoints, liveUniversalOwner{proof: proof}, liveUniversalCatalog{}, graph)
	if err != nil {
		t.Fatal(err)
	}
	user, _ := domain.NewSubject(a.Subject().Type(), a.Subject().ID())
	check := func(want bool) {
		t.Helper()
		got, err := guard.Check(ctx, key.TenantID, user, cap.ResourceType, cap.ResourceID, cap.Action)
		if err != nil || got != want {
			t.Fatalf("want %v got %v, %v", want, got, err)
		}
	}
	check(false)
	activated, err := s.ActivateWithPolicy(ctx, checkpoints, a.TenantID(), a.ID(), 1, bootHead, "activate-v1", []projection.ProfileCapability{cap})
	if err != nil {
		t.Fatal(err)
	}
	if activated.PolicyRevision != 2 || activated.Assignment.Status() != role.ProfileAssignmentActive {
		t.Fatal("activation not journaled")
	}
	if _, err := guard.Check(ctx, key.TenantID, user, cap.ResourceType, cap.ResourceID, cap.Action); !errors.Is(err, consistency.ErrPolicyProjectionLag) {
		t.Fatalf("allow before checkpoint: %v", err)
	}
	if err := journalWorker.Process(ctx, policyjournal.Event{PolicyRevision: activated.PolicyRevision, Projection: activated.ProjectionEvent}); err != nil {
		t.Fatal(err)
	}
	check(true)

	pending, err := s.PendingPolicy(ctx, a.TenantID(), a.ID(), 0, 10)
	if err != nil || len(pending) != 1 || pending[0].PolicyRevision != 2 {
		t.Fatalf("activation outbox: %+v %v", pending, err)
	}
	if err := s.MarkPolicyPublished(ctx, pending[0]); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkPolicyPublished(ctx, pending[0]); err != nil {
		t.Fatal(err)
	}

	revoked, err := s.RevokeActiveWithPolicy(ctx, checkpoints, a.TenantID(), a.ID(), 2, 2, "revoke-v1")
	if err != nil || revoked.PolicyRevision != 3 {
		t.Fatalf("revoke journal: %v", err)
	}
	if _, err := guard.Check(ctx, key.TenantID, user, cap.ResourceType, cap.ResourceID, cap.Action); !errors.Is(err, consistency.ErrPolicyProjectionLag) {
		t.Fatalf("in-flight revoke permitted: %v", err)
	}
	if err := journalWorker.Process(ctx, policyjournal.Event{PolicyRevision: revoked.PolicyRevision, Projection: revoked.ProjectionEvent}); err != nil {
		t.Fatal(err)
	}
	check(false)
	cp, err := checkpoints.LoadCheckpoint(ctx, key)
	if err != nil || cp.Revision != 3 || len(cp.Edges) != 0 {
		t.Fatalf("revocation didn't delete graph state: %+v %v", cp, err)
	}
}
