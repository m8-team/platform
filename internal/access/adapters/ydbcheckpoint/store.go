// Package ydbcheckpoint implements distributed checkpoint leases with
// transactional compare-and-swap on YDB. It must be paired with a graph-side
// fencing writer: a YDB lease alone cannot stop late writes to SpiceDB.
package ydbcheckpoint

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/app/projection"
)

var (
	ErrInvalidStore           = errors.New("invalid YDB checkpoint store")
	ErrLeaseBusy              = errors.New("projection checkpoint leased by another worker")
	ErrLeaseLost              = errors.New("projection checkpoint fencing lease lost")
	ErrCheckpointConflict     = errors.New("projection checkpoint CAS conflict")
	ErrCheckpointNotConverged = errors.New("projection checkpoint has not converged to the active fence")
)

const maxEdgesBytes = 4 << 20

var safePrefix = regexp.MustCompile(`^/[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*$`)

type Store struct {
	db      *sql.DB
	prefix  string
	ownerID string
	ttl     time.Duration
}

func New(db *sql.DB, prefix, ownerID string, ttl time.Duration) (*Store, error) {
	if db == nil || !safePrefix.MatchString(prefix) || ownerID == "" || len(ownerID) > 128 || ttl < 3*time.Second || ttl > 5*time.Minute {
		return nil, ErrInvalidStore
	}
	return &Store{db: db, prefix: prefix, ownerID: ownerID, ttl: ttl}, nil
}

var _ projection.CheckpointStore = (*Store)(nil)
var _ projection.CheckpointReader = (*Store)(nil)
var _ projection.FencedCheckpointStore = (*Store)(nil)

func (s *Store) q(stmt string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s", s.prefix, stmt)
}
func keyArgs(k projection.Key) []any {
	return []any{sql.Named("tenant", string(k.TenantID)), sql.Named("kind", string(k.Kind)), sql.Named("aggregate", k.AggregateID)}
}

func (s *Store) serverTime(ctx context.Context, tx *sql.Tx) (time.Time, error) {
	var now time.Time
	err := tx.QueryRowContext(ctx, s.q("SELECT CurrentUtcTimestamp();")).Scan(&now)
	return now.UTC(), err
}

func (s *Store) tx(ctx context.Context, work func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := work(tx); err != nil {
		return err
	}
	return tx.Commit()
}

type row struct {
	cp          projection.Checkpoint
	leaseEpoch  uint64
	leaseHolder string
	leaseUntil  time.Time
	exists      bool
}

const selectCheckpoint = `DECLARE $tenant AS Utf8; DECLARE $kind AS Utf8; DECLARE $aggregate AS Utf8;
SELECT revision,event_id,zed_token,fence_epoch,edges_json,updated_at,lease_epoch,lease_holder,lease_until
FROM `+"`m8-access/projection_checkpoints`"+` WHERE tenant_id=$tenant AND projection_kind=$kind AND aggregate_id=$aggregate;`

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

// WithLocked intentionally cannot be used: using it would bypass the required
// graph-side fencing handshake. Worker dispatches to WithFenced instead.
func (s *Store) WithLocked(context.Context, projection.Key, func(projection.Checkpoint) (projection.Checkpoint, error)) error {
	return ErrLeaseLost
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
	if !observed.exists || observed.cp.Revision == 0 || observed.cp.FenceEpoch == 0 || observed.cp.FenceEpoch != observed.leaseEpoch || observed.cp.ZedToken == "" {
		return projection.Checkpoint{}, ErrCheckpointNotConverged
	}
	return observed.cp, nil
}

