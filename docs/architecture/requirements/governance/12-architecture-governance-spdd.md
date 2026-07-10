---
title: "Requirements Catalog: архитектурное управление и SPDD"
description: "Требования к архитектурному управлению и SPDD."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 12. Архитектурное управление и SPDD {#requirements-architecture-governance-spdd}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 11. Common Operation](../contexts/11-common-operation.md) | [Следующий раздел: 13. Сквозные сценарии и декомпозиция требований](13-cross-cutting-scenarios.md)

{% endnote %}

Владелец раздела: **Architecture Governance**. Требований: **3**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `PLT-GOV-001` | governance | Must | `CAP-GOV-04` | Трассировка требования | `ANALYZED` |
| `PLT-GOV-002` | governance | Must | `CAP-GOV-09` | Structured Prompt для реализации | `ANALYZED` |
| `PLT-GOV-003` | governance | Must | `CAP-GOV-05` | Контрактный design gate | `ANALYZED` |

## Детальные требования

### PLT-GOV-001. Трассировка требования {#plt-gov-001}

| Поле | Значение |
| --- | --- |
| Тип | `governance` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Architecture Governance` / `repository/CI` |
| Business capability | `CAP-GOV-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §20, §21, §23 |

**Требование.**

Каждое approved требование должно быть связано минимум с capability, owner context, acceptance criteria, contract impact, implementation и tests.

**Критерии приёмки.**

- `PLT-GOV-001-AC-01` — CI выявляет отсутствующие или битые ссылки.
- `PLT-GOV-001-AC-02` — Release scope не включает requirement без verification evidence.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-GOV-001
capability: CAP-GOV-04
owner_context: Architecture Governance
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### PLT-GOV-002. Structured Prompt для реализации {#plt-gov-002}

| Поле | Значение |
| --- | --- |
| Тип | `governance` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Architecture Governance` / `SPDD tooling` |
| Business capability | `CAP-GOV-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §22 |

**Требование.**

Реализация требования с использованием ИИ-агента должна выполняться по versioned Structured Prompt, наследующему PADS, Context Prompt и acceptance criteria.

**Критерии приёмки.**

- `PLT-GOV-002-AC-01` — Prompt содержит include/exclude scope и forbidden dependencies.
- `PLT-GOV-002-AC-02` — Результат включает implementation manifest и проходит независимый Review Prompt.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-GOV-002
capability: CAP-GOV-09
owner_context: Architecture Governance
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### PLT-GOV-003. Контрактный design gate {#plt-gov-003}

| Поле | Значение |
| --- | --- |
| Тип | `governance` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Architecture Governance` / `contract owners` |
| Business capability | `CAP-GOV-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §12, §13, §22, §23 |

**Требование.**

Новое или несовместимо изменяемое публичное API/Event должно пройти отдельный design review до implementation prompt.

**Критерии приёмки.**

- `PLT-GOV-003-AC-01` — Contract ID и owner присвоены до кода.
- `PLT-GOV-003-AC-02` — buf breaking/schema compatibility checks добавлены.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-GOV-003
capability: CAP-GOV-05
owner_context: Architecture Governance
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```
