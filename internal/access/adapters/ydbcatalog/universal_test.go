package ydbcatalog

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/app/catalog"
	"github.com/m8-team/platform/internal/access/domain"
)

type installedStub struct {
	pin TenantModuleInstallation
	err error
}

func (s installedStub) ResolveInstalledModule(context.Context, domain.TenantID, string) (TenantModuleInstallation, error) {
	return s.pin, s.err
}

type manifestStub struct {
	record catalog.PublishedRecord
	err    error
}

func (s manifestStub) Get(context.Context, string, string) (catalog.PublishedRecord, error) {
	return s.record, s.err
}
func TestUniversalCatalogRequiresInstalledPinnedExactPermission(t *testing.T) {
	tenant, _ := domain.NewTenantID("11111111-1111-4111-8111-111111111111")
	m := fixture(t, "1.0.0")
	installed := installedStub{pin: TenantModuleInstallation{ModuleID: "finance", Version: "1.0.0", Revision: 4}}
	release := manifestStub{record: catalog.PublishedRecord{Manifest: m, Revision: 1}}
	check := func(inst installedStub, reader manifestStub, typ, action string) (bool, error) {
		t.Helper()
		catalogReader, err := NewUniversalPermissionReader(inst, reader)
		if err != nil {
			t.Fatal(err)
		}
		return catalogReader.HasCatalogPermission(context.Background(), tenant, typ, action)
	}
	ok, err := check(installed, release, "finance.invoice", "finance.invoice.view")
	if err != nil || !ok {
		t.Fatalf("valid installed action denied: %v", err)
	}
	for _, tc := range []struct{ typ, action string }{
		{"finance.invoice", "finance.invoice.edit"},
		{"finance.customer", "finance.invoice.view"},
		{"identity.user", "finance.invoice.view"},
		{"finance.invoice", "identity.user.read"},
		{"finance/invoice", "finance.invoice.view"},
	} {
		ok, err = check(installed, release, tc.typ, tc.action)
		if ok || err != nil && tc.typ != "finance/invoice" {
			t.Fatalf("incorrect permission for %s / %s: %v %v", tc.typ, tc.action, ok, err)
		}
	}
	installed.err = ErrTenantModuleUnavailable
	if ok, err = check(installed, release, "finance.invoice", "finance.invoice.view"); ok || !errors.Is(err, ErrTenantModuleUnavailable) {
		t.Fatalf("uninstalled module could grant rights: %v", err)
	}
	installed.err = nil
	installed.pin.Version = "1.0.1"
	if ok, err = check(installed, release, "finance.invoice", "finance.invoice.view"); ok || !errors.Is(err, catalog.ErrCatalogStoreInvalid) {
		t.Fatalf("mismatched pinned release accepted: %v", err)
	}
}
func TestUniversalCatalogRejectsMissingDependencies(t *testing.T) {
	if _, err := NewUniversalPermissionReader(nil, nil); !errors.Is(err, ErrInvalidStore) {
		t.Fatal(err)
	}
}
