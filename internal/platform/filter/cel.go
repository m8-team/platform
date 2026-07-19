package filter

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
)

// Default parser resource limits apply when the corresponding config value is
// zero.
const (
	// DefaultMaxExpressionRunes bounds the CEL source length.
	DefaultMaxExpressionRunes = 1024
	// DefaultMaxRecursionDepth bounds parser nesting.
	DefaultMaxRecursionDepth = 32
	// DefaultMaxPredicates bounds normalized AND terms.
	DefaultMaxPredicates = 32
	// DefaultMaxListItems bounds literals in one membership predicate.
	DefaultMaxListItems = 100
)

// Parser and expression errors are separated so startup configuration failures
// are not reported to API clients as invalid filter input.
var (
	// ErrInvalidConfiguration identifies an invalid parser schema or limit.
	ErrInvalidConfiguration = errors.New("invalid filter parser configuration")
	// ErrInvalidExpression identifies invalid client filter input.
	ErrInvalidExpression = errors.New("invalid filter expression")
)

// ValueKind identifies scalar CEL literal types supported by the normalized
// predicate representation.
type ValueKind uint8

const (
	// UnspecifiedKind is not a valid declared variable or literal kind.
	UnspecifiedKind ValueKind = iota
	// StringKind represents CEL string values.
	StringKind
	// BoolKind represents CEL bool values.
	BoolKind
	// IntKind represents CEL int values.
	IntKind
	// UintKind represents CEL uint values.
	UintKind
	// DoubleKind represents CEL double values.
	DoubleKind
	// BytesKind represents CEL bytes values.
	BytesKind
)

// Variable describes one top-level CEL variable. Use ScalarVariable or
// MapVariable to construct valid declarations.
type Variable struct {
	name         string
	scalarKind   ValueKind
	mapKeyKind   ValueKind
	mapValueKind ValueKind
}

// ScalarVariable declares a top-level scalar field available to CEL clients.
func ScalarVariable(name string, kind ValueKind) Variable {
	return Variable{name: name, scalarKind: kind}
}

// MapVariable declares a top-level map available through select or index
// syntax. CEL map keys support string, bool, int, and uint kinds.
func MapVariable(name string, keyKind, valueKind ValueKind) Variable {
	return Variable{name: name, mapKeyKind: keyKind, mapValueKind: valueKind}
}

func (v Variable) isMap() bool {
	return v.mapKeyKind != UnspecifiedKind || v.mapValueKind != UnspecifiedKind
}

// CELParserConfig controls the public variables and parser resource limits.
type CELParserConfig struct {
	Variables          []Variable
	MaxExpressionRunes int
	MaxRecursionDepth  int
	MaxPredicates      int
	MaxListItems       int
}

// CELParser compiles a restricted CEL expression into predicates suitable for
// module-specific validation and persistence pushdown. It is safe for
// concurrent use after construction.
type CELParser struct {
	environment        *cel.Env
	variables          map[string]Variable
	maxExpressionRunes int
	maxPredicates      int
	maxListItems       int
}

// Conjunction is an immutable AND-only sequence of normalized predicates. A
// zero conjunction represents an empty API filter.
type Conjunction struct {
	predicates []Predicate
}

// Predicates returns a defensive copy in source expression order.
func (c Conjunction) Predicates() []Predicate {
	return slices.Clone(c.predicates)
}

// Operator identifies a comparison supported by the normalized subset.
type Operator uint8

const (
	// UnspecifiedOperator is not emitted by CELParser.
	UnspecifiedOperator Operator = iota
	// EqualsOperator compares one field with one literal.
	EqualsOperator
	// InOperator compares one field with a non-empty literal list.
	InOperator
)

// Predicate compares a scalar field or map element with literal values.
// EqualsOperator always has one value; InOperator has one or more values.
type Predicate struct {
	field    FieldReference
	operator Operator
	values   []Literal
}

// Field returns the compared scalar field or map element.
func (p Predicate) Field() FieldReference {
	return p.field
}

// Operator returns the normalized comparison operator.
func (p Predicate) Operator() Operator {
	return p.operator
}

// Values returns a defensive copy of the comparison literals.
func (p Predicate) Values() []Literal {
	return slices.Clone(p.values)
}

// FieldReference names either a scalar variable or one element of a map
// variable.
type FieldReference struct {
	name string
	key  *Literal
}

// Name returns the configured top-level variable name.
func (f FieldReference) Name() string {
	return f.name
}

// Key returns the map key. The second result is false for scalar fields.
func (f FieldReference) Key() (Literal, bool) {
	if f.key == nil {
		return Literal{}, false
	}
	return *f.key, true
}

