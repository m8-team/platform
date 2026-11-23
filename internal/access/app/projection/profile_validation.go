package projection

import (
	"github.com/m8-team/platform/internal/access/domain/role"
)

// ValidateProfileCapabilities verifies a staged assignment's proposed graph
// tuple set without making it effective. The authoritative write and Policy
// mutation must still be committed together in one YDB transaction.
func ValidateProfileCapabilities(pending role.ProfileAssignment,caps []ProfileCapability)error{
	if pending.Validate()!=nil||pending.Status()!=role.ProfileAssignmentPending{return ErrInvalidProfileProjection}
	active,err:=pending.Activate(pending.Revision())
	if err!=nil{return err}
	return (ProfileAssignmentState{Assignment:active,Capabilities:caps}).Validate()
}
