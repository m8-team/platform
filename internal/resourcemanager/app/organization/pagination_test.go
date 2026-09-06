package organizationapp

import "testing"

func TestPageTokenScopeIncludesFilterIDs(t *testing.T) {
	parser, err := newFilterParser()
	if err != nil {
		t.Fatal(err)
	}
	_, first, err := normalizeListQuery(parser, ListOrganizations{Filter: `id == "00000000-0000-4000-8000-000000000001"`}, "caller")
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := normalizeListQuery(parser, ListOrganizations{Filter: `id == "00000000-0000-4000-8000-000000000002"`}, "caller")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("different IDs have the same page-token scope")
	}
}
