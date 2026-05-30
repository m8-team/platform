package types

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

const validID = "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"

func TestIDCreationAndMethods(t *testing.T) {
	raw := uuid.MustParse(validID)

	fromUUID, err := NewIDFromUUID(raw)
	if err != nil {
		t.Fatalf("NewIDFromUUID() error = %v", err)
	}

	parsed, err := ParseID(validID)
	if err != nil {
		t.Fatalf("ParseID() error = %v", err)
	}

	ids := map[string]ID{
		"new":        NewID(),
		"from uuid":  fromUUID,
		"parsed":     parsed,
		"must parse": MustParseID(validID),
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
	if !parsed.Equal(fromUUID) {
		t.Fatal("Equal() = false, want true")
	}
	if parsed.Equal(NewID()) {
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
			run:     func() error { _, err := ParseID("not-a-uuid"); return err },
			wantErr: ErrInvalidID,
		},
		{
			name:    "parse zero uuid",
			run:     func() error { _, err := ParseID(uuid.Nil.String()); return err },
			wantErr: ErrZeroID,
		},
		{
			name:    "create from zero uuid",
			run:     func() error { _, err := NewIDFromUUID(uuid.Nil); return err },
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

func TestMustParseIDPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("panic = nil, want non-nil")
		}
	}()

	MustParseID("not-a-uuid")
}
