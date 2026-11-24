package ydbroles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var _ projection.SnapshotReader = (*Store)(nil)

// LoadCurrent reads authoritative assignment and its compiled capability plan
// from one SERIALIZABLE YDB snapshot. Neither Kafka payloads nor an obsolete
// CQRS index can supply the graph's desired permissions.
func (s *Store) LoadCurrent(ctx context.Context, key projection.Key) (projection.Snapshot, error) {
	if s == nil || s.db == nil {
		return projection.Snapshot{}, ErrInvalidStore
	}
	if key.Validate() != nil || key.Kind != projection.KindProfileAssignment {
		return projection.Snapshot{}, ErrInvalidCommand
	}
	tenant, err := shared.NewTenantID(string(key.TenantID))
	if err != nil {
		return projection.Snapshot{}, err
	}
	var result projection.Snapshot
	err = s.tx(ctx, func(tx *sql.Tx) error {
		a, err := s.getAssignment(ctx, tx, tenant, key.AggregateID)
		if err != nil {
			return err
		}
		var revision int64
		var payload string
		err = tx.QueryRowContext(ctx, s.q(selectProfileProjection),
			sql.Named("tenant", string(tenant)), sql.Named("id", key.AggregateID)).
			Scan(&revision, &payload)
		var caps []projection.ProfileCapability
		if errors.Is(err, sql.ErrNoRows) {
			if a.Status() == role.ProfileAssignmentActive {
				return ErrPolicySourceDrift
			}
		} else if err != nil {
			return err
		} else {
			if revision != a.Revision() {
				return ErrPolicySourceDrift
			}
			if err := json.Unmarshal([]byte(payload), &caps); err != nil {
				return ErrPolicySourceDrift
			}
		}
		state := projection.ProfileAssignmentState{Assignment: a, Capabilities: caps}
		result = projection.Snapshot{Key: key, Revision: a.Revision(), Profile: &state}
		return result.Validate()
	})
	if err != nil {
		return projection.Snapshot{}, err
	}
	return result, nil
}

// PolicyOutboxEvent ties one immutable YDB source mutation to one tenant-wide
// policy head. A dispatcher emits it at least once; consumers dedupe EventID.
type PolicyOutboxEvent struct {
	TenantID       shared.TenantID
	AssignmentID   string
	Revision       int64
	PolicyRevision uint64
	EventID        string
}

func (e PolicyOutboxEvent) PolicyEvent() (projection.Event, uint64, error) {
	key, err := profileProjectionKey(e.TenantID, e.AssignmentID)
	if err != nil || e.PolicyRevision == 0 ||
		e.EventID != fmt.Sprintf("profile/%s/%s/%d", e.TenantID, e.AssignmentID, e.Revision) {
		return projection.Event{}, 0, ErrPolicySourceDrift
	}
	event := projection.Event{Key: key, Revision: e.Revision, EventID: e.EventID}
	if event.Validate() != nil {
		return projection.Event{}, 0, ErrPolicySourceDrift
	}
	return event, e.PolicyRevision, nil
}

func (s *Store) PendingPolicy(ctx context.Context, tenant shared.TenantID, id string, after int64, limit int) ([]PolicyOutboxEvent, error) {
	if s == nil || s.db == nil {
		return nil, ErrInvalidStore
	}
	if validateTenant(tenant) != nil || !shared.ValidUUID(id) || after < 0 || limit < 1 || limit > 100 {
		return nil, ErrInvalidCommand
	}
	rows, err := s.db.QueryContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $after AS Int64;DECLARE $limit AS Uint64;\nSELECT revision,policy_revision,event_id,payload_json FROM `m8-access/profile_policy_events` WHERE tenant_id=$tenant AND assignment_id=$id AND revision>$after AND published=FALSE ORDER BY revision LIMIT $limit;"),
		sql.Named("tenant", string(tenant)), sql.Named("id", id), sql.Named("after", after), sql.Named("limit", uint64(limit)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []PolicyOutboxEvent
	for rows.Next() {
		var e PolicyOutboxEvent
		e.TenantID, e.AssignmentID = tenant, id
		var payload string
		if err := rows.Scan(&e.Revision, &e.PolicyRevision, &e.EventID, &payload); err != nil {
			return nil, err
		}
		a, err := decodeAssignment(payload)
		if err != nil || a.ID() != id || a.TenantID() != tenant || a.Revision() != e.Revision ||
			(a.Status() != role.ProfileAssignmentActive && a.Status() != role.ProfileAssignmentRevoked) {
			return nil, ErrPolicySourceDrift
		}
		if _, _, err := e.PolicyEvent(); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) MarkPolicyPublished(ctx context.Context, e PolicyOutboxEvent) error {
	if s == nil || s.db == nil {
		return ErrInvalidStore
	}
	if _, _, err := e.PolicyEvent(); err != nil {
		return err
	}
	return s.tx(ctx, func(tx *sql.Tx) error {
		var gotID string
		var gotPolicy uint64
		var published bool
		err := tx.QueryRowContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;\nSELECT event_id,policy_revision,published FROM `m8-access/profile_policy_events` WHERE tenant_id=$tenant AND assignment_id=$id AND revision=$revision;"),
			sql.Named("tenant", string(e.TenantID)), sql.Named("id", e.AssignmentID), sql.Named("revision", e.Revision)).
			Scan(&gotID, &gotPolicy, &published)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if gotID != e.EventID || gotPolicy != e.PolicyRevision {
			return ErrPolicySourceDrift
		}
		if published {
			return nil
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $published AS Bool;\nUPDATE `m8-access/profile_policy_events` SET published=$published WHERE tenant_id=$tenant AND assignment_id=$id AND revision=$revision;"),
			sql.Named("tenant", string(e.TenantID)), sql.Named("id", e.AssignmentID), sql.Named("revision", e.Revision), sql.Named("published", true))
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $policy AS Uint64;DECLARE $published AS Bool;\nUPDATE `m8-access/profile_policy_dispatch` SET published=$published WHERE tenant_id=$tenant AND policy_revision=$policy AND assignment_id=$id;"),
			sql.Named("tenant", string(e.TenantID)), sql.Named("id", e.AssignmentID),
			sql.Named("policy", e.PolicyRevision), sql.Named("published", true))
		return err
	})
}
