---
title: "API-AUTH-017 contract"
description: "API contract для StartAuthentication."
keywords:
  - "M8 Platform"
  - "AUTH-FR-017"
---

# API-AUTH-017 — StartAuthentication contract {#api-auth-017}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [Pilot index](index.md) | [Requirements: AUTH-FR-017](../../../architecture/requirements/contexts/06-authentication.md#auth-fr-017) | [Traceability](../../traceability/traceability-registry.md) | [SPDD](../../spdd/index.md)

{% endnote %}

_API-AUTH-001-SPEC · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `API-AUTH-001-SPEC` |
| Версия | `0.1` |
| Статус | Проект |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | AUTH-FR-001, AUTH-FR-017, AUTH-FR-018 |

# RPC

```proto
service AuthenticationService {
  rpc StartAuthentication(StartAuthenticationRequest)
      returns (google.longrunning.Operation);
}

message StartAuthenticationRequest {
  m8.common.v1.RequestContext request_context = 1;
  string client_id = 2 [(buf.validate.field).string.min_len = 1];
  AuthenticationSubject subject = 3 [(buf.validate.field).required = true];
  AuthenticationIntent intent = 4;
  ReauthenticationReason reauthentication_reason = 5;
  AuthenticationMethod requested_method = 6;
  AssuranceLevel requested_assurance_level = 7;
  string requested_provider_id = 8;
}

enum AuthenticationIntent {
  AUTHENTICATION_INTENT_UNSPECIFIED = 0;
  LOGIN = 1;
  REAUTHENTICATE = 2;
  STEP_UP = 3;
}

enum ReauthenticationReason {
  REAUTHENTICATION_REASON_UNSPECIFIED = 0;
  REFRESH_UNAVAILABLE = 1;
  REFRESH_EXPIRED = 2;
  REFRESH_REVOKED = 3;
  SECURITY_POLICY = 4;
}

message StartAuthenticationOperationMetadata {
  string authentication_id = 1;
  AuthenticationState state = 2;
  google.protobuf.Timestamp expires_at = 3;
  AuthenticationChallengeSummary current_challenge = 4;
  m8.common.v1.OperationProgress progress = 5;
}
```

# Metadata

- `Idempotency-Key` MUST присутствовать для mutation.
- `Authorization`/service identity MUST подтверждать Client caller.
- Timeout API ограничивает только принятие команды, не CIBA lifecycle.

# Результат

Operation response после успешной аутентификации содержит типизированный `AuthenticationResult`; до завершения клиент читает/ожидает Operation или Authentication state. Raw refresh/access token в этом контракте отсутствует.

# Permission

`m8.authentication.authentication.start` на Client/Project scope.

# Compatibility

Поля 1–8 не переиспользуются. Новые intent/reason добавляются только как enum values с безопасным поведением для unknown value; изменение default semantics является breaking change.
