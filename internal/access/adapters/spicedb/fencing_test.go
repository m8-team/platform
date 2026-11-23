package spicedb

import (
	"context"
	"errors"
	"io"
	"strconv"
	"testing"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeFenceReadStream struct {
	grpc.ServerStreamingClient[pb.ReadRelationshipsResponse]
	items []*pb.ReadRelationshipsResponse
}

func (s *fakeFenceReadStream) Recv() (*pb.ReadRelationshipsResponse, error) {
	if len(s.items) == 0 {
		return nil, io.EOF
	}
	item := s.items[0]
	s.items = s.items[1:]
	return item, nil
}

type fakeFencePermissions struct {
	*fakePermissions
	read func(*pb.ReadRelationshipsRequest) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error)
}

func (f *fakeFencePermissions) ReadRelationships(_ context.Context, req *pb.ReadRelationshipsRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error) {
	return f.read(req)
}
func testFence(t *testing.T, epoch uint64) ports.GraphFence {
	t.Helper()
	tn, e := domain.NewTenantID(tenantValue)
	if e != nil {
		t.Fatal(e)
	}
	return ports.GraphFence{TenantID: tn, Kind: "ACCESS_GRANT", AggregateID: "11111111-1111-4111-8111-111111111111", Epoch: epoch}
}
func testFencedSDK(t *testing.T, s pb.PermissionsServiceClient) *Client {
	t.Helper()
	basic := testClient(t, &fakePermissions{})
	c, e := NewWithPermissionsClient("model-v1", basic.catalog, s)
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestEstablishGraphFenceFromEmptyGraph(t *testing.T) {
	fence := testFence(t, 1)
	fake := &fakeFencePermissions{fakePermissions: &fakePermissions{write: func(req *pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		if len(req.Updates) != 1 || len(req.OptionalPreconditions) != 1 || req.OptionalPreconditions[0].Operation != pb.Precondition_OPERATION_MUST_NOT_MATCH {
			t.Fatalf("unsafe fence install %+v", req)
		}
		if req.Updates[0].Relationship.GetSubject().GetObject().GetObjectType() != "projection_epoch" {
			t.Fatal("missing monotonic epoch subject")
		}
		return &pb.WriteRelationshipsResponse{WrittenAt: &pb.ZedToken{Token: "fence-zed"}}, nil
	}}}
	fake.read = func(req *pb.ReadRelationshipsRequest) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error) {
		if !req.GetConsistency().GetFullyConsistent() {
			t.Fatal("fence read must be strong")
		}
		return &fakeFenceReadStream{}, nil
	}
	zed, err := testFencedSDK(t, fake).EstablishFence(context.Background(), fence)
	if err != nil || zed != "fence-zed" || fake.writes != 1 {
		t.Fatalf("fence not installed: %q %v", zed, err)
	}
}
func TestStaleEpochCannotReplaceNewerGraphFence(t *testing.T) {
	older, newer := testFence(t, 4), testFence(t, 5)
	obj, e := fenceResource(newer)
	if e != nil {
		t.Fatal(e)
	}
	subject, e := fenceSubject(newer)
	if e != nil {
		t.Fatal(e)
	}
	fake := &fakeFencePermissions{fakePermissions: &fakePermissions{write: func(*pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		t.Fatal("stale writer called graph write")
		return nil, nil
	}}}
	fake.read = func(*pb.ReadRelationshipsRequest) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error) {
		return &fakeFenceReadStream{items: []*pb.ReadRelationshipsResponse{{Relationship: &pb.Relationship{Resource: obj, Relation: "owner", Subject: subject}, ReadAt: &pb.ZedToken{Token: "read"}}}}, nil
	}
	_, err := testFencedSDK(t, fake).EstablishFence(context.Background(), older)
	if !errors.Is(err, ErrGraphFenceConflict) {
		t.Fatalf("stale epoch installed: %v", err)
	}
}
func TestGraphWriteRequiresMatchingEpochPrecondition(t *testing.T) {
	fence := testFence(t, 3)
	grant, e := domain.NewResource("access_grant", fence.AggregateID)
	if e != nil {
		t.Fatal(e)
	}
	user, e := domain.NewSubject("user", "alice")
	if e != nil {
		t.Fatal(e)
	}
	edge, e := domain.NewRelationship(domain.RelationshipInput{TenantID: fence.TenantID, Resource: grant, Relation: "holder", Subject: user})
	if e != nil {
		t.Fatal(e)
	}
	fake := &fakePermissions{write: func(req *pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		if len(req.OptionalPreconditions) != 1 || req.OptionalPreconditions[0].Operation != pb.Precondition_OPERATION_MUST_MATCH {
			t.Fatalf("missing atomic write precondition %+v", req)
		}
		return &pb.WriteRelationshipsResponse{WrittenAt: &pb.ZedToken{Token: "written"}}, nil
	}}
	ctx, e := ports.WithGraphFence(context.Background(), fence)
	if e != nil {
		t.Fatal(e)
	}
	_, e = testFencedSDK(t, fake).WriteRelationships(ctx, fence.TenantID, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: edge}})
	if e != nil {
		t.Fatal(e)
	}
	other, e := domain.NewResource("access_grant", "22222222-2222-4222-8222-222222222222")
	if e != nil {
		t.Fatal(e)
	}
	foreign, e := domain.NewRelationship(domain.RelationshipInput{TenantID: fence.TenantID, Resource: other, Relation: "holder", Subject: user})
	if e != nil {
		t.Fatal(e)
	}
	_, e = testFencedSDK(t, fake).WriteRelationships(ctx, fence.TenantID, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: foreign}})
	if !errors.Is(e, ports.ErrInvalidGraphFence) {
		t.Fatalf("fence applied to wrong aggregate: %v", e)
	}
}
func TestFailedFenceWriteCannotPassAsSuccess(t *testing.T) {
	fence := testFence(t, 3)
	grant, _ := domain.NewResource("access_grant", fence.AggregateID)
	user, _ := domain.NewSubject("user", "alice")
	edge, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: fence.TenantID, Resource: grant, Relation: "holder", Subject: user})
	fake := &fakePermissions{write: func(*pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		return nil, status.Error(codes.FailedPrecondition, "older epoch")
	}}
	ctx, _ := ports.WithGraphFence(context.Background(), fence)
	_, err := testFencedSDK(t, fake).WriteRelationships(ctx, fence.TenantID, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: edge}})
	if !errors.Is(err, ErrGraphFenceConflict) {
		t.Fatalf("stale write not rejected: %v", err)
	}
}

