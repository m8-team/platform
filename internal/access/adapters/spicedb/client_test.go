package spicedb

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const tenantValue = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
const otherTenantValue = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"

type fakePermissions struct {
	pb.PermissionsServiceClient
	check  func(*pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error)
	write  func(*pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error)
	checks int
	writes int
}

func (f *fakePermissions) CheckPermission(_ context.Context, req *pb.CheckPermissionRequest, _ ...grpc.CallOption) (*pb.CheckPermissionResponse, error) {
	f.checks++
	return f.check(req)
}
func (f *fakePermissions) WriteRelationships(_ context.Context, req *pb.WriteRelationshipsRequest, _ ...grpc.CallOption) (*pb.WriteRelationshipsResponse, error) {
	f.writes++
	return f.write(req)
}
func testClient(t *testing.T, f *fakePermissions) *Client {
	t.Helper()
	p, err := domain.NewPermission("m8.workspace.read")
	if err != nil {
		t.Fatal(err)
	}
	d, err := domain.NewPermissionDefinition(p, "workspace", "view")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewStaticPermissionCatalog([]domain.PermissionDefinition{d})
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewWithPermissionsClient("model-v1", catalog, f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func testRequest(t *testing.T, tenantString string, zed string) domain.CheckPermissionRequest {
	t.Helper()
	tn, e := domain.NewTenantID(tenantString)
	if e != nil {
		t.Fatal(e)
	}
	subject, e := domain.NewSubject("user", "same-id")
	if e != nil {
		t.Fatal(e)
	}
	resource, e := domain.NewResource("workspace", "same-id")
	if e != nil {
		t.Fatal(e)
	}
	permission, e := domain.NewPermission("m8.workspace.read")
	if e != nil {
		t.Fatal(e)
	}
	revision, e := domain.NewModelRevision("model-v1")
	if e != nil {
		t.Fatal(e)
	}
	req, e := domain.NewCheckPermissionRequest(domain.CheckPermissionInput{
		TenantID: tn, Subject: subject, Resource: resource, Permission: permission,
		ModelRevision: revision, MinZedToken: zed, Critical: true,
	})
	if e != nil {
		t.Fatal(e)
	}
	return req
}
func TestSDKCheckTenantIsolationAndConsistency(t *testing.T) {
	var observed []*pb.CheckPermissionRequest
	f := &fakePermissions{check: func(r *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
		observed = append(observed, r)
		return &pb.CheckPermissionResponse{CheckedAt: &pb.ZedToken{Token: "zed-v1"}, Permissionship: pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION}, nil
	}}
	client := testClient(t, f)
	for _, tenant := range []string{tenantValue, otherTenantValue} {
		got, err := client.CheckPermission(context.Background(), testRequest(t, tenant, ""))
		if err != nil || !got.Allowed() || got.ZedToken() != "zed-v1" {
			t.Fatalf("decision=%+v error=%v", got, err)
		}
	}
	if len(observed) != 2 || observed[0].Resource.ObjectId == observed[1].Resource.ObjectId {
		t.Fatal("tenants share graph IDs")
	}
	if !strings.HasPrefix(observed[0].Resource.ObjectId, tenantValue+"/") || !strings.HasPrefix(observed[1].Subject.Object.ObjectId, otherTenantValue+"/") {
		t.Fatal("graph identifiers not isolated")
	}
	for _, r := range observed {
		if r.GetConsistency().GetFullyConsistent() != true || r.Permission != "view" || r.Resource.ObjectId != r.Subject.Object.ObjectId {
			t.Fatalf("wrong M8 mapping: %+v", r)
		}
	}
}
func TestSDKCheckZedTokenAndConditionalDenial(t *testing.T) {
	f := &fakePermissions{check: func(r *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
		if r.GetConsistency().GetAtLeastAsFresh().GetToken() != "after-revoke" {
			t.Fatalf("missing watermark: %+v", r)
		}
		return &pb.CheckPermissionResponse{CheckedAt: &pb.ZedToken{Token: "later"}, Permissionship: pb.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION}, nil
	}}
	c := testClient(t, f)
	d, err := c.CheckPermission(context.Background(), testRequest(t, tenantValue, "after-revoke"))
	if err != nil || d.Allowed() || d.Reason() != "spicedb_missing_caveat_context" {
		t.Fatalf("conditional allow: %v %+v", err, d)
	}
}
func TestSDKCheckFailsClosedOnUnknownMissingRevisionAndRPCFailure(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response *pb.CheckPermissionResponse
		failure  error
		want     error
	}{
		{"unknown permissionship", &pb.CheckPermissionResponse{CheckedAt: &pb.ZedToken{Token: "z"}}, nil, ErrInvalidResponse},
		{"missing zed token", &pb.CheckPermissionResponse{Permissionship: pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION}, nil, ErrInvalidResponse},
		{"unavailable", nil, status.Error(codes.Unavailable, "secret from infrastructure"), ports.ErrPermissionEngineUnavailable},
		{"timeout", nil, status.Error(codes.DeadlineExceeded, "secret"), ports.ErrPermissionEngineTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakePermissions{check: func(*pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) { return tc.response, tc.failure }}
			_, err := testClient(t, f).CheckPermission(context.Background(), testRequest(t, tenantValue, ""))
			if !errors.Is(err, tc.want) || strings.Contains(err.Error(), "secret") {
				t.Fatalf("unsafe result: %v", err)
			}
		})
	}
}
func TestSDKCheckRejectsModelMismatchBeforeRPC(t *testing.T) {
	f := &fakePermissions{check: func(*pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
		t.Fatal("unexpected remote call")
		return nil, nil
	}}
	c := testClient(t, f)
	c.revision, _ = domain.NewModelRevision("model-v2")
	_, err := c.CheckPermission(context.Background(), testRequest(t, tenantValue, ""))
	if !errors.Is(err, ErrSchemaRevisionMismatch) || f.checks != 0 {
		t.Fatalf("unexpected: %v", err)
	}
}
func TestSDKWriteAtomicRelationshipsAndToken(t *testing.T) {
	tn, _ := domain.NewTenantID(tenantValue)
	r, _ := domain.NewResource("workspace", "w-1")
	s, _ := domain.NewSubject("user", "u-1")
	edge, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: tn, Resource: r, Relation: "viewer", Subject: s, ConditionRef: "mfa_verified"})
	f := &fakePermissions{write: func(req *pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		if len(req.Updates) != 1 || req.Updates[0].Operation != pb.RelationshipUpdate_OPERATION_TOUCH || req.Updates[0].Relationship.OptionalCaveat.CaveatName != "mfa_verified" {
			t.Fatalf("bad tuple: %+v", req)
		}
		return &pb.WriteRelationshipsResponse{WrittenAt: &pb.ZedToken{Token: "written"}}, nil
	}}
	got, err := testClient(t, f).WriteRelationships(context.Background(), tn, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: edge}})
	if err != nil || got.ZedToken != "written" || f.writes != 1 {
		t.Fatalf("write: %v %+v", err, got)
	}
}
func TestSDKWriteRejectsWrongTenantDuplicateAndBadToken(t *testing.T) {
	tn, _ := domain.NewTenantID(tenantValue)
	other, _ := domain.NewTenantID(otherTenantValue)
	r, _ := domain.NewResource("workspace", "w")
	s, _ := domain.NewSubject("user", "u")
	otherEdge, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: other, Resource: r, Relation: "viewer", Subject: s})
	validEdge, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: tn, Resource: r, Relation: "viewer", Subject: s})
	f := &fakePermissions{write: func(*pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		return &pb.WriteRelationshipsResponse{}, nil
	}}
	c := testClient(t, f)
	if _, err := c.WriteRelationships(context.Background(), tn, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: otherEdge}}); !errors.Is(err, ports.ErrInvalidRelationshipChange) {
		t.Fatalf("cross tenant: %v", err)
	}
	if _, err := c.WriteRelationships(context.Background(), tn, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: validEdge}, {Operation: ports.RelationshipDelete, Relationship: validEdge}}); !errors.Is(err, ports.ErrInvalidRelationshipChange) {
		t.Fatalf("duplicate tuple: %v", err)
	}
	if f.writes != 0 {
		t.Fatal("invalid batch reached remote")
	}
	if _, err := c.WriteRelationships(context.Background(), tn, []ports.RelationshipChange{{Operation: ports.RelationshipTouch, Relationship: validEdge}}); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("missing token: %v", err)
	}
}
func TestSDKConnectionDisallowsRemoteInsecureTransport(t *testing.T) {
	permission, _ := domain.NewPermission("m8.workspace.read")
	def, _ := domain.NewPermissionDefinition(permission, "workspace", "view")
	catalog, _ := NewStaticPermissionCatalog([]domain.PermissionDefinition{def})
	_, err := NewClient(Config{Endpoint: "spicedb.internal:50051", BearerToken: "secret", ModelRevision: "v1", AllowInsecureLoopback: true}, catalog)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("allowed plaintext to remote endpoint: %v", err)
	}
}
