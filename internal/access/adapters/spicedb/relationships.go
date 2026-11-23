package spicedb

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

// WriteRelationships writes a single atomic SpiceDB transaction via the
// official Go SDK. No source watermark is advanced by this adapter.
func (c *Client) WriteRelationships(ctx context.Context, tenant domain.TenantID, changes []ports.RelationshipChange) (ports.RelationshipWriteResult, error) {
	if err := ctx.Err(); err != nil {
		return ports.RelationshipWriteResult{}, err
	}
	if err := tenant.Validate(); err != nil {
		return ports.RelationshipWriteResult{}, err
	}
	if len(changes) == 0 || len(changes) > 1000 {
		return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
	}
	updates := make([]*pb.RelationshipUpdate, 0, len(changes))
	seen := make(map[domain.RelationshipKey]struct{}, len(changes))
	for _, change := range changes {
		if change.Relationship.TenantID() != tenant {
			return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
		}
		key := change.Relationship.Key()
		if _, ok := seen[key]; ok {
			return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
		}
		seen[key] = struct{}{}
		var operation pb.RelationshipUpdate_Operation
		switch change.Operation {
		case ports.RelationshipCreate:
			operation = pb.RelationshipUpdate_OPERATION_CREATE
		case ports.RelationshipTouch:
			operation = pb.RelationshipUpdate_OPERATION_TOUCH
		case ports.RelationshipDelete:
			operation = pb.RelationshipUpdate_OPERATION_DELETE
		default:
			return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
		}
		if fence, guarded := ports.GraphFenceFromContext(ctx); guarded {
			r := change.Relationship
			kind, ownerID, err := fence.OwnedObject()
			if err != nil {return ports.RelationshipWriteResult{}, err}
			owns := (r.Resource().Type() == kind && r.Resource().ID() == ownerID) || (kind != "rm_projection" && r.Subject().Type() == kind && r.Subject().ID() == ownerID)
			if !owns {
				return ports.RelationshipWriteResult{}, ports.ErrInvalidGraphFence
			}
		}
		rel, err := mapRelationship(change.Relationship)
		if err != nil {
			return ports.RelationshipWriteResult{}, err
		}
		if change.Operation == ports.RelationshipDelete {
			// Delete by tuple identity; expiry/caveat metadata must be absent.
			rel.OptionalExpiresAt = nil
			rel.OptionalCaveat = nil
		}
		updates = append(updates, &pb.RelationshipUpdate{Operation: operation, Relationship: rel})
	}
	request := &pb.WriteRelationshipsRequest{Updates: updates}
	if fence, ok := ports.GraphFenceFromContext(ctx); ok {
		if err := fence.Validate(); err != nil || fence.TenantID != tenant {
			return ports.RelationshipWriteResult{}, ports.ErrInvalidGraphFence
		}
		guard, err := fenceResource(fence)
		if err != nil {
			return ports.RelationshipWriteResult{}, err
		}
		owner, err := fenceSubject(fence)
		if err != nil {
			return ports.RelationshipWriteResult{}, err
		}
		request.OptionalPreconditions = []*pb.Precondition{{Operation: pb.Precondition_OPERATION_MUST_MATCH, Filter: fenceFilter(guard, owner)}}
	}
	response, err := c.permissions.WriteRelationships(ctx, request)
	if err != nil {
		if _, guarded := ports.GraphFenceFromContext(ctx); guarded && status.Code(err) == codes.FailedPrecondition {
			return ports.RelationshipWriteResult{}, ErrGraphFenceConflict
		}
		return ports.RelationshipWriteResult{}, classifyRPCFailure(ctx, err)
	}
	if response == nil || response.WrittenAt == nil || response.WrittenAt.Token == "" {
		return ports.RelationshipWriteResult{}, fmt.Errorf("%w: missing writtenAt", ErrInvalidResponse)
	}
	return ports.RelationshipWriteResult{ZedToken: response.WrittenAt.Token}, nil
}
