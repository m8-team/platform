package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrOutboxMismatch = errors.New("role management outbox event does not match")

// OutboxEvent is a durable *fact*, never permission activation evidence.
// Consumer must deduplicate event IDs and use a fenced policy projection
// before any future PDP allows rights from a staged assignment.
type OutboxEvent struct {
	Stream      string
	TenantID    string
	AggregateID string
	Revision    int64
	EventID     string
	Kind        string
	Payload     string
	CreatedAt   time.Time
}

func (e OutboxEvent) Validate() error {
	if e.Revision < 1 || e.EventID == "" || e.AggregateID == "" || e.CreatedAt.IsZero() {
		return ErrOutboxMismatch
	}
	switch e.Stream {
	case "bundle":
		if e.TenantID != "global" || e.Kind != "RoleBundlePublished" {
			return ErrOutboxMismatch
		}
		b, err := decodeBundle(e.Payload)
		if err != nil || b.Ref().Name != e.AggregateID {
			return ErrOutboxMismatch
		}
	case "assignment":
		tenant, err := shared.NewTenantID(e.TenantID)
		if err != nil {
			return ErrOutboxMismatch
		}
		a, err := decodeAssignment(e.Payload)
		if err != nil || a.TenantID() != tenant || a.ID() != e.AggregateID || a.Revision() != e.Revision {
			return ErrOutboxMismatch
		}
		if e.Kind != "ProfileAssignmentStaged" && e.Kind != "ProfileAssignmentRevoked" {
			return ErrOutboxMismatch
		}
		if e.Kind == "ProfileAssignmentStaged" && a.Status() != "PENDING_PROJECTION" ||
			e.Kind == "ProfileAssignmentRevoked" && a.Status() != "REVOKED" {
			return ErrOutboxMismatch
		}
	default:
		return ErrOutboxMismatch
	}
	return nil
}

// Pending is per aggregate. A future dispatcher MUST implement durable
// per-partition leases/checkpoints before sharing a Kafka consumer group.
// The returned order is per aggregate revision, never a global event order.
func (s *Store) Pending(ctx context.Context, stream, tenant, aggregate string, after int64, limit int) ([]OutboxEvent, error) {
	if s == nil || s.db == nil {
		return nil, ErrInvalidStore
	}
	if after < 0 || limit < 1 || limit > 100 || aggregate == "" ||
		(stream != "bundle" && stream != "assignment") {
		return nil, ErrInvalidCommand
	}
	if stream == "bundle" {
		if tenant != "global" || !shared.ValidQualifiedName(aggregate) {
			return nil, ErrInvalidCommand
		}
	} else {
		if _, err := shared.NewTenantID(tenant); err != nil {
			return nil, ErrInvalidCommand
		}
		if !shared.ValidUUID(aggregate) {
			return nil, ErrInvalidCommand
		}
	}
	rows, err := s.db.QueryContext(ctx, s.q("DECLARE $stream AS Utf8;DECLARE $tenant AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $after AS Int64;DECLARE $limit AS Uint64;\nSELECT revision,event_id,event_kind,payload_json,created_at FROM `m8-access/role_management_outbox` WHERE stream=$stream AND tenant_id=$tenant AND aggregate_id=$aggregate AND revision>$after AND published=FALSE ORDER BY revision LIMIT $limit;"),
		sql.Named("stream", stream), sql.Named("tenant", tenant), sql.Named("aggregate", aggregate),
		sql.Named("after", after), sql.Named("limit", uint64(limit)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []OutboxEvent
	for rows.Next() {
		e := OutboxEvent{Stream: stream, TenantID: tenant, AggregateID: aggregate}
		if err := rows.Scan(&e.Revision, &e.EventID, &e.Kind, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		if err := e.Validate(); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// MarkPublished is an internal dispatcher operation only. Do not expose it in
// public API. It commits a local ACK only after the real broker confirms.
func (s *Store) MarkPublished(ctx context.Context, e OutboxEvent) error {
	if s == nil || s.db == nil {
		return ErrInvalidStore
	}
	if err := e.Validate(); err != nil {
		return err
	}
	return s.tx(ctx, func(tx *sql.Tx) error {
		var eventID string
		var published bool
		err := tx.QueryRowContext(ctx, s.q("DECLARE $stream AS Utf8;DECLARE $tenant AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $revision AS Int64;\nSELECT event_id,published FROM `m8-access/role_management_outbox` WHERE stream=$stream AND tenant_id=$tenant AND aggregate_id=$aggregate AND revision=$revision;"),
			sql.Named("stream", e.Stream), sql.Named("tenant", e.TenantID), sql.Named("aggregate", e.AggregateID),
			sql.Named("revision", e.Revision)).Scan(&eventID, &published)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if eventID != e.EventID {
			return ErrOutboxMismatch
		}
		if published {
			return nil
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $stream AS Utf8;DECLARE $tenant AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $revision AS Int64;DECLARE $published AS Bool;\nUPDATE `m8-access/role_management_outbox` SET published=$published WHERE stream=$stream AND tenant_id=$tenant AND aggregate_id=$aggregate AND revision=$revision;"),
			sql.Named("stream", e.Stream), sql.Named("tenant", e.TenantID), sql.Named("aggregate", e.AggregateID),
			sql.Named("revision", e.Revision), sql.Named("published", true))
		if err != nil {
			return fmt.Errorf("ack role management outbox: %w", err)
		}
		return nil
	})
}