// Literal contains a CEL scalar without exposing cel-go runtime types.
type Literal struct {
	kind  ValueKind
	value any
}

// Kind returns the normalized CEL scalar kind.
func (v Literal) Kind() ValueKind {
	return v.kind
}

// AsString returns the value when Kind is StringKind.
func (v Literal) AsString() (string, bool) {
	value, ok := v.value.(string)
	return value, ok
}

// AsBool returns the value when Kind is BoolKind.
func (v Literal) AsBool() (bool, bool) {
	value, ok := v.value.(bool)
	return value, ok
}

// AsInt returns the value when Kind is IntKind.
func (v Literal) AsInt() (int64, bool) {
	value, ok := v.value.(int64)
	return value, ok
}

// AsUint returns the value when Kind is UintKind.
func (v Literal) AsUint() (uint64, bool) {
	value, ok := v.value.(uint64)
	return value, ok
}

// AsDouble returns the value when Kind is DoubleKind.
func (v Literal) AsDouble() (float64, bool) {
	value, ok := v.value.(float64)
	return value, ok
}

// AsBytes returns a defensive copy when Kind is BytesKind.
func (v Literal) AsBytes() ([]byte, bool) {
	value, ok := v.value.([]byte)
	return bytes.Clone(value), ok
}

// NewCELParser validates a field schema and constructs a reusable immutable
// parser. Configuration errors wrap ErrInvalidConfiguration.
func NewCELParser(config CELParserConfig) (*CELParser, error) {
	maxExpressionRunes := config.MaxExpressionRunes
	if maxExpressionRunes == 0 {
		maxExpressionRunes = DefaultMaxExpressionRunes
	}
	if maxExpressionRunes < 1 {
		return nil, fmt.Errorf("%w: max expression runes must be positive", ErrInvalidConfiguration)
	}
	maxRecursionDepth := config.MaxRecursionDepth
	if maxRecursionDepth == 0 {
		maxRecursionDepth = DefaultMaxRecursionDepth
	}
	if maxRecursionDepth < 1 {
		return nil, fmt.Errorf("%w: max recursion depth must be positive", ErrInvalidConfiguration)
	}
	maxPredicates := config.MaxPredicates
	if maxPredicates == 0 {
		maxPredicates = DefaultMaxPredicates
	}
	if maxPredicates < 1 {
		return nil, fmt.Errorf("%w: max predicates must be positive", ErrInvalidConfiguration)
	}
	maxListItems := config.MaxListItems
	if maxListItems == 0 {
		maxListItems = DefaultMaxListItems
	}
	if maxListItems < 1 {
		return nil, fmt.Errorf("%w: max list items must be positive", ErrInvalidConfiguration)
	}
	if len(config.Variables) == 0 {
		return nil, fmt.Errorf("%w: at least one variable is required", ErrInvalidConfiguration)
	}

	variables := make(map[string]Variable, len(config.Variables))
	options := []cel.EnvOption{
		cel.ClearMacros(),
		cel.ParserExpressionSizeLimit(maxExpressionRunes),
		cel.ParserRecursionLimit(maxRecursionDepth),
	}
	for _, variable := range config.Variables {
		if variable.name == "" {
			return nil, fmt.Errorf("%w: variable name is required", ErrInvalidConfiguration)
		}
		if _, duplicate := variables[variable.name]; duplicate {
			return nil, fmt.Errorf(
				"%w: duplicate variable %q",
				ErrInvalidConfiguration,
				variable.name,
			)
		}
		variableType, err := celVariableType(variable)
		if err != nil {
			return nil, err
		}
		variables[variable.name] = variable
		options = append(options, cel.Variable(variable.name, variableType))
	}

	environment, err := cel.NewEnv(options...)
	if err != nil {
		return nil, fmt.Errorf("%w: create CEL environment: %v", ErrInvalidConfiguration, err)
	}
	for _, variable := range config.Variables {
		parsed, issues := environment.Parse(variable.name)
		if parsed == nil || issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf(
				"%w: variable name %q is not a CEL identifier",
				ErrInvalidConfiguration,
				variable.name,
			)
		}
		expression := parsed.NativeRep().Expr()
		if expression.Kind() != ast.IdentKind || expression.AsIdent() != variable.name {
			return nil, fmt.Errorf(
				"%w: variable name %q is not a CEL identifier",
				ErrInvalidConfiguration,
				variable.name,
			)
		}
	}
	return &CELParser{
		environment:        environment,
		variables:          variables,
		maxExpressionRunes: maxExpressionRunes,
		maxPredicates:      maxPredicates,
		maxListItems:       maxListItems,
	}, nil
}

