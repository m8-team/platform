package uninstall

import (
	"testing"

	"github.com/m8platform/platform/internal/installer/planner"
)

func TestArgoApplicationNames(t *testing.T) {
	plan := planner.InstallationPlan{
		Steps: []planner.InstallationStep{
			{
				ChangeSet: planner.ChangeSet{
					ArgoApplications: []planner.ArgoApplicationChange{
						{Name: "data-operators"},
						{Name: "observability"},
						{Name: "m8-shared-services"},
					},
				},
			},
			{
				ChangeSet: planner.ChangeSet{
					ArgoApplications: []planner.ArgoApplicationChange{
						{Name: "data-operators"},
						{Name: ""},
					},
				},
			},
		},
	}

	got := argoApplicationNames(plan)
	want := []string{"m8-data-operators", "m8-observability", "m8-m8-shared-services"}
	if len(got) != len(want) {
		t.Fatalf("len(argoApplicationNames) = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argoApplicationNames[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
