package ydbcheckpoint

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrWatermarkMissing  = errors.New("resource projection watermark missing")
	ErrWatermarkConflict = errors.New("resource projection watermark revision conflict")
)

var _ consistency.WatermarkReader = (*Store)(nil)

func resourceArgs(tenant domain.TenantID, resource domain.Resource) []any {
	return []any{sql.Named("tenant", string(tenant)), sql.Named("kind", resource.Type()), sql.Named("id", resource.ID())}
}

const selectWatermark = `DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;
SELECT source_revision, zed_token FROM \x60m8-access/resource_projection_watermarks\x60
WHERE tenant_id=$tenant AND resource_kind=$kind AND resource_id=$id;`

func (s *Store) LoadWatermark(ctx context.Context, tenant domain.TenantID, resource domain.Resource) (consistency.Watermark, error) {
	if s == nil || tenant.Validate() != nil || resource.Validate() != nil {
		return consistency.Watermark{}, ErrInvalidStore
	}
	var w consistency.Watermark
	err := s.tx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, s.q(selectWatermark), resourceArgs(tenant, resource)...).Scan(&w.Revision, &w.ZedToken)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWatermarkMissing
		}
		return err
	})
	if err != nil {
		return consistency.Watermark{}, err
	}
	if w.Revision == 0 || w.ZedToken == "" {
		return consistency.Watermark{}, ErrWatermarkMissing
	}
	return w, nil
}

// AdvanceWatermark is an internal-only call from the authenticated Resource
// Manager graph projection worker AFTER a successful SpiceDB batch. Revisions
// are strictly contiguous and checkpointed transactionally in YDB.
func (s *Store) AdvanceWatermark(ctx context.Context, tenant domain.TenantID, resource domain.Resource, revision uint64, zedToken string) error {
	if s == nil || tenant.Validate() != nil || resource.Validate() != nil || revision == 0 || zedToken == "" || len(zedToken) > 4096 {
		return ErrInvalidStore
	}
	return s.tx(ctx, func(tx *sql.Tx) error {
		var current uint64
		var previousToken string
		err := tx.QueryRowContext(ctx, s.q(selectWatermark), resourceArgs(tenant, resource)...).Scan(&current, &previousToken)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if errors.Is(err, sql.ErrNoRows) && revision != 1 {
			return ErrWatermarkConflict
		}
		if err == nil && revision == current && zedToken == previousToken {
			return nil
		}
		if err == nil && (revision <= current || revision != current+1) {
			return ErrWatermarkConflict
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Uint64;DECLARE $zed AS Utf8;DECLARE $updated AS Timestamp;
UPSERT INTO \x60m8-access/resource_projection_watermarks\x60(tenant_id,resource_kind,resource_id,source_revision,zed_token,updated_at)
VALUES ($tenant,$kind,$id,$revision,$zed,$updated);`), append(resourceArgs(tenant, resource), sql.Named("revision", revision), sql.Named("zed", zedToken), sql.Named("updated", now))...)
		if err != nil {
			return fmt.Errorf("commit resource watermark: %w", err)
		}
		return nil
	})
}
