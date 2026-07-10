---
title: "AUTH-FR-017 error contract"
description: "Error contract пилота AUTH-FR-017."
keywords:
  - "M8 Platform"
  - "AUTH-FR-017"
---

# AUTH-FR-017 — Error Contract {#auth-fr-017-error-contract}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [Pilot index](index.md) | [Requirements: AUTH-FR-017](../../../architecture/requirements/contexts/06-authentication.md#auth-fr-017) | [Traceability](../../traceability/traceability-registry.md) | [SPDD](../../spdd/index.md)

{% endnote %}

_AUTH-FR-017-ERRORS · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `AUTH-FR-017-ERRORS` |
| Версия | `0.1` |
| Статус | Проект |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Canonical errors StartAuthentication |

| Code | Когда | Transport | Retry |
| --- | --- | --- | --- |
| CLIENT_NOT_FOUND | Client отсутствует или не раскрывается | NOT_FOUND | нет |
| CLIENT_DISABLED | Client отключён | FAILED_PRECONDITION | нет |
| FLOW_NOT_ALLOWED | REAUTHENTICATE/CIBA запрещён policy | PERMISSION_DENIED | нет |
| SUBJECT_NOT_FOUND | Subject не разрешён | NOT_FOUND | нет |
| POLICY_DENIED | Risk Decision = DENY | PERMISSION_DENIED | нет |
| STEP_UP_REQUIRED | Нужен иной challenge/AAL | FAILED_PRECONDITION | после challenge |
| RISK_DECISION_UNAVAILABLE | Обязательный Risk недоступен | UNAVAILABLE | да |
| PROVIDER_UNAVAILABLE | CIBA provider недоступен после старта | UNAVAILABLE/Operation error | да |
| IDEMPOTENCY_CONFLICT | Ключ использован с другим request hash | ALREADY_EXISTS | нет |
| RATE_LIMITED | Превышен лимит стартов | RESOURCE_EXHAUSTED | да после retry_after |

Provider error body, stack trace и внутренний Keycloak ID клиенту не возвращаются.
