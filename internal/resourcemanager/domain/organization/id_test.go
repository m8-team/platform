package organization

import (
	"errors"
	"testing"
)

const testOrganizationID = "018f3f16-9950-7a48-9d12-9fb6d8f4c8f2"

func TestID(t *testing.T) {
	t.Parallel()

	parsed, err := ParseID(testOrganizationID)
	if err != nil {
		t.Fatalf("ParseID() error = %v", err)
	}
	if parsed.String() != testOrganizationID {
		t.Fatalf("String() = %q, want %q", parsed.String(), testOrganizationID)
	}
	if parsed.IsZero() {
		t.Fatal("IsZero() = true, want false")
	}
	if err := parsed.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !parsed.Equal(MustParseID(testOrganizationID)) {
		t.Fatal("Equal() = false, want true")
	}
	if parsed.Equal(NewID()) {
		t.Fatal("Equal(new ID) = true, want false")
	}
	if id := NewID(); id.IsZero() || id.Validate() != nil {
		t.Fatalf("NewID() = %v, want a valid non-zero ID", id)
	}
}

func TestParseIDErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{name: "empty", value: "", wantErr: ErrEmptyOrganizationID},
		{name: "malformed", value: "not-a-uuid", wantErr: ErrInvalidOrganizationID},
		{name: "zero UUID", value: "00000000-0000-0000-0000-000000000000", wantErr: ErrInvalidOrganizationID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseID(tt.value); !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseID() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	var zero ID
	if err := zero.Validate(); !errors.Is(err, ErrEmptyOrganizationID) {
		t.Fatalf("zero ID Validate() error = %v, want %v", err, ErrEmptyOrganizationID)
	}
}

func TestMustParseIDPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("MustParseID() did not panic")
		}
	}()
	MustParseID("invalid")
}
