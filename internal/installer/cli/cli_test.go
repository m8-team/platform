package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusCommandIsImplemented(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.Run(context.Background(), []string{
		"status",
		"--kubeconfig",
		filepath.Join(t.TempDir(), "missing-kubeconfig"),
	})

	if code != ExitError {
		t.Fatalf("exit code = %d, want %d", code, ExitError)
	}
	if strings.Contains(stderr.String(), "not implemented") {
		t.Fatalf("status still returned stub output: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "create Kubernetes client") {
		t.Fatalf("expected Kubernetes client error, got: %s", stderr.String())
	}
}
