package projection

import (
	"context"
	"errors"
	"regexp"
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
	KindRoleBinding       Kind = "ROLE_BINDING"
	KindGrant             Kind = "ACCESS_GRANT"
	KindRMOrganization    Kind = "RM_ORGANIZATION"
	KindRMWorkspace       Kind = "RM_WORKSPACE"
	KindRMEnvironment     Kind = "RM_ENVIRONMENT"
	KindRMWorkloadService Kind = "RM_WORKLOAD_SERVICE"
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
	if k.Kind != KindRoleBinding && k.Kind != KindGrant && !IsResourceKind(k.Kind) {
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
	Resource    *ResourceState
}

func (s Snapshot) Validate() error {
	if err := s.Key.Validate(); err != nil {
		return err
	}
	if s.Revision < 1 {
		return ErrInvalidEvent
	}
	if s.Key.Kind == KindRoleBinding {
		if s.Grant != nil || s.Resource != nil {
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
	} else if s.Key.Kind == KindGrant {
		if s.RoleBinding != nil || s.Resource != nil {
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
	} else {
		if s.RoleBinding != nil || s.Grant != nil || s.Resource == nil {
			return ErrInvalidEvent
		}
		if err := s.Resource.Validate(s.Key, s.Revision); err != nil {
			return err
		}
	}
	return nil
}

type SnapshotReader interface {
	LoadCurrent(ctx context.Context, key Key) (Snapshot, error)
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
	if fenced, ok := w.store.(FencedCheckpointStore); ok {
		installer, ok := w.projector.writer.(ports.GraphFenceInstaller)
		if !ok {
			return ports.ErrInvalidGraphFence
		}
		return fenced.WithFenced(ctx, key, func(leaseCtx context.Context, cp Checkpoint, fence ports.GraphFence) (Checkpoint, error) {
			if err := fence.Validate(); err != nil || fence.TenantID != key.TenantID || fence.Kind != string(key.Kind) || fence.AggregateID != key.AggregateID {
				return cp, ports.ErrInvalidGraphFence
			}
			graphToken, err := installer.EstablishFence(leaseCtx, fence)
			if err != nil {
				return cp, err
			}
			if graphToken == "" {
				return cp, ErrMissingZedToken
			}
			guardedCtx, err := ports.WithGraphFence(leaseCtx, fence)
			if err != nil {
				return cp, err
			}
			// Read the real graph after fencing. A previous attempt may have
			// written tuples and crashed before its checkpoint commit.
			reader, ok := w.projector.writer.(ports.GraphProjectionReader)
			if !ok {
				return cp, ports.ErrInvalidGraphFence
			}
			actual, err := reader.ReadProjectionEdges(guardedCtx, fence)
			if err != nil {
				return cp, err
			}
			// A new epoch always performs authoritative full reconciliation.
			return w.runTransition(guardedCtx, key, eventRevision, eventID, true, cp, fence.Epoch, graphToken, actual)
		})
	}
	return w.store.WithLocked(ctx, key, func(cp Checkpoint) (Checkpoint, error) {
		return w.runTransition(ctx, key, eventRevision, eventID, force, cp, 0, "", nil)
	})
}

func (w *Worker) runTransition(ctx context.Context, key Key, eventRevision int64, eventID string, force bool, cp Checkpoint, fenceEpoch uint64, graphFenceToken string, graphEdges []domain.Relationship) (Checkpoint, error) {
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
	if fenceEpoch != 0 {
		old = graphEdges
	}
	var desired []domain.Relationship
	switch key.Kind {
	case KindRoleBinding:
		desired, err = w.projector.bindingRelationships(ctx, snapshot.RoleBinding)
	case KindGrant:
		desired, err = w.projector.grantRelationships(ctx, snapshot.Grant, true)
	case KindRMOrganization, KindRMWorkspace, KindRMEnvironment, KindRMWorkloadService:
		desired, err = resourceRelationships(key, snapshot.Resource)
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
	result, err := w.projector.apply(ctx, key.TenantID, old, desired, force)
	if err != nil {
		return cp, err
	}
	token := cp.ZedToken
	if graphFenceToken != "" {
		token = graphFenceToken
	}
	if result.Changed {
		token = result.ZedToken
	}
	newCP := Checkpoint{Key: key, FenceEpoch: fenceEpoch, Revision: snapshot.Revision, EventID: eventID, ZedToken: token, Edges: recordEdges(desired), UpdatedAt: w.now().UTC()}
	if err := newCP.Validate(); err != nil {
		return cp, err
	}
	return newCP, nil
}

func (w *Worker) Run(ctx context.Context, feed EventFeed) error {
	if feed == nil {
		return ErrInvalidProjection
	}
	return consumeEvents(ctx, feed, w.Process)
}
