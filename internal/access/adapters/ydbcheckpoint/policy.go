package ydbcheckpoint

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/app/policyjournal"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrPolicyConflict         = errors.New("Access policy journal CAS or idempotency conflict")
	ErrPolicyManifestMissing  = errors.New("Access policy fanout job not published")
	ErrPolicyCheckpointBehind = errors.New("Access policy projection checkpoint is not certified")
)

var _ consistency.PolicyStateReader = (*Store)(nil)
var _ policyjournal.JobCommitter = (*Store)(nil)

const queryPolicyHead = `DECLARE $tenant AS Utf8;
SELECT head_revision,applied_revision FROM ` + "`m8-access/policy_heads`" + ` WHERE tenant_id=$tenant;`
const queryPolicyRequest = `DECLARE $tenant AS Utf8;DECLARE $request AS Utf8;
SELECT policy_revision,request_hash FROM ` + "`m8-access/policy_request_ids`" + ` WHERE tenant_id=$tenant AND mutation_id=$request;`
const queryPolicyJob = `DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;
SELECT required_revision,certified FROM ` + "`m8-access/policy_jobs`" + `
WHERE tenant_id=$tenant AND policy_revision=$revision AND projection_kind=$kind AND aggregate_id=$aggregate;`
const queryIncompleteJobs = `DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;
SELECT COUNT(*) FROM ` + "`m8-access/policy_jobs`" + `
WHERE tenant_id=$tenant AND policy_revision=$revision AND certified=FALSE;`
const queryMutationCompleted = `DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;
SELECT completed FROM ` + "`m8-access/policy_mutations`" + ` WHERE tenant_id=$tenant AND policy_revision=$revision;`

func tenantArg(tenant domain.TenantID) any { return sql.Named("tenant", string(tenant)) }
func policyArgs(tenant domain.TenantID, rev uint64) []any {
	return []any{tenantArg(tenant), sql.Named("revision", rev)}
}

func (s *Store) LoadPolicyState(ctx context.Context, tenant domain.TenantID) (consistency.PolicyState, error) {
	if s == nil || tenant.Validate() != nil {
		return consistency.PolicyState{}, ErrInvalidStore
	}
	var state consistency.PolicyState
	err := s.tx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, s.q(queryPolicyHead), tenantArg(tenant)).Scan(&state.Head, &state.Applied)
		if errors.Is(err, sql.ErrNoRows) {
			return consistency.ErrPolicyUninitialized
		}
		return err
	})
	if err != nil {
		return consistency.PolicyState{}, err
	}
	return state, nil
}

// PublishPolicyMutationTx is the ONLY authorized policy-head writer.
// Call it on the SAME *sql.Tx as the authoritative Access mutation, complete
// binding/grant fanout and Outbox append; the caller owns COMMIT/ROLLBACK.
// A partial role fanout or a stand-alone journal call is a correctness bug.
// First tenant publication is a separate audited Bootstrap manifest.
func (s *Store) PublishPolicyMutationTx(ctx context.Context, tx *sql.Tx, m policyjournal.Mutation) (uint64, error) {
	if s == nil || tx == nil {
		return 0, ErrInvalidStore
	}
	digest, err := m.Digest()
	if err != nil {
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	var previousRevision uint64
	var previousDigest string
	err = tx.QueryRowContext(ctx, s.q(queryPolicyRequest), tenantArg(m.TenantID), sql.Named("request", m.ID)).
		Scan(&previousRevision, &previousDigest)
	if err == nil {
		if previousDigest == digest {
			return previousRevision, nil
		}
		return 0, ErrPolicyConflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	var head, applied uint64
	err = tx.QueryRowContext(ctx, s.q(queryPolicyHead), tenantArg(m.TenantID)).Scan(&head, &applied)
	if errors.Is(err, sql.ErrNoRows) {
		if !m.Bootstrap || m.ExpectedHead != 0 {
			return 0, consistency.ErrPolicyUninitialized
		}
	} else if err != nil {
		return 0, err
	} else if m.Bootstrap {
		return 0, ErrPolicyConflict
	}
	if head != m.ExpectedHead || head == math.MaxUint64 || applied > head {
		return 0, ErrPolicyConflict
	}
	revision := head + 1
	now, err := s.serverTime(ctx, tx)
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $head AS Uint64;DECLARE $applied AS Uint64;DECLARE $updated AS Timestamp;
UPSERT INTO `+"`m8-access/policy_heads`"+`(tenant_id,head_revision,applied_revision,updated_at)
VALUES ($tenant,$head,$applied,$updated);`),
		tenantArg(m.TenantID), sql.Named("head", revision), sql.Named("applied", applied), sql.Named("updated", now))
	if err != nil {
		return 0, fmt.Errorf("write Access policy head: %w", err)
	}
	_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;DECLARE $request AS Utf8;DECLARE $jobs AS Uint32;DECLARE $completed AS Bool;
UPSERT INTO `+"`m8-access/policy_mutations`"+`(tenant_id,policy_revision,mutation_id,total_jobs,completed)
VALUES ($tenant,$revision,$request,$jobs,$completed);`),
		tenantArg(m.TenantID), sql.Named("revision", revision), sql.Named("request", m.ID),
		sql.Named("jobs", uint32(len(m.Jobs))), sql.Named("completed", false))
	if err != nil {
		return 0, fmt.Errorf("write Access policy manifest: %w", err)
	}
	for _, job := range m.Jobs {
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $required AS Int64;DECLARE $certified AS Bool;DECLARE $zed AS Utf8;DECLARE $epoch AS Uint64;
UPSERT INTO `+"`m8-access/policy_jobs`"+`(tenant_id,policy_revision,projection_kind,aggregate_id,required_revision,certified,evidence_zed_token,evidence_epoch)
VALUES ($tenant,$revision,$kind,$aggregate,$required,$certified,$zed,$epoch);`),
			append(policyArgs(m.TenantID, revision), sql.Named("kind", string(job.Key.Kind)),
				sql.Named("aggregate", job.Key.AggregateID), sql.Named("required", job.RequiredRevision),
				sql.Named("certified", false), sql.Named("zed", ""), sql.Named("epoch", uint64(0)))...)
		if err != nil {
			return 0, fmt.Errorf("publish Access policy job: %w", err)
		}
	}
	_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $request AS Utf8;DECLARE $revision AS Uint64;DECLARE $digest AS Utf8;
UPSERT INTO `+"`m8-access/policy_request_ids`"+`(tenant_id,mutation_id,policy_revision,request_hash)
VALUES ($tenant,$request,$revision,$digest);`),
		tenantArg(m.TenantID), sql.Named("request", m.ID), sql.Named("revision", revision), sql.Named("digest", digest))
	if err != nil {
		return 0, fmt.Errorf("publish Access idempotency receipt: %w", err)
	}
	return revision, nil
}

