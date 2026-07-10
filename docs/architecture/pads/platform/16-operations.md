---
title: "PADS: длительные операции"
description: "Каноническая модель Operation, progress, cancellation, idempotency и Temporal."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# Длительные операции {#lro-operations}


## Назначение


Длительная операция — публичный ресурс, представляющий принятый запрос, выполнение которого продолжается после завершения исходного API-вызова. M8 использует модель, совместимую с `google.longrunning.Operation`, расширенную типизированными metadata, progress, error details и authorization rules.

## 16.2. Когда операция обязательна

Operation MUST использоваться, если:

- выполнение может превысить обычный API deadline;
- есть несколько шагов или external side effects;
- нужны retry/compensation;
- требуется approval/signal/timer;
- клиент должен наблюдать progress;
- операция отменяема;
- результат появляется асинхронно;
- массовое действие имеет частичные результаты.

## 16.3. Принципы

| ID | Правило |
| --- | --- |
| `OPS-001` | Operation MUST иметь стабильный уникальный resource name. |
| `OPS-002` | Operation owner — сервис, принявший бизнес-команду. |
| `OPS-003` | Operation MUST быть создана durable до запуска невосстанавливаемого side effect. |
| `OPS-004` | Повтор request ID MUST возвращать ту же Operation. |
| `OPS-005` | Public Operation MUST NOT раскрывать Temporal run ID или vendor-specific state. |
| `OPS-006` | State transition MUST быть монотонным и валидируемым. |
| `OPS-007` | Terminal result и terminal error взаимоисключающи. |
| `OPS-008` | Cancellation является запросом, а не гарантией мгновенной отмены. |
| `OPS-009` | Progress MUST быть честным и не должен откатываться без reason/version semantics. |
| `OPS-010` | Operation access MUST проверяться как доступ к отдельному ресурсу. |
| `OPS-011` | Operation MUST иметь retention policy. |
| `OPS-012` | Каждый terminal state MUST формировать Audit evidence. |
| `OPS-013` | Workflow retry MUST не создавать новую Operation. |
| `OPS-014` | Child operations MUST быть связаны с parent operation/process. |
| `OPS-015` | Structured Prompt MUST определить state machine, cancellation и failure semantics. |

## 16.4. Канонический ресурс

```protobuf
message Operation {
  string name = 1;
  string operation_type = 2;
  string owner_service = 3;
  OperationState state = 4;
  google.protobuf.Any metadata = 5;
  OperationProgress progress = 6;
  google.protobuf.Any response = 7;
  google.rpc.Status error = 8;
  google.protobuf.Timestamp create_time = 9;
  google.protobuf.Timestamp update_time = 10;
  google.protobuf.Timestamp end_time = 11;
  bool cancellation_requested = 12;
  string etag = 13;
  m8.platform.common.v1.ResourceReference target = 14;
  string request_id = 15;
  string correlation_id = 16;
}
```

## 16.5. Состояния

```text
ACCEPTED
→ QUEUED
→ RUNNING
→ SUCCEEDED

RUNNING → CANCELLING → CANCELLED
RUNNING → FAILED
RUNNING → REQUIRES_ATTENTION
```

Допустимые состояния MUST быть общими, а предметная стадия хранится в metadata/progress.

Terminal states:

- `SUCCEEDED`;
- `FAILED`;
- `CANCELLED`.

`REQUIRES_ATTENTION` MAY быть non-terminal, если ожидается remediation/signal.

## 16.6. Progress

```protobuf
message OperationProgress {
  int32 percent = 1;
  string stage = 2;
  string message_code = 3;
  repeated OperationStep steps = 4;
  google.protobuf.Timestamp estimated_completion_time = 5;
}
```

Правила:

- percent MAY быть неизвестен;
- пользовательский текст SHOULD формироваться из message code;
- stage является стабильным машинным кодом;
- estimated time является best effort;
- progress update SHOULD быть rate-limited;
- завершение MUST устанавливать 100%, если percentage применим.

## 16.7. Metadata

Metadata MUST быть типизирована по operation type. Пример:

```protobuf
message CreateManagedResourceMetadata {
  string managed_resource = 1;
  string stage = 2;
  string provider_operation_id = 3;
  repeated Condition conditions = 4;
}
```

