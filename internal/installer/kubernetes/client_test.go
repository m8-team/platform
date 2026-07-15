package kubernetes

import (
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
