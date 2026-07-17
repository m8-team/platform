---
title: "SP-CONTEXT-ACC-001. Context Prompt: Access"
description: "SPDD context prompt: SP-CONTEXT-ACC-001. Context Prompt: Access."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# SP-CONTEXT-ACC-001. Context Prompt: Access {#access}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [SPDD](../index.md) | [Requirements Catalog](../../../architecture/requirements/index.md)

{% endnote %}

```yaml
prompt_id: SP-CONTEXT-ACC-001
kind: context
context: Access
service: m8-access
requirements_namespace: ACC-*
normative_sources:
  - PADS-000@1.0
  - M8-REQ-000@0.1
  - M8-SPDD-CONSTITUTION@1.0
```

## Миссия

Реализовывать только ответственность контекста **Access**, сохраняя его ubiquitous language, инварианты, ownership и публичные контракты.

## Владеет

- PermissionDefinition
- Role
- RoleBinding
- AccessRelationship
- AuthorizationModel

## Не владеет

- user lifecycle
- authentication
- risk assessment

## Разрешённые зависимости

- Resource Manager facts
- Identity facts
- SpiceDB ACL
- Audit

## Запрещено

- Resource Manager DB
- Identity DB
- business data replication

## Обязательные правила

- Domain не импортирует transport/storage/provider packages.
- Aggregate command проверяет invariants и expected revision.
- Mutation идемпотентна; обязательный integration event записывается в Outbox в той же транзакции.
- Межконтекстные типы переводятся через ports/ACL и typed references.
- Permission, risk, audit и telemetry выполняются по PADS.
- Public contract change предваряется API/Event design и compatibility review.

## Ожидаемый ответ агента

План изменения, затронутые requirements/contracts, изменённые файлы, тесты, traceability update, открытые вопросы и остаточные риски. Код без такой сводки считается неполным результатом.
