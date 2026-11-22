package spicedb

import (
	"context"
	"testing"
	"time"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

func TestSDKExpirationAndDeletionByTupleIdentity(t *testing.T) {
	tn, _ := domain.NewTenantID(tenantValue)
	obj, _ := domain.NewResource("access_grant", "11111111-1111-4111-8111-111111111111")
	subject, _ := domain.NewSubject("user", "alice")
	expiry := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	rel, _ := domain.NewRelationship(domain.RelationshipInput{TenantID: tn, Resource: obj, Relation: "holder", Subject: subject, ExpiresAt: expiry})
	calls := 0
	f := &fakePermissions{write: func(req *pb.WriteRelationshipsRequest) (*pb.WriteRelationshipsResponse, error) {
		calls++
		got := req.Updates[0].Relationship
		if calls == 1 {
			if got.OptionalExpiresAt == nil || !got.OptionalExpiresAt.AsTime().Equal(expiry) {
				t.Fatalf("missing expiry: %+v", got)
			}
		} else {
			if got.OptionalExpiresAt != nil || got.OptionalCaveat != nil || req.Updates[0].Operation != pb.RelationshipUpdate_OPERATION_DELETE {
				t.Fatalf("delete identity: %+v", got)
			}
		}
		return &pb.WriteRelationshipsResponse{WrittenAt: &pb.ZedToken{Token: "zed-write"}}, nil
	}}
	c := testClient(t, f)
	for _, op := range []ports.RelationshipOperation{ports.RelationshipTouch, ports.RelationshipDelete} {
		if _, err := c.WriteRelationships(context.Background(), tn, []ports.RelationshipChange{{Operation: op, Relationship: rel}}); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}
