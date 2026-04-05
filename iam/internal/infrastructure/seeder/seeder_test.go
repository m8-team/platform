package seeder

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDatasetsMergesFilesAndInfersUserGroups(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "01-groups.json"), []byte(`{
  "groups": [
    {
      "group_id": "grp-demo-ops",
      "tenant_id": "tenant-demo",
      "display_name": "Operations"
    }
  ],
  "group_members": [
    {
      "group_id": "grp-demo-ops",
      "subject_id": "user-demo-admin",
      "subject_type": "SUBJECT_TYPE_USER_ACCOUNT"
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "02-users.json"), []byte(`{
  "users": [
    {
      "user_id": "user-demo-admin",
      "tenant_id": "tenant-demo",
      "primary_email": "admin@example.com",
      "display_name": "Demo Admin",
      "group_ids": ["grp-demo-ops", "grp-demo-ops"]
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	data, files, err := LoadDatasets(dir)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := files, []string{"01-groups.json", "02-users.json"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("files mismatch: got %v want %v", got, want)
	}
	if got, want := len(data.Groups), 1; got != want {
		t.Fatalf("groups count mismatch: got %d want %d", got, want)
	}
	if got, want := len(data.Users), 1; got != want {
		t.Fatalf("users count mismatch: got %d want %d", got, want)
	}
	if got, want := data.Users[0].GroupIDs, []string{"grp-demo-ops"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("group ids mismatch: got %v want %v", got, want)
	}
}

func TestLoadDatasetsFailsOnUnknownGroupReference(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{
  "users": [
    {
      "user_id": "user-demo-admin",
      "tenant_id": "tenant-demo",
      "primary_email": "admin@example.com",
      "display_name": "Demo Admin",
      "group_ids": ["grp-missing"]
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := LoadDatasets(dir); err == nil {
		t.Fatal("expected error for unknown group reference")
	}
}
