package spicedb

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrGraphFenceConflict = errors.New("SpiceDB projection graph fence conflict")

var _ ports.GraphFenceInstaller = (*Client)(nil)

func fenceResource(f ports.GraphFence) (*pb.ObjectReference, error) {
	return graphObject(f.TenantID, "projection_guard", f.Kind+":"+f.AggregateID)
}

func fenceSubject(f ports.GraphFence) (*pb.SubjectReference, error) {
	obj, err := graphObject(f.TenantID, "projection_epoch", strconv.FormatUint(f.Epoch, 10))
	if err != nil {
		return nil, err
	}
	return &pb.SubjectReference{Object: obj}, nil
}

func fenceFilter(resource *pb.ObjectReference, subject *pb.SubjectReference) *pb.RelationshipFilter {
	filter := &pb.RelationshipFilter{ResourceType: resource.ObjectType, OptionalResourceId: resource.ObjectId, OptionalRelation: "owner"}
	if subject != nil {
		filter.OptionalSubjectFilter = &pb.SubjectFilter{SubjectType: subject.Object.ObjectType, OptionalSubjectId: subject.Object.ObjectId}
	}
	return filter
}

func parseFenceEpoch(f ports.GraphFence, rel *pb.Relationship) (uint64, error) {
	if rel == nil || rel.Subject == nil || rel.Subject.Object == nil || rel.Subject.Object.ObjectType != "projection_epoch" || rel.Subject.OptionalRelation != "" || rel.Relation != "owner" {
		return 0, ErrInvalidResponse
	}
	encoded := rel.Subject.Object.ObjectId
	prefix := string(f.TenantID) + "/"
	if !strings.HasPrefix(encoded, prefix) {
		return 0, ErrGraphFenceConflict
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(encoded, prefix))
	if err != nil {
		return 0, ErrGraphFenceConflict
	}
	epoch, err := strconv.ParseUint(string(raw), 10, 64)
	if err != nil || epoch == 0 {
		return 0, ErrGraphFenceConflict
	}
	return epoch, nil
}

// EstablishFence installs a strictly increasing epoch using a graph-side CAS.
// An outdated worker can neither install its epoch after a takeover nor write
// relationships because all subsequent writes require the current guard tuple.
func (c *Client) EstablishFence(ctx context.Context, fence ports.GraphFence) (string, error) {
	if err := fence.Validate(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	resource, err := fenceResource(fence)
	if err != nil {
		return "", err
	}
	stream, err := c.permissions.ReadRelationships(ctx, &pb.ReadRelationshipsRequest{
		Consistency:        &pb.Consistency{Requirement: &pb.Consistency_FullyConsistent{FullyConsistent: true}},
		RelationshipFilter: fenceFilter(resource, nil), OptionalLimit: 2,
	})
	if err != nil {
		return "", classifyRPCFailure(ctx, err)
	}
	var observed *pb.Relationship
	var readToken string
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", classifyRPCFailure(ctx, err)
		}
		if msg == nil || msg.Relationship == nil || observed != nil {
			return "", ErrGraphFenceConflict
		}
		observed = msg.Relationship
		if msg.ReadAt != nil {
			readToken = msg.ReadAt.Token
		}
	}
	target, err := fenceSubject(fence)
	if err != nil {
		return "", err
	}
	if observed != nil {
		epoch, err := parseFenceEpoch(fence, observed)
		if err != nil {
			return "", err
		}
		if epoch > fence.Epoch {
			return "", ErrGraphFenceConflict
		}
		if epoch == fence.Epoch {
			if readToken == "" {
				return "", ErrInvalidResponse
			}
			return readToken, nil
		}
	}
	var updates []*pb.RelationshipUpdate
	var precondition *pb.Precondition
	if observed != nil {
		precondition = &pb.Precondition{Operation: pb.Precondition_OPERATION_MUST_MATCH, Filter: fenceFilter(resource, observed.Subject)}
		updates = append(updates, &pb.RelationshipUpdate{Operation: pb.RelationshipUpdate_OPERATION_DELETE, Relationship: observed})
	} else {
		precondition = &pb.Precondition{Operation: pb.Precondition_OPERATION_MUST_NOT_MATCH, Filter: fenceFilter(resource, nil)}
	}
	updates = append(updates, &pb.RelationshipUpdate{Operation: pb.RelationshipUpdate_OPERATION_TOUCH, Relationship: &pb.Relationship{
		Resource: resource, Relation: "owner", Subject: target,
	}})
	resp, err := c.permissions.WriteRelationships(ctx, &pb.WriteRelationshipsRequest{Updates: updates, OptionalPreconditions: []*pb.Precondition{precondition}})
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return "", ErrGraphFenceConflict
		}
		return "", classifyRPCFailure(ctx, err)
	}
	if resp == nil || resp.WrittenAt == nil || resp.WrittenAt.Token == "" {
		return "", fmt.Errorf("%w: missing fence write token", ErrInvalidResponse)
	}
	return resp.WrittenAt.Token, nil
}
