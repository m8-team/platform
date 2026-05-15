package authenticationdecision

import (
	"errors"
	"testing"
	"time"

	iam "github.com/m8-team/go-genproto/m8/platform/iam/v1"
	riskdecision "github.com/m8-team/go-genproto/m8/platform/riskdecision/v1"
)

func TestCloneContextRejectsNil(t *testing.T) {
	_, err := CloneContext(nil)
	if !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext, got %v", err)
	}
}

func TestCloneContextReturnsDetachedCopy(t *testing.T) {
	in := &iam.AuthenticationDecisionContext{
		Id: "ctx-1",
		Attributes: map[string]string{
			"source": "test",
		},
	}

	out, err := CloneContext(in)
	if err != nil {
		t.Fatalf("clone context: %v", err)
	}
	out.Attributes["source"] = "changed"

	if in.Attributes["source"] != "test" {
		t.Fatalf("expected source to remain test, got %q", in.Attributes["source"])
	}
}

func TestNewMobileIDSMSFallbackDecision(t *testing.T) {
	decision := NewMobileIDSMSFallbackDecision("decision-1", time.Unix(1700000000, 0).UTC())

	if decision.GetAction() != riskdecision.AuthenticationDecisionAction_AUTHENTICATION_DECISION_ACTION_FALLBACK {
		t.Fatalf("unexpected action: %v", decision.GetAction())
	}
	if decision.GetSelectedChallenge() != iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_MOBILE_ID {
		t.Fatalf("unexpected selected challenge: %v", decision.GetSelectedChallenge())
	}
	if decision.GetCurrentChallenge() != iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_OTP {
		t.Fatalf("unexpected current challenge: %v", decision.GetCurrentChallenge())
	}
	if decision.GetDeliveryChannel() != iam.ChallengeDeliveryChannel_CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SMS {
		t.Fatalf("unexpected delivery channel: %v", decision.GetDeliveryChannel())
	}
	if decision.GetRequiredAcr() != "m8:aal2" {
		t.Fatalf("unexpected required acr: %q", decision.GetRequiredAcr())
	}
}

func TestNewMobileIDSIMPushDecision(t *testing.T) {
	decision := NewMobileIDSIMPushDecision("decision-1", time.Unix(1700000000, 0).UTC())

	if decision.GetAction() != riskdecision.AuthenticationDecisionAction_AUTHENTICATION_DECISION_ACTION_CHALLENGE {
		t.Fatalf("unexpected action: %v", decision.GetAction())
	}
	if decision.GetSelectedChallenge() != iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_MOBILE_ID {
		t.Fatalf("unexpected selected challenge: %v", decision.GetSelectedChallenge())
	}
	if decision.GetCurrentChallenge() != iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_MOBILE_ID {
		t.Fatalf("unexpected current challenge: %v", decision.GetCurrentChallenge())
	}
	if decision.GetDeliveryChannel() != iam.ChallengeDeliveryChannel_CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SIM_PUSH {
		t.Fatalf("unexpected delivery channel: %v", decision.GetDeliveryChannel())
	}
	if decision.GetRequiredAcr() != "m8:aal2+" {
		t.Fatalf("unexpected required acr: %q", decision.GetRequiredAcr())
	}
}

func TestNewMobileIDProviderContext(t *testing.T) {
	provider := NewMobileIDProviderContext(
		"channel-1",
		"provider-1",
		"provider-transaction-1",
		iam.MobileIdMode_MOBILE_ID_MODE_SMS_OTP,
	)

	if provider.GetProviderType() != iam.ProviderType_PROVIDER_TYPE_MOBILE_ID {
		t.Fatalf("unexpected provider type: %v", provider.GetProviderType())
	}
	if provider.GetMobileIdMode() != iam.MobileIdMode_MOBILE_ID_MODE_SMS_OTP {
		t.Fatalf("unexpected mobile id mode: %v", provider.GetMobileIdMode())
	}
}

func TestNewWebAuthnStepUpDecision(t *testing.T) {
	decision := NewWebAuthnStepUpDecision("decision-1", time.Unix(1700000000, 0).UTC())

	if decision.GetAction() != riskdecision.AuthenticationDecisionAction_AUTHENTICATION_DECISION_ACTION_STEP_UP {
		t.Fatalf("unexpected action: %v", decision.GetAction())
	}
	if decision.GetSelectedChallenge() != iam.AuthenticationChallenge_AUTHENTICATION_CHALLENGE_WEBAUTHN {
		t.Fatalf("unexpected selected challenge: %v", decision.GetSelectedChallenge())
	}
	if decision.GetRequiredAcr() != "m8:aal3" {
		t.Fatalf("unexpected required acr: %q", decision.GetRequiredAcr())
	}

	expectedAMR := []string{"webauthn", "passkey", "user_verification"}
	if got := decision.GetExpectedAmr(); len(got) != len(expectedAMR) {
		t.Fatalf("unexpected expected amr length: %d", len(got))
	}
	for i, want := range expectedAMR {
		if decision.GetExpectedAmr()[i] != want {
			t.Fatalf("expected amr[%d] = %q, got %q", i, want, decision.GetExpectedAmr()[i])
		}
	}
}
