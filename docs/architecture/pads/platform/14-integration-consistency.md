---
title: "PADS: интеграция и согласованность"
description: "Синхронные и асинхронные интеграции, согласованность и деградация."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 14. Модель интеграции и согласованности {#pads-integration-consistency}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 13. Правила проектирования событий](13-events.md) | [Следующий раздел: 15. Архитектура безопасности](15-security.md)

{% endnote %}

## 14.1. Назначение главы

Настоящая глава определяет, когда M8 Platform использует синхронные вызовы, события, локальные проекции, Saga, Process Manager, Temporal Workflow и reconciliation. Она также устанавливает транзакционные границы, модель отказов и допустимые уровни согласованности.

## 14.2. Базовая модель

M8 использует:

- **строгую локальную согласованность** внутри одного owner service;
- **итоговую межконтекстную согласованность** для распространения фактов;
- **синхронные decision API** для решений, которые нельзя принимать на устаревшей копии;
- **Temporal orchestration** для длительных, повторяемых и компенсируемых процессов;
- **reconciliation** для desired/observed state внешних ресурсов.

Распределённые ACID-транзакции между контекстами не применяются.

## 14.3. Принципы

| ID | Правило |
| --- | --- |
| `INT-001` | Один use case MUST иметь одну локальную транзакционную точку фиксации владельца. |
| `INT-002` | Межсервисный вызов MUST NOT выполняться внутри открытой database transaction, кроме обоснованного read-only precheck до mutation. |
| `INT-003` | Событие публикуется после commit через Outbox. |
| `INT-004` | Синхронный вызов используется только когда caller не может безопасно продолжить без немедленного ответа. |
| `INT-005` | Долгий или многошаговый процесс MUST иметь устойчивого владельца состояния. |
| `INT-006` | Retry MUST быть ограничен, идемпотентен и классифицирован по ошибкам. |
| `INT-007` | Timeout MUST быть короче caller deadline и учитывать budget всей цепочки. |
| `INT-008` | Синхронные циклы между контекстами запрещены. |
| `INT-009` | Локальная проекция используется только при объявленной свежести и безопасной семантике. |
| `INT-010` | Compensation MUST быть предметным действием, а не техническим откатом чужой транзакции. |
| `INT-011` | Temporal Workflow MUST не содержать недетерминированную логику вне Activities. |
| `INT-012` | Activity MUST быть идемпотентной или иметь deduplication token. |
| `INT-013` | Process state MUST быть наблюдаемым через Operation. |
| `INT-014` | External system state MUST reconciliate с desired state, а не считаться мгновенно согласованным. |
| `INT-015` | Failure mode и degraded behavior MUST быть определены до реализации интеграции. |
| `INT-016` | Integration contract MUST иметь owner, SLO, retry policy и runbook. |
| `INT-017` | Change of interaction pattern MUST проходить ADR и impact analysis. |
| `INT-018` | Caller MUST различать accepted, completed и observed states. |
| `INT-019` | Нельзя подтверждать бизнес-успех, если критичный локальный commit не выполнен. |
| `INT-020` | Structured Prompt MUST явно указывать consistency class и transaction boundary. |

## 14.4. Классы согласованности

| Класс | Описание | Пример |
| --- | --- | --- |
| `C0 Local Strong` | все инварианты фиксируются одной транзакцией владельца | создание User и Outbox event |
| `C1 Synchronous Decision` | требуется актуальное решение другого владельца | Access check, Risk Decision |
| `C2 Read-your-write` | caller сразу видит подтверждённое локальное состояние | Get Project после успешного local update |
| `C3 Eventual Projection` | производная копия обновляется событиями | Project metadata в Provisioning |
| `C4 Orchestrated` | многошаговый процесс с retry/compensation | удаление Project со всеми ресурсами |
| `C5 Reconciled External` | desired/observed state сходятся со временем | Kubernetes/Cloud resource provisioning |
| `C6 Analytical` | данные согласуются пакетно или потоково для анализа | platform metrics warehouse |

Каждое requirement MUST указывать требуемый класс, если рассогласование влияет на поведение.

## 14.5. Выбор механизма взаимодействия

| Вопрос | Если «да» | Механизм |
| --- | --- | --- |
| Нужен немедленный ответ для продолжения безопасной операции? | Да | синхронный API |
| Распространяется факт после commit? | Да | Integration Event |
| Нужна локальная низколатентная копия? | Да | Event projection + reconciliation |
| Процесс длится дольше request deadline? | Да | Operation + Temporal/worker |
| Есть несколько шагов и compensation? | Да | Saga/Temporal orchestration |
| Внешнее состояние может меняться независимо? | Да | reconciliation loop |
| Нужно собрать представление для UI? | Да | BFF/query composition/read model |
| Нужна массовая историческая обработка? | Да | export/batch/analytics pipeline |

## 14.6. Синхронные вызовы

Синхронный вызов допускается для:

- `CheckPermission`;
- `EvaluateRisk`;
- resolve active Subject;
- get authoritative resource state;
- validate external precondition;
- получить Operation status.

Синхронный вызов MUST иметь:

- deadline;
- retry classification;
- circuit breaker или concurrency limit;
- telemetry;
- fallback policy;
- permission/service identity;
- bounded response size.

## 14.7. Запрещённые синхронные цепочки

Запрещены:

- A → B → A;
- Authentication → Access → Resource Manager → Authentication;
- long-running provisioning через удержание HTTP connection без Operation;
- fan-out на неограниченное число downstream services;
- retries на всех уровнях без общего budget;
- вызов внешнего провайдера внутри локальной транзакции.

Глубина критичного синхронного пути SHOULD быть минимальной и контролироваться architecture tests/telemetry.

## 14.8. Timeout budget

Для входящего deadline `D` caller распределяет budget:

```text
D = local_processing + downstream_calls + serialization + safety_margin
```

Каждый downstream timeout MUST быть меньше оставшегося deadline. Retry MUST учитывать суммарный budget, а не повторять полный timeout.

## 14.9. Retry policy

Retry разрешён при:

- transient network failure;
- `UNAVAILABLE`;
- safe `DEADLINE_EXCEEDED` с идемпотентностью;
- provider-specific throttling с `retry_after`;
- optimistic conflict только после повторного чтения и повторной оценки команды.

Retry запрещён для:

- validation errors;
- permission denied;
- business precondition failure;
- idempotency conflict;
- non-idempotent external action без token;
- permanent provider rejection.

## 14.10. Circuit breaker и bulkhead

Критичные integrations SHOULD применять:

- per dependency circuit breaker;
- concurrency limit;
- queue limit;
- separate worker pool;
- degraded mode;
- health signal.

Circuit breaker state MUST быть observable. Он не должен скрывать data corruption или authorization failures.

## 14.11. Асинхронные факты

Асинхронный consumer MUST:

1. проверить envelope/schema;
2. выполнить Inbox deduplication;
3. проверить revision/order;
4. применить локальное изменение в транзакции;
5. записать Inbox success;
6. инициировать собственное событие только после commit;
7. обработать retry/DLQ.

## 14.12. Saga

Saga — последовательность локальных транзакций с предметными compensation actions.

Пример удаления Project:

```text
Request Project deletion
→ mark Project DELETING
→ revoke new mutations
→ delete/retain Identity resources
→ remove Access bindings
→ deprovision Managed Resources
→ finalize Audit evidence
→ mark Project DELETED
```

Если deprovisioning не завершилось, Project остаётся `DELETION_FAILED` или `DELETING`, а не возвращается автоматически в исходное состояние без предметного решения.

## 14.13. Choreography и orchestration

| Подход | Использовать когда | Ограничение |
| --- | --- | --- |
| Choreography | простое распространение независимых фактов | не подходит для сложного прогресса и compensation |
| Process Manager | нужно координировать события и состояние процесса | владелец процесса должен быть явным |
| Temporal orchestration | долгий процесс, retries, timers, signals, compensation | workflow determinism и versioning обязательны |

Скрытая Saga, распределённая по consumers без владельца и наблюдаемого состояния, запрещена.

## 14.14. Владелец процесса

Владелец определяется по бизнес-результату:

| Процесс | Владелец |
| --- | --- |
| создание/удаление Organization/Workspace/Project | Resource Manager |
| authentication flow и step-up | Authentication |
| identity deletion/privacy coordination | Identity |
| role review/revocation campaign | Access |
| risk review lifecycle | Risk Decision |
| managed resource lifecycle | Provisioning |
| audit export | Audit |

Process owner хранит state, Operation и correlation; участники владеют только своими локальными действиями.

## 14.15. Temporal Workflow

Workflow MUST:

- иметь стабильный workflow ID, связанный с Operation;
- быть детерминированным;
- использовать Activities для I/O;
- определять retry policy per activity;
- обрабатывать cancellation signal;
- поддерживать version markers при изменении логики;
- не хранить секреты в history;
- ограничивать history size через continue-as-new;
- публиковать progress в owner state;
- иметь search attributes для diagnostics.

## 14.16. Activities

Activity MUST иметь:

- idempotency key;
- timeout categories (`start_to_close`, `schedule_to_close`, heartbeat);
- retryable error classification;
- heartbeat для долгой работы;
- cancellation handling;
- bounded payload;
- security context reference, а не полный token;
- audit/telemetry correlation.

## 14.17. Compensation

Compensation определяет допустимое обратное действие:

- revoke created credential;
- delete partially created external resource;
- restore previous desired state;
- release reservation;
- mark manual review required.

Compensation MAY быть невозможна. Тогда процесс MUST перейти в явное состояние `FAILED_REQUIRES_INTERVENTION` и сформировать runbook/action item.

## 14.18. Operation и workflow

Operation является публичным контрактом наблюдения. Workflow — внутренний механизм координации. Между ними нет обязательного отношения один-к-одному, но owner MUST поддерживать связь:

```yaml
operation_id: operations/op_123
workflow:
  namespace: m8-platform
  workflow_id: provisioning/mr_123/create
  run_id: ...
```

Клиент не должен знать Temporal run ID.

## 14.19. Reconciliation

Provisioning reconciliation loop сравнивает:

```text
desired state
vs
observed external state
```

Результат:

- `IN_SYNC`;
- `PROGRESSING`;
- `DRIFTED`;
- `DEGRADED`;
- `UNKNOWN`;
- `DELETION_BLOCKED`.

Reconciler MUST быть повторяемым, идемпотентным и устойчивым к внешним изменениям.

## 14.20. Local projections

Проекция может заменить синхронный lookup, если:

- freshness class достаточен;
- stale decision безопасен;
- owner event contract устойчив;
- предусмотрены tombstones;
- есть reconciliation;
- degraded behavior определён.

Для разрешения высокорискового действия stale projection обычно недостаточна.

## 14.21. Read composition

UI и reporting queries MAY агрегировать несколько контекстов через:

- BFF fan-out с timeout/degraded fields;
- материализованную read model;
- аналитическую витрину;
- query service без владения mutations.

Ответ SHOULD помечать `as_of`, partial errors и stale components.

## 14.22. Согласованность удаления

Удаление является отдельным процессом. Owner:

- блокирует новые несовместимые mutations;
- публикует deletion requested/state events;
- ожидает критичные подтверждения;
- повторяет неуспешные шаги;
- сохраняет tombstone;
- завершает hard delete после retention.

Не все consumers обязаны подтверждать удаление синхронно, но критичные копии MUST иметь доказательство обработки.

## 14.23. Multi-region consistency

Для multi-region deployment MUST быть определены:

- home region ресурса;
- routing policy;
- conflict policy;
- replication lag;
- failover procedure;
- write fencing;
- region recovery;
- data residency.

Active-active write одного агрегата без детерминированной conflict model запрещён.

## 14.24. Failure matrix

| Dependency failure | Базовое поведение |
| --- | --- |
| Access unavailable | deny/controlled fail-closed для mutation; read policy определяется отдельно |
| Risk Decision unavailable | fail-closed либо required review для high-risk flow |
| Identity unavailable | не создавать новую authentication transaction без разрешения subject |
| Audit temporarily unavailable | mutation MAY commit при durable local audit outbox; direct loss запрещён |
| Event transport unavailable | local commit + Outbox backlog |
| Temporal unavailable | создать Operation и durable start request либо вернуть controlled unavailable до side effect |
| External provider unavailable | Operation remains retrying/degraded; desired state сохраняется |
| Projection stale | fallback к owner API или explicit stale response |

Конкретный режим MUST быть связан с risk classification.

## 14.25. Recovery и reconciliation

Каждая критичная интеграция MUST иметь:

- health metric;
- backlog/lag metric;
- replay/retry tool;
- reconciliation job;
- manual remediation path;
- audit trail;
- runbook;
- owner/on-call.

## 14.26. Integration registry

```yaml
integration:
  id: INT-AUTH-RISK-EVALUATE
  caller: m8-authentication
  provider: m8-risk-decision
  mode: synchronous
  contract: RiskDecisionService.EvaluateAuthenticationRisk
  consistency: C1
  deadline_ms: 250
  retry: none_within_interactive_request
  failure_mode: fail_closed
  permission: risk.assessments.evaluate
  slo: 99.95
  requirement_ids: [AUTH-FR-001, RISK-FR-001]
```

## 14.27. Тестирование

Обязательны:

- timeout tests;
- retry/idempotency tests;
- dependency outage tests;
- duplicate/out-of-order event tests;
- workflow replay tests;
- compensation tests;
- cancellation tests;
- projection rebuild tests;
- stale data tests;
- partial failure tests;
- disaster recovery exercises для критичных процессов.

## 14.28. SPDD-требования

Structured Prompt MUST включать:

```yaml
integration_model:
  consistency_class: C4
  process_owner: ResourceManager
  local_transaction:
    writes: [Project, Operation, Outbox]
  synchronous_dependencies: []
  asynchronous_participants: [Identity, Access, Provisioning, Audit]
  workflow: temporal
  idempotency_key: request_id
  compensations:
    - action: cancel_deprovisioning
      when: before_external_delete_commit
  failure_states:
    - DELETION_FAILED
    - MANUAL_INTERVENTION_REQUIRED
```

## 14.29. Критерии соответствия главы

Интеграция соответствует PADS, если owner процесса и транзакционная граница явны, механизм взаимодействия выбран по требуемой согласованности, retries идемпотентны, циклы отсутствуют, failure/degraded behavior описано, длительные процессы наблюдаемы через Operation, а внешнее состояние reconciliate.

---
