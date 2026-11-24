package ydbroles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

// PolicyCommitter MUST publish the entire tenant mutation in the very same
// *sql.Tx as source state, CQRS index, projector source and outbox. Use
// ydbcheckpoint.Store with the EXACT same database namespace/prefix.
type PolicyCommitter interface {
	PublishPolicyMutationTx(context.Context, *sql.Tx, policyjournal.Mutation) (uint64, error)
}

var (
	ErrPolicyTransition  = errors.New("invalid or unsupported profile activation transition")
	ErrPolicySourceDrift = errors.New("profile projection source revision mismatch")
)

type PolicyChange struct {
	Assignment      role.ProfileAssignment
	PolicyRevision  uint64
	ProjectionEvent projection.Event
}

const selectProfilePolicyEvent = "DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64; SELECT policy_revision,event_id,payload_json FROM `m8-access/profile_policy_events` WHERE tenant_id=$tenant AND assignment_id=$id AND revision=$revision;"
const selectProfileProjection = "DECLARE $tenant AS Utf8;DECLARE $id AS Utf8; SELECT revision,capabilities_json FROM `m8-access/profile_assignment_projection` WHERE tenant_id=$tenant AND assignment_id=$id;"

func profileProjectionKey(tenant shared.TenantID, id string) (projection.Key, error) {
	td, err := domain.NewTenantID(string(tenant))
	if err != nil {
		return projection.Key{}, err
	}
	key := projection.Key{TenantID: td, Kind: projection.KindProfileAssignment, AggregateID: id}
	return key, key.Validate()
}

// ValidateActivationCaps is a non-authorizing structural guard; the ONLY
// application caller must compile this input from verified immutable catalog
// templates and resource-owner revision proof. Never accept it from HTTP/Kafka.
func ValidateActivationCaps(a role.ProfileAssignment, caps []projection.ProfileCapability) error {
	if a.Validate() != nil || a.Status() != role.ProfileAssignmentPending {
		return ErrPolicyTransition
	}
	active, err := a.Activate(a.Revision())
	if err != nil {
		return err
	}
	state := projection.ProfileAssignmentState{Assignment: active, Capabilities: caps}
	return state.Validate()
}

func (s *Store) queryPolicyReply(ctx context.Context, tx *sql.Tx, tenant shared.TenantID, id string, revision int64) (PolicyChange, error) {
	var policyRevision uint64
	var eventID, payload string
	err := tx.QueryRowContext(ctx, s.q(selectProfilePolicyEvent),
		sql.Named("tenant", string(tenant)), sql.Named("id", id), sql.Named("revision", revision)).
		Scan(&policyRevision, &eventID, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return PolicyChange{}, ErrPolicySourceDrift
	}
	if err != nil {
		return PolicyChange{}, err
	}
	a, err := decodeAssignment(payload)
	if err != nil || a.ID() != id || a.TenantID() != tenant || a.Revision() != revision || policyRevision == 0 {
		return PolicyChange{}, ErrPolicySourceDrift
	}
	key, err := profileProjectionKey(tenant, id)
	if err != nil {
		return PolicyChange{}, err
	}
	event := projection.Event{EventID: eventID, Key: key, Revision: revision}
	if err := event.Validate(); err != nil {
		return PolicyChange{}, ErrPolicySourceDrift
	}
	return PolicyChange{Assignment: a, PolicyRevision: policyRevision, ProjectionEvent: event}, nil
}

