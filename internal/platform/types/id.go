package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const canonicalUUIDLength = 36

var (
	// ErrInvalidID indicates that an ID cannot be parsed from its external form.
	ErrInvalidID = errors.New("invalid id")
	// ErrZeroID indicates that an ID contains the reserved all-zero UUID.
	ErrZeroID = errors.New("id is zero")
)

// ID is a UUID-backed identifier value object.
//
// Its representation is intentionally private so callers cannot bypass Parse
// or NewFromUUID with a direct type conversion. The zero Go value remains
// available for optional fields and is rejected by Validate.
type ID struct {
	value uuid.UUID
}

// New returns a new random, non-zero ID.
func New() ID {
	return ID{value: uuid.New()}
}

// NewFromUUID converts a non-zero UUID to an ID.
func NewFromUUID(value uuid.UUID) (ID, error) {
	id := ID{value: value}
	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

// Parse parses a canonical hyphenated UUID into an ID.
// Letter casing is accepted and String always returns the normalized lowercase
// representation. Alternative UUID encodings such as raw hex, braces, and URN
// are rejected to keep external identifiers unambiguous.
func Parse(value string) (ID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return ID{}, fmt.Errorf("%w: %q: %w", ErrInvalidID, value, err)
	}
	if len(value) != canonicalUUIDLength || !strings.EqualFold(parsed.String(), value) {
		return ID{}, fmt.Errorf("%w: %q: expected canonical UUID", ErrInvalidID, value)
	}

	return NewFromUUID(parsed)
}

// MustParse is Parse for constants and fixtures. It panics on invalid input.
func MustParse(value string) ID {
	id, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return id
}

// UUID returns the underlying UUID value.
func (id ID) UUID() uuid.UUID {
	return id.value
}

// String returns the canonical lowercase UUID representation.
func (id ID) String() string {
	return id.value.String()
}

// IsZero reports whether id is the zero Go value.
func (id ID) IsZero() bool {
	return id.value == uuid.Nil
}

// Validate verifies that id is non-zero.
func (id ID) Validate() error {
	if id.IsZero() {
		return ErrZeroID
	}

	return nil
}

// Equal reports whether id and other contain the same UUID.
func (id ID) Equal(other ID) bool {
	return id.value == other.value
}

// MarshalText implements encoding.TextMarshaler using the canonical UUID.
func (id ID) MarshalText() ([]byte, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}

	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler using Parse. The receiver
// is changed only after the complete input has been validated.
func (id *ID) UnmarshalText(text []byte) error {
	if id == nil {
		return fmt.Errorf("%w: nil receiver", ErrInvalidID)
	}

	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}