func TestReadProjectionEdgesReturnsOwnedTuples(t *testing.T) {
	fence := testFence(t, 4)
	grantObject, _ := graphObject(fence.TenantID, "access_grant", fence.AggregateID)
	userObject, _ := graphObject(fence.TenantID, "user", "alice")
	workspaceObject, _ := graphObject(fence.TenantID, "workspace", "workspace-id")
	fake := &fakeFencePermissions{fakePermissions: &fakePermissions{}}
	fake.read = func(req *pb.ReadRelationshipsRequest) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error) {
		if !req.GetConsistency().GetFullyConsistent() {
			t.Fatal("reconciliation read must be fully consistent")
		}
		filter := req.GetRelationshipFilter()
		var rel *pb.Relationship
		switch filter.ResourceType {
		case "access_grant":
			rel = &pb.Relationship{Resource: grantObject, Relation: "holder", Subject: &pb.SubjectReference{Object: userObject}}
		case "workspace":
			rel = &pb.Relationship{Resource: workspaceObject, Relation: "viewer", Subject: &pb.SubjectReference{Object: grantObject, OptionalRelation: "holder"}}
		}
		if rel == nil {
			return &fakeFenceReadStream{}, nil
		}
		return &fakeFenceReadStream{items: []*pb.ReadRelationshipsResponse{{Relationship: rel}}}, nil
	}
	edges, err := testFencedSDK(t, fake).ReadProjectionEdges(context.Background(), fence)
	if err != nil || len(edges) != 2 {
		t.Fatalf("lost graph edges: %v %d", err, len(edges))
	}
}
func TestReadProjectionEdgesRejectsCrossTenantTuple(t *testing.T) {
	fence := testFence(t, 4)
	obj, _ := graphObject(fence.TenantID, "access_grant", fence.AggregateID)
	other, _ := domain.NewTenantID(otherTenantValue)
	user, _ := graphObject(other, "user", "alice")
	fake := &fakeFencePermissions{fakePermissions: &fakePermissions{}}
	fake.read = func(req *pb.ReadRelationshipsRequest) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error) {
		if req.GetRelationshipFilter().ResourceType != "access_grant" {
			return &fakeFenceReadStream{}, nil
		}
		return &fakeFenceReadStream{items: []*pb.ReadRelationshipsResponse{{Relationship: &pb.Relationship{Resource: obj, Relation: "holder", Subject: &pb.SubjectReference{Object: user}}}}}, nil
	}
	_, err := testFencedSDK(t, fake).ReadProjectionEdges(context.Background(), fence)
	if !errors.Is(err, ErrInvalidGraphReference) {
		t.Fatalf("cross-tenant edge accepted: %v", err)
	}
}

