---
title: "AUTH-FR-017 event contracts"
description: "Event contracts пилота AUTH-FR-017."
keywords:
  - "M8 Platform"
  - "AUTH-FR-017"
---

# AUTH-FR-017 — Event Contracts {#auth-fr-017-event-contracts}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [Pilot index](index.md) | [Requirements: AUTH-FR-017](../../../architecture/requirements/contexts/06-authentication.md#auth-fr-017) | [Traceability](../../traceability/traceability-registry.md) | [SPDD](../../spdd/index.md)

{% endnote %}

_AUTH-FR-017-EVENTS · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `AUTH-FR-017-EVENTS` |
| Версия | `0.1` |
| Статус | Проект |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | AuthenticationStarted и последующие lifecycle facts |

# EVT-AUTH-001 — AuthenticationStarted.v1

```yaml
event_type: m8.authentication.authentication_started.v1
producer: m8-authentication
topic: m8.authentication.events.v1
partition_key: authentication_id
ordering_scope: authentication_id
payload:
  authentication_id: string
  client_id: string
  subject_ref: minimal typed reference
  intent: REAUTHENTICATE
  reason: REFRESH_UNAVAILABLE|REFRESH_EXPIRED|REFRESH_REVOKED|SECURITY_POLICY
  requested_method: CIBA|...
  requested_assurance_level: AAL1|AAL2|AAL3
  state: CREATED|CHALLENGE_PENDING
  expires_at: timestamp
  operation_name: string
  aggregate_revision: uint64
```

Событие не содержит refresh token, OTP, Keycloak auth_req_id, IP в открытом виде или профиль пользователя.

# Дополнительные события процесса

- `AuthenticationChallengeCreated.v1` — challenge создан после provider start.
- `AuthenticationFailed.v1` — terminal failure с безопасным reason category.
- `AuthenticationCompleted.v1` — achieved AAL и result reference без токена.

# Delivery

At-least-once; consumer deduplicates by event_id and ignores revision lower/equal to applied revision. Outbox and aggregate commit are atomic.
