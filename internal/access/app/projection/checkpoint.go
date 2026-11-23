package projection

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
)

// RecordedEdge is the durable, typed representation of an M8 relationship.
// Saving a full edge snapshot makes crash replay and delta reconstruction
// independent of historical role/grant revisions.
type RecordedEdge struct {
	ResourceType    string    `json:"resource_type"`
	ResourceID      string    `json:"resource_id"`
	Relation        string    `json:"relation"`
	SubjectType     string    `json:"subject_type"`
	SubjectID       string    `json:"subject_id"`
	SubjectRelation string    `json:"subject_relation,omitempty"`
	ConditionRef    string    `json:"condition_ref,omitempty"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
}

func recordEdges(edges []domain.Relationship) []RecordedEdge {
	out := make([]RecordedEdge, 0, len(edges))
	for _, r := range edges {
		out = append(out, RecordedEdge{
			ResourceType: r.Resource().Type(), ResourceID: r.Resource().ID(), Relation: r.Relation(), SubjectType: r.Subject().Type(), SubjectID: r.Subject().ID(), SubjectRelation: r.SubjectRelation(), ConditionRef: r.ConditionRef(), ExpiresAt: r.ExpiresAt(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]) < fmt.Sprint(out[j]) })
	return out
}
func (e RecordedEdge) decode(tenant domain.TenantID) (domain.Relationship, error) {
	r, err := domain.NewResource(e.ResourceType, e.ResourceID)
	if err != nil {
		return domain.Relationship{}, err
	}
	s, err := domain.NewSubject(e.SubjectType, e.SubjectID)
	if err != nil {
		return domain.Relationship{}, err
	}
	return domain.NewRelationship(domain.RelationshipInput{TenantID: tenant, Resource: r, Relation: e.Relation, Subject: s, SubjectRelation: e.SubjectRelation, ConditionRef: e.ConditionRef, ExpiresAt: e.ExpiresAt})
}

// Checkpoint is committed ONLY after the entire SpiceDB batch returned a
// ZedToken. Replaying a write after a crash is safe (TOUCH/DELETE are idempotent).
type Checkpoint struct {
	Key        Key            `json:"key"`
	Revision   int64          `json:"revision"`
	EventID    string         `json:"event_id,omitempty"`
	ZedToken   string         `json:"zed_token,omitempty"`
	FenceEpoch uint64         `json:"fence_epoch,omitempty"`
	Edges      []RecordedEdge `json:"edges,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty"`
}

func (c Checkpoint) Validate() error {
	if err := c.Key.Validate(); err != nil {
		return err
	}
	if c.Revision < 0 || (c.Revision > 0 && c.EventID == "") || (c.FenceEpoch != 0 && c.ZedToken == "") {
		return ErrInvalidCheckpoint
	}
	if len(c.Edges) > 1000 {
		return ErrInvalidCheckpoint
	}
	seen := map[domain.RelationshipKey]struct{}{}
	for _, edge := range c.Edges {
		r, err := edge.decode(c.Key.TenantID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidCheckpoint, err)
		}
		if _, ok := seen[r.Key()]; ok {
			return ErrInvalidCheckpoint
		}
		seen[r.Key()] = struct{}{}
	}
	return nil
}
func (c Checkpoint) relationships() ([]domain.Relationship, error) {
	out := make([]domain.Relationship, 0, len(c.Edges))
	for _, edge := range c.Edges {
		r, err := edge.decode(c.Key.TenantID)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// CheckpointStore MUST hold an exclusive lock for Key for the full callback,
// including the remote graph write and a durable commit. It must not return nil
// until the checkpoint has been fsynced/committed. No distributed transaction
// with SpiceDB is assumed.
type CheckpointStore interface {
	WithLocked(ctx context.Context, key Key, fn func(Checkpoint) (Checkpoint, error)) error
}

// CheckpointReader is used by consistency barriers to retrieve the durable
// graph ZedToken corresponding to a source revision. A source watermark must
// be assessed separately; this token alone does not prove source freshness.
type CheckpointReader interface {
	LoadCheckpoint(ctx context.Context, key Key) (Checkpoint, error)
}
