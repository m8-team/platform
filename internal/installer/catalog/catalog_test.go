package catalog

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileCatalogLoadsExampleRelease(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	catalog := NewFileCatalog(filepath.Join(repoRoot, "catalog", "releases"))

	release, err := catalog.Resolve(context.Background(), "1.0.0")
	if err != nil {
		t.Fatalf("resolve release: %v", err)
	}
	if err := catalog.Verify(context.Background(), release); err != nil {
		t.Fatalf("verify release: %v", err)
	}
}
