package ydb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Apply the exact resource and corporate relationship DDL migrations in
// disposable integration tests; never use a simplified schema facsimile.
func prepareResourceCatalogSchema(ctx context.Context, db *sql.DB, prefix string) error {
	for _, migration := range []string{"003_unified_resources.sql", "004_organization_relationships.sql"} {
		if err := prepareResourceMigration(ctx, db, prefix, migration); err != nil {
			return err
		}
	}
	return nil
}

func prepareResourceMigration(ctx context.Context, db *sql.DB, prefix, migration string) error {
	file := filepath.Join("..", "..", "..", "..", "docs", "executable-baseline", "data", "ydb", "m8-resource-manager", migration)
	src, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	for _, piece := range strings.Split(string(src), ";") {
		var lines []string
		for _, line := range strings.Split(piece, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}
			lines = append(lines, line)
		}
		statement := strings.Join(lines, "\n")
		if statement == "" || strings.HasPrefix(statement, "PRAGMA TablePathPrefix") {
			continue
		}
		if _, err = db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s;", prefix, statement)); err != nil {
			return fmt.Errorf("execute migration 003: %w", err)
		}
	}
	return nil
}
