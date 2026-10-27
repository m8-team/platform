package grpcadapter

import (
	"context"
	"net"
	"testing"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// Exercises real generated gRPC clients, service registration, and transport.
// The interceptor is a TEST authentication stub, not an HTTP metadata bypass.
func TestOrgStructureGeneratedGRPCTransport(t *testing.T) {
	handler := orgGRPCFixture(t, true)
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		next grpc.UnaryHandler,
	) (any, error) {
		return next(WithOrgStructureActor(ctx, orgstructure.Actor{
			TenantID:    "tenant",
			PrincipalID: "test-service",
		}), request)
	}))
	if err := newOrgStructureRegistration(handler)(server); err != nil {
		t.Fatal(err)
	}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	conn, err := grpc.NewClient(
		"passthrough:///org-structure-test",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	queries := pb.NewOrgStructureServiceClient(conn)
	writes := pb.NewOrgUnitChangeSetServiceClient(conn)
	ctx := context.Background()

	before, err := queries.GetOrgUnit(ctx, &pb.GetOrgUnitRequest{
		OrganizationId: "org",
		OrgUnitId:      "a",
	})
	if err != nil || before.GetMetadata().GetName() != "Sales" {
		t.Fatalf("unexpected transport Get: %#v %v", before, err)
	}

	plan := &pb.OrgUnitChangeSet{
		Id:                        "cs_transport",
		OrganizationId:            "org",
		ExpectedStructureRevision: 1,
		Mutations: []*pb.OrgUnitMutation{{
			Kind:        pb.OrgUnitMutationKind_ORG_UNIT_MUTATION_KIND_RENAME,
			OrgUnitId:   "a",
			DisplayName: "Revenue Operations",
		}},
	}
	applied, err := writes.ApplyOrgUnitChangeSet(ctx, &pb.ApplyOrgUnitChangeSetRequest{
		ChangeSet:      plan,
		IdempotencyKey: plan.GetId(),
	})
	if err != nil || applied.GetStructureRevision() != 2 {
		t.Fatalf("unexpected transport Apply: %#v %v", applied, err)
	}

	after, err := queries.GetOrgUnit(ctx, &pb.GetOrgUnitRequest{
		OrganizationId: "org",
		OrgUnitId:      "a",
	})
	if err != nil || after.GetMetadata().GetName() != "Revenue Operations" {
		t.Fatalf("unexpected transport Get after Apply: %#v %v", after, err)
	}
}
