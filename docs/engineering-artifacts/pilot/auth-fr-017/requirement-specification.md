---
title: "AUTH-FR-017 requirement specification"
description: "Уточнённая спецификация требования AUTH-FR-017."
keywords:
  - "M8 Platform"
  - "AUTH-FR-017"
---

# AUTH-FR-017 — Уточнённая спецификация требования {#auth-fr-017-requirement-specification}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [Pilot index](index.md) | [Requirements: AUTH-FR-017](../../../architecture/requirements/contexts/06-authentication.md#auth-fr-017) | [Traceability](../../traceability/traceability-registry.md) | [SPDD](../../spdd/index.md)

{% endnote %}

_AUTH-FR-017-SPEC · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `AUTH-FR-017-SPEC` |
| Версия | `0.1` |
| Статус | READY_FOR_REVIEW |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Authentication / m8-authentication |

# 1. Формулировка

Если refresh token отсутствует, просрочен, отозван либо признан непригодным доверенным authorization component, система MUST создать новую независимую `AuthenticationTransaction` с intent `REAUTHENTICATE`. Для Client, разрешающего CIBA, SHOULD запускаться CIBA flow; выбор окончательного challenge зависит от Client Policy и Risk Decision.

# 2. Предусловия

- Client зарегистрирован и активен.
- Вызывающая сторона аутентифицирована как разрешённый Client/service.
- Subject задан устойчивой ссылкой или разрешаемым идентификатором.
- Raw refresh token не передаётся в M8 Authentication.

# 3. Инварианты

- Новая транзакция не продолжает и не меняет failed refresh attempt.
- Один `(client_id, idempotency_key, normalized request hash)` создаёт не более одной AuthenticationTransaction.
- Повтор того же ключа с другим request hash возвращает `IDEMPOTENCY_CONFLICT`.
- Aggregate и Outbox `AuthenticationStarted` фиксируются атомарно.
- До старта provider workflow получены Client Policy, Subject resolution и обязательный Risk Decision.
- DENY завершает Operation ошибкой; CHALLENGE определяет требуемый AAL/method.
- Секреты и provider payload не попадают в domain event, audit или telemetry.

# 4. Критерии приёмки

- `AUTH-FR-017-AC-01`: новая транзакция имеет новый authentication_id и не ссылается на failed refresh как parent continuation.
- `AUTH-FR-017-AC-02`: два эквивалентных запроса с одним idempotency key возвращают одну Operation/authentication_id.
- `AUTH-FR-017-AC-03`: тот же ключ с отличающимся Subject/Client/intent отклоняется `IDEMPOTENCY_CONFLICT`.
- `AUTH-FR-017-AC-04`: commit AuthenticationTransaction без Outbox либо Outbox без aggregate невозможен.
- `AUTH-FR-017-AC-05`: DENY не вызывает Keycloak CIBA adapter.
- `AUTH-FR-017-AC-06`: dependency unavailable не приводит к silent allow.
- `AUTH-FR-017-AC-07`: audit/log/trace fixtures не содержат refresh token, OTP или provider secret.

# 5. NFR

- API p95 ≤ 300 ms до возврата Operation, без ожидания пользовательского подтверждения.
- Доступность старта ≥ 99.95% в основном регионе.
- Trace coverage ≥ 99%; обязательны request_id, correlation_id, authentication_id и operation_id.
