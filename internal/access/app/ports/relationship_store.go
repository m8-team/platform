package ports

import (
	"context"
	"errors"

	"github.com/m8-team/platform/internal/access/domain"
)

var ErrInvalidRelationshipChange = errors.New("invalid relationship change batch")

type RelationshipOperation string

const (
	RelationshipCreate RelationshipOperation = "CREATE"
	RelationshipTouch  RelationshipOperation = "TOUCH"
	RelationshipDelete RelationshipOperation = "DELETE"
)

type RelationshipChange struct {
	Operation    RelationshipOperation
	Relationship domain.Relationship
}

// RelationshipWriteResult carries the SpiceDB graph revision produced by an
// atomic write. The application must durably associate it with source progress
// before claiming a projection is available to readers.
type RelationshipWriteResult struct {
	ZedToken string
}

// RelationshipWriter accepts bounded, single-tenant, atomic change sets. The
// application owns authorization, outbox/inbox, retries, and source revisions.
type RelationshipWriter interface {
	WriteRelationships(ctx context.Context, tenant domain.TenantID, changes []RelationshipChange) (RelationshipWriteResult, error)
}
