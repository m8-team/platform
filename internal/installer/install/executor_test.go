package install

import (
	"testing"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
)

func TestCiliumHelmRelease(t *testing.T) {
	installation := installerv1alpha1.PlatformInstallation{
		Spec: installerv1alpha1.PlatformInstallationSpec{
			Network: installerv1alpha1.NetworkSpec{
				Cilium: installerv1alpha1.CiliumSpec{
					HubbleRelay: true,
					HubbleUI:    true,
				},
			},
		},
	}
	releaseCatalog := installerv1alpha1.PlatformRelease{
		Spec: installerv1alpha1.PlatformReleaseSpec{
			Components: map[string]installerv1alpha1.ComponentRelease{
				"cilium": {
					Version: "1.17.0",
					Chart: installerv1alpha1.ArtifactRef{
						Version: "1.17.0",
					},
				},
			},
		},
	}

	release, err := ciliumHelmRelease(installation, releaseCatalog)
	if err != nil {
		t.Fatalf("ciliumHelmRelease returned error: %v", err)
	}

	if release.Name != "cilium" {
		t.Fatalf("release.Name = %q, want cilium", release.Name)
	}
	if release.Namespace != "kube-system" {
		t.Fatalf("release.Namespace = %q, want kube-system", release.Namespace)
	}
	if release.Chart != "cilium" {
		t.Fatalf("release.Chart = %q, want cilium", release.Chart)
	}
	if release.Repository != "https://helm.cilium.io" {
		t.Fatalf("release.Repository = %q, want https://helm.cilium.io", release.Repository)
	}
	if release.Version != "1.17.0" {
		t.Fatalf("release.Version = %q, want 1.17.0", release.Version)
	}

	hubble := release.Values["hubble"].(map[string]any)
	relay := hubble["relay"].(map[string]any)
	ui := hubble["ui"].(map[string]any)
	if relay["enabled"] != true {
		t.Fatalf("hubble.relay.enabled = %v, want true", relay["enabled"])
	}
	if ui["enabled"] != true {
		t.Fatalf("hubble.ui.enabled = %v, want true", ui["enabled"])
	}
}

func TestCertManagerHelmRelease(t *testing.T) {
	releaseCatalog := installerv1alpha1.PlatformRelease{
		Spec: installerv1alpha1.PlatformReleaseSpec{
			Components: map[string]installerv1alpha1.ComponentRelease{
				"cert-manager": {
					Version: "v1.17.2",
				},
			},
		},
	}

	release, err := certManagerHelmRelease(releaseCatalog)
	if err != nil {
		t.Fatalf("certManagerHelmRelease returned error: %v", err)
	}

	if release.Name != "cert-manager" {
		t.Fatalf("release.Name = %q, want cert-manager", release.Name)
	}
	if release.Namespace != "m8-security" {
		t.Fatalf("release.Namespace = %q, want m8-security", release.Namespace)
	}
	if release.Chart != "cert-manager" {
		t.Fatalf("release.Chart = %q, want cert-manager", release.Chart)
	}
	if release.Repository != "https://charts.jetstack.io" {
		t.Fatalf("release.Repository = %q, want https://charts.jetstack.io", release.Repository)
	}
	if release.Version != "v1.17.2" {
		t.Fatalf("release.Version = %q, want v1.17.2", release.Version)
	}

	crds := release.Values["crds"].(map[string]any)
	if crds["enabled"] != true {
		t.Fatalf("crds.enabled = %v, want true", crds["enabled"])
	}
}
