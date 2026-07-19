package filter

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestCELParserParseLiteralKinds(t *testing.T) {
	parser := newTestCELParser(t)

	tests := []struct {
		name       string
		expression string
		kind       ValueKind
		assert     func(*testing.T, Literal)
	}{
		{
			name:       "string",
			expression: `name == "production"`,
			kind:       StringKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsString()
				if !ok || value != "production" {
					t.Fatalf("AsString() = %q, %v; want %q, true", value, ok, "production")
				}
			},
		},
		{
			name:       "bool",
			expression: `enabled == true`,
			kind:       BoolKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsBool()
				if !ok || !value {
					t.Fatalf("AsBool() = %v, %v; want true, true", value, ok)
				}
			},
		},
		{
			name:       "int",
			expression: `attempts == 7`,
			kind:       IntKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsInt()
				if !ok || value != 7 {
					t.Fatalf("AsInt() = %d, %v; want 7, true", value, ok)
				}
			},
		},
		{
			name:       "negative int",
			expression: `attempts == -1`,
			kind:       IntKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsInt()
				if !ok || value != -1 {
					t.Fatalf("AsInt() = %d, %v; want -1, true", value, ok)
				}
			},
		},
		{
			name:       "minimum int",
			expression: `attempts == -9223372036854775808`,
			kind:       IntKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsInt()
				if !ok || value != -9223372036854775808 {
					t.Fatalf("AsInt() = %d, %v; want minimum int, true", value, ok)
				}
			},
		},
		{
			name:       "uint",
			expression: `revision == 9u`,
			kind:       UintKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsUint()
				if !ok || value != 9 {
					t.Fatalf("AsUint() = %d, %v; want 9, true", value, ok)
				}
			},
		},
		{
			name:       "double",
			expression: `ratio == 1.5`,
			kind:       DoubleKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsDouble()
				if !ok || value != 1.5 {
					t.Fatalf("AsDouble() = %v, %v; want 1.5, true", value, ok)
				}
			},
		},
		{
			name:       "bytes",
			expression: `payload == b"ok"`,
			kind:       BytesKind,
			assert: func(t *testing.T, literal Literal) {
				t.Helper()
				value, ok := literal.AsBytes()
				if !ok || !reflect.DeepEqual(value, []byte("ok")) {
					t.Fatalf("AsBytes() = %q, %v; want %q, true", value, ok, []byte("ok"))
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parser.Parse(test.expression)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.expression, err)
			}

			predicate := requireSinglePredicate(t, expression)
			if predicate.Operator() != EqualsOperator {
				t.Fatalf("operator = %v; want %v", predicate.Operator(), EqualsOperator)
			}
			values := predicate.Values()
			if len(values) != 1 {
				t.Fatalf("value count = %d; want 1", len(values))
			}
			literal := values[0]
			if literal.Kind() != test.kind {
				t.Fatalf("kind = %v; want %v", literal.Kind(), test.kind)
			}
			test.assert(t, literal)
			assertOtherLiteralAccessorsUnavailable(t, literal, test.kind)
		})
	}
}

func TestCELParserParseConjunctionPreservesPredicateOrder(t *testing.T) {
	parser := newTestCELParser(t)

	expression, err := parser.Parse(
		`name == "production" && enabled == true && labels.environment == "prod"`,
	)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	predicates := expression.Predicates()
	if len(predicates) != 3 {
		t.Fatalf("predicate count = %d; want 3", len(predicates))
	}

	wantFields := []string{"name", "enabled", "labels"}
	for index, wantField := range wantFields {
		predicate := predicates[index]
		if predicate.Field().Name() != wantField {
			t.Errorf("predicate %d field = %q; want %q", index, predicate.Field().Name(), wantField)
		}
		if predicate.Operator() != EqualsOperator {
			t.Errorf("predicate %d operator = %v; want %v", index, predicate.Operator(), EqualsOperator)
		}
	}

	key := requireFieldKey(t, predicates[2].Field())
	if value, ok := key.AsString(); !ok || value != "environment" {
		t.Fatalf("map key = %q, %v; want %q, true", value, ok, "environment")
	}
}

