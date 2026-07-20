package usecase

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func TestParseOrganizationFilter(t *testing.T) {
	t.Parallel()

	production := "Production"
	cyrillic := "л"
	firstID := organization.MustParseID("018f3f16-9950-7a48-9d12-9fb6d8f4c8f2")
	secondID := organization.MustParseID("018f3f16-9950-7a48-9d12-9fb6d8f4c8f3")
	tests := []struct {
		name    string
		raw     string
		want    ports.OrganizationFilter
		wantErr bool
	}{
		{name: "empty"},
		{
			name: "id membership",
			raw:  `id in ["018f3f16-9950-7a48-9d12-9fb6d8f4c8f3", "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"]`,
			want: ports.OrganizationFilter{IDs: []organization.ID{firstID, secondID}},
		},
		{
			name: "unicode name",
			raw:  `name == "л"`,
			want: ports.OrganizationFilter{NameEquals: &cyrillic},
		},
		{
			name: "equalities with both label access forms",
			raw:  `state == "DELETED" && name == "Production" && labels.environment == "prod" && labels["example.com/team"] == "platform"`,
			want: ports.OrganizationFilter{
				States:     []organization.State{organization.StateDeleted},
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
			want: ports.OrganizationFilter{
				States: []organization.State{organization.StateActive, organization.StateSuspended},
			},
		},
		{
			name: "reversed equality",
			raw:  `"Production" == name`,
			want: ports.OrganizationFilter{NameEquals: &production},
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
		{name: "invalid id", raw: `id == "invalid"`, wantErr: true},
		{name: "duplicate state", raw: `state == "ACTIVE" && state == "SUSPENDED"`, wantErr: true},
		{name: "duplicate state in list", raw: `state in ["ACTIVE", "ACTIVE"]`, wantErr: true},
		{name: "duplicate label", raw: `labels.team == "one" && labels["team"] == "two"`, wantErr: true},
		{name: "empty label key", raw: `labels[""] == "value"`, wantErr: true},
		{name: "oversized", raw: strings.Repeat("x", maximumOrganizationFilterRunes+1), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOrganizationFilter(test.raw)
			if test.wantErr {
				if !errors.Is(err, ErrInvalidOrganizationFilter) {
					t.Fatalf("parseOrganizationFilter() error = %v, want %v", err, ErrInvalidOrganizationFilter)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOrganizationFilter() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("parseOrganizationFilter() = %#v, want %#v", got, test.want)
			}
		})
	}
}