func (s *Store) acquire(ctx context.Context, key projection.Key) (row, error) {
	var acquired row
	err := s.tx(ctx, func(tx *sql.Tx) error {
		current, err := s.load(ctx, tx, key)
		if err != nil {
			return err
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		if current.exists && current.leaseHolder != "" && current.leaseUntil.After(now) {
			return ErrLeaseBusy
		}
		if current.leaseEpoch == ^uint64(0) {
			return ErrLeaseLost
		}
		current.leaseEpoch++
		current.leaseHolder = s.ownerID
		current.leaseUntil = now.Add(s.ttl)
		edges, err := json.Marshal(current.cp.Edges)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $revision AS Int64;DECLARE $event AS Utf8;DECLARE $zed AS Utf8;DECLARE $fence AS Uint64;DECLARE $edges AS Utf8;DECLARE $updated AS Timestamp;DECLARE $epoch AS Uint64;DECLARE $owner AS Utf8;DECLARE $until AS Timestamp;
UPSERT INTO `+"`m8-access/projection_checkpoints`"+` (tenant_id,projection_kind,aggregate_id,revision,event_id,zed_token,fence_epoch,edges_json,updated_at,lease_epoch,lease_holder,lease_until)
VALUES ($tenant,$kind,$aggregate,$revision,$event,$zed,$fence,$edges,$updated,$epoch,$owner,$until);`),
			append(keyArgs(key), sql.Named("revision", current.cp.Revision), sql.Named("event", current.cp.EventID), sql.Named("zed", current.cp.ZedToken), sql.Named("fence", current.cp.FenceEpoch), sql.Named("edges", string(edges)), sql.Named("updated", now), sql.Named("epoch", current.leaseEpoch), sql.Named("owner", s.ownerID), sql.Named("until", current.leaseUntil))...)
		if err != nil {
			return err
		}
		acquired = current
		return nil
	})
	return acquired, err
}

func (s *Store) renew(ctx context.Context, key projection.Key, epoch uint64) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		current, err := s.load(ctx, tx, key)
		if err != nil {
			return err
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		if !current.exists || current.leaseEpoch != epoch || current.leaseHolder != s.ownerID || !current.leaseUntil.After(now) {
			return ErrLeaseLost
		}
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $until AS Timestamp;
UPDATE `+"`m8-access/projection_checkpoints`"+` SET lease_until=$until WHERE tenant_id=$tenant AND projection_kind=$kind AND aggregate_id=$aggregate;`), append(keyArgs(key), sql.Named("until", now.Add(s.ttl)))...)
		return err
	})
}

func (s *Store) finish(ctx context.Context, key projection.Key, epoch uint64, next *projection.Checkpoint) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		old, err := s.load(ctx, tx, key)
		if err != nil {
			return err
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		if !old.exists || old.leaseEpoch != epoch || old.leaseHolder != s.ownerID || !old.leaseUntil.After(now) {
			return ErrLeaseLost
		}
		if next != nil {
			if next.Key != key || next.Revision < old.cp.Revision || next.FenceEpoch != epoch || next.ZedToken == "" || next.Validate() != nil {
				return ErrCheckpointConflict
			}
			edges, err := json.Marshal(next.Edges)
			if err != nil {
				return err
			}
			if len(edges) > maxEdgesBytes {
				return ErrCheckpointConflict
			}
			_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $revision AS Int64;DECLARE $event AS Utf8;DECLARE $zed AS Utf8;DECLARE $fence AS Uint64;DECLARE $edges AS Utf8;DECLARE $updated AS Timestamp;
UPDATE `+"`m8-access/projection_checkpoints`"+` SET revision=$revision,event_id=$event,zed_token=$zed,fence_epoch=$fence,edges_json=$edges,updated_at=$updated WHERE tenant_id=$tenant AND projection_kind=$kind AND aggregate_id=$aggregate;`),
				append(keyArgs(key), sql.Named("revision", next.Revision), sql.Named("event", next.EventID), sql.Named("zed", next.ZedToken), sql.Named("fence", next.FenceEpoch), sql.Named("edges", string(edges)), sql.Named("updated", next.UpdatedAt.UTC()))...)
			if err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $until AS Timestamp;
UPDATE `+"`m8-access/projection_checkpoints`"+` SET lease_holder='',lease_until=$until WHERE tenant_id=$tenant AND projection_kind=$kind AND aggregate_id=$aggregate;`), append(keyArgs(key), sql.Named("until", now))...)
		return err
	})
}

// WithFenced acquires an exclusive distributed generation. A remote graph write
// is not transactional with YDB; the worker MUST first establish the graph
// guard and all subsequent writes MUST carry the graph-side CAS precondition.
func (s *Store) WithFenced(ctx context.Context, key projection.Key, fn func(context.Context, projection.Checkpoint, ports.GraphFence) (projection.Checkpoint, error)) error {
	if s == nil || fn == nil {
		return ErrInvalidStore
	}
	if err := key.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	acquired, err := s.acquire(ctx, key)
	if err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(ctx)
	var once sync.Once
	stop := func() { once.Do(cancel) }
	errorsCh := make(chan error, 1)
	go func() {
		tick := time.NewTicker(s.ttl / 3)
		defer tick.Stop()
		for {
			select {
			case <-runCtx.Done():
				errorsCh <- nil
				return
			case <-tick.C:
				if err := s.renew(runCtx, key, acquired.leaseEpoch); err != nil {
					stop()
					errorsCh <- err
					return
				}
			}
		}
	}()
	fence := ports.GraphFence{TenantID: key.TenantID, Kind: string(key.Kind), AggregateID: key.AggregateID, Epoch: acquired.leaseEpoch}
	next, workErr := fn(runCtx, acquired.cp, fence)
	stop()
	renewErr := <-errorsCh
	if renewErr != nil {
		return fmt.Errorf("%w: %v", ErrLeaseLost, renewErr)
	}
	if workErr != nil {
		return workErr
	} // lease left to expire: uncertified epoch blocks PDP
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.finish(ctx, key, acquired.leaseEpoch, &next)
}