// Parse validates a CEL filter and normalizes it into an immutable AND-only
// conjunction. Client expression errors wrap ErrInvalidExpression.
func (p *CELParser) Parse(raw string) (Conjunction, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Conjunction{}, nil
	}
	if utf8.RuneCountInString(raw) > p.maxExpressionRunes {
		return Conjunction{}, fmt.Errorf(
			"%w: exceeds %d characters",
			ErrInvalidExpression,
			p.maxExpressionRunes,
		)
	}

	checked, issues := p.environment.Compile(raw)
	if issues != nil && issues.Err() != nil {
		return Conjunction{}, fmt.Errorf("%w: CEL expression: %v", ErrInvalidExpression, issues.Err())
	}
	if !checked.OutputType().IsEquivalentType(cel.BoolType) {
		return Conjunction{}, fmt.Errorf(
			"%w: CEL expression must return bool, got %s",
			ErrInvalidExpression,
			checked.OutputType(),
		)
	}

	decoder := predicateDecoder{
		variables:     p.variables,
		maxPredicates: p.maxPredicates,
		maxListItems:  p.maxListItems,
	}
	if err := decoder.decode(checked.NativeRep().Expr()); err != nil {
		return Conjunction{}, fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}
	return Conjunction{predicates: decoder.predicates}, nil
}

type predicateDecoder struct {
	variables     map[string]Variable
	maxPredicates int
	maxListItems  int
	predicates    []Predicate
}

func (d *predicateDecoder) decode(expression ast.Expr) error {
	if expression.Kind() != ast.CallKind {
		return fmt.Errorf("expected equality, membership, or conjunction")
	}

	call := expression.AsCall()
	if call.IsMemberFunction() {
		return fmt.Errorf("member functions are not supported")
	}
	switch call.FunctionName() {
	case operators.LogicalAnd:
		if len(call.Args()) < 2 {
			return fmt.Errorf("invalid CEL conjunction")
		}
		for _, argument := range call.Args() {
			if err := d.decode(argument); err != nil {
				return err
			}
		}
		return nil
	case operators.Equals:
		predicate, err := d.decodeEquality(call.Args())
		if err != nil {
			return err
		}
		return d.appendPredicate(predicate)
	case operators.In, operators.OldIn:
		predicate, err := d.decodeMembership(call.Args())
		if err != nil {
			return err
		}
		return d.appendPredicate(predicate)
	default:
		return fmt.Errorf("unsupported CEL operator or function %q", call.FunctionName())
	}
}

func (d *predicateDecoder) appendPredicate(predicate Predicate) error {
	if len(d.predicates) >= d.maxPredicates {
		return fmt.Errorf("exceeds %d predicates", d.maxPredicates)
	}
	d.predicates = append(d.predicates, predicate)
	return nil
}

func (d *predicateDecoder) decodeEquality(arguments []ast.Expr) (Predicate, error) {
	if len(arguments) != 2 {
		return Predicate{}, fmt.Errorf("invalid CEL equality")
	}

	field, fieldOK := d.parseField(arguments[0])
	value, valueOK := parseLiteral(arguments[1])
	if !fieldOK || !valueOK {
		field, fieldOK = d.parseField(arguments[1])
		value, valueOK = parseLiteral(arguments[0])
	}
	if !fieldOK || !valueOK {
		return Predicate{}, fmt.Errorf("equality must compare a configured field with a scalar literal")
	}
	return Predicate{
		field:    field,
		operator: EqualsOperator,
		values:   []Literal{value},
	}, nil
}

func (d *predicateDecoder) decodeMembership(arguments []ast.Expr) (Predicate, error) {
	if len(arguments) != 2 {
		return Predicate{}, fmt.Errorf("invalid CEL membership expression")
	}
	field, ok := d.parseField(arguments[0])
	if !ok || arguments[1].Kind() != ast.ListKind {
		return Predicate{}, fmt.Errorf("membership must compare a configured field with a literal list")
	}

	list := arguments[1].AsList()
	if list.Size() == 0 {
		return Predicate{}, fmt.Errorf("membership list must not be empty")
	}
	if list.Size() > d.maxListItems {
		return Predicate{}, fmt.Errorf("membership list exceeds %d items", d.maxListItems)
	}
	if len(list.OptionalIndices()) != 0 {
		return Predicate{}, fmt.Errorf("optional list elements are not supported")
	}
	values := make([]Literal, 0, list.Size())
	for _, element := range list.Elements() {
		value, ok := parseLiteral(element)
		if !ok {
			return Predicate{}, fmt.Errorf("membership list must contain only scalar literals")
		}
		values = append(values, value)
	}
	return Predicate{
		field:    field,
		operator: InOperator,
		values:   values,
	}, nil
}

