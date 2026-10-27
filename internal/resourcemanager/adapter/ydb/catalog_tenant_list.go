package ydb

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

// ListTenantOrganizations loads only direct child Organizations of a known
// Tenant. Never interpret an empty tenant ID as permission to list globally.
func (c *Catalog) ListTenantOrganizations(ctx context.Context, tenantID string) ([]v1engine.Record, error) {
	if tenantID == "" {
		return nil, v1engine.ErrInvalid
	}
	if _, err := c.Get(ctx, v1engine.Tenant, tenantID); err != nil {
		return nil, err
	}
	rows, err := c.store.db.QueryContext(ctx, c.stmt(fmt.Sprintf(
		"DECLARE $tenant AS Utf8; DECLARE $kind AS Utf8; "+
			"SELECT kind,id,tenant_id,owner_kind,owner_id,organization_id,display_name,description,state,revision,environment_key,created_at,updated_at "+
			"FROM %s WHERE tenant_id=$tenant AND kind=$kind AND state!='DELETED';",
		resourceTable(),
	)), sql.Named("tenant", tenantID), sql.Named("kind", string(v1engine.Organization)))
	if err != nil {
		return nil, fmt.Errorf("list tenant organizations: %w", err)
	}
	defer rows.Close()

	out := make([]v1engine.Record, 0)
	for rows.Next() {
		var r v1engine.Record
		var kind, ownerKind string
		if err := rows.Scan(&kind, &r.ID, &r.TenantID, &ownerKind, &r.Owner.ID,
			&r.OrganizationID, &r.Name, &r.Description, &r.State, &r.Revision,
			&r.EnvironmentKey, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Kind = v1engine.Kind(kind)
		r.Owner.Kind = v1engine.Kind(ownerKind)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}
