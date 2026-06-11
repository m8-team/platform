# Authentication Decision Context

`AuthenticationDecisionContext` is a normalized, immutable decision snapshot created by M8 Authentication and consumed by M8 Risk Decision, Policy Engine, and Challenge Selector.

The authentication protobuf package owns only authentication workflow data: lifecycle, challenges, assurance, provider execution context, CIBA callback context, DPoP token-binding context, and safe references to Identity, Resource Manager, Access, and Audit data. Risk scores, risk signals, policy evaluation, and `AuthenticationDecision` results are owned by the separate `m8.platform.riskdecision.v1` package.

The context is safe for policy evaluation and audit because it carries normalized values, masks, hashes, and references instead of raw OTPs, raw passwords, OAuth tokens, private keys, or unnecessary full PII.

## Platform Module Boundaries

M8 Authentication owns authentication operations, lifecycle state, challenge orchestration, Keycloak CIBA callback integration, and `AuthenticationEvidence` creation after successful verification.

M8 Identity owns users, identities, user pools, identity links, authenticators, verified phone/email/username bindings, federated subject links, Mobile ID subjects, and WebAuthn credential metadata.

M8 Resource Manager owns resource hierarchy and configuration references: organizations, workspaces, projects, applications, user pools, clients, and provider configuration references.

M8 Access owns permissions, roles, relationships, and business authorization checks. Authentication can ask Access whether a sensitive auth-related operation may start, but Access decides what the user can do.

M8 Risk Decision owns adaptive security decisions, risk signals, authentication policy evaluation, and `AuthenticationDecision` results. Authentication builds the context and executes the selected challenge; Risk Decision owns adaptive risk and policy logic.

M8 Audit owns immutable audit history. Authentication emits audit events, but it does not own long-term audit storage.

## Authentication vs Identity vs Access

Authentication proves who the subject is by executing challenges such as OTP, approval, Mobile ID, WebAuthn, OIDC, SAML, or password.

Identity resolves and owns who the subject is in the identity system. `SubjectContext` may contain subject ids, masked values, hashes, status, or references, but Identity remains the source of truth.

Access decides what the authenticated subject can do. `TransactionContext` and `ResourceRef` can describe a protected operation for assurance and risk decisions, but they must not become an authorization model inside Authentication.

## Challenge vs State vs acr vs amr

`AuthenticationChallenge` describes what the user must do now.

`AuthenticationState` describes the lifecycle of the authentication operation.

`acr` describes the requested or achieved authentication assurance class, such as `m8:aal1`, `m8:aal2`, `m8:aal2+`, or `m8:aal3`.

`amr` describes the actual authentication methods used, such as `otp`, `sms`, `push`, `approval`, `mobile_id`, `sim_push`, `webauthn`, `passkey`, or `user_verification`.

These concepts must not be mixed with permissions or risk. For example, `AuthenticationChallenge = OTP`, `AuthenticationState = WAITING_FOR_USER`, `acr = m8:aal2`, and `amr = ["otp", "sms"]` describe authentication only. Access still decides whether the user may delete an organization.

## Risk Decision Role

Risk Decision evaluates device, network, velocity, behavior, provider health, operation sensitivity, and risk signals. Its contracts live in `m8.platform.riskdecision.v1`, not in the authentication package. It returns actions such as `ALLOW`, `DENY`, `CHALLENGE`, `STEP_UP`, `FALLBACK`, or `REVIEW`.

Authentication does not embed adaptive risk rules or store risk-owned fields in `AuthenticationDecisionContext`. It sends the authentication snapshot to Risk Decision and then executes the returned decision.

## Audit Role

Authentication emits audit events for lifecycle transitions and challenge results. M8 Audit owns immutable event storage and security event history.

Audit-safe values use masked, hashed, or referenced data. Raw OTPs, raw passwords, access tokens, refresh tokens, private keys, and unnecessary full PII must not be stored in the context or audit events.

## Keycloak CIBA Role

Keycloak is the OAuth2/OIDC authorization server. M8 Authentication is the external authentication channel and orchestrator.

CIBA is represented as delegated OAuth context:

```text
oauth.flow = OAUTH_FLOW_CIBA
oauth.ciba.auth_req_id = "<auth request id>"
oauth.ciba.binding_message = "<message shown to user>"
oauth.ciba.callback_token_ref = "<stored callback token reference>"
```

M8 Authentication proves identity and calls the Keycloak CIBA callback. Keycloak issues `access_token`, `refresh_token`, and `id_token`; Authentication does not issue or store those raw tokens.

## DPoP Role

DPoP is token binding, not a user challenge. It is represented in `DpopContext`:

```text
dpop.required = true
dpop.jkt = "<jwk thumbprint>"
dpop.proof_verified = true
dpop.nonce = "<nonce>"
dpop.proof_jti = "<proof id>"
dpop.proof_iat = "<issued-at time>"
```

## Mobile ID SIM-Push And SMS OTP Fallback

Mobile ID SIM-push keeps Mobile ID as both the selected and current challenge:

```text
selected_challenge = AUTHENTICATION_CHALLENGE_MOBILE_ID
current_challenge = AUTHENTICATION_CHALLENGE_MOBILE_ID
delivery_channel = CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SIM_PUSH
provider_type = PROVIDER_TYPE_MOBILE_ID
mobile_id_mode = MOBILE_ID_MODE_SIM_PUSH
expected_amr = ["mobile_id", "sim_push", "approval"]
required_acr = "m8:aal2+"
```

If the operator cannot send SIM-push, Mobile ID may fall back to SMS OTP without losing the selected Mobile ID path:

```text
selected_challenge = AUTHENTICATION_CHALLENGE_MOBILE_ID
current_challenge = AUTHENTICATION_CHALLENGE_OTP
delivery_channel = CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SMS
provider_type = PROVIDER_TYPE_MOBILE_ID
mobile_id_mode = MOBILE_ID_MODE_SMS_OTP
expected_amr = ["mobile_id", "otp", "sms"]
required_acr = "m8:aal2"
```

## WebAuthn/Passkey Step-Up

For sensitive operations, Risk Decision or assurance policy can require WebAuthn/passkey step-up:

```text
action = AUTHENTICATION_DECISION_ACTION_STEP_UP
selected_challenge = AUTHENTICATION_CHALLENGE_WEBAUTHN
current_challenge = AUTHENTICATION_CHALLENGE_WEBAUTHN
delivery_channel = CHALLENGE_DELIVERY_CHANNEL_WEBAUTHN
required_acr = "m8:aal3"
expected_amr = ["webauthn", "passkey", "user_verification"]
```

## Example Decision Contexts

Low-risk sign-in can return `ALLOW` with existing session evidence if `max_age` and assurance requirements are satisfied.

Passwordless OTP sign-in can return `CHALLENGE` with `selected_challenge = AUTHENTICATION_CHALLENGE_OTP`, `current_challenge = AUTHENTICATION_CHALLENGE_OTP`, and `expected_amr = ["otp", "sms"]` or another delivery channel.

Mobile ID provider degradation can return `FALLBACK`, keeping `selected_challenge = AUTHENTICATION_CHALLENGE_MOBILE_ID` while switching `current_challenge` to OTP over `CHALLENGE_DELIVERY_CHANNEL_MOBILE_ID_SMS`.

High-risk privileged operations can return `STEP_UP` with WebAuthn and `required_acr = "m8:aal3"`.

Critical risk can return `DENY` or `REVIEW` with reason codes suitable for audit and user-safe explanations.
