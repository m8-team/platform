---
title: "Requirements Catalog: Resource Manager"
description: "Требования Resource Manager."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 4. Resource Manager {#requirements-resource-manager}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 3. Платформенные и сквозные требования](03-platform-cross-cutting.md) | [Следующий раздел: 5. Identity](05-identity.md)

{% endnote %}

Владелец раздела: **Resource Manager**. Требований: **26**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `RM-FR-001` | functional | Must | `CAP-RM-01` | Создать Organization | `ANALYZED` |
| `RM-FR-002` | functional | Must | `CAP-RM-01` | Изменить Organization | `ANALYZED` |
| `RM-FR-003` | functional | Must | `CAP-RM-07` | Приостановить Organization | `ANALYZED` |
| `RM-FR-004` | functional | Must | `CAP-RM-07` | Восстановить Organization | `ANALYZED` |
| `RM-FR-005` | functional | Must | `CAP-RM-07` | Архивировать Organization | `ANALYZED` |
| `RM-FR-010` | functional | Must | `CAP-RM-02` | Создать Workspace | `ANALYZED` |
| `RM-FR-011` | functional | Must | `CAP-RM-02` | Изменить Workspace | `ANALYZED` |
| `RM-FR-012` | functional | Must | `CAP-RM-07` | Приостановить или восстановить Workspace | `ANALYZED` |
| `RM-FR-020` | functional | Must | `CAP-RM-03` | Создать Project | `ANALYZED` |
| `RM-FR-021` | functional | Must | `CAP-RM-08` | Переместить Project | `ANALYZED` |
| `RM-FR-022` | functional | Must | `CAP-RM-07` | Удалить Project с зависимостями | `ANALYZED` |
| `RM-FR-023` | functional | Must | `CAP-RM-07` | Приостановить Project | `ANALYZED` |
| `RM-FR-024` | functional | Must | `CAP-RM-07` | Восстановить Project | `ANALYZED` |
| `RM-FR-030` | functional | Must | `CAP-RM-04` | Зарегистрировать Service | `ANALYZED` |
| `RM-FR-031` | functional | Must | `CAP-RM-04` | Изменить регистрацию Service | `ANALYZED` |
| `RM-FR-032` | functional | Must | `CAP-RM-04` | Снять Service с регистрации | `ANALYZED` |
| `RM-FR-040` | functional | Must | `CAP-RM-09` | Получить ресурс по ID | `ANALYZED` |
| `RM-FR-041` | functional | Must | `CAP-RM-05` | Перечислить дочерние ресурсы | `ANALYZED` |
| `RM-FR-042` | functional | Must | `CAP-RM-05` | Получить полный путь ресурса | `ANALYZED` |
| `RM-FR-043` | functional | Must | `CAP-RM-09` | Искать ресурсы | `ANALYZED` |
| `RM-FR-044` | functional | Must | `CAP-RM-06` | Обновить labels через FieldMask | `ANALYZED` |
| `RM-FR-050` | functional | Must | `CAP-RM-10` | Публиковать lifecycle events | `ANALYZED` |
| `RM-DATA-001` | data | Must | `CAP-RM-05` | Источник истины ресурсной иерархии | `ANALYZED` |
| `RM-DATA-002` | data | Must | `CAP-RM-07` | Исторические ссылки на удалённые ресурсы | `ANALYZED` |
| `RM-SEC-001` | security | Must | `CAP-RM-01` | Авторизация управления иерархией | `ANALYZED` |
| `RM-NFR-001` | non-functional | Must | `CAP-RM-09` | Задержка чтения ресурса | `ANALYZED` |

## Детальные требования

### RM-FR-001. Создать Organization {#rm-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Создать Organization как верхнюю административную границу с уникальным идентификатором и начальным состоянием ACTIVE.

**Критерии приёмки.**

- `RM-FR-001-AC-01` — Создаётся один агрегат и событие OrganizationCreated.
- `RM-FR-001-AC-02` — Повтор идемпотентного запроса возвращает исходный ресурс.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-001
capability: CAP-RM-01
owner_context: Resource Manager
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

### RM-FR-002. Изменить Organization {#rm-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Изменять отображаемое имя, описание, labels и допустимые атрибуты Organization с optimistic concurrency.

**Критерии приёмки.**

- `RM-FR-002-AC-01` — Несовпадающая revision возвращает конфликт.
- `RM-FR-002-AC-02` — Изменённые поля отражаются в событии без передачи секретных данных.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-002
capability: CAP-RM-01
owner_context: Resource Manager
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

### RM-FR-003. Приостановить Organization {#rm-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Перевести Organization в SUSPENDED и запретить новые дочерние мутации по политике.

**Критерии приёмки.**

- `RM-FR-003-AC-01` — Существующие данные сохраняются.
- `RM-FR-003-AC-02` — Новая mutation в приостановленной области отклоняется, кроме разрешённых recovery operations.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-003
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-004. Восстановить Organization {#rm-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Вернуть приостановленную Organization в ACTIVE после проверки прав и риска.

**Критерии приёмки.**

- `RM-FR-004-AC-01` — Восстановление возможно только из SUSPENDED.
- `RM-FR-004-AC-02` — Действие фиксируется в Audit.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-004
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-005. Архивировать Organization {#rm-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Перевести Organization в архивное состояние после проверки отсутствия активных зависимостей.

**Критерии приёмки.**

- `RM-FR-005-AC-01` — Активные Workspace блокируют завершение или запускают управляемый процесс.
- `RM-FR-005-AC-02` — Архивная Organization недоступна для новых ресурсов.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-005
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-010. Создать Workspace {#rm-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Создать Workspace внутри существующей активной Organization.

**Критерии приёмки.**

- `RM-FR-010-AC-01` — Родитель существует и ACTIVE.
- `RM-FR-010-AC-02` — Workspace получает стабильный resource path.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-010
capability: CAP-RM-02
owner_context: Resource Manager
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

### RM-FR-011. Изменить Workspace {#rm-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Изменять разрешённые атрибуты Workspace с проверкой revision.

**Критерии приёмки.**

- `RM-FR-011-AC-01` — Parent organization не меняется обычным update.
- `RM-FR-011-AC-02` — Update публикует WorkspaceUpdated.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-011
capability: CAP-RM-02
owner_context: Resource Manager
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

### RM-FR-012. Приостановить или восстановить Workspace {#rm-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Управлять состоянием Workspace без физического удаления дочерних Project.

**Критерии приёмки.**

- `RM-FR-012-AC-01` — Состояние дочерних ресурсов явно учитывается при policy evaluation.
- `RM-FR-012-AC-02` — Решение не меняет данные Access напрямую.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-012
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-020. Создать Project {#rm-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Создать Project как основную границу изоляции внутри Workspace.

**Критерии приёмки.**

- `RM-FR-020-AC-01` — Project имеет уникальный ID и parent reference.
- `RM-FR-020-AC-02` — Создание доступно только с требуемым permission.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-020
capability: CAP-RM-03
owner_context: Resource Manager
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

### RM-FR-021. Переместить Project {#rm-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Переместить Project между Workspace совместимого scope через длительную Operation.

**Критерии приёмки.**

- `RM-FR-021-AC-01` — До commit проверены политики исходной и целевой областей.
- `RM-FR-021-AC-02` — Публикуется ProjectMoved с old_parent и new_parent.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-021
capability: CAP-RM-08
owner_context: Resource Manager
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

### RM-FR-022. Удалить Project с зависимостями {#rm-fr-022}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Координировать завершение Project, deprovision ресурсов и очистку зависимых данных без прямой записи в чужие сервисы.

**Критерии приёмки.**

- `RM-FR-022-AC-01` — Project переходит в DELETING до завершения участников.
- `RM-FR-022-AC-02` — Timeout или отказ участника виден в Operation и допускает retry/manual resolution.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-022
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-023. Приостановить Project {#rm-fr-023}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Приостановить Project и запретить новые управляемые действия в его области по policy.

**Критерии приёмки.**

- `RM-FR-023-AC-01` — Read-only операции остаются доступны согласно Access.
- `RM-FR-023-AC-02` — Состояние распространяется потребителям событием.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-023
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-024. Восстановить Project {#rm-fr-024}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Восстановить Project после устранения причины приостановки.

**Критерии приёмки.**

- `RM-FR-024-AC-01` — Восстановление проходит Access и при необходимости Risk Decision.
- `RM-FR-024-AC-02` — Сохраняется прежний Project ID и hierarchy path.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-024
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-FR-030. Зарегистрировать Service {#rm-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Зарегистрировать Service как ресурс, принадлежащий Project.

**Критерии приёмки.**

- `RM-FR-030-AC-01` — Project существует и разрешает регистрацию.
- `RM-FR-030-AC-02` — Service name уникально в документированной области.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-030
capability: CAP-RM-04
owner_context: Resource Manager
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

### RM-FR-031. Изменить регистрацию Service {#rm-fr-031}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Изменять метаданные и состояние регистрации Service с optimistic concurrency.

**Критерии приёмки.**

- `RM-FR-031-AC-01` — Нельзя сменить Project обычным update.
- `RM-FR-031-AC-02` — Изменения публикуются versioned event.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-031
capability: CAP-RM-04
owner_context: Resource Manager
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

### RM-FR-032. Снять Service с регистрации {#rm-fr-032}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Завершить регистрацию Service с проверкой зависимостей и сохранением исторической ссылки.

**Критерии приёмки.**

- `RM-FR-032-AC-01` — Service становится RETIRED или DELETED согласно policy.
- `RM-FR-032-AC-02` — Исторический Audit продолжает разрешать ссылку на ID.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-032
capability: CAP-RM-04
owner_context: Resource Manager
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

### RM-FR-040. Получить ресурс по ID {#rm-fr-040}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Возвращать Organization, Workspace, Project или Service по типизированному идентификатору в разрешённой области.

**Критерии приёмки.**

- `RM-FR-040-AC-01` — Несуществующий ресурс возвращает NOT_FOUND без раскрытия чужого scope.
- `RM-FR-040-AC-02` — Ответ содержит revision и lifecycle state.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-040
capability: CAP-RM-09
owner_context: Resource Manager
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

### RM-FR-041. Перечислить дочерние ресурсы {#rm-fr-041}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Получать дочерние ресурсы с пагинацией, фильтрацией и стабильным порядком.

**Критерии приёмки.**

- `RM-FR-041-AC-01` — Page token нельзя использовать с другим фильтром.
- `RM-FR-041-AC-02` — Результат ограничен effective access scope.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-041
capability: CAP-RM-05
owner_context: Resource Manager
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

### RM-FR-042. Получить полный путь ресурса {#rm-fr-042}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Возвращать каноническую цепочку Organization/Workspace/Project/Service.

**Критерии приёмки.**

- `RM-FR-042-AC-01` — Путь строится из authoritative hierarchy.
- `RM-FR-042-AC-02` — Удалённые исторические родители обозначаются без восстановления активного состояния.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-042
capability: CAP-RM-05
owner_context: Resource Manager
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

### RM-FR-043. Искать ресурсы {#rm-fr-043}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Искать ресурсы по разрешённым полям и labels в пределах доступной области.

**Критерии приёмки.**

- `RM-FR-043-AC-01` — Поиск не раскрывает наличие недоступных ресурсов.
- `RM-FR-043-AC-02` — Результаты имеют documented freshness.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-043
capability: CAP-RM-09
owner_context: Resource Manager
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

### RM-FR-044. Обновить labels через FieldMask {#rm-fr-044}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Частично обновлять пользовательские labels и annotations через FieldMask.

**Критерии приёмки.**

- `RM-FR-044-AC-01` — Неуказанные поля не изменяются.
- `RM-FR-044-AC-02` — Системные labels защищены от пользовательской записи.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-044
capability: CAP-RM-06
owner_context: Resource Manager
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

### RM-FR-050. Публиковать lifecycle events {#rm-fr-050}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Access, Audit |
| Данные | Resource Manager aggregate |
| Безопасность | permission check |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.8, §7, §9.2, §11 |

**Требование.**

Публиковать versioned Integration Events обо всех значимых изменениях ресурсной иерархии.

**Критерии приёмки.**

- `RM-FR-050-AC-01` — Event содержит resource_id, parent, state и revision.
- `RM-FR-050-AC-02` — Событие фиксируется через Outbox в транзакции агрегата.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-FR-050
capability: CAP-RM-10
owner_context: Resource Manager
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

### RM-DATA-001. Источник истины ресурсной иерархии {#rm-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Resource Manager должен быть единственным источником истины для parent-child отношений Organization, Workspace, Project и Service.

**Критерии приёмки.**

- `RM-DATA-001-AC-01` — Другие сервисы хранят только ссылки или проекции.
- `RM-DATA-001-AC-02` — Любое расхождение проекции исправляется из событий или reconciliation с Resource Manager.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-DATA-001
capability: CAP-RM-05
owner_context: Resource Manager
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

### RM-DATA-002. Исторические ссылки на удалённые ресурсы {#rm-data-002}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Удаление ресурса должно сохранять минимальные tombstone-данные, необходимые для аудита, дедупликации и ссылочной истории.

**Критерии приёмки.**

- `RM-DATA-002-AC-01` — Tombstone не позволяет выполнять новые операции.
- `RM-DATA-002-AC-02` — Retention и окончательное уничтожение документированы.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-DATA-002
capability: CAP-RM-07
owner_context: Resource Manager
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

### RM-SEC-001. Авторизация управления иерархией {#rm-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Каждая mutation ресурсной иерархии должна проверять permission на целевой и, где требуется, родительский ресурс.

**Критерии приёмки.**

- `RM-SEC-001-AC-01` — Move проверяет права в исходной и целевой областях.
- `RM-SEC-001-AC-02` — Недоступный parent не раскрывается через различие ошибок.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-SEC-001
capability: CAP-RM-01
owner_context: Resource Manager
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

### RM-NFR-001. Задержка чтения ресурса {#rm-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Resource Manager` / `m8-resource-manager` |
| Business capability | `CAP-RM-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | p95<=150ms, p99<=400ms |
| Основание PADS | §19 |

**Требование.**

Get resource должен иметь p95 не более 150 мс и p99 не более 400 мс при номинальной нагрузке в основном регионе.

**Критерии приёмки.**

- `RM-NFR-001-AC-01` — SLI измеряется на серверной стороне без client/network time.
- `RM-NFR-001-AC-02` — Нагрузка и размер ресурса определены в performance profile.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RM-NFR-001
capability: CAP-RM-09
owner_context: Resource Manager
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
