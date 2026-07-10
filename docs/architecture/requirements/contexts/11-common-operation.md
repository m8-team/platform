---
title: "Requirements Catalog: Common Operation"
description: "Требования Common Operation."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 11. Common Operation {#requirements-common-operation}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 10. Audit](10-audit.md) | [Следующий раздел: 12. Архитектурное управление и SPDD](../governance/12-architecture-governance-spdd.md)

{% endnote %}

Владелец раздела: **Common Operation**. Требований: **12**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `OPS-FR-001` | functional | Must | `CAP-OPS-02` | Получить Operation | `ANALYZED` |
| `OPS-FR-002` | functional | Must | `CAP-OPS-02` | Перечислить Operations | `ANALYZED` |
| `OPS-FR-003` | functional | Must | `CAP-OPS-04` | Ожидать Operation | `ANALYZED` |
| `OPS-FR-004` | functional | Must | `CAP-OPS-05` | Запросить Cancellation | `ANALYZED` |
| `OPS-FR-005` | functional | Must | `CAP-OPS-08` | Удалить запись Operation | `ANALYZED` |
| `OPS-FR-006` | functional | Must | `CAP-OPS-03` | Обновить Progress | `ANALYZED` |
| `OPS-FR-007` | functional | Must | `CAP-OPS-06` | Завершить Operation результатом | `ANALYZED` |
| `OPS-FR-008` | functional | Must | `CAP-OPS-07` | Завершить Operation ошибкой | `ANALYZED` |
| `OPS-FR-009` | functional | Must | `CAP-OPS-09` | Связать Operation с Workflow | `ANALYZED` |
| `OPS-FR-010` | functional | Must | `CAP-OPS-10` | Идемпотентно начать длительную работу | `ANALYZED` |
| `OPS-DATA-001` | data | Must | `CAP-OPS-01` | Владение Operation предметным контекстом | `ANALYZED` |
| `OPS-NFR-001` | non-functional | Must | `CAP-OPS-02` | Доступность чтения Operation | `ANALYZED` |

## Детальные требования

### OPS-FR-001. Получить Operation {#ops-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Получить Operation по имени/ID с type metadata, progress, result или error.

**Критерии приёмки.**

- `OPS-FR-001-AC-01` — Caller имеет доступ к owner resource scope.
- `OPS-FR-001-AC-02` — Ответ не раскрывает internal workflow ID без необходимости.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-001
capability: CAP-OPS-02
owner_context: Common Operation
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

### OPS-FR-002. Перечислить Operations {#ops-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Перечислить операции по owner, resource, type, state и time range.

**Критерии приёмки.**

- `OPS-FR-002-AC-01` — Пагинация стабильна.
- `OPS-FR-002-AC-02` — Список ограничен доступным scope.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-002
capability: CAP-OPS-02
owner_context: Common Operation
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

### OPS-FR-003. Ожидать Operation {#ops-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Ожидать изменение или завершение операции до client deadline.

**Критерии приёмки.**

- `OPS-FR-003-AC-01` — Timeout возвращает текущую Operation.
- `OPS-FR-003-AC-02` — Wait не создаёт новую работу.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-003
capability: CAP-OPS-04
owner_context: Common Operation
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

### OPS-FR-004. Запросить Cancellation {#ops-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Запросить отмену операции, если предметный owner допускает её в текущей стадии.

**Критерии приёмки.**

- `OPS-FR-004-AC-01` — Cancellation request отделён от факта CANCELLED.
- `OPS-FR-004-AC-02` — После commit point возвращается CANNOT_CANCEL.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-004
capability: CAP-OPS-05
owner_context: Common Operation
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

### OPS-FR-005. Удалить запись Operation {#ops-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Удалить завершённую запись после retention, если owner policy допускает.

**Критерии приёмки.**

- `OPS-FR-005-AC-01` — Удаление не удаляет Audit и предметный ресурс.
- `OPS-FR-005-AC-02` — Незавершённая Operation не удаляется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-005
capability: CAP-OPS-08
owner_context: Common Operation
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

### OPS-FR-006. Обновить Progress {#ops-fr-006}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Обновлять stage, message и percent монотонно и с ограниченной частотой.

**Критерии приёмки.**

- `OPS-FR-006-AC-01` — Percent находится 0..100 и не уменьшается без нового attempt semantics.
- `OPS-FR-006-AC-02` — Progress не используется как authoritative resource state.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-006
capability: CAP-OPS-03
owner_context: Common Operation
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

### OPS-FR-007. Завершить Operation результатом {#ops-fr-007}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Завершить Operation типизированным result ровно один раз.

**Критерии приёмки.**

- `OPS-FR-007-AC-01` — Terminal state immutable.
- `OPS-FR-007-AC-02` — Result type соответствует operation type.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-007
capability: CAP-OPS-06
owner_context: Common Operation
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

### OPS-FR-008. Завершить Operation ошибкой {#ops-fr-008}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Завершить Operation канонической ошибкой с retry/remediation metadata.

**Критерии приёмки.**

- `OPS-FR-008-AC-01` — Ошибка безопасна для caller.
- `OPS-FR-008-AC-02` — Internal cause доступна только telemetry.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-008
capability: CAP-OPS-07
owner_context: Common Operation
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

### OPS-FR-009. Связать Operation с Workflow {#ops-fr-009}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Сохранить correlation с Temporal workflow/run без утечки внешнего lifecycle в domain API.

**Критерии приёмки.**

- `OPS-FR-009-AC-01` — Workflow restart не создаёт новую Operation.
- `OPS-FR-009-AC-02` — Один owner operation может иметь несколько workflow runs по retry policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-009
capability: CAP-OPS-09
owner_context: Common Operation
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

### OPS-FR-010. Идемпотентно начать длительную работу {#ops-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Concrete owner context, Temporal adapter where used, Audit |
| Данные | Operation |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Связать idempotency key, owner command и Operation до запуска workflow.

**Критерии приёмки.**

- `OPS-FR-010-AC-01` — Повтор возвращает исходную Operation.
- `OPS-FR-010-AC-02` — Несовместимый payload с тем же key отклоняется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-FR-010
capability: CAP-OPS-10
owner_context: Common Operation
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

### OPS-DATA-001. Владение Operation предметным контекстом {#ops-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §11, §16 |

**Требование.**

Конкретная Operation должна храниться и управляться контекстом, владеющим предметным результатом; общий пакет задаёт только контракт.

**Критерии приёмки.**

- `OPS-DATA-001-AC-01` — Нет центрального сервиса, принимающего предметные решения за owner.
- `OPS-DATA-001-AC-02` — Owner определяет cancelability, result и retention.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-DATA-001
capability: CAP-OPS-01
owner_context: Common Operation
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

### OPS-NFR-001. Доступность чтения Operation {#ops-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Common Operation` / `operation owner service` |
| Business capability | `CAP-OPS-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | availability>=99.99%, p95<=100ms |
| Основание PADS | §19 |

**Требование.**

GetOperation критических процессов должен иметь доступность не ниже 99,99% и p95 не более 100 мс.

**Критерии приёмки.**

- `OPS-NFR-001-AC-01` — Read path не требует доступности Temporal.
- `OPS-NFR-001-AC-02` — Последнее committed состояние возвращается при недоступном workflow engine.

**Трассировка для следующего этапа:**

```yaml
requirement_id: OPS-NFR-001
capability: CAP-OPS-02
owner_context: Common Operation
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
