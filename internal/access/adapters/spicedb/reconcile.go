package spicedb

import (
	"context"
	"errors"
	"io"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

var _ ports.GraphProjectionReader = (*Client)(nil)

const maxProjectionEdges = 1000

// ReadProjectionEdges reads all graph tuples owned by a single Access binding
// or grant. This repairs crash windows where SpiceDB committed but YDB did not.
// Every query is fully consistent and limited to prevent unbounded scans.
func (c *Client) ReadProjectionEdges(ctx context.Context, f ports.GraphFence) ([]domain.Relationship, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	kind, ownerID, err := f.OwnedObject()
	if err != nil {
		return nil, err
	}
	obj, err := graphObject(f.TenantID, kind, ownerID)
	if err != nil {
		return nil, err
	}
	filters := []*pb.RelationshipFilter{{ResourceType: kind, OptionalResourceId: obj.ObjectId}}
	// v1 schema resource kinds. Do not query all namespaces or trust
	// client-supplied resource-type strings.
	if kind != "rm_projection" {
		for _, resourceType := range []string{"organization", "workspace", "environment", "workload_service"} {
			filters = append(filters, &pb.RelationshipFilter{
				ResourceType:          resourceType,
				OptionalSubjectFilter: &pb.SubjectFilter{SubjectType: kind, OptionalSubjectId: obj.ObjectId},
			})
		}
	}
	collected := make(map[domain.RelationshipKey]domain.Relationship)
	for _, filter := range filters {
		stream, err := c.permissions.ReadRelationships(ctx, &pb.ReadRelationshipsRequest{
			Consistency:        &pb.Consistency{Requirement: &pb.Consistency_FullyConsistent{FullyConsistent: true}},
			RelationshipFilter: filter, OptionalLimit: maxProjectionEdges + 1,
		})
		if err != nil {
			return nil, classifyRPCFailure(ctx, err)
		}
		for {
			item, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return nil, classifyRPCFailure(ctx, err)
			}
			if item == nil || item.Relationship == nil {
				return nil, ErrInvalidResponse
			}
			edge, err := decodeGraphRelationship(f.TenantID, item.Relationship)
			if err != nil {
				return nil, err
			}
			// Do not accept adjacent but unowned tuples in query results.
			r, s := edge.Resource(), edge.Subject()
			if !((r.Type() == kind && r.ID() == ownerID) || (kind != "rm_projection" && s.Type() == kind && s.ID() == ownerID)) {
				return nil, ErrInvalidResponse
			}
			if _, exists := collected[edge.Key()]; exists {
				return nil, ErrInvalidResponse
			}
			collected[edge.Key()] = edge
			if len(collected) > maxProjectionEdges {
				return nil, ErrInvalidResponse
			}
		}
	}
	result := make([]domain.Relationship, 0, len(collected))
	for _, edge := range collected {
		result = append(result, edge)
	}
	return result, nil
}
