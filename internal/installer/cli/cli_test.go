package cli

import (
	"bytes"
	"context"
	"os"
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

func TestInstallCommandRequiresInputWhenNoDefaultExists(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New(&stdout, &stderr)

	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousWorkingDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	code := app.Run(context.Background(), []string{"install"})

	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d", code, ExitUsage)
	}
	if strings.Contains(stderr.String(), "not implemented") {
		t.Fatalf("install still returned stub output: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "install requires --plan") {
		t.Fatalf("expected input error, got: %s", stderr.String())
	}
}

func TestInstallCommandReadsPlanPreview(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New(&stdout, &stderr)
	planPath := filepath.Join(t.TempDir(), "plan.yaml")
	err := os.WriteFile(planPath, []byte(`apiVersion: installer.m8.io/v1alpha1
kind: InstallationPlan
metadata:
  name: m8-test-plan
installation:
  name: m8-test
release:
  name: 1.0.0
  version: 1.0.0
configDigest: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
releaseCatalogDigest: sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
profile: production
steps:
  - id: crds
    wave: -100
    title: Install required CRDs
    phase: bootstrap
    changeSet: {}
    rollback:
      supported: false
`), 0o600)
	if err != nil {
		t.Fatalf("write plan: %v", err)
	}

	code := app.Run(context.Background(), []string{"install", "--plan", planPath})

	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "not implemented") {
		t.Fatalf("install still returned stub output: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Plan-only preview") {
		t.Fatalf("expected preview message, got: %s", stdout.String())
	}
}

func TestInstallCommandAcceptsGitOpsHandoffFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New(&stdout, &stderr)
	planPath := filepath.Join(t.TempDir(), "plan.yaml")
	err := os.WriteFile(planPath, []byte(`apiVersion: installer.m8.io/v1alpha1
kind: InstallationPlan
metadata:
  name: m8-test-plan
installation:
  name: m8-test
release:
  name: 1.0.0
  version: 1.0.0
configDigest: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
releaseCatalogDigest: sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
profile: production
steps:
  - id: crds
    wave: -100
    title: Install required CRDs
    phase: bootstrap
    changeSet: {}
    rollback:
      supported: false
`), 0o600)
	if err != nil {
		t.Fatalf("write plan: %v", err)
	}

	code := app.Run(context.Background(), []string{"install", "--plan", planPath, "--wait-gitops-handoff", "--gitops-handoff-timeout", "1s"})

	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "not implemented") {
		t.Fatalf("install returned stub output: %s", stderr.String())
	}
}

func TestUninstallCommandIsImplemented(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.Run(context.Background(), []string{"uninstall", "--dry-run", "--name", "m8-test"})

	if code != ExitOK {
		t.Fatalf("exit code = %d, want %d; stderr=%s", code, ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "not implemented") {
		t.Fatalf("uninstall still returned stub output: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Dry run only") {
		t.Fatalf("expected dry-run message, got: %s", stdout.String())
	}
}

func TestUninstallDeleteNetworkRequiresConfirmation(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New(&stdout, &stderr)

	code := app.Run(context.Background(), []string{"uninstall", "--dry-run", "--name", "m8-test", "--delete-network"})

	if code != ExitUsage {
		t.Fatalf("exit code = %d, want %d", code, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "requires --confirmation m8-test") {
		t.Fatalf("expected confirmation error, got: %s", stderr.String())
	}
}
