package organizationapp

import (
	"fmt"
	"slices"
	"strings"

	platformfilter "github.com/m8-team/platform/internal/platform/filter"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

const maximumOrganizationFilterRunes = 1024

func newFilterParser() (*platformfilter.CELParser, error) {
	return platformfilter.NewCELParser(
		platformfilter.CELParserConfig{
			MaxExpressionRunes: maximumOrganizationFilterRunes,
			Variables: []platformfilter.Variable{
				platformfilter.ScalarVariable("id", platformfilter.StringKind),
				platformfilter.ScalarVariable("state", platformfilter.StringKind),
				platformfilter.ScalarVariable("name", platformfilter.StringKind),
				platformfilter.MapVariable(
					"labels",
					platformfilter.StringKind,
					platformfilter.StringKind,
				),
			},
		},
	)
}

type organizationFilterBuilder struct {
	filter   ports.OrganizationFilter
	seenID   bool
	seenName bool
}

func parseOrganizationFilter(parser *platformfilter.CELParser, raw string) (ports.OrganizationFilter, error) {
	expression, err := parser.Parse(raw)
	if err != nil {
		return ports.OrganizationFilter{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidOrganizationFilter,
			err,
		)
	}

	builder := organizationFilterBuilder{
		filter: ports.OrganizationFilter{LabelsEqual: make(map[string]string)},
	}
	for _, predicate := range expression.Predicates() {
		if err := builder.add(predicate); err != nil {
			return ports.OrganizationFilter{}, fmt.Errorf(
				"%w: %v",
				ErrInvalidOrganizationFilter,
				err,
			)
		}
	}

	if len(builder.filter.LabelsEqual) == 0 {
		builder.filter.LabelsEqual = nil
	}
	slices.Sort(builder.filter.States)
	slices.SortFunc(builder.filter.IDs, func(a, b organization.ID) int { return strings.Compare(a.String(), b.String()) })
	return builder.filter, nil
}

func (b *organizationFilterBuilder) add(predicate platformfilter.Predicate) error {
	switch predicate.Operator() {
	case platformfilter.EqualsOperator:
		return b.addEquality(predicate)
	case platformfilter.InOperator:
		return b.addStateMembership(predicate)
	default:
		return fmt.Errorf("unsupported filter operator %d", predicate.Operator())
	}
}

func (b *organizationFilterBuilder) addEquality(predicate platformfilter.Predicate) error {
	values := predicate.Values()
	if len(values) != 1 {
		return fmt.Errorf("equality requires exactly one value")
	}
	field := predicate.Field()
	value, ok := values[0].AsString()
	if !ok {
		return fmt.Errorf("%s must be compared with a string literal", field.Name())
	}

	switch field.Name() {
	case "id":
		if _, hasKey := field.Key(); hasKey || b.seenID {
			return fmt.Errorf("invalid or duplicate id predicate")
		}
		id, err := organization.ParseID(value)
		if err != nil {
			return fmt.Errorf("invalid organization id: %w", err)
		}
		b.seenID = true
		b.filter.IDs = []organization.ID{id}
	case "state":
		if _, hasKey := field.Key(); hasKey {
			return fmt.Errorf("state does not support a map key")
		}
		if len(b.filter.States) != 0 {
			return fmt.Errorf("duplicate state predicate")
		}
		state, err := parseOrganizationFilterState(value)
		if err != nil {
			return err
		}
		b.filter.States = []organization.State{state}
	case "name":
		if _, hasKey := field.Key(); hasKey {
			return fmt.Errorf("name does not support a map key")
		}
		if b.seenName {
			return fmt.Errorf("duplicate name predicate")
		}
		b.seenName = true
		b.filter.NameEquals = &value
	case "labels":
		key, err := organizationLabelKey(field)
		if err != nil {
			return err
		}
		if _, duplicate := b.filter.LabelsEqual[key]; duplicate {
			return fmt.Errorf("duplicate label predicate %q", key)
		}
		b.filter.LabelsEqual[key] = value
	default:
		return fmt.Errorf("unsupported field %q", field.Name())
	}

	return nil
}

func (b *organizationFilterBuilder) addStateMembership(predicate platformfilter.Predicate) error {
	field := predicate.Field()
	_, hasKey := field.Key()
	if field.Name() == "id" && !hasKey {
		if b.seenID {
			return fmt.Errorf("duplicate id predicate")
		}
		b.seenID = true
		for _, literal := range predicate.Values() {
			value, ok := literal.AsString()
			if !ok {
				return fmt.Errorf("id membership values must be string literals")
			}
			id, err := organization.ParseID(value)
			if err != nil {
				return fmt.Errorf("invalid organization id: %w", err)
			}
			b.filter.IDs = append(b.filter.IDs, id)
		}
		if len(b.filter.IDs) == 0 {
			return fmt.Errorf("id membership requires a non-empty string list")
		}
		return nil
	}
	if field.Name() != "state" || hasKey {
		return fmt.Errorf("membership is supported only for state")
	}
	if len(b.filter.States) != 0 {
		return fmt.Errorf("duplicate state predicate")
	}
	values := predicate.Values()
	if len(values) == 0 {
		return fmt.Errorf("state membership requires a non-empty string list")
	}

	seen := make(map[organization.State]struct{}, len(values))
	states := make([]organization.State, 0, len(values))
	for _, literal := range values {
		value, ok := literal.AsString()
		if !ok {
			return fmt.Errorf("state membership values must be string literals")
		}
		state, err := parseOrganizationFilterState(value)
		if err != nil {
			return err
		}
		if _, duplicate := seen[state]; duplicate {
			return fmt.Errorf("duplicate state %q", state)
		}
		seen[state] = struct{}{}
		states = append(states, state)
	}
	b.filter.States = states
	return nil
}

func organizationLabelKey(field platformfilter.FieldReference) (string, error) {
	literal, ok := field.Key()
	if !ok {
		return "", fmt.Errorf("labels require a map key")
	}
	key, ok := literal.AsString()
	if !ok {
		return "", fmt.Errorf("label key must be a string literal")
	}
	if key == "" {
		return "", fmt.Errorf("label key is empty")
	}
	return key, nil
}

func parseOrganizationFilterState(value string) (organization.State, error) {
	switch strings.ToUpper(value) {
	case "CREATING":
		return organization.StateCreating, nil
	case "ACTIVE":
		return organization.StateActive, nil
	case "SUSPENDED":
		return organization.StateSuspended, nil
	case "DELETING":
		return organization.StateDeleting, nil
	case "DELETED":
		return organization.StateDeleted, nil
	case "FAILED":
		return organization.StateFailed, nil
	default:
		return organization.StateUnspecified, fmt.Errorf("unsupported state %q", value)
	}
}
