---
title: "Requirements Catalog: Identity"
description: "Требования Identity."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 5. Identity {#requirements-identity}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 4. Resource Manager](04-resource-manager.md) | [Следующий раздел: 6. Authentication](06-authentication.md)

{% endnote %}

Владелец раздела: **Identity**. Требований: **28**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `ID-FR-001` | functional | Must | `CAP-ID-01` | Создать User Pool | `ANALYZED` |
| `ID-FR-002` | functional | Must | `CAP-ID-01` | Изменить User Pool | `ANALYZED` |
| `ID-FR-003` | functional | Must | `CAP-ID-01` | Приостановить User Pool | `ANALYZED` |
| `ID-FR-010` | functional | Must | `CAP-ID-02` | Создать User | `ANALYZED` |
| `ID-FR-011` | functional | Must | `CAP-ID-03` | Изменить профиль User | `ANALYZED` |
| `ID-FR-012` | functional | Must | `CAP-ID-09` | Отключить User | `ANALYZED` |
| `ID-FR-013` | functional | Must | `CAP-ID-09` | Восстановить User | `ANALYZED` |
| `ID-FR-014` | functional | Must | `CAP-ID-09` | Заблокировать User по безопасности | `ANALYZED` |
| `ID-FR-015` | functional | Must | `CAP-ID-02` | Получить User | `ANALYZED` |
| `ID-FR-016` | functional | Must | `CAP-ID-07` | Искать User | `ANALYZED` |
| `ID-FR-020` | functional | Must | `CAP-ID-06` | Связать External Identity | `ANALYZED` |
| `ID-FR-021` | functional | Must | `CAP-ID-06` | Обнаружить конфликт External Identity | `ANALYZED` |
| `ID-FR-022` | functional | Must | `CAP-ID-06` | Отвязать External Identity | `ANALYZED` |
| `ID-FR-023` | functional | Must | `CAP-ID-07` | Разрешить Subject по issuer+subject | `ANALYZED` |
| `ID-FR-024` | functional | Must | `CAP-ID-07` | Разрешить Subject по email или phone | `ANALYZED` |
| `ID-FR-030` | functional | Must | `CAP-ID-04` | Создать Group | `ANALYZED` |
| `ID-FR-031` | functional | Must | `CAP-ID-04` | Изменить Group | `ANALYZED` |
| `ID-FR-032` | functional | Must | `CAP-ID-05` | Добавить Membership | `ANALYZED` |
| `ID-FR-033` | functional | Must | `CAP-ID-05` | Удалить Membership | `ANALYZED` |
| `ID-FR-034` | functional | Must | `CAP-ID-05` | Перечислить Membership | `ANALYZED` |
| `ID-FR-040` | functional | Must | `CAP-ID-08` | Объединить дубли User | `ANALYZED` |
| `ID-FR-041` | functional | Must | `CAP-ID-08` | Обнаружить потенциальные дубли | `ANALYZED` |
| `ID-FR-050` | functional | Must | `CAP-ID-10` | Обезличить или удалить User | `ANALYZED` |
| `ID-FR-060` | functional | Must | `CAP-ID-10` | Публиковать факты жизненного цикла | `ANALYZED` |
| `ID-DATA-001` | data | Must | `CAP-ID-07` | Источник истины внутреннего Subject | `ANALYZED` |
| `ID-DATA-002` | data | Must | `CAP-ID-03` | Классификация атрибутов профиля | `ANALYZED` |
| `ID-SEC-001` | security | Must | `CAP-ID-08` | Защита операций merge и deletion | `ANALYZED` |
| `ID-NFR-001` | non-functional | Must | `CAP-ID-07` | Задержка разрешения Subject | `ANALYZED` |

## Детальные требования

### ID-FR-001. Создать User Pool {#id-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Создать изолированный User Pool в Project с заданной схемой профиля и политикой жизненного цикла.

**Критерии приёмки.**

- `ID-FR-001-AC-01` — Project существует и ACTIVE.
- `ID-FR-001-AC-02` — User Pool ID уникален и публикуется UserPoolCreated.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-001
capability: CAP-ID-01
owner_context: Identity
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

### ID-FR-002. Изменить User Pool {#id-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Изменять настройки User Pool с version check и контролем совместимости схемы.

**Критерии приёмки.**

- `ID-FR-002-AC-01` — Несовместимое удаление обязательного атрибута требует migration plan.
- `ID-FR-002-AC-02` — Изменение не затрагивает другие User Pool.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-002
capability: CAP-ID-01
owner_context: Identity
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

### ID-FR-003. Приостановить User Pool {#id-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Приостановить создание и аутентификацию новых субъектов в User Pool без удаления пользователей.

**Критерии приёмки.**

- `ID-FR-003-AC-01` — Существующие ссылки сохраняются.
- `ID-FR-003-AC-02` — Authentication получает состояние через API или projection event.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-003
capability: CAP-ID-01
owner_context: Identity
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

### ID-FR-010. Создать User {#id-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Создать User в одном User Pool с уникальным внутренним ID и валидным профилем.

**Критерии приёмки.**

- `ID-FR-010-AC-01` — Профиль соответствует активной schema version.
- `ID-FR-010-AC-02` — Повтор с тем же idempotency key не создаёт дубль.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-010
capability: CAP-ID-02
owner_context: Identity
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

### ID-FR-011. Изменить профиль User {#id-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Частично изменять разрешённые атрибуты профиля через FieldMask.

**Критерии приёмки.**

- `ID-FR-011-AC-01` — Неуказанные атрибуты не изменяются.
- `ID-FR-011-AC-02` — Чувствительные изменения фиксируются в Audit без старого секретного значения.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-011
capability: CAP-ID-03
owner_context: Identity
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

### ID-FR-012. Отключить User {#id-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Перевести User в DISABLED с указанием причины и прекратить новые authentication flows.

**Критерии приёмки.**

- `ID-FR-012-AC-01` — Identity публикует UserDisabled.
- `ID-FR-012-AC-02` — Существующие Access relationships не удаляются скрыто.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-012
capability: CAP-ID-09
owner_context: Identity
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

### ID-FR-013. Восстановить User {#id-fr-013}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Вернуть допустимого User в ACTIVE после авторизации и необходимых проверок.

**Критерии приёмки.**

- `ID-FR-013-AC-01` — Восстановление невозможно после окончательной privacy erasure.
- `ID-FR-013-AC-02` — Действие аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-013
capability: CAP-ID-09
owner_context: Identity
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

### ID-FR-014. Заблокировать User по безопасности {#id-fr-014}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Установить security lock с причиной, источником и сроком или ручным снятием.

**Критерии приёмки.**

- `ID-FR-014-AC-01` — Authentication отклоняет новый flow.
- `ID-FR-014-AC-02` — Снятие lock требует отдельного permission.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-014
capability: CAP-ID-09
owner_context: Identity
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

### ID-FR-015. Получить User {#id-fr-015}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Получить User и профиль в пределах разрешённого User Pool.

**Критерии приёмки.**

- `ID-FR-015-AC-01` — Ответ минимизирован по permission и purpose.
- `ID-FR-015-AC-02` — Недоступный User не раскрывается.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-015
capability: CAP-ID-02
owner_context: Identity
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

### ID-FR-016. Искать User {#id-fr-016}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Искать User по нормализованным поддерживаемым идентификаторам и атрибутам.

**Критерии приёмки.**

- `ID-FR-016-AC-01` — Результаты ограничены scope и permission.
- `ID-FR-016-AC-02` — Нечёткий поиск по чувствительным данным выключен по умолчанию.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-016
capability: CAP-ID-07
owner_context: Identity
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

### ID-FR-020. Связать External Identity {#id-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Связать пару issuer+subject с одним внутренним User.

**Критерии приёмки.**

- `ID-FR-020-AC-01` — Пара issuer+subject уникальна глобально в области политики.
- `ID-FR-020-AC-02` — Конфликт не приводит к автоматическому merge.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-020
capability: CAP-ID-06
owner_context: Identity
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

### ID-FR-021. Обнаружить конфликт External Identity {#id-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Обнаружить попытку связать уже занятую внешнюю идентичность и создать контролируемый conflict record.

**Критерии приёмки.**

- `ID-FR-021-AC-01` — Возвращается доменная ошибка IDENTITY_CONFLICT.
- `ID-FR-021-AC-02` — Событие не раскрывает данные другого пользователя вызывающему клиенту.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-021
capability: CAP-ID-06
owner_context: Identity
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

### ID-FR-022. Отвязать External Identity {#id-fr-022}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Удалить связь внешней идентичности при наличии альтернативного безопасного способа восстановления.

**Критерии приёмки.**

- `ID-FR-022-AC-01` — Нельзя оставить пользователя без разрешённого способа входа, если policy это запрещает.
- `ID-FR-022-AC-02` — Отвязка аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-022
capability: CAP-ID-06
owner_context: Identity
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

### ID-FR-023. Разрешить Subject по issuer+subject {#id-fr-023}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Вернуть SubjectReference и состояние пользователя по проверенному внешнему идентификатору.

**Критерии приёмки.**

- `ID-FR-023-AC-01` — Disabled/erased состояния возвращаются явно для доверенного caller.
- `ID-FR-023-AC-02` — Профиль не возвращается без отдельной необходимости.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-023
capability: CAP-ID-07
owner_context: Identity
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

### ID-FR-024. Разрешить Subject по email или phone {#id-fr-024}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Разрешать субъекта по нормализованному адресу только в политиках, где идентификатор считается подтверждённым.

**Критерии приёмки.**

- `ID-FR-024-AC-01` — Неподтверждённый адрес не используется как надёжная внешняя идентичность.
- `ID-FR-024-AC-02` — Неоднозначный результат возвращает AMBIGUOUS_SUBJECT.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-024
capability: CAP-ID-07
owner_context: Identity
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

### ID-FR-030. Создать Group {#id-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Создать Group внутри User Pool или ресурсной области.

**Критерии приёмки.**

- `ID-FR-030-AC-01` — Group имеет стабильный ID.
- `ID-FR-030-AC-02` — Имя уникально в документированном scope.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-030
capability: CAP-ID-04
owner_context: Identity
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

### ID-FR-031. Изменить Group {#id-fr-031}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Изменять метаданные Group с optimistic concurrency.

**Критерии приёмки.**

- `ID-FR-031-AC-01` — Membership не меняется через update Group.
- `ID-FR-031-AC-02` — Revision увеличивается один раз на commit.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-031
capability: CAP-ID-04
owner_context: Identity
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

### ID-FR-032. Добавить Membership {#id-fr-032}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Добавить User или Group membership в допустимую область.

**Критерии приёмки.**

- `ID-FR-032-AC-01` — Дубликат не создаёт вторую связь.
- `ID-FR-032-AC-02` — Membership не является RoleBinding и не предоставляет permission сам по себе.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-032
capability: CAP-ID-05
owner_context: Identity
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

### ID-FR-033. Удалить Membership {#id-fr-033}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Удалить membership без прямого изменения Access relationships.

**Критерии приёмки.**

- `ID-FR-033-AC-01` — Публикуется MembershipRemoved.
- `ID-FR-033-AC-02` — Потребители применяют собственную policy обработки.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-033
capability: CAP-ID-05
owner_context: Identity
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

### ID-FR-034. Перечислить Membership {#id-fr-034}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Получать memberships субъекта или области с пагинацией и access filtering.

**Критерии приёмки.**

- `ID-FR-034-AC-01` — Результат не раскрывает недоступные области.
- `ID-FR-034-AC-02` — Page token связан с фильтром.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-034
capability: CAP-ID-05
owner_context: Identity
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

### ID-FR-040. Объединить дубли User {#id-fr-040}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Провести контролируемое объединение двух User с выбором surviving identity и планом переноса ссылок.

**Критерии приёмки.**

- `ID-FR-040-AC-01` — Операция является длительной и обратимой до commit point.
- `ID-FR-040-AC-02` — Конфликты профиля и external identity разрешаются явно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-040
capability: CAP-ID-08
owner_context: Identity
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

### ID-FR-041. Обнаружить потенциальные дубли {#id-fr-041}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Формировать кандидатов на merge без автоматического объединения.

**Критерии приёмки.**

- `ID-FR-041-AC-01` — Кандидат содержит объяснимые matching signals.
- `ID-FR-041-AC-02` — Доступ к candidates ограничен специальным permission.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-041
capability: CAP-ID-08
owner_context: Identity
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

### ID-FR-050. Обезличить или удалить User {#id-fr-050}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Выполнить privacy deletion или anonymization по policy с сохранением минимальных исторических ссылок.

**Критерии приёмки.**

- `ID-FR-050-AC-01` — Активные credentials/sessions отзываются через интеграционный процесс.
- `ID-FR-050-AC-02` — Audit сохраняет actor-independent evidence без лишних персональных данных.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-050
capability: CAP-ID-10
owner_context: Identity
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

### ID-FR-060. Публиковать факты жизненного цикла {#id-fr-060}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Resource Manager, Access, Audit |
| Данные | Identity aggregate |
| Безопасность | permission check, privacy minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.9, §7, §9.3, §11, §15 |

**Требование.**

Публиковать versioned события о User Pool, User, External Identity, Group и Membership.

**Критерии приёмки.**

- `ID-FR-060-AC-01` — События содержат owner revision.
- `ID-FR-060-AC-02` — Профильные атрибуты включаются только при документированной необходимости.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-FR-060
capability: CAP-ID-10
owner_context: Identity
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

### ID-DATA-001. Источник истины внутреннего Subject {#id-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Identity должен быть единственным владельцем связи внутренних пользователей, User Pool и внешних идентичностей.

**Критерии приёмки.**

- `ID-DATA-001-AC-01` — Authentication хранит SubjectReference, а не копию профиля.
- `ID-DATA-001-AC-02` — Access хранит subject ID и тип, но не состояние профиля как authoritative data.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-DATA-001
capability: CAP-ID-07
owner_context: Identity
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

### ID-DATA-002. Классификация атрибутов профиля {#id-data-002}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11, §15 |

**Требование.**

Каждый профильный атрибут должен иметь classification, validation, mutability, retention и disclosure policy.

**Критерии приёмки.**

- `ID-DATA-002-AC-01` — Неизвестный атрибут отклоняется или хранится только в явно разрешённом extension namespace.
- `ID-DATA-002-AC-02` — Чувствительные поля не появляются в событиях по умолчанию.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-DATA-002
capability: CAP-ID-03
owner_context: Identity
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

### ID-SEC-001. Защита операций merge и deletion {#id-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Merge, unlink last identity и privacy deletion должны требовать усиленного permission и, по политике, step-up.

**Критерии приёмки.**

- `ID-SEC-001-AC-01` — Недостаточный AAL приводит к CHALLENGE, а не к частичному выполнению.
- `ID-SEC-001-AC-02` — Каждый шаг длительной операции аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-SEC-001
capability: CAP-ID-08
owner_context: Identity
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

### ID-NFR-001. Задержка разрешения Subject {#id-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Identity` / `m8-identity` |
| Business capability | `CAP-ID-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | p95<=120ms, availability>=99.95% |
| Основание PADS | §19 |

**Требование.**

ResolveSubject должен иметь p95 не более 120 мс и доступность не ниже 99,95% в основном регионе.

**Критерии приёмки.**

- `ID-NFR-001-AC-01` — Метрика разделяется по типу идентификатора.
- `ID-NFR-001-AC-02` — Кэш не возвращает пользователя после получения UserDisabled beyond freshness target.

**Трассировка для следующего этапа:**

```yaml
requirement_id: ID-NFR-001
capability: CAP-ID-07
owner_context: Identity
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
