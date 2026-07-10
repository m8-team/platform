---
title: "PADS: правила проектирования API"
description: "API-first правила, совместимость, авторизация, ошибки и валидация."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 12. Правила проектирования API {#pads-api-design}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 11. Владение данными](11-data-ownership.md) | [Следующий раздел: 13. Правила проектирования событий](13-events.md)

{% endnote %}

## 12.1. Назначение главы

Настоящая глава задаёт обязательные правила проектирования публичных и внутренних API M8 Platform. Базовый подход — Protobuf-first с публикацией через ConnectRPC, строгим разделением transport, application и domain model.

API является долгоживущим контрактом между владельцем способности и потребителями. Удобство текущей реализации не является основанием раскрывать внутреннюю модель хранения.

## 12.2. Основные принципы

| ID | Нормативное правило |
| --- | --- |
| `API-001` | Публичные сервисные контракты MUST проектироваться Protobuf-first. |
| `API-002` | Основным транспортом SHOULD быть ConnectRPC с поддержкой gRPC и HTTP semantics, где это применимо. |
| `API-003` | API MUST использовать предметный язык контекста-владельца. |
| `API-004` | API message MUST NOT совпадать с persistence model только ради удобства маппинга. |
| `API-005` | Публичный контракт MUST NOT содержать типы YDB, Redis, Temporal, Keycloak, SpiceDB или Kubernetes SDK. |
| `API-006` | Mutating request MUST иметь механизм идемпотентности, если повтор запроса возможен. |
| `API-007` | Long-running mutation MUST возвращать Operation. |
| `API-008` | Синхронная mutation MAY возвращать ресурс, если все значимые эффекты завершены в локальной транзакции. |
| `API-009` | Authorization context MUST формироваться доверенным middleware, а не полями запроса клиента. |
| `API-010` | Validation MUST быть выражена в контракте через Protovalidate, когда правило не требует обращения к состоянию. |
| `API-011` | List API MUST иметь стабильную пагинацию и детерминированную сортировку. |
| `API-012` | Update API SHOULD использовать FieldMask для частичных изменений. |
| `API-013` | Конкурентные изменения MUST защищаться revision/ETag или явным precondition. |
| `API-014` | Ошибки MUST иметь стабильный машинный service-owned code. |
| `API-015` | Breaking change MUST публиковаться в новой major-версии. |
| `API-016` | Deprecated field или method MUST иметь срок, миграционную инструкцию и telemetry использования. |
| `API-017` | API MUST поддерживать deadlines и cancellation propagation. |
| `API-018` | Batch API MUST возвращать результат по каждому элементу и не скрывать частичный успех. |
| `API-019` | API MUST NOT принимать произвольные claims, roles или permissions как доверенные входы. |
| `API-020` | Каждый API method MUST быть связан с requirement ID, permission и audit policy. |

## 12.3. Структура Protobuf-пакетов

Рекомендуемая схема:

```text
m8.platform.<context>.<version>

m8.platform.resourcemanager.v1
m8.platform.identity.v1
m8.platform.authentication.v1
m8.platform.access.v1
m8.platform.riskdecision.v1
m8.platform.provisioning.v1
m8.platform.audit.v1
m8.platform.common.v1
```

Правила:

- package name MUST содержать major API version;
- Go package option MUST быть стабильным и не зависеть от расположения временного репозитория;
- service и message names MUST отражать domain language;
- common package используется только для типов главы 10;
- импорт контракта другого контекста SHOULD заменяться стабильной reference-моделью, если полная модель не требуется.

## 12.4. Ресурсоориентированный API

Для управляемых ресурсов SHOULD использоваться стандартные операции:

```protobuf
service ProjectsService {
  rpc GetProject(GetProjectRequest) returns (Project);
  rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse);
  rpc CreateProject(CreateProjectRequest) returns (google.longrunning.Operation);
  rpc UpdateProject(UpdateProjectRequest) returns (google.longrunning.Operation);
  rpc DeleteProject(DeleteProjectRequest) returns (google.longrunning.Operation);
}
```

Дополнительные предметные действия оформляются custom methods:

```protobuf
rpc SuspendProject(SuspendProjectRequest) returns (google.longrunning.Operation);
rpc RestoreProject(RestoreProjectRequest) returns (google.longrunning.Operation);
```

Custom method MUST обозначать предметное действие, а не технический CRUD workaround.

## 12.5. Имена ресурсов

Канонические имена SHOULD иметь иерархическую форму:

```text
organizations/{organization_id}
organizations/{organization_id}/workspaces/{workspace_id}
organizations/{organization_id}/workspaces/{workspace_id}/projects/{project_id}
projects/{project_id}/services/{service_id}
projects/{project_id}/userPools/{user_pool_id}
```

Внутренний ID MAY быть отдельным полем, но API MUST однозначно определять канонический `name`.

Правила:

- resource name MUST быть неизменяемым идентификатором;
- display name MUST храниться отдельно;
- parent MUST быть явным в create/list request;
- parsing resource name выполняется общей безопасной библиотекой;
- произвольная конкатенация строк в domain layer запрещена.

## 12.6. Команды и ресурсы

API SHOULD быть ресурсоориентированным, но команда допустима, когда:

- действие не сводится к изменению одного поля;
- требуется отдельная авторизация;
- есть собственный жизненный цикл;
- действие запускает workflow;
- результат имеет предметный смысл.

Примеры корректных команд:

- `StartAuthentication`;
- `CheckPermission`;
- `EvaluateRisk`;
- `ReconcileManagedResource`;
- `ExportAuditEvents`.

## 12.7. Формат запросов

Create request SHOULD содержать:

```protobuf
message CreateProjectRequest {
  string parent = 1 [(buf.validate.field).string.min_len = 1];
  Project project = 2 [(buf.validate.field).required = true];
  string project_id = 3;
  string request_id = 4;
  bool validate_only = 5;
}
```

Update request SHOULD содержать:

```protobuf
message UpdateProjectRequest {
  Project project = 1 [(buf.validate.field).required = true];
  google.protobuf.FieldMask update_mask = 2;
  string request_id = 3;
  string etag = 4;
  bool validate_only = 5;
}
```

Delete request SHOULD явно задавать:

- resource name;
- request ID;
- ETag/revision при конкурентном доступе;
- force/cascade только если такая семантика разрешена;
- validate_only;
- reason при чувствительных операциях.

## 12.8. Валидация

Валидация разделяется на уровни:

| Уровень | Примеры | Место |
| --- | --- | --- |
| Синтаксический | длина, формат UUID, enum defined_only | Protobuf/Protovalidate |
| Структурный | обязательные сочетания полей, oneof | contract/application boundary |
| Предметный | допустимый переход состояния, уникальность | domain aggregate/policy |
| Межконтекстный | существование Project, активность User | gateway к владельцу |
| Авторизационный | право выполнить действие | Access/AuthGuard |
| Рисковый | необходимость step-up | Risk Decision |

Контрактная валидация MUST NOT дублировать изменяемую бизнес-политику, если для неё нужен отдельный владелец.

## 12.9. FieldMask

FieldMask MUST:

- применяться только к разрешённым mutable полям;
- отклонять неизвестные и output-only пути;
- иметь определённую семантику пустой маски;
- различать очистку значения и отсутствие изменения;
- обрабатываться application use case, а не адаптером хранения;
- отражаться в audit change set.

`*` MAY быть разрешён только при однозначной и документированной семантике.

## 12.10. Field behavior

Поля SHOULD иметь annotations или документацию:

- `REQUIRED`;
- `OPTIONAL`;
- `OUTPUT_ONLY`;
- `INPUT_ONLY`;
- `IMMUTABLE`;
- `IDENTIFIER`.

Output-only поля MUST игнорироваться или отклоняться в input согласно контракту; выбранная семантика должна быть единообразной.

## 12.11. Пагинация

List API MUST использовать opaque page token. Token SHOULD кодировать:

- последний стабильный sort key;
- filter hash;
- API version;
- expiry или snapshot marker, если требуется консистентный снимок.

Запрещено:

- раскрывать внутренний database offset как контракт;
- менять сортировку между страницами;
- принимать page token с другим filter;
- возвращать дубликаты без документированной причины.

Базовый контракт:

```protobuf
message ListProjectsRequest {
  string parent = 1;
  int32 page_size = 2;
  string page_token = 3;
  string filter = 4;
  string order_by = 5;
  bool show_deleted = 6;
}
```

## 12.12. Фильтрация и сортировка

Поддерживаемые поля, операторы и стоимость фильтра MUST быть задокументированы. Сервер MUST отклонять неизвестные поля и потенциально неограниченные выражения.

Сортировка MUST иметь стабильный tie-breaker, обычно resource ID. Например:

```text
order_by = "create_time desc, name asc"
```

## 12.13. Идемпотентность

Идемпотентность mutation обеспечивается `request_id` или `idempotency_key`.

Владелец MUST хранить:

- caller scope;
- operation/method;
- canonical request hash;
- created resource/operation ID;
- final status;
- expiry.

Повтор с тем же ключом и другим значимым payload MUST возвращать `IDEMPOTENCY_CONFLICT`.

## 12.14. Оптимистическая конкуренция

Для изменяемых ресурсов SHOULD использоваться:

- `revision` — монотонная версия внутри домена;
- `etag` — opaque representation для API;
- `if_match` semantics для update/delete.

Ошибка конкуренции MUST различаться от validation error и SHOULD возвращать текущий ETag, если это безопасно.

## 12.15. Длительные операции

API MUST возвращать Operation, если:

- вызов внешней системы может превышать обычный deadline;
- операция имеет несколько шагов;
- требуется retry, compensation или approval;
- результат появляется после асинхронного процесса;
- клиент должен наблюдать прогресс или отменять выполнение.

Operation response MUST быть создан до запуска невосстанавливаемой работы и связан с `request_id`.

## 12.16. Batch-операции

Batch SHOULD использоваться для эффективности, но MUST ограничивать:

- максимальное число элементов;
- суммарный размер;
- параллелизм;
- время выполнения;
- scope авторизации.

Atomic batch MAY быть предоставлен только внутри одного владельца и одной транзакционной границы. В иных случаях результат должен отражать частичный успех.

## 12.17. Check и decision API

API принятия решений, например `CheckPermission` и `EvaluateRisk`, MUST:

- принимать полный типизированный context;
- возвращать decision ID;
- содержать reason codes;
- иметь явную freshness semantics;
- не раскрывать внутренние правила сверх допустимого уровня;
- поддерживать explain/simulate отдельным защищённым методом, если требуется.

## 12.18. Security context

Trusted context MAY включать:

```text
actor
subject
client
service_identity
organization/workspace/project scope
assurance_level
authentication_id
session_id
request_id
trace_id
source_network/device digest
```

Он формируется AuthGuard из проверенных credentials. Клиентские поля с теми же именами MUST NOT переопределять trusted context.

## 12.19. Deadlines, cancellation и retries

Каждый client MUST устанавливать deadline. Сервер SHOULD:

- останавливать отменяемую работу после cancellation;
- не откатывать уже подтверждённую локальную транзакцию;
- возвращать Operation, если работа продолжается независимо;
- публиковать retry hints только для безопасных ошибок.

Автоматический retry разрешён только для идемпотентных запросов или запросов с idempotency key.

## 12.20. Ошибки

Transport status отражает категорию, а service-owned error code — предметную причину.

Пример:

```yaml
status: FAILED_PRECONDITION
error_code: PROJECT_NOT_ACTIVE
message: Project must be active to register a service.
details:
  resource: projects/prj_123
  current_state: SUSPENDED
  required_state: ACTIVE
retryable: false
```

Подробная модель определена в главе 17.

## 12.21. Версионирование

Совместимые изменения v1:

- добавление optional field;
- добавление enum value при корректной обработке unknown values;
- добавление нового method;
- ослабление validation, не нарушающее безопасность;
- добавление error detail.

Несовместимые изменения:

- удаление/переиспользование field number;
- изменение смысла поля;
- изменение resource name pattern;
- усиление обязательности без перехода;
- изменение default behavior;
- изменение типа идентификатора;
- удаление enum value.

Reserved field numbers и names MUST сохраняться после удаления.

## 12.22. Deprecation

Deprecation lifecycle:

1. пометить элемент deprecated;
2. опубликовать replacement;
3. добавить migration guide;
4. измерять активных consumers;
5. уведомить владельцев;
6. установить not-before removal date;
7. удалить только в новой major-версии либо по согласованной политике internal API.

## 12.23. Streaming и subscriptions

Streaming API MAY использоваться для:

- ожидания Operation;
- подписки UI на прогресс;
- bounded export;
- low-latency decision stream.

Он MUST иметь:

- backpressure;
- heartbeat/idle timeout;
- resume token;
- authorization revalidation;
- max message size;
- правила reconnect.

Для долговременного распределения фактов предпочтительнее event stream главы 13.

## 12.24. Webhooks

Webhook является внешней интеграцией и MUST иметь:

- подписанную доставку;
- timestamp и replay protection;
- event ID;
- versioned payload;
- retry policy;
- dead-letter visibility;
- secret rotation;
- endpoint verification;
- subscription scope;
- delivery audit.

## 12.25. Rate limits и quotas

API MAY применять:

- per client;
- per subject;
- per project;
- per organization;
- per method;
- cost-based quota.

Ответ MUST содержать стабильную error category и допустимый retry delay. Квота не должна заменять Risk Decision для security-sensitive abuse checks.

## 12.26. API-документация

Каждый method MUST документировать:

- цель;
- owner context;
- required permission;
- risk policy;
- idempotency;
- consistency;
- input/output semantics;
- errors;
- audit event;
- SLO class;
- requirement IDs;
- пример запроса и ответа.

## 12.27. Contract testing

Обязательные проверки:

- buf lint;
- buf breaking;
- Protovalidate tests;
- golden JSON/Connect serialization tests;
- compatibility tests с поддерживаемыми SDK;
- authorization matrix tests;
- idempotency tests;
- pagination stability tests;
- error mapping tests;
- load tests для SLO-critical methods.

## 12.28. API-реестр

Минимальная запись:

```yaml
api:
  id: API-AUTH-START-AUTH-V1
  owner_context: Authentication
  service: m8-authentication
  proto_method: m8.platform.authentication.v1.AuthenticationService.StartAuthentication
  requirement_ids: [AUTH-FR-001, AUTH-FR-017]
  permission: authentication.transactions.create
  consistency: local_strong_plus_external_decision
  idempotency: required
  result: operation
  audit_policy: required
  slo_class: interactive
```

## 12.29. SPDD-требования

Structured Prompt для API change MUST содержать:

- существующий package и version;
- compatibility classification;
- resource/method pattern;
- field behavior;
- validation;
- permission;
- error codes;
- idempotency;
- LRO behavior;
- contract tests;
- запрещённые изменения field numbers.

## 12.30. Критерии соответствия главы

API соответствует PADS, если контракт Protobuf-first, не раскрывает инфраструктуру, имеет owner, permission, validation, stable errors, idempotency, compatibility checks и трассировку до requirement/tests.

---
