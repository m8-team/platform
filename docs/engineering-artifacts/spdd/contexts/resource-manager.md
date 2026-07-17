---
title: "SP-CONTEXT-RM-001. Context Prompt: Resource Manager"
description: "SPDD context prompt: SP-CONTEXT-RM-001. Context Prompt: Resource Manager."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# SP-CONTEXT-RM-001. Context Prompt: Resource Manager {#resource-manager}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [SPDD](../index.md) | [Requirements Catalog](../../../architecture/requirements/index.md)

{% endnote %}

```yaml
prompt_id: SP-CONTEXT-RM-001
kind: context
context: Resource Manager
service: m8-resource-manager
requirements_namespace: RM-*
normative_sources:
  - PADS-000@1.0
  - M8-REQ-000@0.1
  - M8-SPDD-CONSTITUTION@1.0
```

## Миссия

Реализовывать только ответственность контекста **Resource Manager**, сохраняя его ubiquitous language, инварианты, ownership и публичные контракты.

## Владеет

- Organization
- Workspace
- Project
- ServiceRegistration
- resource hierarchy

## Не владеет

- users
- authentication
- permissions
- cloud resource provisioning

## Разрешённые зависимости

- Access
- Audit
- Provisioning events

## Запрещено

- Identity database
- SpiceDB direct writes
- provider SDK in domain

## Обязательные правила

- Domain не импортирует transport/storage/provider packages.
- Aggregate command проверяет invariants и expected revision.
- Mutation идемпотентна; обязательный integration event записывается в Outbox в той же транзакции.
- Межконтекстные типы переводятся через ports/ACL и typed references.
- Permission, risk, audit и telemetry выполняются по PADS.
- Public contract change предваряется API/Event design и compatibility review.

## Ожидаемый ответ агента

План изменения, затронутые requirements/contracts, изменённые файлы, тесты, traceability update, открытые вопросы и остаточные риски. Код без такой сводки считается неполным результатом.
