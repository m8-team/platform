---
title: "PADS: распределение требований"
description: "Классы требований, пространства идентификаторов, DoR, DoD и начальное распределение."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 20. Распределение требований {#pads-requirements}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 19. Атрибуты качества](../platform/19-quality-attributes.md) | [Следующий раздел: 21. Модель трассировки](21-traceability.md)

{% endnote %}

## 20.1. Назначение главы

Настоящая глава определяет единый формат требований M8 Platform, правила присвоения владельца, распределение между ограниченными контекстами и сервисами, жизненный цикл, критерии приёмки и связь с контрактами, реализацией и SPDD.

Требование принадлежит контексту, который владеет изменяемым инвариантом или принимает нормативное решение. Наличие зависимости не делает supporting service совладельцем требования.

## 20.2. Классы требований {#requirement-classes}

| Тип | Префикс | Назначение |
| --- | --- | --- |
| Business capability | `CAP-*` | устойчивая способность платформы |
| Platform functional | `PLT-FR-*` | сквозное функциональное требование |
| Context functional | `<CTX>-FR-*` | поведение конкретного owner context |
| Non-functional | `<CTX>-NFR-*`/`PLT-NFR-*` | измеримое качество |
| Security | `<CTX>-SEC-*`/`PLT-SEC-*` | security/privacy control |
| Data | `<CTX>-DATA-*` | ownership, retention, quality |
| Integration | `<CTX>-INT-*` | контракт зависимости и согласованность |
| Compliance | `<CTX>-COMP-*` | юридические/регуляторные требования |
| Architecture constraint | `ARC-*` | обязательное архитектурное ограничение |
| Acceptance criterion | `<REQ>-AC-*` | проверяемый исход сценария |

## 20.3. Пространства идентификаторов

| Prefix | Owner context |
| --- | --- |
| `RM-*` | Resource Manager |
| `ID-*` | Identity |
| `AUTH-*` | Authentication |
| `ACC-*` | Access |
| `RISK-*` | Risk Decision |
| `PROV-*` | Provisioning |
| `AUD-*` | Audit |
| `OPS-*` | Common Operation contract / concrete operation owner |
| `PLT-*` | platform-wide governance/shared property |

ID MUST быть неизменяемым после публикации. Удалённый ID не переиспользуется.

## 20.4. Нормативный формат требования

```yaml
requirement:
  id: AUTH-FR-017
  title: Повторная аутентификация после невозможности refresh
  type: functional
  status: approved
  priority: must
  owner_context: Authentication
  owner_service: m8-authentication
  capability_ids: [CAP-AUTHN-SESSION]
  source:
    stakeholder: ApplicationDeveloper
    business_goal: secure_session_recovery
  statement: >
    Система должна создать новую AuthenticationTransaction, если refresh token
    отсутствует, просрочен, отозван или не может быть безопасно использован.
  rationale: >
    Предыдущая session не должна восстанавливаться неявно.
  preconditions:
    - client is active
    - CIBA is allowed for client
  trigger:
    api: AuthenticationService.StartAuthentication
  business_rules:
    - new transaction does not continue the failed refresh attempt
    - idempotency key maps to one transaction
    - required assurance is evaluated by policy and Risk Decision
  data_ownership:
    reads: [Identity.User, Authentication.Client]
    writes: [Authentication.AuthenticationTransaction, Operation, Outbox]
  security:
    permission: authentication.transactions.create
    risk_evaluation: required
    audit: required
  consistency: C1_then_C0
  outputs:
    - operation
    - authentication_id
  events:
    - AuthenticationStarted.v1
  errors:
    - CLIENT_NOT_ACTIVE
    - SUBJECT_NOT_FOUND
    - RISK_DECISION_UNAVAILABLE
  quality_attributes:
    - QA-AVAIL-AUTH-001
    - QA-PERF-AUTH-001
  acceptance_criteria:
    - AUTH-FR-017-AC-01
    - AUTH-FR-017-AC-02
  traceability:
    adr: [ADR-0012]
    api: [API-AUTH-START-AUTH-V1]
```

## 20.5. Качество формулировки

Требование MUST:

- описывать наблюдаемое поведение или измеримое свойство;
- иметь одного owner context;
- использовать единый язык главы 5;
- не предписывать инфраструктурную реализацию без архитектурной причины;
- содержать условия и границы;
- иметь проверяемые acceptance criteria;
- определять security/data/consistency impact;
- быть атомарным для управления изменением.

Формулировки «система должна быть удобной/быстрой/надёжной» без метрики запрещены.

## 20.6. Определение владельца

Алгоритм:

1. Какой инвариант меняется?
2. Какой контекст определяет смысл результата?
3. Кто является источником истины данных?
4. Кто может отклонить команду по предметной причине?
5. Кто публикует факт завершения?
6. Кто владеет Operation сквозного процесса?

Если ответы указывают на разные контексты, requirement MUST быть разделён на owner requirement и supporting integration requirements.

## 20.7. Сквозное требование

Пример «удалить Project и все связанные данные» декомпозируется:

- `RM-FR-*` — инициировать и координировать deletion lifecycle;
- `ID-FR-*` — удалить/обезличить identity data scope;
- `ACC-FR-*` — отозвать bindings/relationships;
- `PROV-FR-*` — deprovision managed resources;
- `AUD-FR-*` — сохранить evidence и применить retention;
- `PLT-INT-*` — подтверждения, retries, timeout и completion policy.

Resource Manager остаётся process owner, но не изменяет данные участников напрямую.

## 20.8. Состояния требования

```text
PROPOSED
→ ANALYZED
→ APPROVED
→ PLANNED
→ IMPLEMENTING
→ VERIFIED
→ RELEASED
→ DEPRECATED
→ RETIRED
```

`REJECTED` и `SUPERSEDED` являются terminal governance states.

Переход MUST иметь actor, timestamp и reason.

## 20.9. Приоритет

Используется MoSCoW или иной утверждённый механизм:

- Must;
- Should;
- Could;
- Won't for current scope.

Priority не заменяет risk/severity. Security Must может иметь меньший пользовательский объём, но блокировать release.

## 20.10. Acceptance criteria

Критерий SHOULD использовать Given/When/Then:

```yaml
id: AUTH-FR-017-AC-01
given:
  - refresh token is expired
  - client is active
when:
  - StartAuthentication is called with a new request_id
then:
  - a new AuthenticationTransaction is created
  - AuthenticationStarted event is stored in Outbox
  - response contains the same Operation on retry
```

Критерии MUST покрывать happy path, ключевые ошибки, authorization, idempotency и consistency where relevant.

## 20.11. Нефункциональные требования

NFR MUST содержать:

- quality scenario;
- metric/SLI;
- target;
- measurement window;
- load/environment assumptions;
- verification method;
- owner;
- failure consequence.

## 20.12. Security requirements

Security requirement MUST определять:

- asset/resource;
- actor/threat;
- control;
- permission/assurance;
- fail mode;
- audit evidence;
- test.

## 20.13. Data requirements

Data requirement MUST определять:

- owner;
- authoritative fields;
- classification;
- retention;
- replication/projection;
- deletion;
- residency;
- quality checks.

## 20.14. Integration requirements

Integration requirement MUST определять:

- caller/provider;
- API/event contract;
- consistency class;
- deadline/lag;
- retry/idempotency;
- failure/degraded behavior;
- SLO;
- runbook owner.

## 20.15. Начальное распределение функциональных требований

### Resource Manager

| ID | Требование |
| --- | --- |
| `RM-FR-001` | Создать Organization |
| `RM-FR-002` | Изменить Organization с optimistic concurrency |
| `RM-FR-003` | Приостановить/восстановить Organization |
| `RM-FR-010` | Создать Workspace внутри Organization |
| `RM-FR-020` | Создать Project внутри Workspace |
| `RM-FR-021` | Переместить Project управляемым процессом |
| `RM-FR-022` | Удалить Project с координацией зависимостей |
| `RM-FR-030` | Зарегистрировать Service в Project |
| `RM-FR-040` | Публиковать versioned lifecycle events |

### Identity

| ID | Требование |
| --- | --- |
| `ID-FR-001` | Создать User Pool |
| `ID-FR-010` | Создать User |
| `ID-FR-011` | Изменить профиль User |
| `ID-FR-012` | Disable/restore User |
| `ID-FR-020` | Связать External Identity по issuer+subject |
| `ID-FR-021` | Обнаружить конфликт внешней идентичности |
| `ID-FR-030` | Управлять Group и Membership |
| `ID-FR-040` | Выполнить privacy deletion/anonymization |

### Authentication

| ID | Требование |
| --- | --- |
| `AUTH-FR-001` | Начать AuthenticationTransaction |
| `AUTH-FR-002` | Выбрать и создать challenge |
| `AUTH-FR-003` | Получить состояние authentication |
| `AUTH-FR-004` | Повторить отправку challenge по policy |
| `AUTH-FR-005` | Отменить authentication |
| `AUTH-FR-006` | Завершить provider callback |
| `AUTH-FR-010` | Создать session/handoff после успеха |
| `AUTH-FR-017` | Начать re-auth после refresh failure |
| `AUTH-FR-020` | Выполнить step-up до требуемого AAL |
| `AUTH-FR-030` | Revocation session/client access |

### Access

| ID | Требование |
| --- | --- |
| `ACC-FR-001` | Проверить permission на resource |
| `ACC-FR-002` | Batch-check permissions |
| `ACC-FR-003` | Объяснить access decision |
| `ACC-FR-010` | Создать Role |
| `ACC-FR-011` | Назначить RoleBinding |
| `ACC-FR-012` | Отозвать RoleBinding |
| `ACC-FR-020` | Создать/удалить relationship |
| `ACC-FR-030` | Симулировать модель доступа |
| `ACC-FR-040` | Провести access review |

### Risk Decision

| ID | Требование |
| --- | --- |
| `RISK-FR-001` | Оценить риск authentication |
| `RISK-FR-002` | Оценить риск privileged action |
| `RISK-FR-003` | Вернуть ALLOW/DENY/CHALLENGE/REVIEW |
| `RISK-FR-010` | Создать и version RiskPolicy |
| `RISK-FR-011` | Опубликовать/откатить policy |
| `RISK-FR-012` | Симулировать policy на наборе примеров |
| `RISK-FR-020` | Управлять manual review |

### Provisioning

| ID | Требование |
| --- | --- |
| `PROV-FR-001` | Зарегистрировать ResourceDefinition |
| `PROV-FR-010` | Создать ManagedResource desired state |
| `PROV-FR-011` | Изменить desired state |
| `PROV-FR-012` | Удалить ManagedResource |
| `PROV-FR-020` | Выбрать Placement |
| `PROV-FR-030` | Выполнить reconciliation |
| `PROV-FR-031` | Обнаружить drift |
| `PROV-FR-032` | Исправить drift по policy |
| `PROV-FR-040` | Выполнить compensation/manual remediation |

### Audit

| ID | Требование |
| --- | --- |
| `AUD-FR-001` | Принять immutable AuditEvent |
| `AUD-FR-002` | Проверить schema и integrity metadata |
| `AUD-FR-010` | Найти события по разрешённому scope |
| `AUD-FR-011` | Получить цепочку действий по correlation |
| `AUD-FR-012` | Экспортировать audit events через Operation |
| `AUD-FR-020` | Применить retention/legal hold |
| `AUD-FR-021` | Проверить целостность хранения |

### Operations

| ID | Требование |
| --- | --- |
| `OPS-FR-001` | Получить Operation |
| `OPS-FR-002` | Перечислить Operations по scope |
| `OPS-FR-003` | Ожидать изменение/завершение Operation |
| `OPS-FR-004` | Запросить cancellation |
| `OPS-FR-005` | Удалить запись Operation после retention |

## 20.16. Cross-cutting requirements

Cross-cutting requirement MUST иметь platform owner и applicability matrix.

Примеры:

- все mutations имеют idempotency;
- все public APIs имеют authorization;
- все state changes создают AuditEvent;
- все events используют envelope;
- все services экспортируют OpenTelemetry;
- все long-running actions возвращают Operation.

## 20.17. Requirement decomposition rules

Requirement SHOULD быть разделён, если:

- более одного owner context;
- более одного независимого результата;
- разные acceptance/release сроки;
- разные quality/security риски;
- одновременно меняются public contract и implementation без design gate;
- невозможно однозначно определить pass/fail.

## 20.18. Requirement change impact

Изменение MUST запускать анализ:

- domain model/invariant;
- owner/service;
- API/events;
- data ownership;
- security;
- consistency;
- quality attributes;
- migrations;
- tests;
- SPDD prompts;
- documentation/runbooks.

## 20.19. Definition of Ready

Требование готово к реализации, если:

- owner и capability определены;
- statement/rationale ясны;
- acceptance criteria проверяемы;
- data/security/consistency описаны;
- API/event impact классифицирован;
- unresolved architecture decisions отсутствуют или оформлены ADR;
- dependencies и failure mode определены;
- quality targets заданы;
- traceability record создан.

## 20.20. Definition of Done

Требование завершено, если:

- contract/implementation/tests готовы;
- acceptance criteria пройдены;
- architecture/security checks пройдены;
- migrations/reconciliation выполнены;
- telemetry и audit работают;
- SPDD review пройден;
- release evidence связано с requirement;
- documentation/runbook обновлены.

## 20.21. Реестр требований

Реестр SHOULD храниться как version-controlled YAML/Markdown и поддерживать генерацию:

- матрицы владельцев;
- backlog;
- coverage report;
- API/Event catalog links;
- SPDD prompt backlog;
- release scope;
- change impact.

## 20.22. Критерии соответствия главы

Система требований соответствует PADS, если каждый requirement атомарен, имеет ID/owner/criteria, распределён по invariant ownership, содержит cross-cutting impacts и двунаправленно связан с contracts, prompts, code, tests и release evidence.

---
