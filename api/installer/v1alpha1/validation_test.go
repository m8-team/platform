package v1alpha1

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPlatformInstallationDefaultedMinimalProductionIsValid(t *testing.T) {
	installation := PlatformInstallation{
		ObjectMeta: metav1.ObjectMeta{Name: "m8-production"},
		Spec: PlatformInstallationSpec{
			PlatformVersion: "1.0.0",
			Profile:         ProfileProduction,
		},
	}

	defaulted := installation.Defaulted()
	if err := defaulted.Validate(); err != nil {
		t.Fatalf("expected valid installation, got %v", err)
	}
	if defaulted.Spec.Cluster.MinimumNodes != 3 {
		t.Fatalf("expected production minimum nodes to default to 3, got %d", defaulted.Spec.Cluster.MinimumNodes)
	}
	if defaulted.Spec.Secrets.Provider == SecretsProviderKubernetes {
		t.Fatal("production must not default to Kubernetes Secrets")
	}
	if !defaulted.Spec.Security.Cosign.Enforce {
		t.Fatal("production must default to cosign enforcement")
	}
}

func TestPlatformInstallationRejectsInvalidModuleCombination(t *testing.T) {
	installation := PlatformInstallation{
		ObjectMeta: metav1.ObjectMeta{Name: "m8"},
		Spec: PlatformInstallationSpec{
			PlatformVersion: "1.0.0",
			Profile:         ProfileDevelopment,
			Modules: ModulesSpec{
				Authentication: true,
				Identity:       false,
			},
			Identity: IdentitySpec{Keycloak: KeycloakSpec{Mode: ModeOperator}},
			Gateway:  GatewaySpec{EnvoyGateway: EnvoyGatewaySpec{Enabled: true}},
		},
	}

	err := installation.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "authentication requires identity") {
		t.Fatalf("expected authentication dependency error, got %v", err)
	}
}

func TestPlatformReleaseRejectsFloatingArtifactVersion(t *testing.T) {
	release := PlatformRelease{
		ObjectMeta: metav1.ObjectMeta{Name: "1.0.0"},
		Spec: PlatformReleaseSpec{
			Kubernetes: VersionRange{MinVersion: "1.30.0"},
			Components: map[string]ComponentRelease{
				"platform": {
					Version: "1.0.0",
					Chart: ArtifactRef{
						Repository: "oci://registry.example.com/charts/m8-platform",
						Version:    "latest",
						Digest:     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				},
			},
		},
	}

	err := release.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "floating versions are not allowed") {
		t.Fatalf("expected floating version error, got %v", err)
	}
}
