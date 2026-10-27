// Package ydb persists revisioned OrgUnit hierarchies using database/sql backed by YDB.
package ydb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

var (
	ErrInvalidStore     = errors.New("invalid YDB config")
	ErrSnapshotTooLarge = errors.New("snapshot exceeds initial transaction limit")
	ErrCorruptSnapshot  = errors.New("org structure snapshot inconsistent")
)

const maxStagedRows = 10000

var safePrefix = regexp.MustCompile(`^/[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*$`)

type Store struct {
	db     *sql.DB
	prefix string
}

func NewStore(db *sql.DB, prefix string) (*Store, error) {
	if db == nil || !safePrefix.MatchString(prefix) {
		return nil, ErrInvalidStore
	}
	return &Store{db: db, prefix: prefix}, nil
}
func (s *Store) q(query string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n", s.prefix) + query
}
func newID() (string, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	return hex.EncodeToString(b[:]), nil
}
func nilIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
type head struct {
	revision uint64
	snapshot string
}

func (s *Store) readHead(ctx context.Context, q queryer, tenant, org string) (head, error) {
	var h head
	e := q.QueryRowContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8;
 SELECT active_revision,active_snapshot_id FROM `+"`m8-resource-manager/org_structure_heads`"+`
 WHERE tenant_id=$tenant AND organization_id=$org;`), sql.Named("tenant", tenant), sql.Named("org", org)).Scan(&h.revision, &h.snapshot)
	if errors.Is(e, sql.ErrNoRows) {
		return head{}, orgstructure.ErrStoreNotFound
	}
	if e != nil {
		return head{}, e
	}
	if h.revision == 0 || h.snapshot == "" {
		return head{}, ErrCorruptSnapshot
	}
	return h, nil
}
func (s *Store) readSnapshot(ctx context.Context, q queryer, tenant, org string, h head) (orgunit.Structure, error) {
	rows, e := q.QueryContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8; DECLARE $snap AS Utf8;
 SELECT org_unit_id,parent_org_unit_id,display_name,lifecycle_state,authoritative_source_id
 FROM `+"`m8-resource-manager/org_unit_snapshots`"+`
 WHERE tenant_id=$tenant AND organization_id=$org AND snapshot_id=$snap;`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("snap", h.snapshot))
	if e != nil {
		return orgunit.Structure{}, e
	}
	units := []orgunit.Unit{}
	for rows.Next() {
		var u orgunit.Unit
		var p, src sql.NullString
		if e = rows.Scan(&u.ID, &p, &u.DisplayName, &u.State, &src); e != nil {
			break
		}
		u.OrganizationID = org
		u.ParentID = p.String
		u.AuthoritativeSourceID = src.String
		units = append(units, u)
	}
	if e == nil {
		e = rows.Err()
	}
	if c := rows.Close(); e == nil {
		e = c
	}
	if e != nil {
		return orgunit.Structure{}, e
	}
	built, e := orgunit.Build(org, h.revision, units)
	if e != nil {
		return orgunit.Structure{}, fmt.Errorf("%w: %v", ErrCorruptSnapshot, e)
	}
	rows, e = q.QueryContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8; DECLARE $snap AS Utf8;
 SELECT ancestor_id,descendant_id,depth FROM `+"`m8-resource-manager/org_unit_closure`"+`
 WHERE tenant_id=$tenant AND organization_id=$org AND snapshot_id=$snap;`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("snap", h.snapshot))
	if e != nil {
		return orgunit.Structure{}, e
	}
	edges := []orgunit.ClosureRow{}
	for rows.Next() {
		var v orgunit.ClosureRow
		if e = rows.Scan(&v.AncestorID, &v.DescendantID, &v.Depth); e != nil {
			break
		}
		v.OrganizationID = org
		edges = append(edges, v)
	}
	if e == nil {
		e = rows.Err()
	}
	if c := rows.Close(); e == nil {
		e = c
	}
	if e != nil {
		return orgunit.Structure{}, e
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].DescendantID != edges[j].DescendantID {
			return edges[i].DescendantID < edges[j].DescendantID
		}
		return edges[i].Depth < edges[j].Depth
	})
	if !reflect.DeepEqual(edges, built.Closure) {
		return orgunit.Structure{}, ErrCorruptSnapshot
	}
	return built, nil
}
func (s *Store) Load(ctx context.Context, tenant, org string) (orgunit.Structure, error) {
	if tenant == "" || org == "" {
		return orgunit.Structure{}, orgunit.ErrInvalidUnit
	}
	h, e := s.readHead(ctx, s.db, tenant, org)
	if e != nil {
		return orgunit.Structure{}, e
	}
	return s.readSnapshot(ctx, s.db, tenant, org, h)
}
func (s *Store) FindIdempotent(ctx context.Context, tenant, org, csID, hash string) (orgstructure.Result, bool, error) {
	var stored, status, snapshot string
	var revision uint64
	e := s.db.QueryRowContext(ctx, s.q(`DECLARE $tenant AS Utf8; DECLARE $org AS Utf8; DECLARE $id AS Utf8;
 SELECT request_hash,status,staged_snapshot_id,target_revision FROM `+"`m8-resource-manager/org_unit_change_sets`"+`
 WHERE tenant_id=$tenant AND organization_id=$org AND change_set_id=$id;`), sql.Named("tenant", tenant), sql.Named("org", org), sql.Named("id", csID)).Scan(&stored, &status, &snapshot, &revision)
	if errors.Is(e, sql.ErrNoRows) {
		return orgstructure.Result{}, false, nil
	}
	if e != nil {
		return orgstructure.Result{}, false, e
	}
	if stored != hash {
		return orgstructure.Result{}, false, orgstructure.ErrIdempotencyConflict
	}
	if status != "APPLIED" || snapshot == "" || revision == 0 {
		return orgstructure.Result{}, false, ErrCorruptSnapshot
	}
	v, e := s.readSnapshot(ctx, s.db, tenant, org, head{revision: revision, snapshot: snapshot})
	if e != nil {
		return orgstructure.Result{}, false, e
	}
	return orgstructure.Result{Structure: v, Replayed: true}, true, nil
}
