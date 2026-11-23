package grant

import "regexp"

var policyRefPattern = regexp.MustCompile("^[a-z][a-z0-9_]{0,127}$")

// Constraints are requirements to prove from trusted context at decision time.
// They MUST NOT be inferred from an ACTIVE status or caller-provided claims.
type Constraints struct {
	RequireMFA              bool
	RequireActiveCase       bool
	RequireActiveAssignment bool
	ServiceAgreementID      string
	ConditionPolicyRef      string
}

func (c Constraints) Validate() error {
	if c.ConditionPolicyRef != "" && !policyRefPattern.MatchString(c.ConditionPolicyRef) {
		return ErrInvalidGrant
	}
	if len(c.ServiceAgreementID) > 256 {
		return ErrInvalidGrant
	}
	return nil
}
func (c Constraints) RequiresContext() bool {
	return c.RequireMFA || c.RequireActiveCase || c.RequireActiveAssignment || c.ServiceAgreementID != "" || c.ConditionPolicyRef != ""
}