func TestCELParserParseMapAccess(t *testing.T) {
	parser := newTestCELParser(t)

	tests := []struct {
		name       string
		expression string
		field      string
		keyKind    ValueKind
		assertKey  func(*testing.T, Literal)
	}{
		{
			name:       "string key using select syntax",
			expression: `labels.environment == "prod"`,
			field:      "labels",
			keyKind:    StringKind,
			assertKey: func(t *testing.T, key Literal) {
				t.Helper()
				value, ok := key.AsString()
				if !ok || value != "environment" {
					t.Fatalf("key = %q, %v; want %q, true", value, ok, "environment")
				}
			},
		},
		{
			name:       "string key using index syntax",
			expression: `labels["example.com/team"] == "platform"`,
			field:      "labels",
			keyKind:    StringKind,
			assertKey: func(t *testing.T, key Literal) {
				t.Helper()
				value, ok := key.AsString()
				if !ok || value != "example.com/team" {
					t.Fatalf("key = %q, %v; want %q, true", value, ok, "example.com/team")
				}
			},
		},
		{
			name:       "integer key using index syntax",
			expression: `codes[7] == "lucky"`,
			field:      "codes",
			keyKind:    IntKind,
			assertKey: func(t *testing.T, key Literal) {
				t.Helper()
				value, ok := key.AsInt()
				if !ok || value != 7 {
					t.Fatalf("key = %d, %v; want 7, true", value, ok)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parser.Parse(test.expression)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.expression, err)
			}
			predicate := requireSinglePredicate(t, expression)
			if predicate.Field().Name() != test.field {
				t.Fatalf("field = %q; want %q", predicate.Field().Name(), test.field)
			}
			key := requireFieldKey(t, predicate.Field())
			if key.Kind() != test.keyKind {
				t.Fatalf("key kind = %v; want %v", key.Kind(), test.keyKind)
			}
			test.assertKey(t, key)
		})
	}
}

func TestCELParserParseReverseEquality(t *testing.T) {
	parser := newTestCELParser(t)

	tests := []struct {
		name       string
		expression string
		field      string
		key        string
		value      string
	}{
		{
			name:       "scalar",
			expression: `"production" == name`,
			field:      "name",
			value:      "production",
		},
		{
			name:       "map element",
			expression: `"platform" == labels.team`,
			field:      "labels",
			key:        "team",
			value:      "platform",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parser.Parse(test.expression)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.expression, err)
			}
			predicate := requireSinglePredicate(t, expression)
			if predicate.Field().Name() != test.field {
				t.Fatalf("field = %q; want %q", predicate.Field().Name(), test.field)
			}
			if test.key == "" {
				if key, ok := predicate.Field().Key(); ok {
					t.Fatalf("key = %#v, true; want zero, false", key)
				}
			} else {
				key := requireFieldKey(t, predicate.Field())
				if value, ok := key.AsString(); !ok || value != test.key {
					t.Fatalf("key = %q, %v; want %q, true", value, ok, test.key)
				}
			}
			value, ok := predicate.Values()[0].AsString()
			if !ok || value != test.value {
				t.Fatalf("value = %q, %v; want %q, true", value, ok, test.value)
			}
		})
	}
}

