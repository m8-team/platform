package spicedb

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

// WriteRelationships atomically applies changes in a single SpiceDB request.
// It does not publish domain events or advance an M8 source watermark.
func (c *Client) WriteRelationships(ctx context.Context, tenant domain.TenantID, changes []ports.RelationshipChange) (ports.RelationshipWriteResult, error) {
	if err := tenant.Validate(); err != nil {
		return ports.RelationshipWriteResult{}, err
	}
	if len(changes) == 0 || len(changes) > 1000 {
		return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
	}
	type update struct {
		Operation    string            `json:"operation"`
		Relationship graphRelationship `json:"relationship"`
	}
	request := struct {
		Updates []update `json:"updates"`
	}{Updates: make([]update, 0, len(changes))}
	seen := make(map[domain.RelationshipKey]struct{}, len(changes))
	for _, change := range changes {
		if change.Relationship.TenantID() != tenant {
			return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
		}
		if _, exists := seen[change.Relationship.Key()]; exists {
			return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
		}
		seen[change.Relationship.Key()] = struct{}{}
		var operation string
		switch change.Operation {
		case ports.RelationshipCreate:
			operation = "OPERATION_CREATE"
		case ports.RelationshipTouch:
			operation = "OPERATION_TOUCH"
		case ports.RelationshipDelete:
			operation = "OPERATION_DELETE"
		default:
			return ports.RelationshipWriteResult{}, ports.ErrInvalidRelationshipChange
		}
		rel, err := mapRelationship(change.Relationship)
		if err != nil {
			return ports.RelationshipWriteResult{}, err
		}
		request.Updates = append(request.Updates, update{Operation: operation, Relationship: rel})
	}
	var response struct {
		WrittenAt struct {
			Token string `json:"token"`
		} `json:"writtenAt"`
	}
	if err := c.post(ctx, "/v1/relationships/write", request, &response); err != nil {
		return ports.RelationshipWriteResult{}, err
	}
	if response.WrittenAt.Token == "" {
		return ports.RelationshipWriteResult{}, fmt.Errorf("%w: missing writtenAt", ErrInvalidResponse)
	}
	return ports.RelationshipWriteResult{ZedToken: response.WrittenAt.Token}, nil
}
