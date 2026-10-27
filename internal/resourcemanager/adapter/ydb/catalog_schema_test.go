package ydb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// prepareResourceCatalogSchema applies the EXACT 003_unified_resources.sql DDL
// in disposable integration tests, not a simplified facsimile.
func prepareResourceCatalogSchema(ctx context.Context, db *sql.DB, prefix string) error {
	file := filepath.Join("..", "..", "..", "..", "docs", "executable-baseline", "data", "ydb", "m8-resource-manager", "003_unified_resources.sql")
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