func (s *Store) policyRetry(ctx context.Context, tx *sql.Tx, tenant shared.TenantID, id, key, digest string) (PolicyChange, bool, error) {
	var savedHash string
	var revision int64
	err := tx.QueryRowContext(ctx, s.q(selectAssignmentReceipt),
		append(assignmentArgs(tenant, id), sql.Named("key", key))...).Scan(&savedHash, &revision)
	if errors.Is(err, sql.ErrNoRows) {
		return PolicyChange{}, false, nil
	}
	if err != nil {
		return PolicyChange{}, false, err
	}
	if savedHash != digest {
		return PolicyChange{}, false, ErrAssignmentConflict
	}
	record, err := s.queryPolicyReply(ctx, tx, tenant, id, revision)
	return record, true, err
}
func (s *Store) writeProjectionSource(ctx context.Context, tx *sql.Tx, a role.ProfileAssignment, caps []projection.ProfileCapability) error {
	payload, err := encodeJSON(caps)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $capabilities AS Utf8;\nUPSERT INTO `m8-access/profile_assignment_projection` (tenant_id,assignment_id,revision,capabilities_json) VALUES ($tenant,$id,$revision,$capabilities);"),
		sql.Named("tenant", string(a.TenantID())), sql.Named("id", a.ID()),
		sql.Named("revision", a.Revision()), sql.Named("capabilities", payload))
	return err
}
func (s *Store) writePolicyEvent(ctx context.Context, tx *sql.Tx, a role.ProfileAssignment, policyRev uint64, now time.Time) error {
	payload, err := encodeAssignment(a)
	if err != nil {
		return err
	}
	eventID := fmt.Sprintf("profile/%s/%s/%d", a.TenantID(), a.ID(), a.Revision())
	_, err = tx.ExecContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $policy AS Uint64;DECLARE $event AS Utf8;DECLARE $payload AS Utf8;DECLARE $published AS Bool;DECLARE $created AS Timestamp;\nUPSERT INTO `m8-access/profile_policy_events` (tenant_id,assignment_id,revision,policy_revision,event_id,payload_json,published,created_at) VALUES ($tenant,$id,$revision,$policy,$event,$payload,$published,$created);"),
		sql.Named("tenant", string(a.TenantID())), sql.Named("id", a.ID()), sql.Named("revision", a.Revision()),
		sql.Named("policy", policyRev), sql.Named("event", eventID), sql.Named("payload", payload),
		sql.Named("published", false), sql.Named("created", now))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, s.q("DECLARE $tenant AS Utf8;DECLARE $policy AS Uint64;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $event AS Utf8;DECLARE $published AS Bool;\nUPSERT INTO `m8-access/profile_policy_dispatch` (tenant_id,policy_revision,assignment_id,projection_revision,event_id,published) VALUES ($tenant,$policy,$id,$revision,$event,$published);"),
		sql.Named("tenant", string(a.TenantID())), sql.Named("policy", policyRev),
		sql.Named("id", a.ID()), sql.Named("revision", a.Revision()),
		sql.Named("event", eventID), sql.Named("published", false))
	return err
}

func activationMutation(tenant shared.TenantID, id string, expectedHead uint64, key string, rev int64) (policyjournal.Mutation, error) {
	k, err := profileProjectionKey(tenant, id)
	if err != nil {
		return policyjournal.Mutation{}, err
	}
	m := policyjournal.Mutation{TenantID: k.TenantID, ExpectedHead: expectedHead,
		ID: "profile/" + id + "/" + key, Jobs: []policyjournal.Job{{Key: k, RequiredRevision: rev}}}
	// Stable namespace is required to avoid collisions with legacy grants.
	if err := m.Validate(); err != nil {
		return policyjournal.Mutation{}, err
	}
	return m, nil
}

