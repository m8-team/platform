package grant

import (
 "github.com/m8-team/platform/internal/access/domain/shared"
)

// Target references the independently owned resource and its authoritative organization.
// Validation of tenancy/ownership must be performed via Resource Manager at application boundary.
type Target struct {
 OrganizationID string
 Resource shared.Resource
 EnvironmentID string
}
func (t Target) Validate() error {
 if !shared.ValidUUID(t.OrganizationID) {return ErrInvalidGrant}
 if err:=t.Resource.Validate();err!=nil{return err}
 if t.EnvironmentID!=""&&!shared.ValidUUID(t.EnvironmentID){return ErrInvalidGrant}
 return nil
}