func TestCELParserParseMembership(t *testing.T) {
	parser := newTestCELParser(t)

	tests := []struct {
		name       string
		expression string
		field      string
		key        string
		want       []any
	}{
		{
			name:       "strings",
			expression: `name in ["one", "two"]`,
			field:      "name",
			want:       []any{"one", "two"},
		},
		{
			name:       "integers",
			expression: `attempts in [1, 2, 3]`,
			field:      "attempts",
			want:       []any{int64(1), int64(2), int64(3)},
		},
		{
			name:       "signed doubles",
			expression: `ratio in [-1.5, 0.5]`,
			field:      "ratio",
			want:       []any{-1.5, 0.5},
		},
		{
			name:       "map element",
			expression: `labels.environment in ["prod", "staging"]`,
			field:      "labels",
			key:        "environment",
			want:       []any{"prod", "staging"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parser.Parse(test.expression)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.expression, err)
			}
			predicate := requireSinglePredicate(t, expression)
			if predicate.Field().Name() != test.field {
				t.Fatalf("field = %q; want %q", predicate.Field().Name(), test.field)
			}
			if predicate.Operator() != InOperator {
				t.Fatalf("operator = %v; want %v", predicate.Operator(), InOperator)
			}
			if test.key != "" {
				key := requireFieldKey(t, predicate.Field())
				if value, ok := key.AsString(); !ok || value != test.key {
					t.Fatalf("key = %q, %v; want %q, true", value, ok, test.key)
				}
			}
			values := predicate.Values()
			if len(values) != len(test.want) {
				t.Fatalf("value count = %d; want %d", len(values), len(test.want))
			}
			for index, want := range test.want {
				switch want := want.(type) {
				case string:
					got, ok := values[index].AsString()
					if !ok || got != want {
						t.Errorf("value %d = %q, %v; want %q, true", index, got, ok, want)
					}
				case int64:
					got, ok := values[index].AsInt()
					if !ok || got != want {
						t.Errorf("value %d = %d, %v; want %d, true", index, got, ok, want)
					}
				case float64:
					got, ok := values[index].AsDouble()
					if !ok || got != want {
						t.Errorf("value %d = %v, %v; want %v, true", index, got, ok, want)
					}
				default:
					t.Fatalf("unsupported test value type %T", want)
				}
			}
		})
	}
}

func TestCELParserParseEmptyFilter(t *testing.T) {
	parser := newTestCELParser(t)

	for _, raw := range []string{"", " ", "\t\n"} {
		expression, err := parser.Parse(raw)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", raw, err)
			continue
		}
		if predicates := expression.Predicates(); len(predicates) != 0 {
			t.Errorf("Parse(%q) predicates = %#v; want empty", raw, predicates)
		}
	}
}

