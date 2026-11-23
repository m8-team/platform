package access

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// This test protects the bounded-context import direction. Cross-module
// adapters belong to internal/integration, not internal/access.
func TestAccessHasNoResourceManagerDependency(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve Access source root")
	}
	root := filepath.Dir(sourceFile)
	forbidden := []string{
		"github.com/m8-team/platform/internal/resourcemanager",
		"github.com/m8-team/platform/internal/integration",
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			name, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			for _, prefix := range forbidden {
				if name == prefix || strings.HasPrefix(name, prefix+"/") {
					t.Errorf("forbidden direct dependency in %s: %s", path, name)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