Provider operation ID MAY быть доступен только privileged consumers.

## 16.8. Результат

Operation response SHOULD содержать созданный/изменённый ресурс либо типизированный summary. Большие результаты сохраняются как отдельный ресурс/export object, а Operation возвращает reference.

## 16.9. Ошибка

Terminal error MUST содержать:

- canonical transport category;
- service-owned error code;
- retryability;
- failed stage;
- target resource;
- remediation hint/code;
- provider detail только в безопасном operator field;
- correlation ID.

## 16.10. API операций

Минимальный интерфейс:

```protobuf
service OperationsService {
  rpc GetOperation(GetOperationRequest) returns (Operation);
  rpc ListOperations(ListOperationsRequest) returns (ListOperationsResponse);
  rpc WaitOperation(WaitOperationRequest) returns (Operation);
  rpc CancelOperation(CancelOperationRequest) returns (Operation);
  rpc DeleteOperation(DeleteOperationRequest) returns (google.protobuf.Empty);
}
```

`DeleteOperation` удаляет запись наблюдения после retention/policy и MUST NOT отменять бизнес-результат.

## 16.11. Wait

Wait MUST:

- принимать max wait/deadline;
- возвращать текущее состояние при timeout;
- не создавать отдельный workflow;
- поддерживать cancellation connection;
- иметь authorization;
- не гарантировать terminal state, если deadline короткий.

## 16.12. Cancellation

Cancellation states:

1. request accepted;
2. workflow/activity receives signal;
3. safe point reached;
4. compensation, если нужна;
5. terminal `CANCELLED` или `FAILED`.

Если side effect необратим, Operation MAY завершиться `SUCCEEDED` или `FAILED` несмотря на cancellation request; причина MUST быть явной.

## 16.13. Idempotency

Связь:

```text
caller + method + request_id → operation_name
```

Повтор возвращает текущую Operation. Request hash mismatch возвращает idempotency conflict.

## 16.14. Авторизация

Operation может содержать чувствительные metadata. Permission SHOULD различать:

- get own operation;
- get project operation;
- list all operations;
- cancel operation;
- view operator details;
- delete operation record.

## 16.15. Связь с Temporal

Owner хранит mapping, но public API не зависит от Temporal. Workflow updates owner state через activity/application command или durable adapter.

Workflow restart/continue-as-new MUST сохранять одну Operation. Temporal history не является системой записи публичного progress.

## 16.16. Parent/child операции

Сложный процесс MAY иметь child operations. Parent metadata SHOULD отражать:

- child names;
- required/optional status;
- completed count;
- failed count;
- partial result.

Клиенту не обязательно видеть все internal child operations.

## 16.17. Частичный успех

Batch/multi-resource operation MUST определить:

- atomicity scope;
- succeeded items;
- failed items;
- retryable items;
- compensation status;
- final overall outcome.

`SUCCEEDED_WITH_WARNINGS` SHOULD моделироваться через successful response с typed summary, а не новым общим terminal state.

## 16.18. Retention

Operation records сохраняются достаточно долго для:

- client retrieval;
- support investigation;
- audit linkage;
- retries/reconciliation.

После истечения retention MAY оставаться minimal tombstone с request ID, target, outcome и Audit reference.

## 16.19. Наблюдаемость

Метрики:

- operations created/completed;
- duration by type/stage;
- queue wait;
- retry count;
- cancellation latency;
- failed/requires attention;
- stuck operations;
- progress update lag;
- workflow-operation mismatch.

## 16.20. Тестирование

Обязательны:

- state transition tests;
- duplicate request tests;
- cancellation at each stage;
- workflow replay;
- retry without duplicate side effect;
- result/error exclusivity;
- authorization tests;
- retention/delete semantics;
- stuck operation detection;
- parent-child aggregation.

## 16.21. SPDD-требования

Prompt MUST определить operation type, owner, target, state machine, metadata schema, progress stages, idempotency, workflow mapping, cancellation, result/error, audit and tests.

## 16.22. Критерии соответствия главы

Длительная операция соответствует PADS, если она durable, имеет owner и стабильный ID, не раскрывает workflow engine, идемпотентна, наблюдаема, авторизуема, отменяется по определённой семантике и связана с requirement/audit/tests.

---
