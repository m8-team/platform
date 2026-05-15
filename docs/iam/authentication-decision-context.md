# Authentication Decision Context

`AuthenticationDecisionContext` is an immutable normalized snapshot used by M8 Authentication, Risk Decision, Policy Engine, Challenge Selector, and delegated authentication integrations. It is safe to store for audit because it carries normalized values, hashes, masks, and references instead of raw OTPs, passwords, tokens, private keys, or full sensitive PII.

## Challenge, State, ACR, and AMR

`AuthenticationChallenge` describes what the user must do now, such as entering an OTP, approving a request, completing Mobile ID, or using WebAuthn.

`AuthenticationState` describes the lifecycle of the authentication operation, such as created, waiting for user action, verifying, authenticated, expired, or failed.

`acr` describes the requested or achieved assurance class. The proto keeps OIDC-compatible string values such as `m8:aal3` and also provides normalized `AuthenticationAssuranceLevel` values for internal policy logic.

`amr` describes the actual methods used. The proto keeps OIDC-compatible string values such as `webauthn`, `passkey`, and `user_verification`, and also provides normalized `AuthenticationMethodReference` enum values.

## Why The Context Exists

The context gives each decision component the same stable input. Risk, policy, provider selection, and audit can evaluate the same tenant, client, subject, session, transaction, device, network, runtime, OAuth, DPoP, provider, risk, policy, and evidence data without re-reading mutable state from multiple systems.

## Mobile ID Modes

Mobile ID supports SIM-push approval and SMS OTP fallback. The fallback case is represented without losing the original selected provider path:

```text
selected_challenge = AUTHENTICATION_CHALLENGE_MOBILE_ID
current_challenge = AUTHENTICATION_CHALLENGE_OTP
delivery_channel = CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SMS
```

## CIBA

CIBA is represented as a delegated OAuth flow under `OAuthContext`. The context stores `auth_req_id`, `binding_message`, delivery mode, expiration, and `callback_token_ref`. The raw callback token is not stored in the decision context.

## DPoP

DPoP is modeled as token-binding context, not as a user challenge. It records whether proof is required, the JWK thumbprint (`jkt`), proof verification status, nonce, proof JTI, and proof issued-at time.

## Step-Up Authentication

Step-up authentication raises an existing session to a stronger assurance level. For WebAuthn/passkeys, a decision can require:

```text
action = AUTHENTICATION_DECISION_ACTION_STEP_UP
selected_challenge = AUTHENTICATION_CHALLENGE_WEBAUTHN
required_acr = "m8:aal3"
expected_amr = ["webauthn", "passkey", "user_verification"]
```

## Decision Examples

Low risk sign-in can return `ALLOW` with the current session evidence.

High value operations can return `STEP_UP` with WebAuthn and `m8:aal3`.

Provider degradation can return `FALLBACK`, keeping Mobile ID as the selected challenge while switching the current challenge to OTP over `MOBILE_ID_SMS`.

Critical risk can return `DENY` or `REVIEW` with reason codes suitable for audit.
