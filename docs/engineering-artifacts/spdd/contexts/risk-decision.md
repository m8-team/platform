---
title: "SP-CONTEXT-RISK-001. Context Prompt: Risk Decision"
description: "SPDD context prompt: SP-CONTEXT-RISK-001. Context Prompt: Risk Decision."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# SP-CONTEXT-RISK-001. Context Prompt: Risk Decision {#risk-decision}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [SPDD](../index.md) | [Requirements Catalog](../../../architecture/requirements/index.md)

{% endnote %}

```yaml
prompt_id: SP-CONTEXT-RISK-001
kind: context
context: Risk Decision
service: m8-risk-decision
requirements_namespace: RISK-*
normative_sources:
  - PADS-000@1.0
  - M8-REQ-000@0.1
  - M8-SPDD-CONSTITUTION@1.0
```

## Миссия

Реализовывать только ответственность контекста **Risk Decision**, сохраняя его ubiquitous language, инварианты, ownership и публичные контракты.

## Владеет

- RiskAssessment
- RiskPolicy
- RiskSignal
- ManualReview

## Не владеет

- executing authentication challenge
- permission ownership
- device credential ownership

## Разрешённые зависимости

- Authentication context
- Access context
- external signal ACL
- Audit

## Запрещено

- raw secrets
- silent allow on mandatory dependency failure
- model internals in public errors

## Обязательные правила

- Domain не импортирует transport/storage/provider packages.
- Aggregate command проверяет invariants и expected revision.
- Mutation идемпотентна; обязательный integration event записывается в Outbox в той же транзакции.
- Межконтекстные типы переводятся через ports/ACL и typed references.
- Permission, risk, audit и telemetry выполняются по PADS.
- Public contract change предваряется API/Event design и compatibility review.

## Ожидаемый ответ агента

План изменения, затронутые requirements/contracts, изменённые файлы, тесты, traceability update, открытые вопросы и остаточные риски. Код без такой сводки считается неполным результатом.
