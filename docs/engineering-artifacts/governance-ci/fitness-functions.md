---
title: "M8 Architecture Fitness Functions"
description: "Architecture fitness functions for M8 Platform."
keywords:
  - "M8 Platform"
  - "governance CI"
---

# M8 Architecture Fitness Functions {#architecture-fitness-functions}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Governance CI](index.md) | [CI checks](ci-checks.md)

{% endnote %}

_M8-FITNESS-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-FITNESS-000` |
| Версия | `0.1` |
| Статус | Базовая нормативная редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Automated architecture, contract, security and traceability checks |

# Реестр

| ID | Правило | Gate | Механизм | Критерий |
| --- | --- | --- | --- | --- |
| `FF-001` | No cross-context database access | block | go/import + configuration scan | Сервис не подключается к таблицам/credentials другого owner. |
| `FF-002` | Domain dependency direction | block | go list / depguard | domain не импортирует application, transport, storage, provider SDK. |
| `FF-003` | Protobuf lint and breaking | block | buf lint + buf breaking | Публичные API совместимы с baseline. |
| `FF-004` | Event schema compatibility | block | schema registry check | Event type/version/fields/partition key совместимы. |
| `FF-005` | Requirement ID validity | block | artifact validator | Все requirement refs существуют. |
| `FF-006` | Traceability completeness | block for implementing | artifact validator | Implementation имеет prompt/test/contract links. |
| `FF-007` | Outbox atomicity | block | integration test | Aggregate и обязательный Outbox фиксируются атомарно. |
| `FF-008` | Consumer idempotency | block | integration/fault test | Duplicate event не создаёт второй эффект. |
| `FF-009` | Mutation idempotency | block | concurrency test | Одинаковый ключ создаёт один logical result. |
| `FF-010` | Secret leakage | block | gitleaks + fixture scan | Secrets отсутствуют в code/log/event/audit fixtures. |
| `FF-011` | Permission annotation | block | proto/custom lint | Каждый public protected RPC имеет permission. |
| `FF-012` | Risk gate annotation | block | catalog lint | Sensitive action имеет risk policy link. |
| `FF-013` | Audit coverage | block | test/static annotation | Significant mutation формирует AuditEvent. |
| `FF-014` | Telemetry identifiers | warn/block critical | integration test | request/correlation/operation/resource IDs проходят end-to-end. |
| `FF-015` | LRO contract | block | proto lint | Long-running RPC имеет typed metadata/result/cancel semantics. |
| `FF-016` | YDB migration ownership | block | migration path lint | Migration находится в owner service и backward compatible. |
| `FF-017` | ADR required for boundary change | block | diff policy | Новая service dependency/store/technology имеет ADR. |
| `FF-018` | PII classification | block | data registry lint | Новое persisted field имеет classification/retention/deletion. |
| `FF-019` | SLO metadata | warn/block critical | service catalog lint | Critical API имеет SLI/SLO/error budget. |
| `FF-020` | Prompt scope | block | SPDD schema lint | Task Prompt имеет allowed/forbidden paths и один owner context. |

# Рекомендуемый pipeline

```text
artifact-validate → gofmt/lint/unit → architecture-imports
→ buf lint/breaking → schema compatibility → integration/fault tests
→ security/secret scan → acceptance → traceability/release evidence
```

Pull Request не может быть объединён при blocking finding. Временное исключение имеет owner, expiry, issue и ADR/waiver ID.
