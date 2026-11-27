// Package shared contains immutable authorization identifiers shared by Access bounded contexts.
package shared

import (
	"errors"
	"strings"
	"uuid"
)

var ErrInvalidTenantID = errors.New("invalid tenant id")

// TenantID is a trusted authorization partition, never a substitute for authentication.
type TenantID struct {
	value uuid.UUID
	set   bool // The all-zero UUID is a valid, explicitly supplied ID.
}

func NewTenantID(raw string) (TenantID, error) {
	value, ok := parseCanonicalUUID(strings.TrimSpace(raw))
	if !ok {
		return TenantID{}, ErrInvalidTenantID
	}
	return TenantID{value: value, set: true}, nil
}

func (id TenantID) Validate() error {
	if !id.set {
		return ErrInvalidTenantID
	}
	return nil
}

func (id TenantID) String() string {
	if !id.set {
		return ""
	}
	return id.value.String()
}

func (id TenantID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *TenantID) UnmarshalText(text []byte) error {
	value, err := NewTenantID(string(text))
	if err == nil {
		*id = value
	}
	return err
}

// ValidUUID is used for aggregate identifiers without coupling domain entities to transport.
func ValidUUID(id string) bool {
	_, ok := parseCanonicalUUID(id)
	return ok
}

func parseCanonicalUUID(raw string) (uuid.UUID, bool) {
	value, err := uuid.Parse(raw)
	return value, err == nil && value.String() == raw
}
