package usecase

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

const maximumOrganizationFilterRunes = 1024

var organizationFilterEnvironment, organizationFilterEnvironmentError = cel.NewEnv(
	cel.ClearMacros(),
	cel.ParserExpressionSizeLimit(maximumOrganizationFilterRunes),
	cel.ParserRecursionLimit(32),
	cel.Variable("state", cel.StringType),
	cel.Variable("name", cel.StringType),
	cel.Variable("labels", cel.MapType(cel.StringType, cel.StringType)),
)

type organizationFilterFieldKind uint8

const (
	organizationFilterFieldUnknown organizationFilterFieldKind = iota
	organizationFilterFieldState
	organizationFilterFieldName
	organizationFilterFieldLabel
)

type organizationFilterField struct {
	kind     organizationFilterFieldKind
	labelKey string
}

type organizationFilterBuilder struct {
	filter   ports.OrganizationFilter
	seenName bool
}

func parseOrganizationFilter(raw string) (ports.OrganizationFilter, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.OrganizationFilter{}, nil
	}
	if utf8.RuneCountInString(raw) > maximumOrganizationFilterRunes {
		return ports.OrganizationFilter{}, fmt.Errorf(
			"%w: exceeds %d characters",
			ErrInvalidOrganizationFilter,
			maximumOrganizationFilterRunes,
		)
	}
	if organizationFilterEnvironmentError != nil {
		return ports.OrganizationFilter{}, fmt.Errorf(
			"%w: initialize CEL environment: %v",
			ErrInvalidOrganizationFilter,
			organizationFilterEnvironmentError,
		)
	}

	checked, issues := organizationFilterEnvironment.Compile(raw)
	if issues != nil && issues.Err() != nil {
		return ports.OrganizationFilter{}, fmt.Errorf(
			"%w: CEL expression: %v",
			ErrInvalidOrganizationFilter,
			issues.Err(),
		)
	}
	if !checked.OutputType().IsEquivalentType(cel.BoolType) {
		return ports.OrganizationFilter{}, fmt.Errorf(
			"%w: CEL expression must return bool, got %s",
			ErrInvalidOrganizationFilter,
			checked.OutputType(),
		)
	}

	builder := organizationFilterBuilder{
		filter: ports.OrganizationFilter{LabelsEqual: make(map[string]string)},
	}
	if err := builder.add(checked.NativeRep().Expr()); err != nil {
		return ports.OrganizationFilter{}, fmt.Errorf("%w: %v", ErrInvalidOrganizationFilter, err)
	}

	if len(builder.filter.LabelsEqual) == 0 {
		builder.filter.LabelsEqual = nil
	}
	slices.Sort(builder.filter.States)
	return builder.filter, nil
}

func (b *organizationFilterBuilder) add(expression ast.Expr) error {
	if expression.Kind() != ast.CallKind {
		return fmt.Errorf("unsupported CEL expression; expected equality, membership, or conjunction")
	}

	call := expression.AsCall()
	switch call.FunctionName() {
	case operators.LogicalAnd:
		if len(call.Args()) != 2 {
			return fmt.Errorf("invalid CEL conjunction")
		}
		if err := b.add(call.Args()[0]); err != nil {
			return err
		}
		return b.add(call.Args()[1])
	case operators.Equals:
		return b.addEquality(call.Args())
	case operators.In, operators.OldIn:
		return b.addStateMembership(call.Args())
	default:
		return fmt.Errorf("unsupported CEL operator or function %q", call.FunctionName())
	}
}

func (b *organizationFilterBuilder) addEquality(arguments []ast.Expr) error {
	if len(arguments) != 2 {
		return fmt.Errorf("invalid CEL equality")
	}

	field, fieldOK := parseOrganizationFilterField(arguments[0])
	value, valueOK := parseStringLiteral(arguments[1])
	if !fieldOK || !valueOK {
		field, fieldOK = parseOrganizationFilterField(arguments[1])
		value, valueOK = parseStringLiteral(arguments[0])
	}
	if !fieldOK || !valueOK {
		return fmt.Errorf("equality must compare a supported field with a string literal")
	}

	switch field.kind {
	case organizationFilterFieldState:
		if len(b.filter.States) != 0 {
			return fmt.Errorf("duplicate state predicate")
		}
		state, err := parseOrganizationFilterState(value)
		if err != nil {
			return err
		}
		b.filter.States = []organization.State{state}
	case organizationFilterFieldName:
		if b.seenName {
			return fmt.Errorf("duplicate name predicate")
		}
		b.seenName = true
		b.filter.NameEquals = &value
	case organizationFilterFieldLabel:
		if field.labelKey == "" {
			return fmt.Errorf("label key is empty")
		}
		if _, duplicate := b.filter.LabelsEqual[field.labelKey]; duplicate {
			return fmt.Errorf("duplicate label predicate %q", field.labelKey)
		}
		b.filter.LabelsEqual[field.labelKey] = value
	default:
		return fmt.Errorf("unsupported field")
	}

	return nil
}

func (b *organizationFilterBuilder) addStateMembership(arguments []ast.Expr) error {
	if len(arguments) != 2 {
		return fmt.Errorf("invalid CEL membership expression")
	}
	field, ok := parseOrganizationFilterField(arguments[0])
	if !ok || field.kind != organizationFilterFieldState {
		return fmt.Errorf("membership is supported only for state")
	}
	if len(b.filter.States) != 0 {
		return fmt.Errorf("duplicate state predicate")
	}
	if arguments[1].Kind() != ast.ListKind || arguments[1].AsList().Size() == 0 {
		return fmt.Errorf("state membership requires a non-empty string list")
	}

	seen := make(map[organization.State]struct{}, arguments[1].AsList().Size())
	states := make([]organization.State, 0, arguments[1].AsList().Size())
	for _, element := range arguments[1].AsList().Elements() {
		value, ok := parseStringLiteral(element)
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

func parseOrganizationFilterField(expression ast.Expr) (organizationFilterField, bool) {
	switch expression.Kind() {
	case ast.IdentKind:
		switch expression.AsIdent() {
		case "state":
			return organizationFilterField{kind: organizationFilterFieldState}, true
		case "name":
			return organizationFilterField{kind: organizationFilterFieldName}, true
		default:
			return organizationFilterField{}, false
		}
	case ast.SelectKind:
		selection := expression.AsSelect()
		if !selection.IsTestOnly() && isLabelsIdentifier(selection.Operand()) {
			return organizationFilterField{
				kind:     organizationFilterFieldLabel,
				labelKey: selection.FieldName(),
			}, true
		}
	case ast.CallKind:
		call := expression.AsCall()
		if call.FunctionName() != operators.Index || len(call.Args()) != 2 || !isLabelsIdentifier(call.Args()[0]) {
			return organizationFilterField{}, false
		}
		key, ok := parseStringLiteral(call.Args()[1])
		if !ok {
			return organizationFilterField{}, false
		}
		return organizationFilterField{kind: organizationFilterFieldLabel, labelKey: key}, true
	}

	return organizationFilterField{}, false
}

func isLabelsIdentifier(expression ast.Expr) bool {
	return expression.Kind() == ast.IdentKind && expression.AsIdent() == "labels"
}

func parseStringLiteral(expression ast.Expr) (string, bool) {
	if expression.Kind() != ast.LiteralKind {
		return "", false
	}
	value, ok := expression.AsLiteral().Value().(string)
	return value, ok
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
