package ydbcheckpoint

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/access/app/projection"
)

type row struct {
	cp          projection.Checkpoint
	leaseEpoch  uint64
	leaseHolder string
	leaseUntil  time.Time
	exists      bool
}

// certifiedAt reports whether the current lease generation has persisted graph
// evidence at or beyond the requested source revision.
func (r row) certifiedAt(required int64) bool {
	return required > 0 && r.exists && r.cp.Revision >= required && r.cp.ZedToken != "" &&
		r.cp.FenceEpoch != 0 && r.cp.FenceEpoch == r.leaseEpoch
}

const selectCheckpoint = `DECLARE $tenant AS Utf8; DECLARE $kind AS Utf8; DECLARE $aggregate AS Utf8;
SELECT revision,event_id,zed_token,fence_epoch,edges_json,updated_at,lease_epoch,lease_holder,lease_until
FROM ` + "`m8-access/projection_checkpoints`" + ` WHERE tenant_id=$tenant AND projection_kind=$kind AND aggregate_id=$aggregate;`

func (s *Store) load(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, key projection.Key) (row, error) {
	state := row{cp: projection.Checkpoint{Key: key}}
	var edges string
	args := keyArgs(key)
	err := q.QueryRowContext(ctx, s.q(selectCheckpoint), args...).Scan(&state.cp.Revision, &state.cp.EventID, &state.cp.ZedToken, &state.cp.FenceEpoch, &edges, &state.cp.UpdatedAt, &state.leaseEpoch, &state.leaseHolder, &state.leaseUntil)
	if errors.Is(err, sql.ErrNoRows) {
		return state, nil
	}
	if err != nil {
		return row{}, err
	}
	if len(edges) > maxEdgesBytes {
		return row{}, projection.ErrInvalidCheckpoint
	}
	if err := json.Unmarshal([]byte(edges), &state.cp.Edges); err != nil {
		return row{}, fmt.Errorf("decode projection edges: %w", err)
	}
	if err := state.cp.Validate(); err != nil {
		return row{}, err
	}
	state.exists = true
	return state, nil
}

// LoadCheckpoint returns only a checkpoint certified against the most recent
// lease epoch. It fails closed during takeover until graph reconciliation is
// committed at that generation.
func (s *Store) LoadCheckpoint(ctx context.Context, key projection.Key) (projection.Checkpoint, error) {
	if err := key.Validate(); err != nil {
		return projection.Checkpoint{}, err
	}
	var observed row
	err := s.tx(ctx, func(tx *sql.Tx) error {
		var loadErr error
		observed, loadErr = s.load(ctx, tx, key)
		return loadErr
	})
	if err != nil {
		return projection.Checkpoint{}, err
	}
	if !observed.certifiedAt(1) {
		return projection.Checkpoint{}, ErrCheckpointNotConverged
	}
	return observed.cp, nil
}
