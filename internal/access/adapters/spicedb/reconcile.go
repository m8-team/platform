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
		for _, resourceType := range []string{"organization", "workspace", "environment", "workload_service", "access_capability"} {
			filters = append(filters, &pb.RelationshipFilter{
				ResourceType:          resourceType,
				OptionalSubjectFilter: &pb.SubjectFilter{SubjectType: kind, OptionalSubjectId: obj.ObjectId},
			})
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	collected := make(map[domain.RelationshipKey]domain.Relationship)
	for _, filter := range filters {
		request := &pb.ReadRelationshipsRequest{
			Consistency:        &pb.Consistency{Requirement: &pb.Consistency_FullyConsistent{FullyConsistent: true}},
			RelationshipFilter: filter, OptionalLimit: maxProjectionEdges,
		}
		stream, err := c.permissions.ReadRelationships(ctx, request)
		if err != nil {
			return nil, classifyRPCFailure(ctx, err)
		}
		start := len(collected)
		var cursor *pb.Cursor
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
			cursor = item.AfterResultCursor
		}
		// A full page can hide an extra tuple. Resume the same query and
		// snapshot before accepting a projection at the size boundary.
		if len(collected)-start == maxProjectionEdges {
			if cursor.GetToken() == "" {
				return nil, ErrInvalidResponse
			}
			request.OptionalCursor = cursor
			overflow, err := c.permissions.ReadRelationships(ctx, request)
			if err != nil {
				return nil, classifyRPCFailure(ctx, err)
			}
			_, err = overflow.Recv()
			if err == nil {
				return nil, ErrInvalidResponse
			}
			if !errors.Is(err, io.EOF) {
				return nil, classifyRPCFailure(ctx, err)
			}
		}
	}
	result := make([]domain.Relationship, 0, len(collected))
	for _, edge := range collected {
		result = append(result, edge)
	}
	return result, nil
}