func (d *predicateDecoder) parseField(expression ast.Expr) (FieldReference, bool) {
	switch expression.Kind() {
	case ast.IdentKind:
		name := expression.AsIdent()
		variable, exists := d.variables[name]
		if !exists || variable.isMap() {
			return FieldReference{}, false
		}
		return FieldReference{name: name}, true
	case ast.SelectKind:
		selection := expression.AsSelect()
		if selection.IsTestOnly() || selection.Operand().Kind() != ast.IdentKind {
			return FieldReference{}, false
		}
		name := selection.Operand().AsIdent()
		variable, exists := d.variables[name]
		if !exists || !variable.isMap() || variable.mapKeyKind != StringKind {
			return FieldReference{}, false
		}
		key := newLiteral(StringKind, selection.FieldName())
		return FieldReference{name: name, key: &key}, true
	case ast.CallKind:
		call := expression.AsCall()
		if call.IsMemberFunction() || call.FunctionName() != operators.Index || len(call.Args()) != 2 {
			return FieldReference{}, false
		}
		if call.Args()[0].Kind() != ast.IdentKind {
			return FieldReference{}, false
		}
		name := call.Args()[0].AsIdent()
		variable, exists := d.variables[name]
		if !exists || !variable.isMap() {
			return FieldReference{}, false
		}
		key, ok := parseLiteral(call.Args()[1])
		if !ok || key.Kind() != variable.mapKeyKind {
			return FieldReference{}, false
		}
		return FieldReference{name: name, key: &key}, true
	default:
		return FieldReference{}, false
	}
}

func parseLiteral(expression ast.Expr) (Literal, bool) {
	if expression.Kind() == ast.CallKind {
		call := expression.AsCall()
		if call.IsMemberFunction() ||
			call.FunctionName() != operators.Negate ||
			len(call.Args()) != 1 ||
			call.Args()[0].Kind() != ast.LiteralKind {
			return Literal{}, false
		}
		operand, ok := parseLiteral(call.Args()[0])
		if !ok {
			return Literal{}, false
		}
		switch value := operand.value.(type) {
		case int64:
			return newLiteral(IntKind, -value), true
		case float64:
			return newLiteral(DoubleKind, -value), true
		default:
			return Literal{}, false
		}
	}
	if expression.Kind() != ast.LiteralKind {
		return Literal{}, false
	}
	switch value := expression.AsLiteral().Value().(type) {
	case string:
		return newLiteral(StringKind, value), true
	case bool:
		return newLiteral(BoolKind, value), true
	case int64:
		return newLiteral(IntKind, value), true
	case uint64:
		return newLiteral(UintKind, value), true
	case float64:
		return newLiteral(DoubleKind, value), true
	case []byte:
		return newLiteral(BytesKind, bytes.Clone(value)), true
	default:
		return Literal{}, false
	}
}

func newLiteral(kind ValueKind, value any) Literal {
	return Literal{kind: kind, value: value}
}

func celVariableType(variable Variable) (*cel.Type, error) {
	if variable.isMap() {
		if variable.scalarKind != UnspecifiedKind {
			return nil, fmt.Errorf(
				"%w: variable %q mixes scalar and map types",
				ErrInvalidConfiguration,
				variable.name,
			)
		}
		keyType, ok := celMapKeyType(variable.mapKeyKind)
		if !ok {
			return nil, fmt.Errorf(
				"%w: variable %q has unsupported map key kind",
				ErrInvalidConfiguration,
				variable.name,
			)
		}
		valueType, ok := celScalarType(variable.mapValueKind)
		if !ok {
			return nil, fmt.Errorf(
				"%w: variable %q has unsupported map value kind",
				ErrInvalidConfiguration,
				variable.name,
			)
		}
		return cel.MapType(keyType, valueType), nil
	}

	scalarType, ok := celScalarType(variable.scalarKind)
	if !ok {
		return nil, fmt.Errorf(
			"%w: variable %q has unsupported scalar kind",
			ErrInvalidConfiguration,
			variable.name,
		)
	}
	return scalarType, nil
}

func celMapKeyType(kind ValueKind) (*cel.Type, bool) {
	switch kind {
	case StringKind:
		return cel.StringType, true
	case BoolKind:
		return cel.BoolType, true
	case IntKind:
		return cel.IntType, true
	case UintKind:
		return cel.UintType, true
	default:
		return nil, false
	}
}

func celScalarType(kind ValueKind) (*cel.Type, bool) {
	switch kind {
	case StringKind:
		return cel.StringType, true
	case BoolKind:
		return cel.BoolType, true
	case IntKind:
		return cel.IntType, true
	case UintKind:
		return cel.UintType, true
	case DoubleKind:
		return cel.DoubleType, true
	case BytesKind:
		return cel.BytesType, true
	default:
		return nil, false
	}
}
