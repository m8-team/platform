package projection

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrInvalidEvent      = errors.New("invalid projection event")
	ErrRevisionGap       = errors.New("projection source revision gap; reconciliation required")
	ErrSourceBehind      = errors.New("authoritative projection snapshot is behind event")
	ErrSourceAhead       = errors.New("authoritative projection snapshot is ahead of event; reconcile required")
	ErrInvalidCheckpoint = errors.New("invalid projection checkpoint")
)

type Kind string

const (
	KindRoleBinding Kind = "ROLE_BINDING"
	KindGrant       Kind = "ACCESS_GRANT"
)

// Key is the independent graph ownership boundary. Role changes MUST fan out
// one new monotonically revisioned projection event per affected RoleBinding.
type Key struct {
	TenantID    domain.TenantID `json:"tenant_id"`
	Kind        Kind            `json:"kind"`
	AggregateID string          `json:"aggregate_id"`
}

var canonicalAggregateID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (k Key) Validate() error {
	if err := k.TenantID.Validate(); err != nil {
		return err
	}
	if k.Kind != KindRoleBinding && k.Kind != KindGrant {
		return ErrInvalidEvent
	}
	if !canonicalAggregateID.MatchString(k.AggregateID) {
		return ErrInvalidEvent
	}
	return nil
}
func (k Key) String() string { return string(k.TenantID) + "/" + string(k.Kind) + "/" + k.AggregateID }

type Event struct {
	EventID  string `json:"event_id"`
	Key      Key    `json:"key"`
	Revision int64  `json:"revision"`
}

func (e Event) Validate() error {
	if err := e.Key.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(e.EventID) == "" || len(e.EventID) > 256 || e.Revision < 1 {
		return ErrInvalidEvent
	}
	return nil
}

// Snapshot must be read from the authoritative Access store, not trusted from
// Kafka payloads. Revision is a per-Key projection sequence, including role edits.
// A tombstone carries a positive revision and nil RoleBinding / Grant.
type Snapshot struct {
	Key         Key
	Revision    int64
	RoleBinding *RoleBindingState
	Grant       *domain.AccessGrant
}

func (s Snapshot) Validate() error {
	if err := s.Key.Validate(); err != nil {
		return err
	}
	if s.Revision < 1 {
		return ErrInvalidEvent
	}
	if s.Key.Kind == KindRoleBinding {
		if s.Grant != nil {
			return ErrInvalidEvent
		}
		if s.RoleBinding != nil {
			if err := s.RoleBinding.Binding.ValidateForRole(s.RoleBinding.Role); err != nil {
				return err
			}
			if s.RoleBinding.Binding.ID() != s.Key.AggregateID || s.RoleBinding.Binding.TenantID() != s.Key.TenantID {
				return ErrInvalidEvent
			}
		}
	} else {
		if s.RoleBinding != nil {
			return ErrInvalidEvent
		}
		if s.Grant != nil {
			if err := s.Grant.Validate(); err != nil {
				return err
			}
			if s.Grant.ID() != s.Key.AggregateID || s.Grant.TenantID() != s.Key.TenantID {
				return ErrInvalidEvent
			}
		}
	}
	return nil
}