func TestCELParserCopiesBytesReturnedByLiteral(t *testing.T) {
	parser := newTestCELParser(t)
	expression, err := parser.Parse(`payload == b"secret"`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	literal := requireSinglePredicate(t, expression).Values()[0]

	first, ok := literal.AsBytes()
	if !ok {
		t.Fatal("first AsBytes() ok = false; want true")
	}
	first[0] = 'X'

	second, ok := literal.AsBytes()
	if !ok {
		t.Fatal("second AsBytes() ok = false; want true")
	}
	if string(second) != "secret" {
		t.Fatalf("second AsBytes() = %q; want %q", second, "secret")
	}
	if len(first) > 0 && &first[0] == &second[0] {
		t.Fatal("AsBytes() returned slices backed by the same array")
	}
}

func TestCELParserReturnsDefensivePredicateAndValueCopies(t *testing.T) {
	parser := newTestCELParser(t)
	conjunction, err := parser.Parse(`name == "production" && enabled == true`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	predicates := conjunction.Predicates()
	if len(predicates) != 2 {
		t.Fatalf("predicate count = %d; want 2", len(predicates))
	}
	predicates[0] = Predicate{}

	secondRead := conjunction.Predicates()
	if secondRead[0].Field().Name() != "name" {
		t.Fatalf("field after mutating returned predicates = %q; want %q", secondRead[0].Field().Name(), "name")
	}
	values := secondRead[0].Values()
	values[0] = Literal{}

	thirdRead := conjunction.Predicates()
	value, ok := thirdRead[0].Values()[0].AsString()
	if !ok || value != "production" {
		t.Fatalf("value after mutating returned values = %q, %v; want %q, true", value, ok, "production")
	}
}

func TestCELParserSupportsConcurrentParse(t *testing.T) {
	parser := newTestCELParser(t)

	const workers = 16
	errorsChannel := make(chan error, workers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workers)
	for range workers {
		go func() {
			defer waitGroup.Done()
			for range 20 {
				conjunction, err := parser.Parse(
					`name == "production" && labels.environment in ["prod", "staging"]`,
				)
				if err != nil {
					errorsChannel <- err
					return
				}
				if len(conjunction.Predicates()) != 2 {
					errorsChannel <- errors.New("unexpected predicate count")
					return
				}
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatalf("concurrent Parse() error = %v", err)
	}
}

func TestNewCELParserUsesDefaultLimits(t *testing.T) {
	parser, err := NewCELParser(CELParserConfig{
		Variables: []Variable{ScalarVariable("name", StringKind)},
	})
	if err != nil {
		t.Fatalf("NewCELParser() error = %v", err)
	}
	if parser.maxExpressionRunes != DefaultMaxExpressionRunes {
		t.Errorf(
			"maxExpressionRunes = %d; want %d",
			parser.maxExpressionRunes,
			DefaultMaxExpressionRunes,
		)
	}
	if parser.maxPredicates != DefaultMaxPredicates {
		t.Errorf("maxPredicates = %d; want %d", parser.maxPredicates, DefaultMaxPredicates)
	}
	if parser.maxListItems != DefaultMaxListItems {
		t.Errorf("maxListItems = %d; want %d", parser.maxListItems, DefaultMaxListItems)
	}
}

func TestNewCELParserRejectsInvalidConfiguration(t *testing.T) {
	validVariable := ScalarVariable("name", StringKind)
	tests := []struct {
		name   string
		config CELParserConfig
	}{
		{
			name:   "variables are required",
			config: CELParserConfig{},
		},
		{
			name: "variable name is required",
			config: CELParserConfig{
				Variables: []Variable{ScalarVariable("", StringKind)},
			},
		},
		{
			name: "duplicate variable",
			config: CELParserConfig{
				Variables: []Variable{validVariable, validVariable},
			},
		},
		{
			name: "unsupported scalar kind",
			config: CELParserConfig{
				Variables: []Variable{ScalarVariable("name", ValueKind(255))},
			},
		},
		{
			name: "unspecified scalar kind",
			config: CELParserConfig{
				Variables: []Variable{ScalarVariable("name", UnspecifiedKind)},
			},
		},
		{
			name: "unsupported map key kind",
			config: CELParserConfig{
				Variables: []Variable{MapVariable("labels", ValueKind(255), StringKind)},
			},
		},
		{
			name: "double map key kind",
			config: CELParserConfig{
				Variables: []Variable{MapVariable("labels", DoubleKind, StringKind)},
			},
		},
		{
			name: "bytes map key kind",
			config: CELParserConfig{
				Variables: []Variable{MapVariable("labels", BytesKind, StringKind)},
			},
		},
		{
			name: "unsupported map value kind",
			config: CELParserConfig{
				Variables: []Variable{MapVariable("labels", StringKind, ValueKind(255))},
			},
		},
		{
			name: "scalar and map kinds cannot be mixed",
			config: CELParserConfig{
				Variables: []Variable{{
					name:         "mixed",
					scalarKind:   StringKind,
					mapKeyKind:   StringKind,
					mapValueKind: StringKind,
				}},
			},
		},
		{
			name: "negative expression limit",
			config: CELParserConfig{
				Variables:          []Variable{validVariable},
				MaxExpressionRunes: -1,
			},
		},
		{
			name: "negative recursion limit",
			config: CELParserConfig{
				Variables:         []Variable{validVariable},
				MaxRecursionDepth: -1,
			},
		},
		{
			name: "negative predicate limit",
			config: CELParserConfig{
				Variables:     []Variable{validVariable},
				MaxPredicates: -1,
			},
		},
		{
			name: "negative membership list limit",
			config: CELParserConfig{
				Variables:    []Variable{validVariable},
				MaxListItems: -1,
			},
		},
		{
			name: "invalid CEL identifier",
			config: CELParserConfig{
				Variables: []Variable{ScalarVariable("not-valid", StringKind)},
			},
		},
		{
			name: "reserved CEL identifier",
			config: CELParserConfig{
				Variables: []Variable{ScalarVariable("true", BoolKind)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser, err := NewCELParser(test.config)
			if err == nil {
				t.Fatalf("NewCELParser() = %#v, nil; want error", parser)
			}
			if !errors.Is(err, ErrInvalidConfiguration) {
				t.Fatalf("error = %v; want errors.Is(ErrInvalidConfiguration)", err)
			}
		})
	}
}

func TestCELParserEnforcesExpressionSizeLimit(t *testing.T) {
	parser, err := NewCELParser(CELParserConfig{
		Variables:          []Variable{ScalarVariable("name", StringKind)},
		MaxExpressionRunes: 8,
	})
	if err != nil {
		t.Fatalf("NewCELParser() error = %v", err)
	}

	_, err = parser.Parse(`name == "production"`)
	if err == nil {
		t.Fatal("Parse() error = nil; want size limit error")
	}
	if !errors.Is(err, ErrInvalidExpression) {
		t.Fatalf("error = %v; want errors.Is(ErrInvalidExpression)", err)
	}
	if !strings.Contains(err.Error(), "exceeds 8 characters") {
		t.Fatalf("error = %q; want expression-size detail", err)
	}
}

func TestCELParserEnforcesPredicateLimit(t *testing.T) {
	parser, err := NewCELParser(CELParserConfig{
		Variables:     []Variable{ScalarVariable("name", StringKind)},
		MaxPredicates: 2,
	})
	if err != nil {
		t.Fatalf("NewCELParser() error = %v", err)
	}

	if _, err := parser.Parse(`name == "one" && name == "two"`); err != nil {
		t.Fatalf("Parse() at predicate limit error = %v", err)
	}
	_, err = parser.Parse(`name == "one" && name == "two" && name == "three"`)
	if err == nil {
		t.Fatal("Parse() above predicate limit error = nil; want error")
	}
	if !errors.Is(err, ErrInvalidExpression) {
		t.Fatalf("error = %v; want errors.Is(ErrInvalidExpression)", err)
	}
	if !strings.Contains(err.Error(), "exceeds 2 predicates") {
		t.Fatalf("error = %q; want predicate-limit detail", err)
	}
}

func TestCELParserEnforcesMembershipListLimit(t *testing.T) {
	parser, err := NewCELParser(CELParserConfig{
		Variables:    []Variable{ScalarVariable("name", StringKind)},
		MaxListItems: 2,
	})
	if err != nil {
		t.Fatalf("NewCELParser() error = %v", err)
	}

	if _, err := parser.Parse(`name in ["one", "two"]`); err != nil {
		t.Fatalf("Parse() at membership-list limit error = %v", err)
	}
	_, err = parser.Parse(`name in ["one", "two", "three"]`)
	if err == nil {
		t.Fatal("Parse() above membership-list limit error = nil; want error")
	}
	if !errors.Is(err, ErrInvalidExpression) {
		t.Fatalf("error = %v; want errors.Is(ErrInvalidExpression)", err)
	}
	if !strings.Contains(err.Error(), "membership list exceeds 2 items") {
		t.Fatalf("error = %q; want membership-list-limit detail", err)
	}
}

func TestCELParserEnforcesRecursionLimit(t *testing.T) {
	parser, err := NewCELParser(CELParserConfig{
		Variables:         []Variable{ScalarVariable("enabled", BoolKind)},
		MaxRecursionDepth: 2,
	})
	if err != nil {
		t.Fatalf("NewCELParser() error = %v", err)
	}

	_, err = parser.Parse(`((((enabled == true))))`)
	if err == nil {
		t.Fatal("Parse() error = nil; want recursion limit error")
	}
	if !errors.Is(err, ErrInvalidExpression) {
		t.Fatalf("error = %v; want errors.Is(ErrInvalidExpression)", err)
	}
}

func TestCELParserRejectsInvalidExpressions(t *testing.T) {
	parser := newTestCELParser(t)

	tests := []struct {
		name       string
		expression string
	}{
		{name: "legacy equality syntax", expression: `name = "prod"`},
		{name: "unknown variable", expression: `unknown == "prod"`},
		{name: "non boolean result", expression: `name`},
		{name: "logical or", expression: `name == "prod" || name == "dev"`},
		{name: "logical not", expression: `!(name == "prod")`},
		{name: "relational operator", expression: `attempts > 1`},
		{name: "not equals operator", expression: `name != "prod"`},
		{name: "member function", expression: `name.startsWith("prod")`},
		{name: "global function", expression: `size(name) == 4`},
		{name: "field to field equality", expression: `name == alias`},
		{name: "literal to literal equality", expression: `"prod" == "prod"`},
		{name: "computed equality value", expression: `name == ("pro" + "d")`},
		{name: "map variable without key", expression: `labels == {"team": "platform"}`},
		{name: "wrong map key type", expression: `labels[1] == "platform"`},
		{name: "membership list is empty", expression: `name in []`},
		{name: "membership right side is map", expression: `name in labels`},
		{name: "membership has computed item", expression: `name in ["prod", "de" + "v"]`},
		{name: "membership is reversed", expression: `["prod", "dev"] in name`},
		{name: "macro", expression: `["prod"].exists(value, value == name)`},
		{name: "conditional", expression: `true ? name == "prod" : name == "dev"`},
		{name: "wrong scalar literal type", expression: `name == true`},
		{name: "wrong map value type", expression: `labels.team == true`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expression, err := parser.Parse(test.expression)
			if err == nil {
				t.Fatalf("Parse(%q) = %#v, nil; want error", test.expression, expression)
			}
			if !errors.Is(err, ErrInvalidExpression) {
				t.Fatalf("error = %v; want errors.Is(ErrInvalidExpression)", err)
			}
		})
	}
}

func newTestCELParser(t *testing.T) *CELParser {
	t.Helper()
	parser, err := NewCELParser(CELParserConfig{
		Variables: []Variable{
			ScalarVariable("name", StringKind),
			ScalarVariable("alias", StringKind),
			ScalarVariable("enabled", BoolKind),
			ScalarVariable("attempts", IntKind),
			ScalarVariable("revision", UintKind),
			ScalarVariable("ratio", DoubleKind),
			ScalarVariable("payload", BytesKind),
			MapVariable("labels", StringKind, StringKind),
			MapVariable("codes", IntKind, StringKind),
		},
	})
	if err != nil {
		t.Fatalf("NewCELParser() error = %v", err)
	}
	return parser
}

func requireSinglePredicate(t *testing.T, expression Conjunction) Predicate {
	t.Helper()
	predicates := expression.Predicates()
	if len(predicates) != 1 {
		t.Fatalf("predicate count = %d; want 1", len(predicates))
	}
	return predicates[0]
}

func requireFieldKey(t *testing.T, field FieldReference) Literal {
	t.Helper()
	key, ok := field.Key()
	if !ok {
		t.Fatalf("field %q key missing; want key", field.Name())
	}
	return key
}

func assertOtherLiteralAccessorsUnavailable(t *testing.T, literal Literal, kind ValueKind) {
	t.Helper()
	if _, ok := literal.AsString(); ok != (kind == StringKind) {
		t.Errorf("AsString() availability = %v; kind = %v", ok, kind)
	}
	if _, ok := literal.AsBool(); ok != (kind == BoolKind) {
		t.Errorf("AsBool() availability = %v; kind = %v", ok, kind)
	}
	if _, ok := literal.AsInt(); ok != (kind == IntKind) {
		t.Errorf("AsInt() availability = %v; kind = %v", ok, kind)
	}
	if _, ok := literal.AsUint(); ok != (kind == UintKind) {
		t.Errorf("AsUint() availability = %v; kind = %v", ok, kind)
	}
	if _, ok := literal.AsDouble(); ok != (kind == DoubleKind) {
		t.Errorf("AsDouble() availability = %v; kind = %v", ok, kind)
	}
	if _, ok := literal.AsBytes(); ok != (kind == BytesKind) {
		t.Errorf("AsBytes() availability = %v; kind = %v", ok, kind)
	}
}

func FuzzCELParserParseDoesNotPanic(f *testing.F) {
	parser, err := NewCELParser(CELParserConfig{
		Variables: []Variable{
			ScalarVariable("name", StringKind),
			ScalarVariable("enabled", BoolKind),
			MapVariable("labels", StringKind, StringKind),
		},
	})
	if err != nil {
		f.Fatalf("NewCELParser() error = %v", err)
	}
	for _, seed := range []string{
		`name == "production"`,
		`labels.environment in ["prod", "staging"]`,
		`name.startsWith("prod")`,
		"\x00\xff",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		_, err := parser.Parse(raw)
		if err != nil && !errors.Is(err, ErrInvalidExpression) {
			t.Fatalf("Parse(%q) error = %v; want ErrInvalidExpression", raw, err)
		}
	})
}
