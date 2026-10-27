package scope

import "testing"

func TestPlacementMatrix(t *testing.T) {
	cases := []struct {
		name    string
		child   Kind
		parent  Parent
		allowed bool
	}{
		{"tenant root", KindTenant, Parent{}, true},
		{"organization tenant", KindOrganization, Parent{KindTenant, "t1"}, true},
		{"group tenant", KindGroup, Parent{KindTenant, "t1"}, true},
		{"group cannot own organization", KindOrganization, Parent{KindGroup, "g1"}, false},
		{"organization without tenant", KindOrganization, Parent{}, false},
		{"workspace under organization", KindWorkspace, Parent{KindOrganization, "o1"}, true},
		{"workspace under project", KindWorkspace, Parent{KindProject, "p1"}, false},
		{"project under organization", KindProject, Parent{KindOrganization, "o1"}, true},
		{"project under workspace", KindProject, Parent{KindWorkspace, "w1"}, true},
		{"project under project", KindProject, Parent{KindProject, "p1"}, false},
		{"environment under organization", KindEnvironment, Parent{KindOrganization, "o1"}, true},
		{"environment under workspace", KindEnvironment, Parent{KindWorkspace, "w1"}, false},
		{"environment under tenant", KindEnvironment, Parent{KindTenant, "t1"}, false},
		{"workload under project", KindWorkload, Parent{KindProject, "p1"}, true},
		{"workload under workspace", KindWorkload, Parent{KindWorkspace, "w1"}, true},
		{"workload under environment", KindWorkload, Parent{KindEnvironment, "e1"}, false},
		{"descriptor under organization", KindDescriptor, Parent{KindOrganization, "o1"}, true},
		{"descriptor under project", KindDescriptor, Parent{KindProject, "p1"}, true},
		{"descriptor under workspace", KindDescriptor, Parent{KindWorkspace, "w1"}, true},
		{"descriptor under environment", KindDescriptor, Parent{KindEnvironment, "e1"}, false},
		{"unknown kind", Kind("unknown"), Parent{KindOrganization, "o1"}, false},
		{"parent missing id", KindProject, Parent{KindOrganization, ""}, false},
		{"root with id", KindTenant, Parent{"", "x1"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePlacement(tc.child, tc.parent)
			if (err == nil) != tc.allowed {
				t.Fatalf("ValidatePlacement(%s, %+v) = %v; allowed=%v", tc.child, tc.parent, err, tc.allowed)
			}
		})
	}
}
