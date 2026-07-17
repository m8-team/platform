---
title: "SP-CONTEXT-AUTH-001. Context Prompt: Authentication"
description: "SPDD context prompt: SP-CONTEXT-AUTH-001. Context Prompt: Authentication."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# SP-CONTEXT-AUTH-001. Context Prompt: Authentication {#authentication}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [SPDD](../index.md) | [Requirements Catalog](../../../architecture/requirements/index.md)

{% endnote %}

```yaml
prompt_id: SP-CONTEXT-AUTH-001
kind: context
context: Authentication
service: m8-authentication
requirements_namespace: AUTH-*
normative_sources:
  - PADS-000@1.0
  - M8-REQ-000@0.1
  - M8-SPDD-CONSTITUTION@1.0
```

## Миссия

Реализовывать только ответственность контекста **Authentication**, сохраняя его ubiquitous language, инварианты, ownership и публичные контракты.

## Владеет

- Client
- AuthenticationTransaction
- AuthenticationChallenge
- Handoff
- SessionReference

## Не владеет

- user profile ownership
- permission model ownership
- risk policy ownership

## Разрешённые зависимости

- Identity
- Access
- Risk Decision
- Keycloak ACL
- Audit

## Запрещено

- Identity DB
- SpiceDB writes
- Keycloak types in domain
- raw token/OTP in telemetry

## Обязательные правила

- Domain не импортирует transport/storage/provider packages.
- Aggregate command проверяет invariants и expected revision.
- Mutation идемпотентна; обязательный integration event записывается в Outbox в той же транзакции.
- Межконтекстные типы переводятся через ports/ACL и typed references.
- Permission, risk, audit и telemetry выполняются по PADS.
- Public contract change предваряется API/Event design и compatibility review.

## Ожидаемый ответ агента

План изменения, затронутые requirements/contracts, изменённые файлы, тесты, traceability update, открытые вопросы и остаточные риски. Код без такой сводки считается неполным результатом.