// CommitPolicyJob certifies a specific fanout job ONLY after the worker has
// committed its fenced checkpoint. The job is looked up from the authoritative
// transaction manifest; broker messages cannot skip or invent projection work.
// Incomplete lower policy revisions block Applied advancement, even when newer
// mutations have already finished.
func (s *Store) CommitPolicyJob(ctx context.Context, policyRevision uint64, key projection.Key) error {
	if s == nil || policyRevision == 0 || key.Validate() != nil ||
		(key.Kind != projection.KindGrant && key.Kind != projection.KindRoleBinding) {
		return ErrInvalidStore
	}
	return s.tx(ctx, func(tx *sql.Tx) error {
		args := append(policyArgs(key.TenantID, policyRevision), sql.Named("kind", string(key.Kind)), sql.Named("aggregate", key.AggregateID))
		var required int64
		var certified bool
		err := tx.QueryRowContext(ctx, s.q(queryPolicyJob), args...).Scan(&required, &certified)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPolicyManifestMissing
		}
		if err != nil {
			return err
		}
		current, err := s.load(ctx, tx, key)
		if err != nil {
			return err
		}
		if required < 1 || !current.exists || current.cp.Revision < required ||
			current.cp.FenceEpoch == 0 || current.cp.FenceEpoch != current.leaseEpoch || current.cp.ZedToken == "" {
			return ErrPolicyCheckpointBehind
		}
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $certified AS Bool;DECLARE $zed AS Utf8;DECLARE $epoch AS Uint64;
UPDATE `+"`m8-access/policy_jobs`"+` SET certified=$certified,evidence_zed_token=$zed,evidence_epoch=$epoch
WHERE tenant_id=$tenant AND policy_revision=$revision AND projection_kind=$kind AND aggregate_id=$aggregate;`),
			append(args, sql.Named("certified", true), sql.Named("zed", current.cp.ZedToken),
				sql.Named("epoch", current.cp.FenceEpoch))...)
		if err != nil {
			return err
		}
		var incomplete uint64
		err = tx.QueryRowContext(ctx, s.q(queryIncompleteJobs), policyArgs(key.TenantID, policyRevision)...).Scan(&incomplete)
		if err != nil {
			return err
		}
		if incomplete == 0 {
			_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $revision AS Uint64;DECLARE $completed AS Bool;
UPDATE `+"`m8-access/policy_mutations`"+` SET completed=$completed
WHERE tenant_id=$tenant AND policy_revision=$revision;`),
				append(policyArgs(key.TenantID, policyRevision), sql.Named("completed", true))...)
			if err != nil {
				return err
			}
		}
		return s.advancePolicyLocked(ctx, tx, key.TenantID)
	})
}

// advancePolicyLocked uses the SERIALIZABLE journal transaction and only
// advances across contiguous fully-completed mutations. It never interprets
// ZedTokens as ordered or combines tokens from different grants/bindings.
func (s *Store) advancePolicyLocked(ctx context.Context, tx *sql.Tx, tenant domain.TenantID) error {
	var head, applied uint64
	err := tx.QueryRowContext(ctx, s.q(queryPolicyHead), tenantArg(tenant)).Scan(&head, &applied)
	if errors.Is(err, sql.ErrNoRows) {
		return consistency.ErrPolicyUninitialized
	}
	if err != nil {
		return err
	}
	if applied > head {
		return ErrPolicyConflict
	}
	old := applied
	for applied < head {
		next := applied + 1
		var completed bool
		err := tx.QueryRowContext(ctx, s.q(queryMutationCompleted), policyArgs(tenant, next)...).Scan(&completed)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		if !completed {
			break
		}
		applied = next
	}
	if applied == old {
		return nil
	}
	now, err := s.serverTime(ctx, tx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $applied AS Uint64;DECLARE $updated AS Timestamp;
UPDATE `+"`m8-access/policy_heads`"+` SET applied_revision=$applied,updated_at=$updated WHERE tenant_id=$tenant;`),
		tenantArg(tenant), sql.Named("applied", applied), sql.Named("updated", now))
	return err
}
