package migrator

import (
	"slices"
	"testing"
)

func TestParseExpectedTables(t *testing.T) {
	sql := `
-- bootstrap
CREATE TABLE IF NOT EXISTS users (
  id Utf8 NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE groups (
  id Utf8 NOT NULL,
  PRIMARY KEY (id)
);
`

	got := ParseExpectedTables(sql)
	want := []string{"users", "groups"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected tables: got=%v want=%v", got, want)
	}
}

func TestIsCreateOnlyMigration(t *testing.T) {
	createOnly := `
CREATE TABLE IF NOT EXISTS users (
  id Utf8 NOT NULL,
  PRIMARY KEY (id)
);`
	if !IsCreateOnlyMigration(createOnly) {
		t.Fatal("expected create-only migration to be detected")
	}

	alterMigration := `
ALTER TABLE users ADD COLUMN display_name Utf8;`
	if IsCreateOnlyMigration(alterMigration) {
		t.Fatal("expected alter migration not to be create-only")
	}

	mixedMigration := `
CREATE TABLE IF NOT EXISTS users (
  id Utf8 NOT NULL,
  PRIMARY KEY (id)
);
ALTER TABLE users ADD COLUMN display_name Utf8;`
	if IsCreateOnlyMigration(mixedMigration) {
		t.Fatal("expected mixed migration not to be create-only")
	}
}