type SnapshotReader interface {
	LoadCurrent(ctx context.Context, key Key) (Snapshot, error)
}

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
	Key       Key            `json:"key"`
	Revision  int64          `json:"revision"`
	EventID   string         `json:"event_id,omitempty"`
	ZedToken  string         `json:"zed_token,omitempty"`
	Edges     []RecordedEdge `json:"edges,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
}

func (c Checkpoint) Validate() error {
	if err := c.Key.Validate(); err != nil {
		return err
	}
	if c.Revision < 0 || (c.Revision > 0 && c.EventID == "") {
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

type Worker struct {
	projector *Projector
	source    SnapshotReader
	store     CheckpointStore
	now       func() time.Time
}

func NewWorker(p *Projector, source SnapshotReader, store CheckpointStore, now func() time.Time) (*Worker, error) {
	if p == nil || source == nil || store == nil || now == nil {
		return nil, ErrInvalidProjection
	}
	return &Worker{projector: p, source: source, store: store, now: now}, nil
}
func (w *Worker) Process(ctx context.Context, event Event) error {
	if err := event.Validate(); err != nil {
		return err
	}
	return w.process(ctx, event.Key, event.Revision, event.EventID, false)
}

// ReconcileKey always re-TOUCHes the desired edges, repairing missing graph
// tuples even when the source revision did not change. A separate discovery
// controller must enumerate all keys including tombstones.
func (w *Worker) ReconcileKey(ctx context.Context, key Key) error {
	if err := key.Validate(); err != nil {
		return err
	}
	return w.process(ctx, key, 0, "reconcile", true)
}
func (w *Worker) process(ctx context.Context, key Key, eventRevision int64, eventID string, force bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return w.store.WithLocked(ctx, key, func(cp Checkpoint) (Checkpoint, error) {
		if cp.Key != key {
			return cp, ErrInvalidCheckpoint
		}
		if err := cp.Validate(); err != nil {
			return cp, err
		}
		if !force {
			if eventRevision <= cp.Revision {
				return cp, nil
			}
			if eventRevision > cp.Revision+1 {
				return cp, ErrRevisionGap
			}
		}
		snapshot, err := w.source.LoadCurrent(ctx, key)
		if err != nil {
			return cp, err
		}
		if err := snapshot.Validate(); err != nil {
			return cp, err
		}
		if snapshot.Key != key {
			return cp, ErrInvalidEvent
		}
		if !force && snapshot.Revision < eventRevision {
			return cp, ErrSourceBehind
		}
		if !force && snapshot.Revision > eventRevision {
			return cp, ErrSourceAhead
		}
		if snapshot.Revision < cp.Revision {
			return cp, ErrSourceBehind
		}
		old, err := cp.relationships()
		if err != nil {
			return cp, err
		}
		var desired []domain.Relationship
		switch key.Kind {
		case KindRoleBinding:
			desired, err = w.projector.bindingRelationships(ctx, snapshot.RoleBinding)
		case KindGrant:
			desired, err = w.projector.grantRelationships(ctx, snapshot.Grant, true)
		default:
			return cp, ErrInvalidEvent
		}
		if err != nil {
			return cp, err
		}
		if _, err = indexRelationships(key.TenantID, desired); err != nil {
			return cp, err
		}
		if len(desired) > 1000 {
			return cp, ErrProjectionTooLarge
		}
		var result Result
		if force {
			result, err = w.forceApply(ctx, key.TenantID, old, desired)
		} else {
			result, err = w.projector.apply(ctx, key.TenantID, old, desired)
		}
		if err != nil {
			return cp, err
		}
		token := cp.ZedToken
		if result.Changed {
			token = result.ZedToken
		}
		newCP := Checkpoint{Key: key, Revision: snapshot.Revision, EventID: eventID, ZedToken: token, Edges: recordEdges(desired), UpdatedAt: w.now().UTC()}
		if err := newCP.Validate(); err != nil {
			return cp, err
		}
		return newCP, nil
	})
}

// forceApply performs an idempotent repair (touch all desired, delete removed).
// It cannot enumerate unknown, externally-created tuples; those need a separate
// ReadRelationships reconciliation pass before claiming full graph integrity.
func (w *Worker) forceApply(ctx context.Context, tenant domain.TenantID, old, desired []domain.Relationship) (Result, error) {
	prev, err := indexRelationships(tenant, old)
	if err != nil {
		return Result{}, err
	}
	next, err := indexRelationships(tenant, desired)
	if err != nil {
		return Result{}, err
	}
	changes := make([]ports.RelationshipChange, 0, len(prev)+len(next))
	for k, r := range prev {
		if _, ok := next[k]; !ok {
			changes = append(changes, ports.RelationshipChange{Operation: ports.RelationshipDelete, Relationship: r})
		}
	}
	for _, r := range next {
		changes = append(changes, ports.RelationshipChange{Operation: ports.RelationshipTouch, Relationship: r})
	}
	if len(changes) == 0 {
		return Result{}, nil
	}
	if len(changes) > maxUpdates {
		return Result{}, ErrProjectionTooLarge
	}
	sort.Slice(changes, func(i, j int) bool {
		return fmt.Sprint(changes[i].Relationship.Key()) < fmt.Sprint(changes[j].Relationship.Key())
	})
	written, err := w.projector.writer.WriteRelationships(ctx, tenant, changes)
	if err != nil {
		return Result{}, err
	}
	if written.ZedToken == "" {
		return Result{}, ErrMissingZedToken
	}
	return Result{Changed: true, ZedToken: written.ZedToken, Changes: len(changes)}, nil
}

// Delivery is a Kafka or Outbox event whose acknowledgement MUST be deferred
// until Process has durably committed its checkpoint.
type Delivery interface {
	Event() Event
	Ack(context.Context) error
}
type EventFeed interface {
	Next(context.Context) (Delivery, error)
}

func (w *Worker) Run(ctx context.Context, feed EventFeed) error {
	if feed == nil {
		return ErrInvalidProjection
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		d, err := feed.Next(ctx)
		if err != nil {
			return err
		}
		if d == nil {
			return ErrInvalidEvent
		}
		if err := w.Process(ctx, d.Event()); err != nil {
			return err
		} // no ack; caller retries/restarts
		if err := d.Ack(ctx); err != nil {
			return err
		} // replay is safe
	}
}
