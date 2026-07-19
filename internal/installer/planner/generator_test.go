package planner

import (
	"context"
	"testing"
	"time"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratePlanOrdersSyncWaves(t *testing.T) {
	installation := installerv1alpha1.PlatformInstallation{
		ObjectMeta: metav1.ObjectMeta{Name: "m8-production", Namespace: "m8-system"},
		Spec: installerv1alpha1.PlatformInstallationSpec{
			PlatformVersion: "1.0.0",
			Profile:         installerv1alpha1.ProfileProduction,
		},
	}.Defaulted()

	release := testRelease()
	plan, err := Generate(context.Background(), GenerateInput{
		Installation:         installation,
		Release:              release,
		ConfigDigest:         "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ReleaseCatalogDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CreatedAt:            time.Unix(0, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if len(plan.Steps) == 0 {
		t.Fatal("expected steps")
	}
	if plan.Steps[0].ID != "crds" {
		t.Fatalf("first step = %s, want crds", plan.Steps[0].ID)
	}
	if plan.Steps[len(plan.Steps)-1].ID != "smoke-tests" {
		t.Fatalf("last step = %s, want smoke-tests", plan.Steps[len(plan.Steps)-1].ID)
	}
	for i := 1; i < len(plan.Steps); i++ {
		if plan.Steps[i].Wave < plan.Steps[i-1].Wave {
			t.Fatalf("steps not ordered by wave at %d: %d < %d", i, plan.Steps[i].Wave, plan.Steps[i-1].Wave)
		}
	}
}

func testRelease() installerv1alpha1.PlatformRelease {
	componentNames := []string{
		"cert-manager",
		"trust-manager",
		"spire",
		"external-secrets-operator",
		"kyverno",
		"trivy-operator",
		"cilium",
		"cloudnative-pg",
		"ydb-operator",
		"strimzi",
		"redis-operator",
		"platform",
	}
	components := make(map[string]installerv1alpha1.ComponentRelease, len(componentNames))
	for _, name := range componentNames {
		components[name] = installerv1alpha1.ComponentRelease{
			Version: "1.0.0",
			Chart: installerv1alpha1.ArtifactRef{
				Repository: "oci://registry.example.com/charts/" + name,
				Version:    "1.0.0",
				Digest:     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		}
	}
	return installerv1alpha1.PlatformRelease{
		ObjectMeta: metav1.ObjectMeta{Name: "1.0.0"},
		Spec: installerv1alpha1.PlatformReleaseSpec{
			Kubernetes: installerv1alpha1.VersionRange{MinVersion: "1.30.0"},
			Components: components,
			Signature:  installerv1alpha1.SignatureRef{Value: "signed"},
		},
	}
}
