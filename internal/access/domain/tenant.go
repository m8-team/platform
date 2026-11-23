package domain

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidTenantID = errors.New("invalid tenant id")
var canonicalUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// TenantID is an explicit authorization partition. It must come from a trusted
// tenant context; parsing a client-supplied value does not authenticate it.
type TenantID string

func NewTenantID(value string) (TenantID, error) {
	id := TenantID(strings.TrimSpace(value))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func (t TenantID) Validate() error {
	if !canonicalUUIDPattern.MatchString(string(t)) {
		return ErrInvalidTenantID
	}
	return nil
}
