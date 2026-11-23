//go:build integration

package ydbcatalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
)

func TestLiveYDBTenantModulePinAndDisable(t *testing.T) {
	store := testStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	data, err := os.ReadFile("schema_tenant_modules_v1.sql")
	if err != nil {
		t.Fatal(err)
	}
	var clean []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "--") {
			clean = append(clean, line)
		}
	}
	for _, query := range strings.Split(strings.Join(clean, "\n"), ";") {
		if query = strings.TrimSpace(query); query != "" {
			if _, err := store.db.ExecContext(ctx, fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s;", store.prefix, query)); err != nil {
				t.Fatal("installation migration", err)
			}
		}
	}
	if _, err := store.Publish(ctx, fixture(t, "1.0.0"), 0, "publish"); err != nil {
		t.Fatal(err)
	}
	tenant, _ := domain.NewTenantID("11111111-1111-4111-8111-111111111111")
	create := store.q("DECLARE $tenant AS Utf8;DECLARE $module AS Utf8;DECLARE $version AS Utf8;DECLARE $enabled AS Bool;DECLARE $revision AS Uint64;\nUPSERT INTO `m8-access/catalog_tenant_modules` (tenant_id,module_id,release_version,enabled,revision) VALUES ($tenant,$module,$version,$enabled,$revision);")
	setEnabled := func(enabled bool) {
		t.Helper()
		_, err := store.db.ExecContext(ctx, create, sql.Named("tenant", string(tenant)),
			sql.Named("module", "finance"), sql.Named("version", "1.0.0"),
			sql.Named("enabled", enabled), sql.Named("revision", uint64(1)))
		if err != nil {
			t.Fatal(err)
		}
	}
	reader, err := NewUniversalPermissionReader(store, store)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := reader.HasCatalogPermission(ctx, tenant, "finance.invoice", "finance.invoice.view"); ok || !errors.Is(err, ErrTenantModuleUnavailable) {
		t.Fatalf("missing installation granted access: %v", err)
	}
	setEnabled(true)
	if ok, err := reader.HasCatalogPermission(ctx, tenant, "finance.invoice", "finance.invoice.view"); err != nil || !ok {
		t.Fatalf("installed pin not recognized: %v", err)
	}
	setEnabled(false)
	if ok, err := reader.HasCatalogPermission(ctx, tenant, "finance.invoice", "finance.invoice.view"); ok || !errors.Is(err, ErrTenantModuleUnavailable) {
		t.Fatalf("disabled module granted access: %v", err)
	}
}
