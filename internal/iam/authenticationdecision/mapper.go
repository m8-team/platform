package authenticationdecision

import (
	"errors"
	"time"

	iam "github.com/m8-team/go-genproto/m8/platform/iam/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// ErrNilContext is returned when a nil decision context is passed to the mapper boundary.
	ErrNilContext = errors.New("authentication decision context is nil")

	// ErrNilDecision is returned when a nil decision is passed to the mapper boundary.
	ErrNilDecision = errors.New("authentication decision is nil")
)

// Mapper is the boundary for converting decision context and decision values.
//
// The current scaffold clones generated protobuf messages so callers do not
// share mutable decision snapshots across risk, policy, provider, and audit code.
type Mapper struct{}

// NewMapper returns a mapper for authentication decision context values.
func NewMapper() Mapper {
	return Mapper{}
}

// FromProtoContext returns a detached copy of a generated decision context.
func (Mapper) FromProtoContext(in *iam.AuthenticationDecisionContext) (*iam.AuthenticationDecisionContext, error) {
	return CloneContext(in)
}

// ToProtoContext returns a detached copy ready to pass to generated protobuf APIs.
func (Mapper) ToProtoContext(in *iam.AuthenticationDecisionContext) (*iam.AuthenticationDecisionContext, error) {
	return CloneContext(in)
}

// FromProtoDecision returns a detached copy of a generated authentication decision.
func (Mapper) FromProtoDecision(in *iam.AuthenticationDecision) (*iam.AuthenticationDecision, error) {
	return CloneDecision(in)
}

// ToProtoDecision returns a detached copy ready to pass to generated protobuf APIs.
func (Mapper) ToProtoDecision(in *iam.AuthenticationDecision) (*iam.AuthenticationDecision, error) {
	return CloneDecision(in)
}

// CloneContext returns a deep copy of an authentication decision context.
func CloneContext(in *iam.AuthenticationDecisionContext) (*iam.AuthenticationDecisionContext, error) {
	if in == nil {
		return nil, ErrNilContext
	}
	return proto.Clone(in).(*iam.AuthenticationDecisionContext), nil
}

// CloneDecision returns a deep copy of an authentication decision.
func CloneDecision(in *iam.AuthenticationDecision) (*iam.AuthenticationDecision, error) {
	if in == nil {
		return nil, ErrNilDecision
	}
	return proto.Clone(in).(*iam.AuthenticationDecision), nil
}

// NewMobileIDSMSFallbackDecision builds the canonical Mobile ID SMS OTP fallback decision.
func NewMobileIDSMSFallbackDecision(decisionID string, decisionTime time.Time) *iam.AuthenticationDecision {
	return &iam.AuthenticationDecision{
		DecisionId:        decisionID,
		Action:            iam.AuthenticationDecisionAction_AUTHENTICATION_DECISION_ACTION_FALLBACK,
		SelectedChallenge: iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_MOBILE_ID,
		CurrentChallenge:  iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_OTP,
		DeliveryChannel:   iam.ChallengeDeliveryChannel_CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SMS,
		ExpectedAmr:       []string{"mobile_id", "otp", "sms"},
		ExpectedMethodReferences: []iam.AuthenticationMethodReference{
			iam.AuthenticationMethodReference_AUTHENTICATION_METHOD_REFERENCE_MOBILE_ID,
			iam.AuthenticationMethodReference_AUTHENTICATION_METHOD_REFERENCE_OTP,
			iam.AuthenticationMethodReference_AUTHENTICATION_METHOD_REFERENCE_SMS,
		},
		Reasons:      []string{"mobile_id_sms_fallback"},
		DecisionTime: timestamppb.New(decisionTime),
	}
}

// NewWebAuthnStepUpDecision builds the canonical WebAuthn/passkey step-up decision.
func NewWebAuthnStepUpDecision(decisionID string, decisionTime time.Time) *iam.AuthenticationDecision {
	return &iam.AuthenticationDecision{
		DecisionId:        decisionID,
		Action:            iam.AuthenticationDecisionAction_AUTHENTICATION_DECISION_ACTION_STEP_UP,
		SelectedChallenge: iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_WEBAUTHN,
		CurrentChallenge:  iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_WEBAUTHN,
		DeliveryChannel:   iam.ChallengeDeliveryChannel_CHALLENGE_DELIVERY_CHANNEL_WEBAUTHN,
		RequiredAcr:       "m8:aal3",
		ExpectedAmr:       []string{"webauthn", "passkey", "user_verification"},
		ExpectedMethodReferences: []iam.AuthenticationMethodReference{
			iam.AuthenticationMethodReference_AUTHENTICATION_METHOD_REFERENCE_WEBAUTHN,
			iam.AuthenticationMethodReference_AUTHENTICATION_METHOD_REFERENCE_PASSKEY,
			iam.AuthenticationMethodReference_AUTHENTICATION_METHOD_REFERENCE_USER_VERIFICATION,
		},
		Reasons:      []string{"webauthn_step_up_required"},
		DecisionTime: timestamppb.New(decisionTime),
	}
}
