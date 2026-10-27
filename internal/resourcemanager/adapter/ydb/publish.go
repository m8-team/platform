package ydb

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

// stage writes a complete immutable snapshot before it can be published.
// Failed publication leaves an unreachable snapshot for a retention-aware GC.
func (s *Store) stage(ctx context.Context, tenant, org, snapshot string, v orgunit.Structure) error {
	if len(v.Units)+len(v.Closure) > maxStagedRows {
		return ErrSnapshotTooLarge
	}
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if e != nil {
		return e
	}
	defer tx.Rollback()
	for _, u := range v.Units {
		_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8; DECLARE $snap AS Utf8;
  DECLARE $id AS Utf8; DECLARE $parent AS Utf8?; DECLARE $display AS Utf8;
  DECLARE $state AS Utf8; DECLARE $src AS Utf8?;
  INSERT INTO `+"`m8-resource-manager/org_unit_snapshots`"+`
  (tenant_id,organization_id,snapshot_id,org_unit_id,parent_org_unit_id,display_name,lifecycle_state,authoritative_source_id)
  VALUES ($tenant,$org,$snap,$id,$parent,$display,$state,$src);`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("snap", snapshot), sql.Named("id", u.ID), sql.Named("parent", nilIfEmpty(u.ParentID)), sql.Named("display", u.DisplayName), sql.Named("state", string(u.State)), sql.Named("src", nilIfEmpty(u.AuthoritativeSourceID)))
		if e != nil {
			return e
		}
	}
	for _, edge := range v.Closure {
		_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8; DECLARE $snap AS Utf8;
  DECLARE $ancestor AS Utf8; DECLARE $descendant AS Utf8; DECLARE $depth AS Uint32;
  INSERT INTO `+"`m8-resource-manager/org_unit_closure`"+`
  (tenant_id,organization_id,snapshot_id,ancestor_id,descendant_id,depth)
  VALUES ($tenant,$org,$snap,$ancestor,$descendant,$depth);`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("snap", snapshot), sql.Named("ancestor", edge.AncestorID), sql.Named("descendant", edge.DescendantID), sql.Named("depth", edge.Depth))
		if e != nil {
			return e
		}
	}
	return tx.Commit()
}

// Initialize seeds a newly provisioned organization's structure. It is a
// privileged provisioning operation, not exposed as unguarded public API.
func (s *Store) Initialize(ctx context.Context, tenant string, v orgunit.Structure, now time.Time) error {
	if tenant == "" || v.Revision != 1 || now.IsZero() {
		return orgunit.ErrInvalidUnit
	}
	rebuilt, e := orgunit.Build(v.OrganizationID, 1, v.Units)
	if e != nil || !reflect.DeepEqual(rebuilt, v) {
		return orgunit.ErrInvalidUnit
	}
	snapshot, e := newID()
	if e != nil {
		return e
	}
	if e = s.stage(ctx, tenant, v.OrganizationID, snapshot, v); e != nil {
		return e
	}
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8;
 DECLARE $snapshot AS Utf8; DECLARE $at AS Timestamp;
 INSERT INTO `+"`m8-resource-manager/org_structure_heads`"+`
 (tenant_id,organization_id,active_revision,active_snapshot_id,updated_at)
 VALUES ($tenant,$org,1u,$snapshot,$at);`), sql.Named("tenant", tenant), sql.Named("org", v.OrganizationID), sql.Named("snapshot", snapshot), sql.Named("at", now))
	if e != nil {
		return e
	}
	for _, unit := range v.Units {
		_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8;
 DECLARE $id AS Utf8; DECLARE $src AS Utf8?; DECLARE $at AS Timestamp;
 INSERT INTO `+"`m8-resource-manager/org_units`"+`
 (tenant_id,organization_id,org_unit_id,authoritative_source_id,create_time)
 VALUES ($tenant,$org,$id,$src,$at);`), sql.Named("tenant", tenant), sql.Named("org", v.OrganizationID), sql.Named("id", unit.ID), sql.Named("src", nilIfEmpty(unit.AuthoritativeSourceID)), sql.Named("at", now))
		if e != nil {
			return e
		}
	}
	return tx.Commit()
}

