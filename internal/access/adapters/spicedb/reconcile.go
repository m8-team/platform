package spicedb

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"strings"

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

func decodeGraphID(tenant domain.TenantID, id string) (string, error) {
	prefix := string(tenant) + "/"
	if !strings.HasPrefix(id, prefix) {
		return "", ErrInvalidGraphReference
	}
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(id, prefix))
	if err != nil || len(b) == 0 {
		return "", ErrInvalidGraphReference
	}
	return string(b), nil
}

func decodeGraphRelationship(tenant domain.TenantID, rel *pb.Relationship) (domain.Relationship, error) {
	if rel == nil || rel.Resource == nil || rel.Subject == nil || rel.Subject.Object == nil {
		return domain.Relationship{}, ErrInvalidResponse
	}
	resourceID, err := decodeGraphID(tenant, rel.Resource.ObjectId)
	if err != nil {
		return domain.Relationship{}, err
	}
	subjectID, err := decodeGraphID(tenant, rel.Subject.Object.ObjectId)
	if err != nil {
		return domain.Relationship{}, err
	}
	resource, err := domain.NewResource(rel.Resource.ObjectType, resourceID)
	if err != nil {
		return domain.Relationship{}, err
	}
	subject, err := domain.NewSubject(rel.Subject.Object.ObjectType, subjectID)
	if err != nil {
		return domain.Relationship{}, err
	}
	input := domain.RelationshipInput{TenantID: tenant, Resource: resource, Subject: subject, Relation: rel.Relation, SubjectRelation: rel.Subject.OptionalRelation}
	if rel.OptionalCaveat != nil {
		// Caveat context is intentionally unsupported: do not lose it on a
		// delta. Refuse reconciliation until the trusted evaluator is added.
		if rel.OptionalCaveat.Context != nil {
			return domain.Relationship{}, ErrInvalidResponse
		}
		input.ConditionRef = rel.OptionalCaveat.CaveatName
	}
	if rel.OptionalExpiresAt != nil {
		if err := rel.OptionalExpiresAt.CheckValid(); err != nil {
			return domain.Relationship{}, ErrInvalidResponse
		}
		input.ExpiresAt = rel.OptionalExpiresAt.AsTime()
	}
	return domain.NewRelationship(input)
}
