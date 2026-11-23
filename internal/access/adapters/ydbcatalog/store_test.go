package ydbcatalog

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

func fixture(t *testing.T, version string) module.Manifest {
	t.Helper()
	publisher, err := shared.NewSubject("service_account", "finance")
	if err != nil {
		t.Fatal(err)
	}
	kind, err := resource.NewTypeDefinition("finance", "finance.invoice", []authorization.ScopeKind{
		authorization.ScopeInstance, authorization.ScopeOrganization})
	if err != nil {
		t.Fatal(err)
	}
	action, err := permission.NewPermission("finance.invoice.view")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := permission.NewDefinition(action, "finance.invoice", "use")
	if err != nil {
		t.Fatal(err)
	}
	template, err := role.NewTemplate("finance", role.TemplateRef{Name: "finance.accountant", Version: version},
		[]permission.Permission{action}, []authorization.ScopeKind{authorization.ScopeOrganization})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := module.NewManifest(module.ManifestInput{
		ModuleID: "finance", Version: version, Publisher: publisher,
		Resources: []resource.TypeDefinition{kind}, Permissions: []permission.Definition{definition},
		Roles: []role.Template{template},
	})
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func TestManifestRoundTripAndCanonicalDigest(t *testing.T) {
	manifest := fixture(t, "1.0.0")
	data, err := marshalManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := unmarshalManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	again, err := marshalManifest(restored)
	if err != nil || string(data) != string(again) {
		t.Fatalf("noncanonical round trip: %v", err)
	}
	if len(restored.Resources()) != 1 || len(restored.Roles()) != 1 ||
		restored.Publisher() != manifest.Publisher() {
		t.Fatal("lost catalog definitions")
	}
	if requestHash(data, 0) == requestHash(data, 1) {
		t.Fatal("CAS expected revision absent from digest")
	}
	if _, err = unmarshalManifest([]byte("{\"module_id\":\"finance\",\"version\":\"1.0.0\",\"publisher_type\":\"user\",\"publisher_id\":\"alice\"}")); !errors.Is(err, ErrInvalidManifest) {
		t.Fatalf("invalid publisher accepted: %v", err)
	}
	if _, err = unmarshalManifest([]byte("invalid-json")); !errors.Is(err, ErrInvalidManifest) {
		t.Fatal(err)
	}
	if _, err = unmarshalManifest([]byte(strings.Repeat("a", maxManifestSize+1))); !errors.Is(err, ErrInvalidManifest) {
		t.Fatal(err)
	}
}

func TestSignedCursorBindsIdentityNamespaceSnapshotAndExpiry(t *testing.T) {
	p, _ := shared.NewSubject("service_account", "finance")
	other, _ := shared.NewSubject("service_account", "other")
	when := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	store := &Store{tokenKey: []byte(strings.Repeat("k", 32)), cursorTTL: time.Hour, now: func() time.Time { return when }}
	token, err := store.encodeCursor(p, "finance", 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.decodeCursor(token, p, "finance")
	if err != nil || state.After != 1 || state.Snapshot != 3 {
		t.Fatalf("cursor decode: %v", err)
	}
	for _, variant := range []struct {
		token    string
		subject  shared.Subject
		moduleID string
	}{
		{token, other, "finance"},
		{token, p, "identity"},
		{token + "x", p, "finance"},
		{"not-an-authenticated-cursor", p, "finance"},
	} {
		if _, err := store.decodeCursor(variant.token, variant.subject, variant.moduleID); !errors.Is(err, ErrInvalidCursor) {
			t.Fatalf("accepted invalid cursor: %v", err)
		}
	}
	store.now = func() time.Time { return when.Add(time.Hour) }
	if _, err := store.decodeCursor(token, p, "finance"); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("accepted expired cursor: %v", err)
	}
}
func TestInvalidStoreConstruction(t *testing.T) {
	if _, err := New(nil, "/local/fixture", []byte(strings.Repeat("k", 32)), time.Hour); !errors.Is(err, ErrInvalidStore) {
		t.Fatal(err)
	}
	if _, err := New(nil, "/local/../escape", []byte(strings.Repeat("k", 32)), time.Hour); !errors.Is(err, ErrInvalidStore) {
		t.Fatal(err)
	}
}
