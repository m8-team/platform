package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

// PendingTenantPolicy enumerates immutable profile-policy events in tenant
// PolicyHead order, not by assignment UUID. The dispatcher may deliver at
// least once, but it must never skip a pending revision on broker failure.
func (s *Store) PendingTenantPolicy(ctx context.Context, tenant shared.TenantID, limit int) ([]PolicyOutboxEvent, error) {
	if s == nil || s.db == nil {
		return nil, ErrInvalidStore
	}
	if tenant.Validate() != nil || limit < 1 || limit > 100 {
		return nil, ErrInvalidCommand
	}
	rows, err := s.db.QueryContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $limit AS Uint64;\nSELECT policy_revision,assignment_id,projection_revision,event_id FROM `m8-access/profile_policy_dispatch` WHERE tenant_id=$tenant AND published=FALSE ORDER BY policy_revision,assignment_id LIMIT $limit;"),
		sql.Named("tenant", string(tenant)), sql.Named("limit", uint64(limit)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]PolicyOutboxEvent, 0, limit)
	for rows.Next() {
		var e PolicyOutboxEvent
		e.TenantID = tenant
		if err := rows.Scan(&e.PolicyRevision, &e.AssignmentID, &e.Revision, &e.EventID); err != nil {
			return nil, err
		}
		if _, _, err := e.PolicyEvent(); err != nil {
			return nil, ErrPolicySourceDrift
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// PolicyEventPublisher MUST wait for an authenticated broker ACK before
// returning nil. It is an external side effect, never inside a YDB tx.
// The Kafka transport must use a stable event ID/partition key and dedupe.
type PolicyEventPublisher interface {
	PublishPolicyEvent(context.Context, policyjournal.Event) error
}

var ErrDispatcherUnavailable = errors.New("profile policy outbox dispatcher unavailable")

type PolicyDispatcher struct {
	store     *Store
	publisher PolicyEventPublisher
}

func NewPolicyDispatcher(store *Store, publisher PolicyEventPublisher) (*PolicyDispatcher, error) {
	if store == nil || store.db == nil || publisher == nil {
		return nil, ErrDispatcherUnavailable
	}
	return &PolicyDispatcher{store: store, publisher: publisher}, nil
}

// DispatchTenant is a bounded at-least-once delivery pass. Callers must own
// a trusted tenant dispatcher lease; the Store intentionally does not infer
// ownership from a thread ID or Kafka consumer group. Two replicas may
// duplicate sends, which are safe with idempotent event consumers.
func (d *PolicyDispatcher) DispatchTenant(ctx context.Context, tenant shared.TenantID, limit int) (int, error) {
	if d == nil || d.store == nil || d.publisher == nil {
		return 0, ErrDispatcherUnavailable
	}
	pending, err := d.store.PendingTenantPolicy(ctx, tenant, limit)
	if err != nil {
		return 0, err
	}
	delivered := 0
	for _, event := range pending {
		if err := ctx.Err(); err != nil {
			return delivered, err
		}
		projectionEvent, policyRevision, err := event.PolicyEvent()
		if err != nil {
			return delivered, err
		}
		message := policyjournal.Event{PolicyRevision: policyRevision, Projection: projectionEvent}
		if err := message.Validate(); err != nil {
			return delivered, fmt.Errorf("%w: %v", ErrPolicySourceDrift, err)
		}
		if err := d.publisher.PublishPolicyEvent(ctx, message); err != nil {
			// No YDB ACK on ambiguous or failed broker publication.
			return delivered, err
		}
		if err := d.store.MarkPolicyPublished(ctx, event); err != nil {
			// Broker may have committed. A replay is intentionally safe.
			return delivered, err
		}
		delivered++
	}
	return delivered, nil
}
