package ydbcatalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/app/catalog"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidStore = errors.New("invalid YDB authorization catalog store")
var safePrefix = regexp.MustCompile("^/[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*$")

// Store owns the authoritative release state, idempotency receipts,
// materialized read index and durable outbox in one YDB transaction.
type Store struct {
	db        *sql.DB
	prefix    string
	tokenKey  []byte
	cursorTTL time.Duration
	now       func() time.Time
}

var _ catalog.Registry = (*Store)(nil)

func New(db *sql.DB, prefix string, tokenKey []byte, cursorTTL time.Duration) (*Store, error) {
	if db == nil || !safePrefix.MatchString(prefix) || len(tokenKey) < 32 || cursorTTL < time.Minute || cursorTTL > 24*time.Hour {
		return nil, ErrInvalidStore
	}
	return &Store{db: db, prefix: prefix, tokenKey: append([]byte(nil), tokenKey...), cursorTTL: cursorTTL, now: time.Now}, nil
}
func (s *Store) q(query string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s", s.prefix, query)
}
func (s *Store) tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) serverTime(ctx context.Context, tx *sql.Tx) (time.Time, error) {
	var now time.Time
	if err := tx.QueryRowContext(ctx, s.q("SELECT CurrentUtcTimestamp();")).Scan(&now); err != nil {
		return time.Time{}, err
	}
	return now.UTC(), nil
}

const selectReceipt = "DECLARE $module AS Utf8;DECLARE $key AS Utf8;\nSELECT request_hash,version,revision FROM \x60m8-access/catalog_receipts\x60 WHERE module_id=$module AND request_key=$key;"
const selectHead = "DECLARE $module AS Utf8;\nSELECT head_revision FROM \x60m8-access/catalog_heads\x60 WHERE module_id=$module;"
const selectRelease = "DECLARE $module AS Utf8;DECLARE $version AS Utf8;\nSELECT revision,manifest_json FROM \x60m8-access/catalog_releases\x60 WHERE module_id=$module AND version=$version;"

func (s *Store) readRelease(ctx context.Context, db interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, moduleID, version string) (catalog.PublishedRecord, error) {
	var revision int64
	var jsonText string
	err := db.QueryRowContext(ctx, s.q(selectRelease), sql.Named("module", moduleID), sql.Named("version", version)).Scan(&revision, &jsonText)
	if errors.Is(err, sql.ErrNoRows) {
		return catalog.PublishedRecord{}, catalog.ErrCatalogNotFound
	}
	if err != nil {
		return catalog.PublishedRecord{}, err
	}
	manifest, err := unmarshalManifest([]byte(jsonText))
	if err != nil || manifest.ModuleID() != moduleID || manifest.Version() != version {
		return catalog.PublishedRecord{}, catalog.ErrCatalogStoreInvalid
	}
	result := catalog.PublishedRecord{Manifest: manifest, Revision: revision}
	if err := result.Validate(); err != nil {
		return catalog.PublishedRecord{}, err
	}
	return result, nil
}

