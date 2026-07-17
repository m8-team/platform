package output

import (
	"bytes"
	"strings"
	"testing"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWriteInstallationsStatusTable(t *testing.T) {
	var buffer bytes.Buffer

	err := WriteInstallationsStatus(&buffer, FormatTable, []installerv1alpha1.PlatformInstallation{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "m8-production", Namespace: "m8-system"},
			Spec: installerv1alpha1.PlatformInstallationSpec{
				PlatformVersion: "1.0.0",
				Profile:         installerv1alpha1.ProfileProduction,
			},
			Status: installerv1alpha1.PlatformInstallationStatus{
				Phase:           installerv1alpha1.PhaseReady,
				PlatformVersion: "1.0.0",
				Components: []installerv1alpha1.ComponentStatus{
					{Name: "argocd", Ready: true},
					{Name: "keycloak", Ready: false},
				},
				Endpoints: []installerv1alpha1.EndpointStatus{
					{Name: "console", Ready: true},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("write status: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{"m8-production", "m8-system", "Ready", "1/2 ready", "1/1 ready"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}
