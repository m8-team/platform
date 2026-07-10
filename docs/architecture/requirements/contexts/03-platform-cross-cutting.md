---
title: "Requirements Catalog: платформенные и сквозные требования"
description: "Платформенные функциональные, архитектурные, security и NFR требования."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 3. Платформенные и сквозные требования {#requirements-platform-cross-cutting}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 2. Сводная карта распределения](../overview/02-distribution-map.md) | [Следующий раздел: 4. Resource Manager](04-resource-manager.md)

{% endnote %}

Владелец раздела: **Platform**. Требований: **20**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `PLT-FR-001` | functional | Must | `CAP-RM-05` | Единая ресурсная область запроса | `ANALYZED` |
| `PLT-FR-002` | functional | Must | `CAP-OBS-08` | Единый контекст вызова | `ANALYZED` |
| `PLT-FR-003` | functional | Must | `CAP-OPS-01` | Единая модель длительных операций | `ANALYZED` |
| `PLT-FR-004` | functional | Must | `CAP-AUD-01` | Обязательный аудит значимых изменений | `ANALYZED` |
| `PLT-FR-005` | functional | Must | `CAP-AUTHZ-06` | Обязательная проверка доступа | `ANALYZED` |
| `PLT-FR-006` | functional | Must | `CAP-RISK-02` | Проверка риска для чувствительных действий | `ANALYZED` |
| `PLT-FR-007` | functional | Must | `CAP-OPS-10` | Идемпотентные мутации | `ANALYZED` |
| `PLT-FR-008` | functional | Must | `CAP-GOV-05` | Версионируемые публичные контракты | `ANALYZED` |
| `PLT-ARC-001` | architecture | Must | `CAP-GOV-07` | База данных принадлежит одному сервису | `ANALYZED` |
| `PLT-ARC-002` | architecture | Must | `CAP-INT-04` | Transactional Outbox для интеграционных фактов | `ANALYZED` |
| `PLT-ARC-003` | architecture | Must | `CAP-INT-05` | Inbox и дедупликация потребителей | `ANALYZED` |
| `PLT-ARC-004` | architecture | Must | `CAP-INT-09` | Отсутствие распределённых транзакций | `ANALYZED` |
| `PLT-ARC-005` | architecture | Must | `CAP-INT-10` | Антикоррупционный слой внешних систем | `ANALYZED` |
| `PLT-SEC-001` | security | Must | `CAP-OBS-02` | Запрет секретов в наблюдаемости | `ANALYZED` |
| `PLT-SEC-002` | security | Must | `CAP-AUTHZ-06` | Взаимная идентификация сервисов | `ANALYZED` |
| `PLT-SEC-003` | security | Must | `CAP-ID-03` | Минимизация персональных данных | `ANALYZED` |
| `PLT-SEC-004` | security | Must | `CAP-AUD-02` | Неподделываемый actor context | `ANALYZED` |
| `PLT-NFR-001` | non-functional | Must | `CAP-OBS-05` | Доступность критических API | `ANALYZED` |
| `PLT-NFR-002` | non-functional | Must | `CAP-OBS-01` | Сквозная трассировка | `ANALYZED` |
| `PLT-NFR-003` | non-functional | Must | `CAP-INT-07` | Восстановление после сбоя | `ANALYZED` |

## Детальные требования

### PLT-FR-001. Единая ресурсная область запроса {#plt-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-RM-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | scope validation, authorization required |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §1.5, §8, §15 |

**Требование.**

Каждый публичный запрос должен однозначно определять Organization, Workspace или Project scope либо явно работать с глобальной платформенной областью.

**Критерии приёмки.**

- `PLT-FR-001-AC-01` — Запрос без обязательной области отклоняется до выполнения бизнес-операции.
- `PLT-FR-001-AC-02` — Область передаётся в авторизацию, аудит, трассировку и предметную операцию без неявного изменения.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-001
capability: CAP-RM-05
owner_context: Platform
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

### PLT-FR-002. Единый контекст вызова {#plt-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-OBS-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | RequestContext |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §18 |

**Требование.**

Все публичные и межсервисные операции должны поддерживать RequestContext с request_id, correlation_id, actor, client и resource scope.

**Критерии приёмки.**

- `PLT-FR-002-AC-01` — Идентификаторы сохраняются при синхронных и асинхронных переходах.
- `PLT-FR-002-AC-02` — Отсутствующие технические идентификаторы создаются на границе входа, но actor и scope не выдумываются.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-002
capability: CAP-OBS-08
owner_context: Platform
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

### PLT-FR-003. Единая модель длительных операций {#plt-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all operation owners` |
| Business capability | `CAP-OPS-01` |
| Согласованность | C1 — локальная транзакция и последующая длительная работа |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §16 |

**Требование.**

Операция, не завершаемая надёжно в пределах синхронного запроса, должна возвращать ресурс Operation.

**Критерии приёмки.**

- `PLT-FR-003-AC-01` — Повтор команды с тем же idempotency key возвращает ту же Operation.
- `PLT-FR-003-AC-02` — Operation отделена от состояния предметного ресурса и содержит типизированный result или error.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-003
capability: CAP-OPS-01
owner_context: Platform
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

### PLT-FR-004. Обязательный аудит значимых изменений {#plt-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-AUD-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | AuditEvent.v1 |
| Атрибуты качества | — |
| Основание PADS | §14, §15, §18 |

**Требование.**

Каждое значимое изменение состояния, назначение полномочий и решение безопасности должно формировать AuditEvent.

**Критерии приёмки.**

- `PLT-FR-004-AC-01` — AuditEvent содержит actor, target, action, outcome и correlation_id.
- `PLT-FR-004-AC-02` — Секреты и необработанные учётные данные в аудит не попадают.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-004
capability: CAP-AUD-01
owner_context: Platform
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

### PLT-FR-005. Обязательная проверка доступа {#plt-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all public services` |
| Business capability | `CAP-AUTHZ-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | Access check, fail closed |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §8, §15 |

**Требование.**

Каждая публичная операция над защищённым ресурсом должна проверять разрешение до изменения состояния или раскрытия данных.

**Критерии приёмки.**

- `PLT-FR-005-AC-01` — Отсутствие решения Access приводит к fail-closed для mutation и чувствительного чтения.
- `PLT-FR-005-AC-02` — Проверяемые permission и resource reference фиксируются в трассе и аудите.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-005
capability: CAP-AUTHZ-06
owner_context: Platform
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

### PLT-FR-006. Проверка риска для чувствительных действий {#plt-fr-006}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `sensitive action owners` |
| Business capability | `CAP-RISK-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | risk evaluation |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Операции, помеченные как чувствительные, должны получить решение Risk Decision до выполнения необратимого действия.

**Критерии приёмки.**

- `PLT-FR-006-AC-01` — DENY блокирует выполнение.
- `PLT-FR-006-AC-02` — CHALLENGE приводит к step-up и повторной оценке, а не к обходу решения.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-006
capability: CAP-RISK-02
owner_context: Platform
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

### PLT-FR-007. Идемпотентные мутации {#plt-fr-007}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-OPS-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | IdempotencyRecord |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §4, §12, §14 |

**Требование.**

Все внешне вызываемые мутации должны поддерживать идемпотентность в пределах документированного окна.

**Критерии приёмки.**

- `PLT-FR-007-AC-01` — Повтор с тем же ключом и эквивалентным телом возвращает исходный результат.
- `PLT-FR-007-AC-02` — Повтор с тем же ключом и несовместимым телом отклоняется конфликтом.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-007
capability: CAP-OPS-10
owner_context: Platform
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

### PLT-FR-008. Версионируемые публичные контракты {#plt-fr-008}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `contract owners` |
| Business capability | `CAP-GOV-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §12, §13, §23 |

**Требование.**

Публичные API и Integration Events должны иметь стабильную версию и проходить проверку обратной совместимости.

**Критерии приёмки.**

- `PLT-FR-008-AC-01` — Удаление или изменение смысла опубликованного поля блокируется CI.
- `PLT-FR-008-AC-02` — Несовместимое изменение выпускается новой major-версией с планом миграции.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-FR-008
capability: CAP-GOV-05
owner_context: Platform
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

### PLT-ARC-001. База данных принадлежит одному сервису {#plt-arc-001}

| Поле | Значение |
| --- | --- |
| Тип | `architecture` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-GOV-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §4, §11 |

**Требование.**

Сервис не должен читать или изменять таблицы, принадлежащие другому ограниченному контексту.

**Критерии приёмки.**

- `PLT-ARC-001-AC-01` — Запрещённые подключения и импорты выявляются архитектурными проверками.
- `PLT-ARC-001-AC-02` — Межконтекстное чтение выполняется API, событием или локальной проекцией.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-ARC-001
capability: CAP-GOV-07
owner_context: Platform
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

### PLT-ARC-002. Transactional Outbox для интеграционных фактов {#plt-arc-002}

| Поле | Значение |
| --- | --- |
| Тип | `architecture` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `event publishers` |
| Business capability | `CAP-INT-04` |
| Согласованность | C1 — локальная атомарность, затем C2 — итоговая доставка |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §13, §14 |

**Требование.**

Изменение агрегата и намерение публикации обязательного события должны фиксироваться атомарно через Outbox.

**Критерии приёмки.**

- `PLT-ARC-002-AC-01` — Сбой брокера после commit не приводит к потере события.
- `PLT-ARC-002-AC-02` — Публикация до commit агрегата запрещена.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-ARC-002
capability: CAP-INT-04
owner_context: Platform
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

### PLT-ARC-003. Inbox и дедупликация потребителей {#plt-arc-003}

| Поле | Значение |
| --- | --- |
| Тип | `architecture` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `event consumers` |
| Business capability | `CAP-INT-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §13, §14 |

**Требование.**

Потребители интеграционных событий должны быть устойчивы к повторной и неупорядоченной доставке в заявленных границах.

**Критерии приёмки.**

- `PLT-ARC-003-AC-01` — Повтор event_id не создаёт повторный предметный эффект.
- `PLT-ARC-003-AC-02` — Устаревшая revision не перезаписывает более новое состояние проекции.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-ARC-003
capability: CAP-INT-05
owner_context: Platform
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

### PLT-ARC-004. Отсутствие распределённых транзакций {#plt-arc-004}

| Поле | Значение |
| --- | --- |
| Тип | `architecture` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-INT-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §14 |

**Требование.**

Согласованность между сервисами должна достигаться процессом, событиями, компенсацией или reconciliation без двухфазной фиксации.

**Критерии приёмки.**

- `PLT-ARC-004-AC-01` — Межсервисный процесс имеет владельца и явные состояния.
- `PLT-ARC-004-AC-02` — Необратимые шаги имеют порядок, timeout и ручное восстановление.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-ARC-004
capability: CAP-INT-09
owner_context: Platform
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

### PLT-ARC-005. Антикоррупционный слой внешних систем {#plt-arc-005}

| Поле | Значение |
| --- | --- |
| Тип | `architecture` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `integration owners` |
| Business capability | `CAP-INT-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §8, §9, §14 |

**Требование.**

Модели Keycloak, SpiceDB, Temporal, Kubernetes и облачных SDK не должны становиться доменными типами M8.

**Критерии приёмки.**

- `PLT-ARC-005-AC-01` — Внешняя ошибка преобразуется в каноническую ошибку M8.
- `PLT-ARC-005-AC-02` — Замена поставщика не требует изменения агрегатов и публичных контрактов.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-ARC-005
capability: CAP-INT-10
owner_context: Platform
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

### PLT-SEC-001. Запрет секретов в наблюдаемости {#plt-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-OBS-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15, §18 |

**Требование.**

Токены, пароли, OTP, приватные ключи и секреты поставщиков не должны попадать в логи, трассы, метрики и AuditEvent.

**Критерии приёмки.**

- `PLT-SEC-001-AC-01` — Автоматический secret scanning не обнаруживает секретные поля в telemetry fixtures.
- `PLT-SEC-001-AC-02` — Поля маскируются до сериализации события или записи журнала.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-SEC-001
capability: CAP-OBS-02
owner_context: Platform
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

### PLT-SEC-002. Взаимная идентификация сервисов {#plt-sec-002}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `platform runtime` |
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

Межсервисные вызовы должны выполняться от проверенной service identity с минимально необходимыми полномочиями.

**Критерии приёмки.**

- `PLT-SEC-002-AC-01` — Анонимный межсервисный mutation отклоняется.
- `PLT-SEC-002-AC-02` — Ротация service credential не требует остановки платформы.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-SEC-002
capability: CAP-AUTHZ-06
owner_context: Platform
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

### PLT-SEC-003. Минимизация персональных данных {#plt-sec-003}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `data owners` |
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

Каждый контекст должен хранить только персональные данные, необходимые для его ответственности, и использовать ссылки вместо копий профиля.

**Критерии приёмки.**

- `PLT-SEC-003-AC-01` — Новая копия персонального атрибута имеет owner, purpose, retention и deletion rule.
- `PLT-SEC-003-AC-02` — Privacy deletion обновляет или удаляет допустимые проекции.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-SEC-003
capability: CAP-ID-03
owner_context: Platform
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

### PLT-SEC-004. Неподделываемый actor context {#plt-sec-004}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `edge and service middleware` |
| Business capability | `CAP-AUD-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §10, §15 |

**Требование.**

Actor и Client в RequestContext должны происходить из проверенной аутентификации и не могут приниматься из недоверенного пользовательского поля.

**Критерии приёмки.**

- `PLT-SEC-004-AC-01` — Попытка подмены actor field не меняет effective actor.
- `PLT-SEC-004-AC-02` — Сервисные действия различают initiating actor и executing service.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-SEC-004
capability: CAP-AUD-02
owner_context: Platform
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

### PLT-NFR-001. Доступность критических API {#plt-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `critical service owners` |
| Business capability | `CAP-OBS-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | availability >= 99.95% monthly |
| Основание PADS | §18, §19 |

**Требование.**

Критические публичные операции должны достигать месячной доступности не ниже 99,95% за исключением согласованных окон обслуживания.

**Критерии приёмки.**

- `PLT-NFR-001-AC-01` — SLI рассчитывается по серверным результатам и исключает только документированные client errors.
- `PLT-NFR-001-AC-02` — Нарушение error budget блокирует рискованные релизы по принятой политике.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-NFR-001
capability: CAP-OBS-05
owner_context: Platform
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

### PLT-NFR-002. Сквозная трассировка {#plt-nfr-002}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `all services` |
| Business capability | `CAP-OBS-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | trace coverage >= 99% |
| Основание PADS | §18, §19 |

**Требование.**

Не менее 99% успешных и ошибочных критических запросов должны иметь связанную распределённую трассу.

**Критерии приёмки.**

- `PLT-NFR-002-AC-01` — Trace context проходит через брокер и Temporal workflow.
- `PLT-NFR-002-AC-02` — По operation_id можно найти исходный request_id и ключевые downstream spans.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-NFR-002
capability: CAP-OBS-01
owner_context: Platform
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

### PLT-NFR-003. Восстановление после сбоя {#plt-nfr-003}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Platform` / `service owners` |
| Business capability | `CAP-INT-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | RPO for committed domain data = 0 within replicated storage guarantees |
| Основание PADS | §14, §19 |

**Требование.**

После временного сбоя сервисы должны восстанавливать обработку без потери подтверждённых изменений и без повторного предметного эффекта.

**Критерии приёмки.**

- `PLT-NFR-003-AC-01` — Outbox backlog обрабатывается после восстановления.
- `PLT-NFR-003-AC-02` — Повтор workflow или consumer delivery не нарушает инварианты.

**Трассировка для следующего этапа:**

```yaml
requirement_id: PLT-NFR-003
capability: CAP-INT-07
owner_context: Platform
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
