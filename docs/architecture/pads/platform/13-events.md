---
title: "PADS: правила проектирования событий"
description: "События, outbox, envelope, версионирование и потребители."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 13. Правила проектирования событий {#pads-events}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 12. Правила проектирования API](12-api-design.md) | [Следующий раздел: 14. Модель интеграции и согласованности](14-integration-consistency.md)

{% endnote %}

## 13.1. Назначение главы

Настоящая глава определяет модель Domain Events и Integration Events M8 Platform, их схемы, envelope, семантику доставки, ordering, partitioning, versioning, дедупликацию, replay и совместимость.

Событие является неизменяемым утверждением о факте, который уже произошёл. Событие не является удалённым вызовом команды.

## 13.2. Классы сообщений

| Класс | Назначение | Владелец схемы | Пример |
| --- | --- | --- | --- |
| Domain Event | факт внутри ограниченного контекста | domain owner | `AuthenticationChallengeCompleted` |
| Integration Event | опубликованное представление факта для других контекстов | producer context | `UserDisabled` |
| Audit Event | факт ответственности: кто и что сделал | Audit contract + producer | `RoleBindingCreated` с actor/outcome |
| Command Message | просьба выполнить действие | receiving context | `DeleteProjectData` |
| Notification | сообщение пользователю/оператору | notification capability | `ApprovalRequired` |

Domain Event MAY быть преобразован в Integration Event. Их payload и срок хранения не обязаны совпадать.

## 13.3. Нормативные правила

| ID | Правило |
| --- | --- |
| `EVT-001` | Event name MUST обозначать свершившийся факт в прошедшем времени. |
| `EVT-002` | Event MUST быть неизменяемым после публикации. |
| `EVT-003` | Producer MUST владеть фактом, который публикует. |
| `EVT-004` | Команда MUST NOT маскироваться под событие. |
| `EVT-005` | Integration Event MUST публиковаться только после commit авторитетного состояния. |
| `EVT-006` | State change и Outbox record MUST фиксироваться атомарно. |
| `EVT-007` | Consumer MUST быть идемпотентным. |
| `EVT-008` | Event envelope MUST содержать event ID, type, version, producer, occurrence time, correlation и causation. |
| `EVT-009` | Ordering гарантируется только в явно объявленной partition scope. |
| `EVT-010` | Consumer MUST NOT предполагать глобальный порядок событий. |
| `EVT-011` | Duplicate delivery является нормальной ситуацией. |
| `EVT-012` | Exactly-once end-to-end MUST NOT заявляться без доказуемой транзакционной границы. |
| `EVT-013` | Event schema MUST быть backward-compatible внутри major version. |
| `EVT-014` | Payload MUST содержать достаточный факт, но не обязан быть полным snapshot ресурса. |
| `EVT-015` | Секреты, token material и необязательные персональные данные MUST NOT включаться в событие. |
| `EVT-016` | Delete/tombstone events MUST быть поддержаны критичными проекциями. |
| `EVT-017` | Event retention и replay policy MUST быть определены для каждого stream. |
| `EVT-018` | Poison event MUST изолироваться без остановки всей partition навсегда. |
| `EVT-019` | Consumer side effects MUST иметь inbox/deduplication или эквивалент. |
| `EVT-020` | Event contract MUST быть зарегистрирован в Event Catalog и связан с requirements. |

## 13.4. Стандартный envelope

```protobuf
message EventEnvelope {
  string event_id = 1;
  string event_type = 2;
  uint32 schema_version = 3;
  string producer = 4;
  google.protobuf.Timestamp occurred_at = 5;
  google.protobuf.Timestamp published_at = 6;
  string correlation_id = 7;
  string causation_id = 8;
  string trace_id = 9;
  string tenant_scope = 10;
  m8.platform.common.v1.ActorReference actor = 11;
  m8.platform.common.v1.ResourceReference resource = 12;
  map<string, string> attributes = 13;
  bytes payload = 14;
}
```

Реализация MAY использовать typed wrapper вместо raw bytes, но envelope semantics MUST сохраняться.

## 13.5. Event ID

`event_id` MUST:

- быть глобально уникальным;
- генерироваться producer;
- сохраняться при повторной публикации того же логического события;
- использоваться consumer Inbox;
- попадать в telemetry и audit diagnostics.

Создание нового event ID при каждом retry публикации запрещено.

## 13.6. Event type и именование

Рекомендуемый формат:

```text
m8.<context>.<aggregate>.<fact>.v<major>
```

Примеры:

```text
m8.resourcemanager.project.created.v1
m8.identity.user.disabled.v1
m8.authentication.transaction.completed.v1
m8.access.role_binding.created.v1
m8.riskdecision.assessment.decided.v1
m8.provisioning.managed_resource.reconciled.v1
```

Имя MUST быть стабильным и не содержать название команды, таблицы или vendor technology.

## 13.7. Occurred time и published time

- `occurred_at` — момент предметного факта или фиксации транзакции;
- `published_at` — момент отправки transport producer;
- consumer processing time хранится отдельно.

Producer MUST не изменять `occurred_at` при retry. Clock skew SHOULD отслеживаться.

## 13.8. Correlation и causation

- `correlation_id` связывает весь сквозной процесс;
- `causation_id` указывает непосредственный request, command, event или activity;
- `trace_id` связывает событие с distributed trace, но не заменяет correlation.

Событие, вызванное другим событием, MUST указать его `event_id` как causation либо типизированную ссылку на command ID.

## 13.9. Payload design

Payload SHOULD содержать:

- идентификатор ресурса;
- новое состояние или изменённые факты;
- revision;
- минимальные данные для типичных consumers;
- reason code, если он является частью предметного факта.

Payload MUST NOT:

- раскрывать внутреннюю storage schema;
- требовать lookup по неустойчивому внутреннему ключу;
- включать необработанный access token;
- использовать свободный JSON там, где существует стабильная typed schema;
- содержать полные snapshots по умолчанию без оценки privacy и размера.

## 13.10. Delta и snapshot events

| Тип | Когда применять |
| --- | --- |
| Delta event | изменение ограниченного набора полей; consumers умеют применять по revision |
| State event | новое значимое состояние агрегата |
| Snapshot event | bootstrap, восстановление или редкая полная синхронизация |
| Tombstone event | удаление/анонимизация/недоступность |

Delta event MUST содержать revision и SHOULD содержать changed field paths.

## 13.11. Topics и streams

Topic boundary SHOULD соответствовать:

- контексту;
- требованиям retention;
- классу конфиденциальности;
- ordering key;
- профилю throughput;
- кругу consumers.

Нельзя объединять restricted security events и широкодоступные resource lifecycle events только ради уменьшения числа topics.

## 13.12. Partitioning

Partition key выбирается по агрегату или ресурсу, для которого нужен порядок.

| Event family | Рекомендуемый ключ |
| --- | --- |
| Project lifecycle | `project_id` |
| User lifecycle | `user_id` |
| Authentication transaction | `authentication_id` |
| Access relationship | canonical resource scope или relation key |
| Managed resource | `managed_resource_id` |
| Audit | organization/project scope + time bucket при больших объёмах |

Hot partition risk MUST быть оценён. Глобальный organization ID как единственный ключ может быть неприемлем для крупного tenant.

## 13.13. Ordering

Гарантия порядка действует только внутри partition и transport contract. Consumer MUST:

- сравнивать revision;
- игнорировать устаревшее событие;
- обнаруживать пропуски, если sequence обязателен;
- уметь запрашивать snapshot/reconciliation;
- не полагаться на порядок между разными aggregates.

## 13.14. Delivery semantics

Базовая семантика — at-least-once. Это означает:

- duplicate events возможны;
- publish delay возможен;
- reorder между partitions возможен;
- consumer retry возможен после частичного side effect.

Для каждого side effect нужна идемпотентная модель.

## 13.15. Outbox

Outbox record MUST содержать:

```yaml
outbox:
  event_id: evt_...
  aggregate_type: Project
  aggregate_id: prj_...
  aggregate_revision: 42
  event_type: m8.resourcemanager.project.updated.v1
  payload: protobuf-bytes
  created_at: ...
  publish_status: pending
  attempt: 0
```

Dispatcher MUST:

- публиковать без изменения event ID;
- применять retry with backoff;
- фиксировать published_at;
- иметь lag metrics;
- поддерживать безопасный replay;
- не блокировать foreground transaction transport-вызовом.

## 13.16. Inbox и дедупликация

Consumer Inbox SHOULD хранить:

- event ID;
- consumer ID;
- processing status;
- first/last attempt;
- result digest;
- expiry;
- side effect transaction marker.

Если consumer state и Inbox находятся в одной БД, они SHOULD фиксироваться одной транзакцией.

## 13.17. Consumer contract

Каждый consumer MUST определить:

- интересующие event types;
- поддерживаемые schema versions;
- idempotency key;
- ordering assumptions;
- max tolerated lag;
- retryable/non-retryable errors;
- dead-letter procedure;
- rebuild procedure;
- owner team;
- dashboard/runbook.

## 13.18. Schema evolution

Совместимые изменения:

- добавление optional field;
- добавление enum value при unknown-safe consumers;
- расширение metadata;
- публикация нового event type.

Несовместимые:

- изменение смысла поля;
- изменение units;
- удаление обязательного consumer field;
- изменение partition key semantics;
- reuse field number;
- смена идентификатора ресурса.

При несовместимости публикуется новый event type major version. Dual-publish MAY использоваться в ограниченный миграционный период.

## 13.19. Consumer-driven compatibility

Перед изменением schema producer SHOULD запускать:

- buf breaking;
- registered consumer contract tests;
- replay sample validation;
- size regression checks;
- privacy checks.

Список активных consumers MUST быть известен Event Catalog.

## 13.20. Replay

Replay MUST быть отделён от обычной публикации и контролировать:

- диапазон времени/offset;
- target consumer;
- rate limit;
- side effects;
- deduplication mode;
- audit trail;
- dry-run;
- остановку и возобновление.

Consumer MUST различать replay mode, если повтор внешних side effects опасен.

## 13.21. Dead-letter и poison events

Non-retryable event помещается в quarantine/DLQ с:

- original envelope;
- consumer;
- error code;
- attempt history;
- first/last failure time;
- schema information;
- remediation state.

DLQ MUST иметь owner, alert и runbook. Автоматический бесконечный retry запрещён.

## 13.22. Retention

Retention определяется назначением:

- operational integration — достаточный срок для outage recovery;
- projection rebuild — срок, покрывающий полную реконструкцию, либо snapshot;
- audit — отдельная политика Audit;
- analytics — governed warehouse retention.

Transport retention не заменяет долгосрочное нормативное хранение.

## 13.23. Privacy и security

Event producer MUST проводить data minimization. Restricted fields SHOULD заменяться:

- stable pseudonymous ID;
- hash/digest;
- reference;
- reason code;
- redacted representation.

Topics MUST иметь ACL по producer/consumer identity и encryption in transit/at rest.

## 13.24. Audit Events

Audit Event отличается от Integration Event:

- Integration Event сообщает потребителям об изменении состояния;
- Audit Event доказывает действие, actor, authorization/risk context и outcome.

Один use case MAY создавать оба события через разные schemas и retention.

## 13.25. Event Catalog

Запись каталога:

```yaml
event:
  id: EVT-ID-USER-DISABLED-V1
  type: m8.identity.user.disabled.v1
  owner: m8-identity
  aggregate: User
  partition_key: user_id
  retention: 30d
  classification: confidential
  schema: m8.platform.identity.events.v1.UserDisabled
  consumers:
    - m8-authentication
    - m8-access
    - m8-audit
  rebuild_source: Identity Export API
  requirement_ids: [ID-FR-003]
```

## 13.26. Наблюдаемость

Producer metrics:

- outbox pending count;
- publish latency;
- publish errors;
- event size;
- partition skew.

Consumer metrics:

- consumer lag;
- processing latency;
- duplicate count;
- retry count;
- DLQ count;
- stale revision count;
- reconciliation mismatch.

## 13.27. Тестирование

Обязательны:

- schema compatibility tests;
- envelope validation;
- duplicate delivery tests;
- out-of-order tests;
- replay tests;
- tombstone tests;
- poison event tests;
- privacy tests;
- projection rebuild tests;
- producer-consumer contract tests.

## 13.28. SPDD-требования

Event Structured Prompt MUST объявлять:

- fact semantics;
- owner aggregate;
- event type/version;
- envelope;
- payload;
- partition key;
- ordering assumption;
- data classification;
- consumers;
- Outbox/Inbox behavior;
- compatibility and tests.

## 13.29. Критерии соответствия главы

Событийная интеграция соответствует PADS, если опубликованы только подтверждённые факты, schemas versioned, delivery at-least-once учтена, consumers idempotent, ordering scope явна, replay/DLQ предусмотрены, секреты отсутствуют, а событие зарегистрировано и трассируется до requirements.

---