// Commit atomically publishes a staged snapshot, ChangeSet receipt and outbox.
// Serializable conflict on the active head is treated as a failed CAS.
func (s *Store) Commit(ctx context.Context, r orgstructure.CommitRequest) (orgstructure.Result, error) {
	if r.Actor.TenantID == "" || r.Actor.PrincipalID == "" || r.Plan.OrganizationID == "" || r.Plan.ID == "" || r.RequestHash == "" {
		return orgstructure.Result{}, orgunit.ErrInvalidChangeSet
	}
	tenant, org := r.Actor.TenantID, r.Plan.OrganizationID
	beforeHead, e := s.readHead(ctx, s.db, tenant, org)
	if e != nil {
		return orgstructure.Result{}, e
	}
	if beforeHead.revision != r.Plan.ExpectedRevision {
		return orgstructure.Result{}, orgunit.ErrRevisionConflict
	}
	before, e := s.readSnapshot(ctx, s.db, tenant, org, beforeHead)
	if e != nil {
		return orgstructure.Result{}, e
	}
	after, e := orgunit.Apply(before, r.Plan)
	if e != nil {
		return orgstructure.Result{}, e
	}
	if !reflect.DeepEqual(after, r.Snapshot) {
		return orgstructure.Result{}, orgunit.ErrInvalidChangeSet
	}
	nextSnapshot := beforeHead.snapshot
	if after.Revision != before.Revision {
		nextSnapshot, e = newID()
		if e != nil {
			return orgstructure.Result{}, e
		}
		if e = s.stage(ctx, tenant, org, nextSnapshot, after); e != nil {
			return orgstructure.Result{}, e
		}
	}
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if e != nil {
		return orgstructure.Result{}, e
	}
	defer tx.Rollback()
	current, e := s.readHead(ctx, tx, tenant, org)
	if e != nil {
		return orgstructure.Result{}, e
	}
	if current.revision != beforeHead.revision || current.snapshot != beforeHead.snapshot {
		return orgstructure.Result{}, orgunit.ErrRevisionConflict
	}
	changes, e := json.Marshal(r.Plan.Changes)
	if e != nil {
		return orgstructure.Result{}, e
	}
	_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8; DECLARE $id AS Utf8;
 DECLARE $source AS Utf8?; DECLARE $base AS Uint64; DECLARE $next AS Uint64;
 DECLARE $snap AS Utf8; DECLARE $hash AS String; DECLARE $effective AS Timestamp;
 DECLARE $recorded AS Timestamp; DECLARE $changes AS Utf8;
 INSERT INTO `+"`m8-resource-manager/org_unit_change_sets`"+`
 (tenant_id,organization_id,change_set_id,source_id,base_revision,target_revision,
 staged_snapshot_id,request_hash,status,effective_at,recorded_at,changes,created_at,updated_at)
 VALUES ($tenant,$org,$id,$source,$base,$next,$snap,$hash,'APPLIED',
 $effective,$recorded,CAST($changes AS JsonDocument),$recorded,$recorded);`),
		sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("id", r.Plan.ID), sql.Named("source", nilIfEmpty(r.Plan.SourceID)), sql.Named("base", r.Plan.ExpectedRevision), sql.Named("next", after.Revision), sql.Named("snap", nextSnapshot), sql.Named("hash", r.RequestHash), sql.Named("effective", r.Plan.EffectiveAt), sql.Named("recorded", r.Plan.RecordedAt), sql.Named("changes", string(changes)))
	if e != nil {
		return orgstructure.Result{}, e
	}
	if after.Revision != before.Revision {
		for _, m := range r.Plan.Changes {
			if m.Kind != orgunit.ChangeCreate {
				continue
			}
			_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8;
   DECLARE $id AS Utf8; DECLARE $src AS Utf8?; DECLARE $at AS Timestamp;
   INSERT INTO `+"`m8-resource-manager/org_units`"+`
   (tenant_id,organization_id,org_unit_id,authoritative_source_id,create_time)
   VALUES ($tenant,$org,$id,$src,$at);`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("id", m.UnitID), sql.Named("src", nilIfEmpty(r.Plan.SourceID)), sql.Named("at", r.Plan.RecordedAt))
			if e != nil {
				return orgstructure.Result{}, e
			}
		}
		_, e = tx.ExecContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8;
  DECLARE $revision AS Uint64; DECLARE $snap AS Utf8; DECLARE $at AS Timestamp;
  UPDATE `+"`m8-resource-manager/org_structure_heads`"+`
  SET active_revision=$revision,active_snapshot_id=$snap,updated_at=$at
  WHERE tenant_id=$tenant AND organization_id=$org;`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("revision", after.Revision), sql.Named("snap", nextSnapshot), sql.Named("at", r.Plan.RecordedAt))
		if e != nil {
			return orgstructure.Result{}, e
		}
		if e = s.writeOutbox(ctx, tx, r, after.Revision); e != nil {
			return orgstructure.Result{}, e
		}
	}
	if e = tx.Commit(); e != nil {
		return orgstructure.Result{}, e
	}
	return orgstructure.Result{Structure: after}, nil
}
func (s *Store) writeOutbox(ctx context.Context, tx *sql.Tx, r orgstructure.CommitRequest, revision uint64) error {
	eventID, e := newID()
	if e != nil {
		return e
	}
	event, e := json.Marshal(struct {
		TenantID          string `json:"tenant_id"`
		OrganizationID    string `json:"organization_id"`
		ChangeSetID       string `json:"change_set_id"`
		StructureRevision uint64 `json:"structure_revision"`
		PrincipalID       string `json:"actor_principal_id"`
	}{r.Actor.TenantID, r.Plan.OrganizationID, r.Plan.ID, revision, r.Actor.PrincipalID})
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, s.q(`DECLARE $event AS Utf8; DECLARE $org AS Utf8;
 DECLARE $revision AS Uint64; DECLARE $partition AS Utf8;
 DECLARE $payload AS Utf8; DECLARE $at AS Timestamp;
 INSERT INTO `+"`m8-resource-manager/outbox`"+`
 (event_id,aggregate_type,aggregate_id,aggregate_revision,event_type,topic,partition_key,
 payload,status,attempt_count,created_at)
 VALUES ($event,'org_structure',$org,$revision,'m8.resourcemanager.org_structure.changed.v1',
 'm8.platform.resource-manager.org-structure.v1',$partition,
 CAST($payload AS JsonDocument),'PENDING',0u,$at);`), sql.Named("event", eventID), sql.Named("org", r.Plan.OrganizationID), sql.Named("revision", revision), sql.Named("partition", r.Actor.TenantID+"/"+r.Plan.OrganizationID), sql.Named("payload", string(event)), sql.Named("at", r.Plan.RecordedAt))
	return e
}
