---
title: "SP-CONTEXT-AUD-001. Context Prompt: Audit"
description: "SPDD context prompt: SP-CONTEXT-AUD-001. Context Prompt: Audit."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# SP-CONTEXT-AUD-001. Context Prompt: Audit {#audit}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [SPDD](../index.md) | [Requirements Catalog](../../../architecture/requirements/index.md)

{% endnote %}

```yaml
prompt_id: SP-CONTEXT-AUD-001
kind: context
context: Audit
service: m8-audit
requirements_namespace: AUD-*
normative_sources:
  - PADS-000@1.0
  - M8-REQ-000@0.1
  - M8-SPDD-CONSTITUTION@1.0
```

## Миссия

Реализовывать только ответственность контекста **Audit**, сохраняя его ubiquitous language, инварианты, ownership и публичные контракты.

## Владеет

- AuditEvent
- AuditExport
- RetentionPolicy
- LegalHold
- IntegrityProof

## Не владеет

- business state mutation
- authorization policy ownership
- raw secrets

## Разрешённые зависимости

- all producers
- Access
- immutable storage

## Запрещено

- mutating source aggregates
- untrusted actor override
- PII beyond approved audit schema

## Обязательные правила

- Domain не импортирует transport/storage/provider packages.
- Aggregate command проверяет invariants и expected revision.
- Mutation идемпотентна; обязательный integration event записывается в Outbox в той же транзакции.
- Межконтекстные типы переводятся через ports/ACL и typed references.
- Permission, risk, audit и telemetry выполняются по PADS.
- Public contract change предваряется API/Event design и compatibility review.

## Ожидаемый ответ агента

План изменения, затронутые requirements/contracts, изменённые файлы, тесты, traceability update, открытые вопросы и остаточные риски. Код без такой сводки считается неполным результатом.
