---
title: "Requirements Catalog: Access"
description: "Требования Access."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 7. Access {#requirements-access}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 6. Authentication](06-authentication.md) | [Следующий раздел: 8. Risk Decision](08-risk-decision.md)

{% endnote %}

Владелец раздела: **Access**. Требований: **25**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `ACC-FR-001` | functional | Must | `CAP-AUTHZ-06` | Проверить Permission | `ANALYZED` |
| `ACC-FR-002` | functional | Must | `CAP-AUTHZ-07` | Пакетно проверить Permissions | `ANALYZED` |
| `ACC-FR-003` | functional | Must | `CAP-AUTHZ-08` | Объяснить решение Access | `ANALYZED` |
| `ACC-FR-004` | functional | Must | `CAP-AUTHZ-10` | Получить эффективные Permissions субъекта | `ANALYZED` |
| `ACC-FR-005` | functional | Must | `CAP-AUTHZ-10` | Получить Subjects с доступом к ресурсу | `ANALYZED` |
| `ACC-FR-010` | functional | Must | `CAP-AUTHZ-02` | Создать Permission definition | `ANALYZED` |
| `ACC-FR-011` | functional | Must | `CAP-AUTHZ-03` | Создать Role | `ANALYZED` |
| `ACC-FR-012` | functional | Must | `CAP-AUTHZ-03` | Изменить Role | `ANALYZED` |
| `ACC-FR-013` | functional | Must | `CAP-AUTHZ-03` | Удалить Role | `ANALYZED` |
| `ACC-FR-014` | functional | Must | `CAP-AUTHZ-04` | Создать RoleBinding | `ANALYZED` |
| `ACC-FR-015` | functional | Must | `CAP-AUTHZ-04` | Отозвать RoleBinding | `ANALYZED` |
| `ACC-FR-016` | functional | Must | `CAP-AUTHZ-04` | Установить срок RoleBinding | `ANALYZED` |
| `ACC-FR-020` | functional | Must | `CAP-AUTHZ-05` | Создать Relationship | `ANALYZED` |
| `ACC-FR-021` | functional | Must | `CAP-AUTHZ-05` | Удалить Relationship | `ANALYZED` |
| `ACC-FR-022` | functional | Must | `CAP-AUTHZ-05` | Пакетно изменить Relationships | `ANALYZED` |
| `ACC-FR-030` | functional | Must | `CAP-AUTHZ-09` | Симулировать изменение доступа | `ANALYZED` |
| `ACC-FR-031` | functional | Must | `CAP-AUTHZ-09` | Проверить policy impact | `ANALYZED` |
| `ACC-FR-040` | functional | Must | `CAP-AUTHZ-11` | Создать Access Review | `ANALYZED` |
| `ACC-FR-041` | functional | Must | `CAP-AUTHZ-11` | Подтвердить или отозвать доступ | `ANALYZED` |
| `ACC-FR-050` | functional | Must | `CAP-AUTHZ-01` | Опубликовать Authorization Model | `ANALYZED` |
| `ACC-FR-051` | functional | Must | `CAP-AUTHZ-12` | Синхронизировать Access с SpiceDB | `ANALYZED` |
| `ACC-FR-060` | functional | Must | `CAP-AUTHZ-13` | Публиковать факты Access | `ANALYZED` |
| `ACC-DATA-001` | data | Must | `CAP-AUTHZ-05` | Источник истины отношений доступа | `ANALYZED` |
| `ACC-SEC-001` | security | Must | `CAP-AUTHZ-06` | Fail mode проверок доступа | `ANALYZED` |
| `ACC-NFR-001` | non-functional | Must | `CAP-AUTHZ-06` | Задержка CheckPermission | `ANALYZED` |

## Детальные требования

### ACC-FR-001. Проверить Permission {#acc-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Принять решение, может ли Subject выполнить Permission над Resource в заданном context.

**Критерии приёмки.**

- `ACC-FR-001-AC-01` — Результат ALLOW/DENY воспроизводим для revision модели.
- `ACC-FR-001-AC-02` — Timeout или недоступность движка обрабатывается по fail mode операции.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-001
capability: CAP-AUTHZ-06
owner_context: Access
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

### ACC-FR-002. Пакетно проверить Permissions {#acc-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Проверить набор permission/resource/subject tuples с ограниченным размером и частичными результатами.

**Критерии приёмки.**

- `ACC-FR-002-AC-01` — Порядок результатов соответствует входным items.
- `ACC-FR-002-AC-02` — Ошибка одного item не скрывает результаты остальных, если контракт допускает partial response.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-002
capability: CAP-AUTHZ-07
owner_context: Access
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

### ACC-FR-003. Объяснить решение Access {#acc-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Вернуть безопасное объяснение отношений и правил, приведших к решению.

**Критерии приёмки.**

- `ACC-FR-003-AC-01` — Explanation доступно только с отдельным permission.
- `ACC-FR-003-AC-02` — Внешнему пользователю не раскрываются чувствительные отношения третьих лиц.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-003
capability: CAP-AUTHZ-08
owner_context: Access
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

### ACC-FR-004. Получить эффективные Permissions субъекта {#acc-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Получить вычисленный набор разрешений в заданной ресурсной области.

**Критерии приёмки.**

- `ACC-FR-004-AC-01` — Результат имеет model revision/freshness.
- `ACC-FR-004-AC-02` — Пагинация стабильна и не обещает глобальный полный список без scope.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-004
capability: CAP-AUTHZ-10
owner_context: Access
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

### ACC-FR-005. Получить Subjects с доступом к ресурсу {#acc-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Получить субъектов, имеющих заданный permission над ресурсом, если модель поддерживает перечисление.

**Критерии приёмки.**

- `ACC-FR-005-AC-01` — Операция ограничена permission и размером.
- `ACC-FR-005-AC-02` — Неперечислимые отношения возвращают явное ограничение.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-005
capability: CAP-AUTHZ-10
owner_context: Access
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

### ACC-FR-010. Создать Permission definition {#acc-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Добавить permission в versioned authorization model до его использования ролями и проверками.

**Критерии приёмки.**

- `ACC-FR-010-AC-01` — Имя уникально в model version.
- `ACC-FR-010-AC-02` — Несовместимое изменение требует новой model version.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-010
capability: CAP-AUTHZ-02
owner_context: Access
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

### ACC-FR-011. Создать Role {#acc-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Создать именованный Role с набором permission references и допустимым scope.

**Критерии приёмки.**

- `ACC-FR-011-AC-01` — Role не содержит несуществующий permission.
- `ACC-FR-011-AC-02` — Системные роли защищены от несанкционированного изменения.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-011
capability: CAP-AUTHZ-03
owner_context: Access
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

### ACC-FR-012. Изменить Role {#acc-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Изменять состав Role с оценкой влияния на существующие bindings.

**Критерии приёмки.**

- `ACC-FR-012-AC-01` — Change impact показывает затрагиваемые scopes.
- `ACC-FR-012-AC-02` — Изменение публикует RoleUpdated после commit.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-012
capability: CAP-AUTHZ-03
owner_context: Access
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

### ACC-FR-013. Удалить Role {#acc-fr-013}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Удалить или retire Role после проверки активных bindings.

**Критерии приёмки.**

- `ACC-FR-013-AC-01` — Активные bindings блокируют удаление или обрабатываются отдельной Operation.
- `ACC-FR-013-AC-02` — Role ID не переиспользуется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-013
capability: CAP-AUTHZ-03
owner_context: Access
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

### ACC-FR-014. Создать RoleBinding {#acc-fr-014}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Назначить Role субъекту в допустимой resource scope.

**Критерии приёмки.**

- `ACC-FR-014-AC-01` — Subject, Role и ResourceReference валидны.
- `ACC-FR-014-AC-02` — Дубликат binding идемпотентен.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-014
capability: CAP-AUTHZ-04
owner_context: Access
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

### ACC-FR-015. Отозвать RoleBinding {#acc-fr-015}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Отозвать назначение роли и распространить изменение в authorization engine.

**Критерии приёмки.**

- `ACC-FR-015-AC-01` — После подтверждённой синхронизации check больше не разрешает доступ.
- `ACC-FR-015-AC-02` — До синхронизации freshness/degraded state наблюдаемы.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-015
capability: CAP-AUTHZ-04
owner_context: Access
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

### ACC-FR-016. Установить срок RoleBinding {#acc-fr-016}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Создать временное назначение с not_before и expires_at.

**Критерии приёмки.**

- `ACC-FR-016-AC-01` — Просроченное binding не участвует в решении.
- `ACC-FR-016-AC-02` — Истечение публикуется и аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-016
capability: CAP-AUTHZ-04
owner_context: Access
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

### ACC-FR-020. Создать Relationship {#acc-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Создать предметное отношение между Subject и Resource или между ресурсами.

**Критерии приёмки.**

- `ACC-FR-020-AC-01` — Relation допустим текущей model version.
- `ACC-FR-020-AC-02` — Запись в SpiceDB выполняется через Access adapter.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-020
capability: CAP-AUTHZ-05
owner_context: Access
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

### ACC-FR-021. Удалить Relationship {#acc-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Удалить отношение доступа с идемпотентной семантикой.

**Критерии приёмки.**

- `ACC-FR-021-AC-01` — Повтор удаления считается успешным или not present по контракту без побочного эффекта.
- `ACC-FR-021-AC-02` — Публикуется RelationshipDeleted.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-021
capability: CAP-AUTHZ-05
owner_context: Access
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

### ACC-FR-022. Пакетно изменить Relationships {#acc-fr-022}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Атомарно в пределах Access storage применить ограниченный набор relationship mutations.

**Критерии приёмки.**

- `ACC-FR-022-AC-01` — Невалидный item не приводит к частичному commit, если запрошена atomic mode.
- `ACC-FR-022-AC-02` — Размер batch ограничен.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-022
capability: CAP-AUTHZ-05
owner_context: Access
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

### ACC-FR-030. Симулировать изменение доступа {#acc-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Оценить предполагаемые relationships/role changes без их применения.

**Критерии приёмки.**

- `ACC-FR-030-AC-01` — Simulation не изменяет authoritative store.
- `ACC-FR-030-AC-02` — Ответ содержит model revision и differences.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-030
capability: CAP-AUTHZ-09
owner_context: Access
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

### ACC-FR-031. Проверить policy impact {#acc-fr-031}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Показать, какие subject/resource decisions изменятся при новой model version на тестовом наборе.

**Критерии приёмки.**

- `ACC-FR-031-AC-01` — Test corpus версионируется.
- `ACC-FR-031-AC-02` — Изменение критических решений блокирует publish без approval.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-031
capability: CAP-AUTHZ-09
owner_context: Access
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

### ACC-FR-040. Создать Access Review {#acc-fr-040}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Создать кампанию ревизии назначений для scope и reviewers.

**Критерии приёмки.**

- `ACC-FR-040-AC-01` — Snapshot review scope фиксируется.
- `ACC-FR-040-AC-02` — Review не меняет binding до явного решения.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-040
capability: CAP-AUTHZ-11
owner_context: Access
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

### ACC-FR-041. Подтвердить или отозвать доступ {#acc-fr-041}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Записать решение reviewer и применить требуемое изменение.

**Критерии приёмки.**

- `ACC-FR-041-AC-01` — Reviewer не может подтвердить собственный доступ, если segregation policy запрещает.
- `ACC-FR-041-AC-02` — Решение аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-041
capability: CAP-AUTHZ-11
owner_context: Access
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

### ACC-FR-050. Опубликовать Authorization Model {#acc-fr-050}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Опубликовать проверенную model version и синхронизировать её с authorization engine.

**Критерии приёмки.**

- `ACC-FR-050-AC-01` — Breaking model требует migration.
- `ACC-FR-050-AC-02` — Rollback возвращает ранее совместимую version без потери authoritative relationships.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-050
capability: CAP-AUTHZ-01
owner_context: Access
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

### ACC-FR-051. Синхронизировать Access с SpiceDB {#acc-fr-051}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-12` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Надёжно преобразовывать модель и relationships M8 в SpiceDB schema/tuples через ACL.

**Критерии приёмки.**

- `ACC-FR-051-AC-01` — Внутренние SpiceDB types не выходят в public API.
- `ACC-FR-051-AC-02` — Reconciliation обнаруживает пропуски и лишние tuples.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-051
capability: CAP-AUTHZ-12
owner_context: Access
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

### ACC-FR-060. Публиковать факты Access {#acc-fr-060}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-13` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Identity, SpiceDB ACL, Audit |
| Данные | AuthorizationModel, RoleBinding, Relationship |
| Безопасность | administrative permission, least privilege |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.11, §7, §9.5, §15 |

**Требование.**

Публиковать события о model, roles, bindings, relationships и review decisions.

**Критерии приёмки.**

- `ACC-FR-060-AC-01` — Event payload минимален и versioned.
- `ACC-FR-060-AC-02` — Событие следует authoritative commit.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-FR-060
capability: CAP-AUTHZ-13
owner_context: Access
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

### ACC-DATA-001. Источник истины отношений доступа {#acc-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Access должен владеть предметными RoleBinding и Relationship; SpiceDB является вычислительным и/или материализованным представлением через адаптер.

**Критерии приёмки.**

- `ACC-DATA-001-AC-01` — Восстановление SpiceDB возможно из authoritative Access data и event log/outbox.
- `ACC-DATA-001-AC-02` — Другие сервисы не записывают tuples напрямую.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-DATA-001
capability: CAP-AUTHZ-05
owner_context: Access
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

### ACC-SEC-001. Fail mode проверок доступа {#acc-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Mutation и чувствительное чтение должны использовать fail-closed при невозможности получить достоверное решение.

**Критерии приёмки.**

- `ACC-SEC-001-AC-01` — Fail-open допускается только для явно классифицированной некритической функции и принятого ADR.
- `ACC-SEC-001-AC-02` — Degraded decision помечается и наблюдается.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-SEC-001
capability: CAP-AUTHZ-06
owner_context: Access
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

### ACC-NFR-001. Задержка CheckPermission {#acc-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Access` / `m8-access` |
| Business capability | `CAP-AUTHZ-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | p95<=25ms, p99<=75ms |
| Основание PADS | §19 |

**Требование.**

CheckPermission должен иметь p95 не более 25 мс и p99 не более 75 мс внутри региона при номинальной нагрузке.

**Критерии приёмки.**

- `ACC-NFR-001-AC-01` — Метрика измеряется отдельно для cache hit и engine call.
- `ACC-NFR-001-AC-02` — Correctness не жертвуется ради устаревшего allow cache.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ACC-NFR-001
capability: CAP-AUTHZ-06
owner_context: Access
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
