package usecase

import (
	"fmt"
	"slices"
	"strings"

	platformfilter "github.com/m8-team/platform/internal/platform/filter"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

const maximumWorkspaceFilterRunes = 1024

var workspaceFilterParser, workspaceFilterParserError = platformfilter.NewCELParser(
	platformfilter.CELParserConfig{
		MaxExpressionRunes: maximumWorkspaceFilterRunes,
		Variables: []platformfilter.Variable{
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

type workspaceFilterBuilder struct {
	filter   ports.WorkspaceFilter
	seenName bool
}

func parseWorkspaceFilter(raw string) (ports.WorkspaceFilter, error) {
	if workspaceFilterParserError != nil {
		return ports.WorkspaceFilter{}, fmt.Errorf(
			"initialize workspace filter parser: %w",
			workspaceFilterParserError,
		)
	}

	expression, err := workspaceFilterParser.Parse(raw)
	if err != nil {
		return ports.WorkspaceFilter{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidWorkspaceFilter,
			err,
		)
	}

	builder := workspaceFilterBuilder{
		filter: ports.WorkspaceFilter{LabelsEqual: make(map[string]string)},
	}
	for _, predicate := range expression.Predicates() {
		if err := builder.add(predicate); err != nil {
			return ports.WorkspaceFilter{}, fmt.Errorf(
				"%w: %v",
				ErrInvalidWorkspaceFilter,
				err,
			)
		}
	}

	if len(builder.filter.LabelsEqual) == 0 {
		builder.filter.LabelsEqual = nil
	}
	slices.Sort(builder.filter.States)
	return builder.filter, nil
}

func (b *workspaceFilterBuilder) add(predicate platformfilter.Predicate) error {
	switch predicate.Operator() {
	case platformfilter.EqualsOperator:
		return b.addEquality(predicate)
	case platformfilter.InOperator:
		return b.addStateMembership(predicate)
	default:
		return fmt.Errorf("unsupported filter operator %d", predicate.Operator())
	}
}

func (b *workspaceFilterBuilder) addEquality(predicate platformfilter.Predicate) error {
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
	case "state":
		if _, hasKey := field.Key(); hasKey {
			return fmt.Errorf("state does not support a map key")
		}
		if len(b.filter.States) != 0 {
			return fmt.Errorf("duplicate state predicate")
		}
		state, err := parseWorkspaceFilterState(value)
		if err != nil {
			return err
		}
		b.filter.States = []workspace.State{state}
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
		key, err := workspaceLabelKey(field)
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

func (b *workspaceFilterBuilder) addStateMembership(predicate platformfilter.Predicate) error {
	field := predicate.Field()
	_, hasKey := field.Key()
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

	seen := make(map[workspace.State]struct{}, len(values))
	states := make([]workspace.State, 0, len(values))
	for _, literal := range values {
		value, ok := literal.AsString()
		if !ok {
			return fmt.Errorf("state membership values must be string literals")
		}
		state, err := parseWorkspaceFilterState(value)
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

func workspaceLabelKey(field platformfilter.FieldReference) (string, error) {
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

func parseWorkspaceFilterState(value string) (workspace.State, error) {
	switch strings.ToUpper(value) {
	case "CREATING":
		return workspace.StateCreating, nil
	case "ACTIVE":
		return workspace.StateActive, nil
	case "SUSPENDED":
		return workspace.StateSuspended, nil
	case "DELETING":
		return workspace.StateDeleting, nil
	case "DELETED":
		return workspace.StateDeleted, nil
	case "FAILED":
		return workspace.StateFailed, nil
	default:
		return workspace.StateUnspecified, fmt.Errorf("unsupported state %q", value)
	}
}