// ActivateWithPolicy is the ONLY storage writer of ACTIVE profile assignments.
// Requires an initialized policy journal, exact head, a verified instance-only
// capability plan, durable source and outbox in the same YDB transaction.
// PolicyBarrier blocks permits while Applied < Head; graph worker must certify.
func (s *Store) ActivateWithPolicy(ctx context.Context, committer PolicyCommitter, tenant shared.TenantID, id string, expectedRev int64, expectedPolicyHead uint64, key string, caps []projection.ProfileCapability) (PolicyChange, error) {
	if s == nil || s.db == nil || committer == nil {
		return PolicyChange{}, ErrInvalidStore
	}
	if validateTenant(tenant) != nil || !shared.ValidUUID(id) || expectedRev < 1 || !validKey(key) || len(key) > 110 || expectedPolicyHead == 0 {
		return PolicyChange{}, ErrInvalidCommand
	}
	for _, cap := range caps {
		if cap.Validate() != nil {
			return PolicyChange{}, ErrPolicyTransition
		}
	}
	encoded, err := json.Marshal(caps)
	if err != nil {
		return PolicyChange{}, err
	}
	digest := commandDigest("activate/"+fmt.Sprint(expectedPolicyHead), expectedRev, string(encoded))
	var result PolicyChange
	err = s.tx(ctx, func(tx *sql.Tx) error {
		if previous, ok, err := s.policyRetry(ctx, tx, tenant, id, key, digest); err != nil {
			return err
		} else if ok {
			result = previous
			return nil
		}
		a, err := s.getAssignment(ctx, tx, tenant, id)
		if err != nil {
			return err
		}
		if a.Revision() != expectedRev {
			return shared.ErrRevisionConflict
		}
		if err := ValidateActivationCaps(a, caps); err != nil {
			return err
		}
		activated, err := a.Activate(expectedRev)
		if err != nil {
			return err
		}
		payload, err := encodeAssignment(activated)
		if err != nil {
			return err
		}
		m, err := activationMutation(tenant, id, expectedPolicyHead, key, activated.Revision())
		if err != nil {
			return err
		}
		policyRev, err := committer.PublishPolicyMutationTx(ctx, tx, m)
		if err != nil {
			return err
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		if err := s.saveAssignmentTx(ctx, tx, activated, payload); err != nil {
			return err
		}
		if err := s.writeProjectionSource(ctx, tx, activated, caps); err != nil {
			return err
		}
		if err := s.saveReceiptTx(ctx, tx, tenant, id, key, digest, activated.Revision()); err != nil {
			return err
		}
		if err := s.writePolicyEvent(ctx, tx, activated, policyRev, now); err != nil {
			return err
		}
		k, err := profileProjectionKey(tenant, id)
		if err != nil {
			return err
		}
		result = PolicyChange{Assignment: activated, PolicyRevision: policyRev,
			ProjectionEvent: projection.Event{EventID: fmt.Sprintf("profile/%s/%s/%d", tenant, id, activated.Revision()), Key: k, Revision: activated.Revision()}}
		return nil
	})
	if err != nil {
		return PolicyChange{}, err
	}
	return result, nil
}

// RevokeActiveWithPolicy never removes only one side of a graph relationship.
// Revocation updates the tenant PolicyHead BEFORE any positive check can escape;
// the worker then deletes all owned tuples and certifies Applied advancement.
func (s *Store) RevokeActiveWithPolicy(ctx context.Context, committer PolicyCommitter, tenant shared.TenantID, id string, expectedRev int64, expectedPolicyHead uint64, key string) (PolicyChange, error) {
	if s == nil || s.db == nil || committer == nil {
		return PolicyChange{}, ErrInvalidStore
	}
	if validateTenant(tenant) != nil || !shared.ValidUUID(id) || expectedRev < 2 || !validKey(key) || len(key) > 110 || expectedPolicyHead == 0 {
		return PolicyChange{}, ErrInvalidCommand
	}
	digest := commandDigest("revoke-active/"+fmt.Sprint(expectedPolicyHead), expectedRev, string(tenant)+"/"+id)
	var result PolicyChange
	err := s.tx(ctx, func(tx *sql.Tx) error {
		if previous, ok, err := s.policyRetry(ctx, tx, tenant, id, key, digest); err != nil {
			return err
		} else if ok {
			result = previous
			return nil
		}
		a, err := s.getAssignment(ctx, tx, tenant, id)
		if err != nil {
			return err
		}
		if a.Revision() != expectedRev {
			return shared.ErrRevisionConflict
		}
		if a.Status() != role.ProfileAssignmentActive {
			return ErrPolicyTransition
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		revoked, err := a.RevokeActive(expectedRev, now)
		if err != nil {
			return err
		}
		payload, err := encodeAssignment(revoked)
		if err != nil {
			return err
		}
		m, err := activationMutation(tenant, id, expectedPolicyHead, key, revoked.Revision())
		if err != nil {
			return err
		}
		policyRev, err := committer.PublishPolicyMutationTx(ctx, tx, m)
		if err != nil {
			return err
		}
		if err := s.saveAssignmentTx(ctx, tx, revoked, payload); err != nil {
			return err
		}
		if err := s.writeProjectionSource(ctx, tx, revoked, nil); err != nil {
			return err
		}
		if err := s.saveReceiptTx(ctx, tx, tenant, id, key, digest, revoked.Revision()); err != nil {
			return err
		}
		if err := s.writePolicyEvent(ctx, tx, revoked, policyRev, now); err != nil {
			return err
		}
		k, err := profileProjectionKey(tenant, id)
		if err != nil {
			return err
		}
		result = PolicyChange{Assignment: revoked, PolicyRevision: policyRev,
			ProjectionEvent: projection.Event{EventID: fmt.Sprintf("profile/%s/%s/%d", tenant, id, revoked.Revision()), Key: k, Revision: revoked.Revision()}}
		return nil
	})
	if err != nil {
		return PolicyChange{}, err
	}
	return result, nil
}
