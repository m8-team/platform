package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/sugar"
)

const schemaMigrationsTable = "schema_migrations"

type Runner struct {
	driver        *ydb.Driver
	migrationsDir string
	databaseName  string
}

type Migration struct {
	Name           string
	Path           string
	SQL            string
	Checksum       string
	ExpectedTables []string
	CreateOnly     bool
}

type ItemReport struct {
	Name   string
	Status string
}

type Report struct {
	Items      []ItemReport
	Applied    int
	Backfilled int
	Skipped    int
}

func New(cfg config.YDBConfig, migrationsDir string) (*Runner, error) {
	if cfg.DSN == "" {
		return nil, errors.New("IAM_YDB_DSN is required for migrator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driver, err := ydb.Open(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &Runner{
		driver:        driver,
		migrationsDir: migrationsDir,
		databaseName:  driver.Name(),
	}, nil
}

func (r *Runner) Close(ctx context.Context) error {
	if r == nil || r.driver == nil {
		return nil
	}
	return r.driver.Close(ctx)
}

func (r *Runner) Run(ctx context.Context) (*Report, error) {
	if err := r.ensureMetadataTable(ctx); err != nil {
		return nil, err
	}

	migrations, err := LoadMigrations(r.migrationsDir)
	if err != nil {
		return nil, err
	}

	report := &Report{Items: make([]ItemReport, 0, len(migrations))}
	for _, migration := range migrations {
		status, err := r.applyMigration(ctx, migration)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", migration.Name, err)
		}
		report.Items = append(report.Items, ItemReport{Name: migration.Name, Status: status})
		switch status {
		case "applied":
			report.Applied++
		case "backfilled":
			report.Backfilled++
		default:
			report.Skipped++
		}
	}
	return report, nil
}

func LoadMigrations(migrationsDir string) ([]Migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		fullPath := filepath.Join(migrationsDir, entry.Name())
		payload, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		sql := strings.TrimSpace(string(payload))
		migrations = append(migrations, Migration{
			Name:           entry.Name(),
			Path:           fullPath,
			SQL:            sql,
			Checksum:       checksum(payload),
			ExpectedTables: ParseExpectedTables(sql),
			CreateOnly:     IsCreateOnlyMigration(sql),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})
	return migrations, nil
}

func (r *Runner) applyMigration(ctx context.Context, migration Migration) (string, error) {
	appliedChecksum, found, err := r.lookupAppliedMigration(ctx, migration.Name)
	if err != nil {
		return "", err
	}
	if found {
		if appliedChecksum != migration.Checksum {
			return "", fmt.Errorf("migration checksum mismatch: stored=%s file=%s", appliedChecksum, migration.Checksum)
		}
		return "skipped", nil
	}

	allTablesExist, err := r.allExpectedTablesExist(ctx, migration.ExpectedTables)
	if err != nil {
		return "", err
	}
	if migration.CreateOnly && allTablesExist {
		if err := r.recordAppliedMigration(ctx, migration); err != nil {
			return "", err
		}
		return "backfilled", nil
	}

	if err := r.driver.Query().Exec(ctx, prefixedSQL(r.databaseName, migration.SQL)); err != nil {
		return "", err
	}
	if err := r.recordAppliedMigration(ctx, migration); err != nil {
		return "", err
	}
	return "applied", nil
}

func (r *Runner) allExpectedTablesExist(ctx context.Context, tableNames []string) (bool, error) {
	if len(tableNames) == 0 {
		return false, nil
	}
	for _, tableName := range tableNames {
		exists, err := sugar.IsTableExists(ctx, r.driver.Scheme(), r.absoluteTablePath(tableName))
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func (r *Runner) absoluteTablePath(tableName string) string {
	return path.Join(r.databaseName, tableName)
}

func (r *Runner) ensureMetadataTable(ctx context.Context) error {
	sql := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  name Utf8 NOT NULL,
  checksum Utf8 NOT NULL,
  expected_tables Utf8,
  create_only Bool,
  applied_at Timestamp,
  PRIMARY KEY(name)
)`, schemaMigrationsTable)
	return r.driver.Query().Exec(ctx, prefixedSQL(r.databaseName, sql))
}

func (r *Runner) lookupAppliedMigration(ctx context.Context, migrationName string) (string, bool, error) {
	row, err := r.driver.Query().QueryRow(ctx, prefixedSQL(r.databaseName, `
DECLARE $name AS Utf8;

SELECT checksum
FROM schema_migrations
WHERE name = $name;
`),
		query.WithParameters(
			ydb.ParamsBuilder().
				Param("$name").Text(migrationName).
				Build(),
		),
	)
	if err != nil {
		if errors.Is(err, query.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	var checksum string
	if err := row.Scan(&checksum); err != nil {
		return "", false, err
	}
	return checksum, true, nil
}

func (r *Runner) recordAppliedMigration(ctx context.Context, migration Migration) error {
	expectedTables := strings.Join(migration.ExpectedTables, ",")
	return r.driver.Query().Exec(ctx, prefixedSQL(r.databaseName, `
DECLARE $name AS Utf8;
DECLARE $checksum AS Utf8;
DECLARE $expected_tables AS Utf8;
DECLARE $create_only AS Bool;
DECLARE $applied_at AS Timestamp;

UPSERT INTO schema_migrations (name, checksum, expected_tables, create_only, applied_at)
VALUES ($name, $checksum, $expected_tables, $create_only, $applied_at);
`),
		query.WithParameters(
			ydb.ParamsBuilder().
				Param("$name").Text(migration.Name).
				Param("$checksum").Text(migration.Checksum).
				Param("$expected_tables").Text(expectedTables).
				Param("$create_only").Bool(migration.CreateOnly).
				Param("$applied_at").Timestamp(time.Now().UTC()).
				Build(),
		),
	)
}

func prefixedSQL(databaseName string, sql string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n\n%s", databaseName, sql)
}

func checksum(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
