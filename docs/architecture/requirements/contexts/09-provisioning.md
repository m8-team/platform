---
title: "Requirements Catalog: Provisioning"
description: "Требования Provisioning."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 9. Provisioning {#requirements-provisioning}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 8. Risk Decision](08-risk-decision.md) | [Следующий раздел: 10. Audit](10-audit.md)

{% endnote %}

Владелец раздела: **Provisioning**. Требований: **25**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `PROV-FR-001` | functional | Must | `CAP-PROV-01` | Зарегистрировать ResourceDefinition | `ANALYZED` |
| `PROV-FR-002` | functional | Must | `CAP-PROV-01` | Опубликовать ResourceDefinition | `ANALYZED` |
| `PROV-FR-003` | functional | Must | `CAP-PROV-02` | Зарегистрировать Driver | `ANALYZED` |
| `PROV-FR-004` | functional | Must | `CAP-PROV-02` | Отключить Driver | `ANALYZED` |
| `PROV-FR-010` | functional | Must | `CAP-PROV-03` | Создать ManagedResource | `ANALYZED` |
| `PROV-FR-011` | functional | Must | `CAP-PROV-03` | Изменить desired state | `ANALYZED` |
| `PROV-FR-012` | functional | Must | `CAP-PROV-10` | Удалить ManagedResource | `ANALYZED` |
| `PROV-FR-013` | functional | Must | `CAP-PROV-08` | Приостановить reconciliation | `ANALYZED` |
| `PROV-FR-014` | functional | Must | `CAP-PROV-08` | Возобновить reconciliation | `ANALYZED` |
| `PROV-FR-020` | functional | Must | `CAP-PROV-04` | Выбрать Placement | `ANALYZED` |
| `PROV-FR-021` | functional | Must | `CAP-PROV-04` | Закрепить Placement | `ANALYZED` |
| `PROV-FR-030` | functional | Must | `CAP-PROV-05` | Выполнить Reconciliation | `ANALYZED` |
| `PROV-FR-031` | functional | Must | `CAP-PROV-05` | Получить observed state | `ANALYZED` |
| `PROV-FR-032` | functional | Must | `CAP-PROV-07` | Обнаружить Drift | `ANALYZED` |
| `PROV-FR-033` | functional | Must | `CAP-PROV-07` | Исправить Drift | `ANALYZED` |
| `PROV-FR-034` | functional | Must | `CAP-PROV-03` | Принять внешний ресурс под управление | `ANALYZED` |
| `PROV-FR-040` | functional | Must | `CAP-PROV-08` | Повторить неуспешный шаг | `ANALYZED` |
| `PROV-FR-041` | functional | Must | `CAP-PROV-09` | Выполнить Compensation | `ANALYZED` |
| `PROV-FR-042` | functional | Must | `CAP-PROV-09` | Создать Manual Remediation | `ANALYZED` |
| `PROV-FR-050` | functional | Must | `CAP-PROV-11` | Получить Outputs | `ANALYZED` |
| `PROV-FR-051` | functional | Must | `CAP-PROV-12` | Получить состояние ManagedResource | `ANALYZED` |
| `PROV-FR-060` | functional | Must | `CAP-PROV-13` | Публиковать Provisioning events | `ANALYZED` |
| `PROV-DATA-001` | data | Must | `CAP-PROV-05` | Разделение desired и observed state | `ANALYZED` |
| `PROV-SEC-001` | security | Must | `CAP-PROV-11` | Секреты управляемых ресурсов | `ANALYZED` |
| `PROV-NFR-001` | non-functional | Must | `CAP-PROV-03` | Время начала Provisioning Operation | `ANALYZED` |

## Детальные требования

### PROV-FR-001. Зарегистрировать ResourceDefinition {#prov-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-01` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Зарегистрировать versioned тип управляемого ресурса, schema desired state и capabilities driver.

**Критерии приёмки.**

- `PROV-FR-001-AC-01` — Definition проходит validation и compatibility check.
- `PROV-FR-001-AC-02` — Draft version не используется для новых ресурсов до publish.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-001
capability: CAP-PROV-01
owner_context: Provisioning
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

### PROV-FR-002. Опубликовать ResourceDefinition {#prov-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-01` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Опубликовать одобренную definition version для создания ManagedResource.

**Критерии приёмки.**

- `PROV-FR-002-AC-01` — Breaking schema имеет migration strategy.
- `PROV-FR-002-AC-02` — Active version выбирается явно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-002
capability: CAP-PROV-01
owner_context: Provisioning
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

### PROV-FR-003. Зарегистрировать Driver {#prov-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-02` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Зарегистрировать driver implementation, supported definitions, regions и health metadata.

**Критерии приёмки.**

- `PROV-FR-003-AC-01` — Driver проходит conformance tests.
- `PROV-FR-003-AC-02` — Секреты подключения хранятся вне definition.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-003
capability: CAP-PROV-02
owner_context: Provisioning
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

### PROV-FR-004. Отключить Driver {#prov-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-02` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Запретить новые placements на driver без потери управления существующими ресурсами.

**Критерии приёмки.**

- `PROV-FR-004-AC-01` — Существующие resources получают degraded/maintenance policy.
- `PROV-FR-004-AC-02` — Отключение аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-004
capability: CAP-PROV-02
owner_context: Provisioning
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

### PROV-FR-010. Создать ManagedResource {#prov-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-03` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Создать desired state ManagedResource в Project и запустить reconciliation Operation.

**Критерии приёмки.**

- `PROV-FR-010-AC-01` — Project ACTIVE и caller авторизован.
- `PROV-FR-010-AC-02` — Desired state и Outbox/Operation фиксируются атомарно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-010
capability: CAP-PROV-03
owner_context: Provisioning
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

### PROV-FR-011. Изменить desired state {#prov-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-03` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Изменить спецификацию ManagedResource с revision и запустить reconciliation.

**Критерии приёмки.**

- `PROV-FR-011-AC-01` — Невалидное или запрещённое изменение отклоняется до commit.
- `PROV-FR-011-AC-02` — Observed state не перезаписывается клиентом.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-011
capability: CAP-PROV-03
owner_context: Provisioning
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

### PROV-FR-012. Удалить ManagedResource {#prov-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-10` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Перевести ресурс в DELETING и выполнить deprovision через driver.

**Критерии приёмки.**

- `PROV-FR-012-AC-01` — Необратимость подтверждается policy/step-up при необходимости.
- `PROV-FR-012-AC-02` — Partial deletion видима и допускает remediation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-012
capability: CAP-PROV-10
owner_context: Provisioning
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

### PROV-FR-013. Приостановить reconciliation {#prov-fr-013}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-08` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Приостановить автоматические изменения ресурса, сохранив наблюдение состояния.

**Критерии приёмки.**

- `PROV-FR-013-AC-01` — Pause не маскирует drift.
- `PROV-FR-013-AC-02` — Resume продолжает с актуальной desired revision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-013
capability: CAP-PROV-08
owner_context: Provisioning
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

### PROV-FR-014. Возобновить reconciliation {#prov-fr-014}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-08` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Возобновить согласование с идемпотентной обработкой предыдущих попыток.

**Критерии приёмки.**

- `PROV-FR-014-AC-01` — Duplicate resume не создаёт параллельный workflow.
- `PROV-FR-014-AC-02` — Stale workflow не применяет старую revision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-014
capability: CAP-PROV-08
owner_context: Provisioning
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

### PROV-FR-020. Выбрать Placement {#prov-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-04` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Выбрать допустимый cluster/region/provider на основании policy, capacity, residency и constraints.

**Критерии приёмки.**

- `PROV-FR-020-AC-01` — Placement decision объяснимо и versioned.
- `PROV-FR-020-AC-02` — Недоступная capacity возвращает WAIT/FAILED без скрытой смены требований.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-020
capability: CAP-PROV-04
owner_context: Provisioning
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

### PROV-FR-021. Закрепить Placement {#prov-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-04` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Сохранить выбранное размещение и правила допустимого перемещения.

**Критерии приёмки.**

- `PROV-FR-021-AC-01` — Driver получает normalized placement reference.
- `PROV-FR-021-AC-02` — Автоматическая миграция между регионами не выполняется без policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-021
capability: CAP-PROV-04
owner_context: Provisioning
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

### PROV-FR-030. Выполнить Reconciliation {#prov-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-05` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Сравнить desired и observed state и выполнить необходимые действия через driver.

**Критерии приёмки.**

- `PROV-FR-030-AC-01` — Reconciliation идемпотентен.
- `PROV-FR-030-AC-02` — Action plan основан на конкретной desired revision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-030
capability: CAP-PROV-05
owner_context: Provisioning
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

### PROV-FR-031. Получить observed state {#prov-fr-031}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-05` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Получить нормализованное наблюдаемое состояние от driver.

**Критерии приёмки.**

- `PROV-FR-031-AC-01` — Raw provider payload не становится public contract.
- `PROV-FR-031-AC-02` — Freshness и last_observed_at возвращаются.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-031
capability: CAP-PROV-05
owner_context: Provisioning
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

### PROV-FR-032. Обнаружить Drift {#prov-fr-032}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-07` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Определить расхождение desired/observed и классифицировать severity.

**Критерии приёмки.**

- `PROV-FR-032-AC-01` — Drift event не создаётся повторно без изменения fingerprint.
- `PROV-FR-032-AC-02` — Игнорируемые поля определены policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-032
capability: CAP-PROV-07
owner_context: Provisioning
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

### PROV-FR-033. Исправить Drift {#prov-fr-033}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-07` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Автоматически или вручную устранить drift согласно policy.

**Критерии приёмки.**

- `PROV-FR-033-AC-01` — AUTO remediation не выполняет запрещённые destructive actions.
- `PROV-FR-033-AC-02` — Manual approval сохраняется в Audit.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-033
capability: CAP-PROV-07
owner_context: Provisioning
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

### PROV-FR-034. Принять внешний ресурс под управление {#prov-fr-034}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-03` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Импортировать существующий внешний ресурс после discovery и проверки ownership.

**Критерии приёмки.**

- `PROV-FR-034-AC-01` — Import не изменяет внешний ресурс до approval.
- `PROV-FR-034-AC-02` — Конфликты desired state показываются явно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-034
capability: CAP-PROV-03
owner_context: Provisioning
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

### PROV-FR-040. Повторить неуспешный шаг {#prov-fr-040}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-08` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Выполнить retry временной ошибки с backoff, deadline и retry budget.

**Критерии приёмки.**

- `PROV-FR-040-AC-01` — Permanent error не повторяется бесконечно.
- `PROV-FR-040-AC-02` — Количество попыток и next_retry_at наблюдаемы.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-040
capability: CAP-PROV-08
owner_context: Provisioning
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

### PROV-FR-041. Выполнить Compensation {#prov-fr-041}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-09` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Компенсировать завершённые шаги процесса, если это безопасно и предусмотрено workflow.

**Критерии приёмки.**

- `PROV-FR-041-AC-01` — Необратимый шаг обозначается commit point.
- `PROV-FR-041-AC-02` — Ошибка compensation переводит ресурс в NEEDS_ATTENTION.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-041
capability: CAP-PROV-09
owner_context: Provisioning
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

### PROV-FR-042. Создать Manual Remediation {#prov-fr-042}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-09` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Создать операторскую задачу с безопасной диагностикой и допустимыми действиями.

**Критерии приёмки.**

- `PROV-FR-042-AC-01` — Действие оператора проверяется Access/Risk.
- `PROV-FR-042-AC-02` — Результат возвращается process workflow.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-042
capability: CAP-PROV-09
owner_context: Provisioning
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

### PROV-FR-050. Получить Outputs {#prov-fr-050}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-11` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Возвращать типизированные выходы ресурса с маскированием секретов.

**Критерии приёмки.**

- `PROV-FR-050-AC-01` — Secret output возвращается только через secret reference.
- `PROV-FR-050-AC-02` — Output связан с observed revision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-050
capability: CAP-PROV-11
owner_context: Provisioning
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

### PROV-FR-051. Получить состояние ManagedResource {#prov-fr-051}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-12` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Возвращать desired/observed summary, conditions, operation и последнюю ошибку.

**Критерии приёмки.**

- `PROV-FR-051-AC-01` — Resource state не подменяется Operation state.
- `PROV-FR-051-AC-02` — Ошибка имеет канонический код и remediation hint.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-051
capability: CAP-PROV-12
owner_context: Provisioning
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

### PROV-FR-060. Публиковать Provisioning events {#prov-fr-060}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-13` |
| Согласованность | C1 local desired state + C2 external reconciliation |
| Зависимости | Resource Manager, Access, Risk Decision, Temporal ACL, provider drivers, Audit |
| Данные | ManagedResource, ResourceDefinition, Reconciliation |
| Безопасность | permission check, secret references only |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.13, §7, §9.7, §14, §16 |

**Требование.**

Публиковать события desired changes, reconciliation, drift, readiness и deletion.

**Критерии приёмки.**

- `PROV-FR-060-AC-01` — Event содержит resource and revision identifiers.
- `PROV-FR-060-AC-02` — Provider-specific details минимизированы.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-FR-060
capability: CAP-PROV-13
owner_context: Provisioning
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

### PROV-DATA-001. Разделение desired и observed state {#prov-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Desired state должен принадлежать Provisioning, а observed state формироваться только из проверенного driver observation.

**Критерии приёмки.**

- `PROV-DATA-001-AC-01` — Client не может записывать observed fields.
- `PROV-DATA-001-AC-02` — Каждое observed state имеет source, time и desired revision correlation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-DATA-001
capability: CAP-PROV-05
owner_context: Provisioning
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

### PROV-SEC-001. Секреты управляемых ресурсов {#prov-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Секреты и credentials внешних ресурсов должны передаваться ссылками на secret manager и не храниться в открытом виде в desired state, Operation или event.

**Критерии приёмки.**

- `PROV-SEC-001-AC-01` — Public API возвращает secret reference или redacted value.
- `PROV-SEC-001-AC-02` — Driver получает секрет только в момент выполнения и не логирует его.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-SEC-001
capability: CAP-PROV-11
owner_context: Provisioning
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

### PROV-NFR-001. Время начала Provisioning Operation {#prov-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Provisioning` / `m8-provisioning` |
| Business capability | `CAP-PROV-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | ack p95<=500ms |
| Основание PADS | §19 |

**Требование.**

Create/Update ManagedResource должен вернуть принятую Operation не позднее 500 мс при доступном локальном хранилище.

**Критерии приёмки.**

- `PROV-NFR-001-AC-01` — Внешний provider не вызывается в критическом пути ответа.
- `PROV-NFR-001-AC-02` — Operation и desired state уже committed.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PROV-NFR-001
capability: CAP-PROV-03
owner_context: Provisioning
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
