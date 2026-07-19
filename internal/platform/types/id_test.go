package types

import (
	"encoding"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const validID = "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"

var (
	_ encoding.TextMarshaler   = ID{}
	_ encoding.TextUnmarshaler = (*ID)(nil)
)

func TestIDCreationAndMethods(t *testing.T) {
	raw := uuid.MustParse(validID)

	fromUUID, err := NewFromUUID(raw)
	if err != nil {
		t.Fatalf("NewIDFromUUID() error = %v", err)
	}

	parsed, err := Parse(validID)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ids := map[string]ID{
		"new":        New(),
		"from uuid":  fromUUID,
		"parsed":     parsed,
		"must parse": MustParse(validID),
	}

	for name, id := range ids {
		t.Run(name, func(t *testing.T) {
			if id.IsZero() {
				t.Fatal("id is zero")
			}
			if err := id.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}

	if parsed.UUID() != raw {
		t.Fatalf("UUID() = %s, want %s", parsed.UUID(), raw)
	}
	if parsed.String() != validID {
		t.Fatalf("String() = %q, want %q", parsed.String(), validID)
	}

	uppercase, err := Parse(strings.ToUpper(validID))
	if err != nil {
		t.Fatalf("Parse(uppercase) error = %v", err)
	}
	if uppercase.String() != validID {
		t.Fatalf("uppercase String() = %q, want %q", uppercase.String(), validID)
	}
	if !parsed.Equal(fromUUID) {
		t.Fatal("Equal() = false, want true")
	}
	if parsed.Equal(New()) {
		t.Fatal("Equal() = true, want false")
	}
}

func TestIDErrors(t *testing.T) {
	var zero ID

	tests := []struct {
		name    string
		run     func() error
		wantErr error
	}{
		{
			name:    "parse invalid uuid",
			run:     func() error { _, err := Parse("not-a-uuid"); return err },
			wantErr: ErrInvalidID,
		},
		{
			name:    "parse raw hex uuid",
			run:     func() error { _, err := Parse(strings.ReplaceAll(validID, "-", "")); return err },
			wantErr: ErrInvalidID,
		},
		{
			name:    "parse uuid in braces",
			run:     func() error { _, err := Parse("{" + validID + "}"); return err },
			wantErr: ErrInvalidID,
		},
		{
			name:    "parse uuid urn",
			run:     func() error { _, err := Parse("urn:uuid:" + validID); return err },
			wantErr: ErrInvalidID,
		},
		{
			name:    "parse uuid with whitespace",
			run:     func() error { _, err := Parse(" " + validID); return err },
			wantErr: ErrInvalidID,
		},
		{
			name:    "parse zero uuid",
			run:     func() error { _, err := Parse(uuid.Nil.String()); return err },
			wantErr: ErrZeroID,
		},
		{
			name:    "create from zero uuid",
			run:     func() error { _, err := NewFromUUID(uuid.Nil); return err },
			wantErr: ErrZeroID,
		},
		{
			name:    "validate zero id",
			run:     zero.Validate,
			wantErr: ErrZeroID,
		},
	}

	if !zero.IsZero() {
		t.Fatal("zero ID IsZero() = false, want true")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestIDTextAndJSONRoundTrip(t *testing.T) {
	id := MustParse(validID)

	text, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}
	if string(text) != validID {
		t.Fatalf("MarshalText() = %q, want %q", text, validID)
	}

	var fromText ID
	if err := fromText.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText() error = %v", err)
	}
	if !fromText.Equal(id) {
		t.Fatalf("UnmarshalText() = %s, want %s", fromText, id)
	}

	encoded, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(encoded) != `"`+validID+`"` {
		t.Fatalf("json.Marshal() = %s, want %q", encoded, validID)
	}

	var fromJSON ID
	if err := json.Unmarshal(encoded, &fromJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !fromJSON.Equal(id) {
		t.Fatalf("json.Unmarshal() = %s, want %s", fromJSON, id)
	}
}

func TestIDUnmarshalRejectsInvalidInputWithoutMutation(t *testing.T) {
	id := MustParse(validID)
	original := id

	if err := id.UnmarshalText([]byte("not-a-uuid")); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("UnmarshalText() error = %v, want %v", err, ErrInvalidID)
	}
	if !id.Equal(original) {
		t.Fatalf("ID changed after failed unmarshal: got %s, want %s", id, original)
	}

	var nilID *ID
	if err := nilID.UnmarshalText([]byte(validID)); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("nil UnmarshalText() error = %v, want %v", err, ErrInvalidID)
	}

	var zero ID
	if _, err := zero.MarshalText(); !errors.Is(err, ErrZeroID) {
		t.Fatalf("zero MarshalText() error = %v, want %v", err, ErrZeroID)
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("panic = nil, want non-nil")
		}
	}()

	MustParse("not-a-uuid")
}
