package types

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewID(t *testing.T) {
	id := NewID()

	if id.IsZero() {
		t.Fatal("expected generated id to be non-zero")
	}

	if err := id.Validate(); err != nil {
		t.Fatalf("expected generated id to be valid: %v", err)
	}
}

func TestNewIDFromUUID(t *testing.T) {
	raw := uuid.MustParse("018f3f16-9950-7a48-9d12-9fb6d8f4c8f2")

	id, err := NewIDFromUUID(raw)
	if err != nil {
		t.Fatalf("expected id from uuid: %v", err)
	}

	if id.UUID() != raw {
		t.Fatalf("expected uuid %s, got %s", raw, id.UUID())
	}
}

func TestNewIDFromUUIDReturnsErrorForZeroUUID(t *testing.T) {
	_, err := NewIDFromUUID(uuid.Nil)

	if !errors.Is(err, ErrZeroID) {
		t.Fatalf("expected ErrZeroID, got %v", err)
	}
}

func TestParseID(t *testing.T) {
	raw := "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"

	id, err := ParseID(raw)
	if err != nil {
		t.Fatalf("expected parsed id: %v", err)
	}

	if id.String() != raw {
		t.Fatalf("expected string %q, got %q", raw, id.String())
	}
}

func TestParseIDReturnsError(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{
			name:    "invalid uuid",
			value:   "not-a-uuid",
			wantErr: ErrInvalidID,
		},
		{
			name:    "zero uuid",
			value:   uuid.Nil.String(),
			wantErr: ErrZeroID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseID(tt.value)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestMustParseID(t *testing.T) {
	raw := "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"

	id := MustParseID(raw)

	if id.String() != raw {
		t.Fatalf("expected string %q, got %q", raw, id.String())
	}
}

func TestMustParseIDPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	MustParseID("not-a-uuid")
}

func TestIDMethods(t *testing.T) {
	raw := uuid.MustParse("018f3f16-9950-7a48-9d12-9fb6d8f4c8f2")
	id := ID(raw)

	if id.UUID() != raw {
		t.Fatalf("expected uuid %s, got %s", raw, id.UUID())
	}

	if id.String() != raw.String() {
		t.Fatalf("expected string %q, got %q", raw.String(), id.String())
	}

	if id.IsZero() {
		t.Fatal("expected id to be non-zero")
	}

	if err := id.Validate(); err != nil {
		t.Fatalf("expected id to be valid: %v", err)
	}

	if !id.Equal(ID(raw)) {
		t.Fatal("expected ids to be equal")
	}

	if id.Equal(NewID()) {
		t.Fatal("expected different ids to be not equal")
	}
}

func TestZeroID(t *testing.T) {
	var id ID

	if !id.IsZero() {
		t.Fatal("expected zero id")
	}

	if err := id.Validate(); !errors.Is(err, ErrZeroID) {
		t.Fatalf("expected ErrZeroID, got %v", err)
	}
}
