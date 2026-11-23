package ydbcheckpoint

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrWatermarkMissing  = errors.New("resource projection watermark missing")
	ErrWatermarkConflict = errors.New("resource projection watermark revision conflict")
)

var _ consistency.WatermarkReader = (*Store)(nil)
var _ projection.ResourceWatermarkCommitter = (*Store)(nil)

func resourceArgs(tenant domain.TenantID, resource domain.Resource) []any {
	return []any{sql.Named("tenant", string(tenant)), sql.Named("kind", resource.Type()), sql.Named("id", resource.ID())}
}

const selectWatermark = `DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;
SELECT source_revision, zed_token FROM ` + "`m8-access/resource_projection_watermarks`" + `
WHERE tenant_id=$tenant AND resource_kind=$kind AND resource_id=$id;`

// LoadWatermark validates the actual certified aggregate checkpoint IN THE SAME
// serializable read. During a lease takeover a stale watermark is not trusted.
func (s *Store) LoadWatermark(ctx context.Context, tenant domain.TenantID, resource domain.Resource) (consistency.Watermark, error) {
	if s == nil || tenant.Validate() != nil || resource.Validate() != nil {
		return consistency.Watermark{}, ErrInvalidStore
	}
	kind, err := projection.KindForResourceType(resource.Type())
	if err != nil {
		return consistency.Watermark{}, ErrWatermarkMissing
	}
	key := projection.Key{TenantID: tenant, Kind: kind, AggregateID: resource.ID()}
	if err := key.Validate(); err != nil {
		return consistency.Watermark{}, err
	}
	var watermark consistency.Watermark
	err = s.tx(ctx, func(tx *sql.Tx) error {
		cpRow, err := s.load(ctx, tx, key)
		if err != nil {
			return err
		}
		if !cpRow.exists || cpRow.cp.Revision < 1 || cpRow.cp.FenceEpoch == 0 ||
			cpRow.cp.FenceEpoch != cpRow.leaseEpoch || cpRow.cp.ZedToken == "" {
			return ErrCheckpointNotConverged
		}
		err = tx.QueryRowContext(ctx, s.q(selectWatermark), resourceArgs(tenant, resource)...).
			Scan(&watermark.Revision, &watermark.ZedToken)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWatermarkMissing
		}
		if err != nil {
			return err
		}
		if watermark.Revision == 0 || watermark.Revision != uint64(cpRow.cp.Revision) ||
			watermark.ZedToken != cpRow.cp.ZedToken {
			return ErrCheckpointNotConverged
		}
		return nil
	})
	if err != nil {
		return consistency.Watermark{}, err
	}
	return watermark, nil
}

// CommitResourceWatermark derives progress exclusively from the certified
// fenced checkpoint. No caller may pass a revision or an opaque ZedToken.
// It commits after the remote SpiceDB write and checkpoint commit; if this
// transaction fails, the event must not be acknowledged.
func (s *Store) CommitResourceWatermark(ctx context.Context, key projection.Key) error {
	if s == nil || key.Validate() != nil || !projection.IsResourceKind(key.Kind) {
		return ErrInvalidStore
	}
	resourceType, err := projection.ResourceTypeForKind(key.Kind)
	if err != nil {
		return err
	}
	resource, err := domain.NewResource(resourceType, key.AggregateID)
	if err != nil {
		return err
	}
	return s.tx(ctx, func(tx *sql.Tx) error {
		row, err := s.load(ctx, tx, key)
		if err != nil {
			return err
		}
		if !row.exists || row.cp.Revision < 1 || row.cp.ZedToken == "" ||
			row.cp.FenceEpoch == 0 || row.cp.FenceEpoch != row.leaseEpoch {
			return ErrCheckpointNotConverged
		}
		var current uint64
		var previousToken string
		err = tx.QueryRowContext(ctx, s.q(selectWatermark), resourceArgs(key.TenantID, resource)...).
			Scan(&current, &previousToken)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil && current > uint64(row.cp.Revision) {
			return ErrWatermarkConflict
		}
		if err == nil && current == uint64(row.cp.Revision) && previousToken == row.cp.ZedToken {
			return nil
		}
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Uint64;DECLARE $zed AS Utf8;DECLARE $updated AS Timestamp;
UPSERT INTO `+"`m8-access/resource_projection_watermarks`"+`(tenant_id,resource_kind,resource_id,source_revision,zed_token,updated_at)
VALUES ($tenant,$kind,$id,$revision,$zed,$updated);`), append(resourceArgs(key.TenantID, resource),
			sql.Named("revision", uint64(row.cp.Revision)), sql.Named("zed", row.cp.ZedToken), sql.Named("updated", now))...)
		if err != nil {
			return fmt.Errorf("commit certified RM watermark: %w", err)
		}
		return nil
	})
}
