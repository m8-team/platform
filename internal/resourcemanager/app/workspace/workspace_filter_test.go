package workspaceapp

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func TestParseWorkspaceFilter(t *testing.T) {
	parser, err := newFilterParser()
	if err != nil {
		t.Fatal(err)
	}
	t.Parallel()

	production := "Production"
	cyrillic := "л"
	tests := []struct {
		name    string
		raw     string
		want    ports.WorkspaceFilter
		wantErr bool
	}{
		{name: "empty"},
		{
			name: "unicode name",
			raw:  `name == "л"`,
			want: ports.WorkspaceFilter{NameEquals: &cyrillic},
		},
		{
			name: "equalities with both label access forms",
			raw:  `state == "DELETED" && name == "Production" && labels.environment == "prod" && labels["example.com/team"] == "platform"`,
			want: ports.WorkspaceFilter{
				States:     []workspace.State{workspace.StateDeleted},
				NameEquals: &production,
				LabelsEqual: map[string]string{
					"environment":      "prod",
					"example.com/team": "platform",
				},
			},
		},
		{
			name: "state membership is canonicalized",
			raw:  `state in ["SUSPENDED", "ACTIVE"]`,
			want: ports.WorkspaceFilter{
				States: []workspace.State{workspace.StateActive, workspace.StateSuspended},
			},
		},
		{
			name: "reversed equality",
			raw:  `"Production" == name`,
			want: ports.WorkspaceFilter{NameEquals: &production},
		},
		{name: "legacy syntax", raw: `state = "ACTIVE" AND name = "Production"`, wantErr: true},
		{name: "unknown field", raw: `description == "unsupported"`, wantErr: true},
		{name: "function", raw: `name.startsWith("Prod")`, wantErr: true},
		{name: "logical or", raw: `state == "ACTIVE" || state == "SUSPENDED"`, wantErr: true},
		{name: "non boolean", raw: `name`, wantErr: true},
		{name: "empty state list", raw: `state in []`, wantErr: true},
		{name: "membership on name", raw: `name in ["Production"]`, wantErr: true},
		{name: "membership on label", raw: `labels.team in ["platform"]`, wantErr: true},
		{name: "invalid state", raw: `state == "UNKNOWN"`, wantErr: true},
		{name: "duplicate state", raw: `state == "ACTIVE" && state == "SUSPENDED"`, wantErr: true},
		{name: "duplicate state in list", raw: `state in ["ACTIVE", "ACTIVE"]`, wantErr: true},
		{name: "duplicate label", raw: `labels.team == "one" && labels["team"] == "two"`, wantErr: true},
		{name: "empty label key", raw: `labels[""] == "value"`, wantErr: true},
		{name: "oversized", raw: strings.Repeat("x", maximumWorkspaceFilterRunes+1), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseWorkspaceFilter(parser, test.raw)
			if test.wantErr {
				if !errors.Is(err, ErrInvalidWorkspaceFilter) {
					t.Fatalf("parseWorkspaceFilter() error = %v, want %v", err, ErrInvalidWorkspaceFilter)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseWorkspaceFilter() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("parseWorkspaceFilter() = %#v, want %#v", got, test.want)
			}
		})
	}
}
