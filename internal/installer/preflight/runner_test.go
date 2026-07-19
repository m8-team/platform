package preflight

import (
	"context"
	"testing"

	installerv1alpha1 "github.com/m8-team/platform/api/installer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeCluster struct {
	version        string
	nodes          NodeSummary
	storageClasses []string
	resources      map[string]bool
}

func (f fakeCluster) ServerVersion(context.Context) (string, error) {
	return f.version, nil
}

func (f fakeCluster) NodeSummary(context.Context) (NodeSummary, error) {
	return f.nodes, nil
}

func (f fakeCluster) HasAPIResource(_ context.Context, groupVersion string, kind string) (bool, error) {
	return f.resources[groupVersion+"/"+kind], nil
}

func (f fakeCluster) StorageClasses(context.Context) ([]string, error) {
	return f.storageClasses, nil
}

func TestDefaultRunnerReportsProductionZoneWarning(t *testing.T) {
	installation := installerv1alpha1.PlatformInstallation{
		ObjectMeta: metav1.ObjectMeta{Name: "m8-production"},
		Spec: installerv1alpha1.PlatformInstallationSpec{
			PlatformVersion: "1.0.0",
			Profile:         installerv1alpha1.ProfileProduction,
		},
	}.Defaulted()

	report := DefaultRunner(fakeCluster{
		version: "v1.31.1",
		nodes: NodeSummary{
			Total:         3,
			Ready:         3,
			Architectures: map[string]int{"amd64": 3},
			Zones:         map[string]int{"eu-1a": 3},
		},
		storageClasses: []string{"standard"},
		resources:      map[string]bool{"gateway.networking.k8s.io/v1/GatewayClass": true},
	}).Run(context.Background(), installation)

	if report.Summary.Warnings != 1 {
		t.Fatalf("warnings = %d, want 1; report=%+v", report.Summary.Warnings, report)
	}
	if report.Summary.Failed != 0 {
		t.Fatalf("failed = %d, want 0", report.Summary.Failed)
	}
}
