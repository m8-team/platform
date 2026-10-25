package scope

import "testing"

func TestPlacementMatrix(t *testing.T) {
	cases := []struct {
		name   string
		child  Kind
		parent Parent
		valid  bool
	}{
		{"group root", KindGroup, Parent{}, true},
		{"organization root", KindOrganization, Parent{}, true},
		{"organization under group forbidden", KindOrganization, Parent{KindGroup, "g1"}, false},
		{"workspace under organization", KindWorkspace, Parent{KindOrganization, "o1"}, true},
		{"workspace under project forbidden", KindWorkspace, Parent{KindProject, "p1"}, false},
		{"project directly under organization", KindProject, Parent{KindOrganization, "o1"}, true},
		{"project under workspace", KindProject, Parent{KindWorkspace, "w1"}, true},
		{"service under project", KindService, Parent{KindProject, "p1"}, true},
		{"legacy service under workspace", KindService, Parent{KindWorkspace, "w1"}, true},
		{"service under organization forbidden", KindService, Parent{KindOrganization, "o1"}, false},
		{"resource under organization", KindResource, Parent{KindOrganization, "o1"}, true},
		{"resource under workspace", KindResource, Parent{KindWorkspace, "w1"}, true},
		{"resource under project", KindResource, Parent{KindProject, "p1"}, true},
		{"resource under service forbidden", KindResource, Parent{KindService, "s1"}, false},
		{"missing parent ID", KindProject, Parent{KindOrganization, ""}, false},
		{"orphan parent ID", KindOrganization, Parent{"", "o1"}, false},
		{"unknown child kind", Kind("unknown"), Parent{KindOrganization, "o1"}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlacement(tt.child, tt.parent)
			if (err == nil) != tt.valid {
				t.Fatalf("ValidatePlacement(%s, %+v) = %v, want valid=%v", tt.child, tt.parent, err, tt.valid)
			}
		})
	}
}
