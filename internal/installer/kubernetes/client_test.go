package kubernetes

import (
	"context"
	"regexp"
	"testing"
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestInstallerCRDSingularNamesAreDNS1035Labels(t *testing.T) {
	pattern := regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

	for _, crd := range installerCRDs() {
		singular, found, err := unstructured.NestedString(crd.Object, "spec", "names", "singular")
		if err != nil {
			t.Fatalf("read singular for %s: %v", crd.GetName(), err)
		}
		if !found {
			t.Fatalf("singular not found for %s", crd.GetName())
		}
		if !pattern.MatchString(singular) {
			t.Fatalf("singular for %s = %q, want DNS-1035 label", crd.GetName(), singular)
		}
	}
}

func TestToUnstructuredAcceptsTypedValue(t *testing.T) {
	release := installerv1alpha1.PlatformRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: installerv1alpha1.GroupName + "/" + installerv1alpha1.Version,
			Kind:       "PlatformRelease",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "1.0.0"},
		Spec: installerv1alpha1.PlatformReleaseSpec{
			Kubernetes: installerv1alpha1.VersionRange{MinVersion: "1.30.0"},
			Components: map[string]installerv1alpha1.ComponentRelease{
				"platform": {
					Version: "1.0.0",
					Chart: installerv1alpha1.ArtifactRef{
						Repository: "oci://registry.example.com/charts/platform",
						Version:    "1.0.0",
						Digest:     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				},
			},
		},
	}

	object, err := toUnstructured(release)
	if err != nil {
		t.Fatalf("to unstructured: %v", err)
	}
	if object.GetKind() != "PlatformRelease" {
		t.Fatalf("kind = %q, want PlatformRelease", object.GetKind())
	}
	if object.GetName() != "1.0.0" {
		t.Fatalf("name = %q, want 1.0.0", object.GetName())
	}
}

func TestWaitForArgoApplications(t *testing.T) {
	application := &unstructured.Unstructured{}
	application.SetAPIVersion("argoproj.io/v1alpha1")
	application.SetKind("Application")
	application.SetNamespace("argocd")
	application.SetName("m8-data-operators")

	client := &Client{
		dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), application),
	}

	err := client.WaitForArgoApplications(context.Background(), "argocd", []string{"m8-data-operators"}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForArgoApplications returned error: %v", err)
	}
}

func TestDeleteArgoApplicationsIgnoresMissingApplications(t *testing.T) {
	client := &Client{
		dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
	}

	deleted, err := client.DeleteArgoApplications(context.Background(), "argocd", []string{"m8-data-operators"})
	if err != nil {
		t.Fatalf("DeleteArgoApplications returned error: %v", err)
	}
	if len(deleted) != 1 {
		t.Fatalf("len(deleted) = %d, want 1", len(deleted))
	}
}
