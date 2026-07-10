---
title: "AUTH-FR-017 acceptance tests"
description: "Acceptance test specification для AUTH-FR-017."
keywords:
  - "M8 Platform"
  - "AUTH-FR-017"
---

# AUTH-FR-017 Acceptance Test Specification {#auth-fr-017-acceptance-tests}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [Pilot index](index.md) | [Requirements: AUTH-FR-017](../../../architecture/requirements/contexts/06-authentication.md#auth-fr-017) | [Traceability](../../traceability/traceability-registry.md) | [SPDD](../../spdd/index.md)

{% endnote %}

_AT-AUTH-017 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `AT-AUTH-017` |
| Версия | `0.1` |
| Статус | Проект |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Executable behavior and evidence |

# Сценарии

```gherkin
Feature: Reauthentication after refresh cannot be used

  Scenario: Create a new independent transaction
    Given an active Client allowed to use CIBA
    And a resolvable Subject
    And Risk Decision returns ALLOW
    When StartAuthentication is called with intent REAUTHENTICATE and reason REFRESH_UNAVAILABLE
    Then a new authentication_id is returned in Operation metadata
    And AuthenticationTransaction and AuthenticationStarted Outbox record are committed atomically
    And no failed refresh token or provider secret is stored

  Scenario: Same idempotency key returns the same operation
    Given the first request has committed successfully
    When the equivalent request is repeated with the same Idempotency-Key
    Then the same Operation name and authentication_id are returned
    And no second aggregate or Outbox event is created

  Scenario: Idempotency conflict
    Given an Idempotency-Key was used for Subject A
    When it is repeated for Subject B
    Then IDEMPOTENCY_CONFLICT is returned

  Scenario: Risk denial
    Given Risk Decision returns DENY
    When StartAuthentication is executed
    Then Keycloak CIBA is not called
    And an audit outcome DENIED is recorded

  Scenario: Provider temporarily unavailable after commit
    Given the command was committed and Operation returned
    And Keycloak is unavailable
    When Temporal executes provider start
    Then workflow retries according to policy
    And Operation remains observable
    And no duplicate AuthenticationStarted event is emitted
```

# Нагрузочные и fault tests

- 100 concurrent identical requests create one aggregate.
- Broker unavailable after commit: Outbox backlog is later published once logically.
- Workflow replay produces no duplicate provider side effect beyond provider idempotency contract.
- Secret scanner over logs/traces/audit fixtures finds zero raw secrets.
