---
title: "SP-CONTEXT-PROV-001. Context Prompt: Provisioning"
description: "SPDD context prompt: SP-CONTEXT-PROV-001. Context Prompt: Provisioning."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# SP-CONTEXT-PROV-001. Context Prompt: Provisioning {#provisioning}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [SPDD](../index.md) | [Requirements Catalog](../../../architecture/requirements/index.md)

{% endnote %}

```yaml
prompt_id: SP-CONTEXT-PROV-001
kind: context
context: Provisioning
service: m8-provisioning
requirements_namespace: PROV-*
normative_sources:
  - PADS-000@1.0
  - M8-REQ-000@0.1
  - M8-SPDD-CONSTITUTION@1.0
```

## Миссия

Реализовывать только ответственность контекста **Provisioning**, сохраняя его ubiquitous language, инварианты, ownership и публичные контракты.

## Владеет

- ResourceDefinition
- ManagedResource
- DesiredState
- ObservedState
- Placement
- Driver
- Reconciliation

## Не владеет

- Organization hierarchy ownership
- authorization model
- provider resource as M8 domain type

## Разрешённые зависимости

- Resource Manager
- Access
- Risk Decision
- Temporal
- provider ACL
- Audit

## Запрещено

- provider SDK in domain
- secret material in desired state
- cross-service transaction

## Обязательные правила

- Domain не импортирует transport/storage/provider packages.
- Aggregate command проверяет invariants и expected revision.
- Mutation идемпотентна; обязательный integration event записывается в Outbox в той же транзакции.
- Межконтекстные типы переводятся через ports/ACL и typed references.
- Permission, risk, audit и telemetry выполняются по PADS.
- Public contract change предваряется API/Event design и compatibility review.

## Ожидаемый ответ агента

План изменения, затронутые requirements/contracts, изменённые файлы, тесты, traceability update, открытые вопросы и остаточные риски. Код без такой сводки считается неполным результатом.