func TestFenceEpochIsBoundToTenant(t *testing.T) {
	fence := testFence(t, 7)
	other := fence
	other.TenantID = otherTenantValue
	subject, err := fenceSubject(other)
	if err != nil {
		t.Fatal(err)
	}
	_, err = parseFenceEpoch(fence, &pb.Relationship{Relation: "owner", Subject: subject})
	if !errors.Is(err, ErrGraphFenceConflict) {
		t.Fatalf("foreign epoch returned %v, want fence conflict", err)
	}
}

func TestReadProjectionEdgesChecksFullPageForOverflow(t *testing.T) {
	fence := testFence(t, 1)
	grant, err := graphObject(fence.TenantID, "access_grant", fence.AggregateID)
	if err != nil {
		t.Fatal(err)
	}
	page := make([]*pb.ReadRelationshipsResponse, maxProjectionEdges)
	for i := range page {
		user, err := graphObject(fence.TenantID, "user", strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		page[i] = &pb.ReadRelationshipsResponse{Relationship: &pb.Relationship{Resource: grant, Relation: "holder", Subject: &pb.SubjectReference{Object: user}}, AfterResultCursor: &pb.Cursor{Token: strconv.Itoa(i + 1)}}
	}
	for _, size := range []int{maxProjectionEdges, maxProjectionEdges + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			fake := &fakeFencePermissions{fakePermissions: &fakePermissions{}}
			probed := false
			fake.read = func(req *pb.ReadRelationshipsRequest) (grpc.ServerStreamingClient[pb.ReadRelationshipsResponse], error) {
				if req.OptionalLimit != maxProjectionEdges {
					t.Fatalf("unsupported page limit %d", req.OptionalLimit)
				}
				if req.RelationshipFilter.ResourceType != "access_grant" {
					return &fakeFenceReadStream{}, nil
				}
				if req.OptionalCursor == nil {
					return &fakeFenceReadStream{items: page}, nil
				}
				if req.OptionalCursor.Token != page[len(page)-1].AfterResultCursor.Token {
					t.Fatal("wrong continuation cursor")
				}
				probed = true
				return &fakeFenceReadStream{items: page[:size-maxProjectionEdges]}, nil
			}
			edges, err := testFencedSDK(t, fake).ReadProjectionEdges(t.Context(), fence)
			if !probed {
				t.Fatal("full page accepted without checking continuation")
			}
			if size == maxProjectionEdges {
				if err != nil || len(edges) != size {
					t.Fatalf("exact limit: %d edges, %v", len(edges), err)
				}
			} else if !errors.Is(err, ErrInvalidResponse) || edges != nil {
				t.Fatalf("overflow: %d edges, %v", len(edges), err)
			}
		})
	}
}