// Publish serializes a complete authoritative catalog command and its read
// projection and outbox entry. Replays are checked BEFORE the catalog head.
func (s *Store) Publish(ctx context.Context, m module.Manifest, expectedRevision int64, key string) (catalog.PublishedRecord, error) {
	if s == nil || s.db == nil {
		return catalog.PublishedRecord{}, ErrInvalidStore
	}
	if err := ctx.Err(); err != nil {
		return catalog.PublishedRecord{}, err
	}
	if m.Validate() != nil || expectedRevision < 0 || key == "" || strings.TrimSpace(key) != key || len(key) > 256 {
		return catalog.PublishedRecord{}, catalog.ErrCatalogArgument
	}
	for _, p := range m.Permissions() {
		if p.GraphPermission() != "use" {
			return catalog.PublishedRecord{}, catalog.ErrCatalogArgument
		}
	}
	payload, err := marshalManifest(m)
	if err != nil {
		return catalog.PublishedRecord{}, catalog.ErrCatalogArgument
	}
	hash := requestHash(payload, expectedRevision)
	var published catalog.PublishedRecord
	err = s.tx(ctx, func(tx *sql.Tx) error {
		var oldHash, oldVersion string
		var oldRevision int64
		lookupErr := tx.QueryRowContext(ctx, s.q(selectReceipt), sql.Named("module", m.ModuleID()), sql.Named("key", key)).
			Scan(&oldHash, &oldVersion, &oldRevision)
		if lookupErr == nil {
			if oldHash != hash || oldVersion != m.Version() {
				return catalog.ErrCatalogConflict
			}
			record, err := s.readRelease(ctx, tx, m.ModuleID(), oldVersion)
			if err != nil {
				return err
			}
			if record.Revision != oldRevision {
				return catalog.ErrCatalogStoreInvalid
			}
			published = record
			return nil
		}
		if !errors.Is(lookupErr, sql.ErrNoRows) {
			return lookupErr
		}
		var head int64
		headErr := tx.QueryRowContext(ctx, s.q(selectHead), sql.Named("module", m.ModuleID())).Scan(&head)
		if headErr != nil && !errors.Is(headErr, sql.ErrNoRows) {
			return headErr
		}
		if head != expectedRevision {
			return shared.ErrRevisionConflict
		}
		if head == int64(^uint64(0)>>1) {
			return catalog.ErrCatalogConflict
		}
		var previousRev int64
		var previousJSON string
		previousErr := tx.QueryRowContext(ctx, s.q(selectRelease), sql.Named("module", m.ModuleID()), sql.Named("version", m.Version())).
			Scan(&previousRev, &previousJSON)
		if previousErr == nil {
			return catalog.ErrCatalogConflict
		}
		if !errors.Is(previousErr, sql.ErrNoRows) {
			return previousErr
		}
		revision := head + 1
		now, err := s.serverTime(ctx, tx)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $module AS Utf8;DECLARE $revision AS Int64;DECLARE $updated AS Timestamp;\nUPSERT INTO \x60m8-access/catalog_heads\x60 (module_id,head_revision,updated_at) VALUES ($module,$revision,$updated);"),
			sql.Named("module", m.ModuleID()), sql.Named("revision", revision), sql.Named("updated", now))
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $module AS Utf8;DECLARE $version AS Utf8;DECLARE $revision AS Int64;DECLARE $payload AS Utf8;DECLARE $hash AS Utf8;DECLARE $created AS Timestamp;\nUPSERT INTO \x60m8-access/catalog_releases\x60 (module_id,version,revision,manifest_json,manifest_hash,created_at) VALUES ($module,$version,$revision,$payload,$hash,$created);"),
			sql.Named("module", m.ModuleID()), sql.Named("version", m.Version()), sql.Named("revision", revision), sql.Named("payload", string(payload)), sql.Named("hash", hash), sql.Named("created", now))
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $module AS Utf8;DECLARE $revision AS Int64;DECLARE $version AS Utf8;DECLARE $payload AS Utf8;\nUPSERT INTO \x60m8-access/catalog_read_index\x60 (module_id,revision,version,manifest_json) VALUES ($module,$revision,$version,$payload);"),
			sql.Named("module", m.ModuleID()), sql.Named("revision", revision), sql.Named("version", m.Version()), sql.Named("payload", string(payload)))
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, s.q("DECLARE $module AS Utf8;DECLARE $key AS Utf8;DECLARE $hash AS Utf8;DECLARE $version AS Utf8;DECLARE $revision AS Int64;\nUPSERT INTO \x60m8-access/catalog_receipts\x60 (module_id,request_key,request_hash,version,revision) VALUES ($module,$key,$hash,$version,$revision);"),
			sql.Named("module", m.ModuleID()), sql.Named("key", key), sql.Named("hash", hash), sql.Named("version", m.Version()), sql.Named("revision", revision))
		if err != nil {
			return err
		}
		eventID := fmt.Sprintf("catalog/%s/%d", m.ModuleID(), revision)
		_, err = tx.ExecContext(ctx, s.q("DECLARE $module AS Utf8;DECLARE $revision AS Int64;DECLARE $event AS Utf8;DECLARE $kind AS Utf8;DECLARE $payload AS Utf8;DECLARE $published AS Bool;DECLARE $created AS Timestamp;\nUPSERT INTO \x60m8-access/catalog_outbox\x60 (module_id,revision,event_id,event_kind,payload_json,published,created_at) VALUES ($module,$revision,$event,$kind,$payload,$published,$created);"),
			sql.Named("module", m.ModuleID()), sql.Named("revision", revision), sql.Named("event", eventID), sql.Named("kind", "ModuleManifestPublished"), sql.Named("payload", string(payload)), sql.Named("published", false), sql.Named("created", now))
		if err != nil {
			return err
		}
		published = catalog.PublishedRecord{Manifest: m, Revision: revision}
		return nil
	})
	if err != nil {
		return catalog.PublishedRecord{}, err
	}
	return published, nil
}

func (s *Store) Get(ctx context.Context, moduleID, version string) (catalog.PublishedRecord, error) {
	if s == nil || s.db == nil {
		return catalog.PublishedRecord{}, ErrInvalidStore
	}
	if !shared.ValidModuleName(moduleID) || !shared.ValidReleaseVersion(version) {
		return catalog.PublishedRecord{}, catalog.ErrCatalogArgument
	}
	if err := ctx.Err(); err != nil {
		return catalog.PublishedRecord{}, err
	}
	return s.readRelease(ctx, s.db, moduleID, version)
}
