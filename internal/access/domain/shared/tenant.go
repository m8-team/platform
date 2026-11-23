// Package shared contains immutable authorization identifiers shared by Access bounded contexts.
package shared

import (
 "errors"
 "regexp"
 "strings"
)

var ErrInvalidTenantID = errors.New("invalid tenant id")
var canonicalUUID = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

// TenantID is a trusted authorization partition, never a substitute for authentication.
type TenantID string

func NewTenantID(raw string) (TenantID, error) {
 id := TenantID(strings.TrimSpace(raw))
 return id, id.Validate()
}
func (id TenantID) Validate() error {
 if !canonicalUUID.MatchString(string(id)) { return ErrInvalidTenantID }
 return nil
}
func (id TenantID) String() string { return string(id) }

// ValidUUID is used for aggregate identifiers without coupling domain entities to transport.
func ValidUUID(id string) bool { return canonicalUUID.MatchString(id) }
