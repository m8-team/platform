package ydb

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

const relationshipsTable = "`m8-resource-manager/organization_relationships`"
const relationshipRequestsTable = "`m8-resource-manager/organization_relationship_requests`"
const relationshipHistoryTable = "`m8-resource-manager/organization_relationship_history`"
const relationshipHeadsTable = "`m8-resource-manager/organization_relationship_graph_heads`"

const relationColumns = "tenant_id,id,from_organization_id,to_organization_id,relationship_type,ownership_basis_points,valid_from_unix,valid_to_unix,evidence_ref,state,revision,created_at,updated_at"

func scanRelationship(rows interface{ Scan(...any) error }) (v1engine.OrganizationRelationship, error) {
	var r v1engine.OrganizationRelationship
	var typ string
	err := rows.Scan(&r.TenantID, &r.ID, &r.FromOrganizationID, &r.ToOrganizationID, &typ,
		&r.OwnershipBasisPoints, &r.ValidFrom, &r.ValidTo, &r.EvidenceRef, &r.State, &r.Revision, &r.CreatedAt, &r.UpdatedAt)
	r.Type = v1engine.RelationshipType(typ)
	if errors.Is(err, sql.ErrNoRows) {
		return r, v1engine.ErrNotFound
	}
	return r, err
}
func (c *Catalog) relation(ctx context.Context, q queryer, tenant, id string) (v1engine.OrganizationRelationship, error) {
	return scanRelationship(q.QueryRowContext(ctx, c.stmt(
		"DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;SELECT "+relationColumns+
			" FROM "+relationshipsTable+" WHERE tenant_id=$tenant AND id=$id;"),
		sql.Named("tenant", tenant), sql.Named("id", id)))
}
func (c *Catalog) GetRelationship(ctx context.Context, tenant, id string) (v1engine.OrganizationRelationship, error) {
	if tenant == "" || id == "" {
		return v1engine.OrganizationRelationship{}, v1engine.ErrInvalid
	}
	return c.relation(ctx, c.store.db, tenant, id)
}
func (c *Catalog) listRelationships(ctx context.Context, q queryer, tenant string) ([]v1engine.OrganizationRelationship, error) {
	rows, err := q.QueryContext(ctx, c.stmt(
		"DECLARE $tenant AS Utf8;SELECT "+relationColumns+" FROM "+relationshipsTable+" WHERE tenant_id=$tenant;"),
		sql.Named("tenant", tenant))
	if err != nil {
		return nil, fmt.Errorf("list corporate relations: %w", err)
	}
	defer rows.Close()
	var out []v1engine.OrganizationRelationship
	for rows.Next() {
		r, err := scanRelationship(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (c *Catalog) ListRelationships(ctx context.Context, tenant, organization string, includeClosed bool) ([]v1engine.OrganizationRelationship, error) {
	if tenant == "" {
		return nil, v1engine.ErrInvalid
	}
	if _, err := c.Get(ctx, v1engine.Tenant, tenant); err != nil {
		return nil, err
	}
	if organization != "" {
		org, err := c.Get(ctx, v1engine.Organization, organization)
		if err != nil {
			return nil, err
		}
		if org.TenantID != tenant {
			return nil, v1engine.ErrNotFound
		}
	}
	relations, err := c.listRelationships(ctx, c.store.db, tenant)
	if err != nil {
		return nil, err
	}
	return v1engine.FilterRelationships(relations, tenant, organization, includeClosed), nil
}

func (c *Catalog) relationshipReplay(ctx context.Context, q queryer, tenant, key string, expected [32]byte) (string, bool, error) {
	var hash []byte
	var id string
	err := q.QueryRowContext(ctx, c.stmt(
		"DECLARE $tenant AS Utf8;DECLARE $key AS Utf8;SELECT request_hash,relationship_id FROM "+
			relationshipRequestsTable+" WHERE tenant_id=$tenant AND idempotency_key=$key;"),
		sql.Named("tenant", tenant), sql.Named("key", key)).Scan(&hash, &id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if hex.EncodeToString(hash) != hex.EncodeToString(expected[:]) {
		return "", true, v1engine.ErrConflict
	}
	return id, true, nil
}
func (c *Catalog) saveRelationshipReceipt(ctx context.Context, tx *sql.Tx, tenant, key string, hash [32]byte, id string) error {
	_, err := tx.ExecContext(ctx, c.stmt("DECLARE $tenant AS Utf8;DECLARE $key AS Utf8;DECLARE $hash AS String;DECLARE $id AS Utf8;INSERT INTO "+
		relationshipRequestsTable+" (tenant_id,idempotency_key,request_hash,relationship_id) VALUES($tenant,$key,$hash,$id);"),
		sql.Named("tenant", tenant), sql.Named("key", key), sql.Named("hash", hash[:]), sql.Named("id", id))
	return err
}
func (c *Catalog) saveRelationshipHistory(ctx context.Context, tx *sql.Tx, r v1engine.OrganizationRelationship) error {
	_, err := tx.ExecContext(ctx, c.stmt("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;"+
		"DECLARE $from AS Utf8;DECLARE $to AS Utf8;DECLARE $type AS Utf8;DECLARE $bps AS Int32;"+
		"DECLARE $start AS Int64;DECLARE $end AS Int64;DECLARE $evidence AS Utf8;DECLARE $state AS Utf8;DECLARE $updated AS Timestamp;"+
		"INSERT INTO "+relationshipHistoryTable+
		" (tenant_id,relationship_id,revision,from_organization_id,to_organization_id,relationship_type,"+
		"ownership_basis_points,valid_from_unix,valid_to_unix,evidence_ref,state,updated_at) "+
		"VALUES($tenant,$id,$revision,$from,$to,$type,$bps,$start,$end,$evidence,$state,$updated);"),
		sql.Named("tenant", r.TenantID), sql.Named("id", r.ID), sql.Named("revision", r.Revision),
		sql.Named("from", r.FromOrganizationID), sql.Named("to", r.ToOrganizationID),
		sql.Named("type", string(r.Type)), sql.Named("bps", r.OwnershipBasisPoints),
		sql.Named("start", r.ValidFrom), sql.Named("end", r.ValidTo),
		sql.Named("evidence", r.EvidenceRef), sql.Named("state", r.State), sql.Named("updated", r.UpdatedAt))
	return err
}

// All concurrent relationship commands for one Tenant write this same row.
// They must retry the entire serializable transaction on YDB transaction conflicts.
func (c *Catalog) bumpRelationshipHead(ctx context.Context, tx *sql.Tx, tenant string, now time.Time) error {
	var revision int64
	err := tx.QueryRowContext(ctx, c.stmt("DECLARE $tenant AS Utf8;SELECT revision FROM "+relationshipHeadsTable+
		" WHERE tenant_id=$tenant;"), sql.Named("tenant", tenant)).Scan(&revision)
	if err == nil {
		_, err = tx.ExecContext(ctx, c.stmt("DECLARE $tenant AS Utf8;DECLARE $rev AS Int64;DECLARE $at AS Timestamp;UPDATE "+
			relationshipHeadsTable+" SET revision=$rev,updated_at=$at WHERE tenant_id=$tenant;"),
			sql.Named("tenant", tenant), sql.Named("rev", revision+1), sql.Named("at", now))
		return err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = tx.ExecContext(ctx, c.stmt("DECLARE $tenant AS Utf8;DECLARE $at AS Timestamp;INSERT INTO "+
		relationshipHeadsTable+" (tenant_id,revision,updated_at) VALUES($tenant,1,$at);"),
		sql.Named("tenant", tenant), sql.Named("at", now))
	return err
}
func (c *Catalog) insertRelationship(ctx context.Context, tx *sql.Tx, r v1engine.OrganizationRelationship) error {
	_, err := tx.ExecContext(ctx, c.stmt(
		"DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $from AS Utf8;DECLARE $to AS Utf8;DECLARE $type AS Utf8;"+
			"DECLARE $bps AS Int32;DECLARE $start AS Int64;DECLARE $end AS Int64;DECLARE $evidence AS Utf8;"+
			"DECLARE $state AS Utf8;DECLARE $revision AS Int64;DECLARE $created AS Timestamp;DECLARE $updated AS Timestamp;"+
			"INSERT INTO "+relationshipsTable+" ("+relationColumns+") "+
			"VALUES($tenant,$id,$from,$to,$type,$bps,$start,$end,$evidence,$state,$revision,$created,$updated);"),
		sql.Named("tenant", r.TenantID), sql.Named("id", r.ID), sql.Named("from", r.FromOrganizationID),
		sql.Named("to", r.ToOrganizationID), sql.Named("type", string(r.Type)), sql.Named("bps", r.OwnershipBasisPoints),
		sql.Named("start", r.ValidFrom), sql.Named("end", r.ValidTo),
		sql.Named("evidence", r.EvidenceRef), sql.Named("state", r.State),
		sql.Named("revision", r.Revision), sql.Named("created", r.CreatedAt), sql.Named("updated", r.UpdatedAt))
	return err
}
func (c *Catalog) updateRelationship(ctx context.Context, tx *sql.Tx, r v1engine.OrganizationRelationship) error {
	_, err := tx.ExecContext(ctx, c.stmt(
		"DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $bps AS Int32;DECLARE $end AS Int64;"+
			"DECLARE $evidence AS Utf8;DECLARE $state AS Utf8;DECLARE $revision AS Int64;DECLARE $updated AS Timestamp;"+
			"UPDATE "+relationshipsTable+" SET ownership_basis_points=$bps,valid_to_unix=$end,evidence_ref=$evidence,"+
			"state=$state,revision=$revision,updated_at=$updated WHERE tenant_id=$tenant AND id=$id;"),
		sql.Named("tenant", r.TenantID), sql.Named("id", r.ID), sql.Named("bps", r.OwnershipBasisPoints),
		sql.Named("end", r.ValidTo), sql.Named("evidence", r.EvidenceRef),
		sql.Named("state", r.State), sql.Named("revision", r.Revision), sql.Named("updated", r.UpdatedAt))
	return err
}
func (c *Catalog) relationshipEvent(ctx context.Context, tx *sql.Tx, r v1engine.OrganizationRelationship, typ string) error {
	return c.event(ctx, tx, v1engine.Record{
		Kind: v1engine.Kind("organization_relationship"), ID: r.ID, TenantID: r.TenantID, Revision: r.Revision,
	}, typ)
}

func (c *Catalog) CreateRelationship(ctx context.Context, r v1engine.OrganizationRelationship, key string) (v1engine.OrganizationRelationship, error) {
	if key == "" || r.Validate() != nil {
		return v1engine.OrganizationRelationship{}, v1engine.ErrInvalid
	}
	hash, err := v1engine.RelationshipHash("create", r)
	if err != nil {
		return v1engine.OrganizationRelationship{}, err
	}
	var out v1engine.OrganizationRelationship
	err = c.transaction(ctx, func(tx *sql.Tx) error {
		id, found, err := c.relationshipReplay(ctx, tx, r.TenantID, key, hash)
		if err != nil {
			return err
		}
		if found {
			out, err = c.relation(ctx, tx, r.TenantID, id)
			return err
		}
		from, err := c.get(ctx, tx, v1engine.Organization, r.FromOrganizationID)
		if err != nil {
			return err
		}
		to, err := c.get(ctx, tx, v1engine.Organization, r.ToOrganizationID)
		if err != nil {
			return err
		}
		if from.TenantID != r.TenantID || to.TenantID != r.TenantID {
			return v1engine.ErrInvalid
		}
		if from.State != "ACTIVE" || to.State != "ACTIVE" {
			return v1engine.ErrConflict
		}
		all, err := c.listRelationships(ctx, tx, r.TenantID)
		if err != nil {
			return err
		}
		if err = v1engine.ValidateRelationshipGraph(all, r); err != nil {
			return err
		}
		now := time.Now().UTC()
		r.ID = uuid.NewString()
		r.State, r.Revision, r.CreatedAt, r.UpdatedAt = "ACTIVE", 1, now, now
		if err = c.insertRelationship(ctx, tx, r); err != nil {
			return err
		}
		if err = c.saveRelationshipHistory(ctx, tx, r); err != nil {
			return err
		}
		if err = c.saveRelationshipReceipt(ctx, tx, r.TenantID, key, hash, r.ID); err != nil {
			return err
		}
		if err = c.bumpRelationshipHead(ctx, tx, r.TenantID, now); err != nil {
			return err
		}
		if err = c.relationshipEvent(ctx, tx, r, "m8.resourcemanager.organization_relationship.created.v1"); err != nil {
			return err
		}
		out = r
		return nil
	})
	return out, err
}
func (c *Catalog) UpdateRelationship(ctx context.Context, u v1engine.RelationshipUpdate, key string) (v1engine.OrganizationRelationship, error) {
	if key == "" || u.ID == "" || u.TenantID == "" || u.ExpectedRevision <= 0 {
		return v1engine.OrganizationRelationship{}, v1engine.ErrInvalid
	}
	hash := v1engine.RelationshipUpdateHash("update", u)
	var out v1engine.OrganizationRelationship
	err := c.transaction(ctx, func(tx *sql.Tx) error {
		id, found, err := c.relationshipReplay(ctx, tx, u.TenantID, key, hash)
		if err != nil {
			return err
		}
		if found {
			out, err = c.relation(ctx, tx, u.TenantID, id)
			return err
		}
		r, err := c.relation(ctx, tx, u.TenantID, u.ID)
		if err != nil {
			return err
		}
		if r.Revision != u.ExpectedRevision || r.State != "ACTIVE" {
			return v1engine.ErrConflict
		}
		r.OwnershipBasisPoints = u.OwnershipBasisPoints
		if u.ValidTo != 0 {
			r.ValidTo = u.ValidTo
		}
		r.EvidenceRef = u.EvidenceRef
		if err = r.Validate(); err != nil {
			return err
		}
		existing, err := c.listRelationships(ctx, tx, u.TenantID)
		if err != nil {
			return err
		}
		if err = v1engine.ValidateRelationshipGraph(existing, r); err != nil {
			return err
		}
		r.Revision++
		r.UpdatedAt = time.Now().UTC()
		if err = c.updateRelationship(ctx, tx, r); err != nil {
			return err
		}
		if err = c.saveRelationshipHistory(ctx, tx, r); err != nil {
			return err
		}
		if err = c.saveRelationshipReceipt(ctx, tx, u.TenantID, key, hash, r.ID); err != nil {
			return err
		}
		if err = c.bumpRelationshipHead(ctx, tx, u.TenantID, r.UpdatedAt); err != nil {
			return err
		}
		if err = c.relationshipEvent(ctx, tx, r, "m8.resourcemanager.organization_relationship.updated.v1"); err != nil {
			return err
		}
		out = r
		return nil
	})
	return out, err
}
func (c *Catalog) CloseRelationship(ctx context.Context, tenant, id string, revision, closedAt int64, key string) (v1engine.OrganizationRelationship, error) {
	if tenant == "" || id == "" || revision <= 0 || key == "" || closedAt <= 0 {
		return v1engine.OrganizationRelationship{}, v1engine.ErrInvalid
	}
	hash := v1engine.RelationshipUpdateHash("close", v1engine.RelationshipUpdate{TenantID: tenant, ID: id, ExpectedRevision: revision, ValidTo: closedAt})
	var out v1engine.OrganizationRelationship
	err := c.transaction(ctx, func(tx *sql.Tx) error {
		replayID, found, err := c.relationshipReplay(ctx, tx, tenant, key, hash)
		if err != nil {
			return err
		}
		if found {
			out, err = c.relation(ctx, tx, tenant, replayID)
			return err
		}
		r, err := c.relation(ctx, tx, tenant, id)
		if err != nil {
			return err
		}
		if r.Revision != revision || r.State != "ACTIVE" {
			return v1engine.ErrConflict
		}
		if closedAt <= r.ValidFrom || (r.ValidTo != 0 && closedAt > r.ValidTo) ||
			closedAt > time.Now().Unix()+60 {
			return v1engine.ErrInvalid
		}
		r.ValidTo, r.State, r.Revision, r.UpdatedAt = closedAt, "CLOSED", r.Revision+1, time.Now().UTC()
		if err = c.updateRelationship(ctx, tx, r); err != nil {
			return err
		}
		if err = c.saveRelationshipHistory(ctx, tx, r); err != nil {
			return err
		}
		if err = c.saveRelationshipReceipt(ctx, tx, tenant, key, hash, id); err != nil {
			return err
		}
		if err = c.bumpRelationshipHead(ctx, tx, tenant, r.UpdatedAt); err != nil {
			return err
		}
		if err = c.relationshipEvent(ctx, tx, r, "m8.resourcemanager.organization_relationship.closed.v1"); err != nil {
			return err
		}
		out = r
		return nil
	})
	return out, err
}
