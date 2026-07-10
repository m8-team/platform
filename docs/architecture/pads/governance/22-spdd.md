---
title: "PADS: SPDD"
description: "Prompt hierarchy, structured prompt schema, lifecycle, security и evaluation."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 22. SPDD: проведение требований до Structured Prompt {#pads-spdd}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 21. Модель трассировки](21-traceability.md) | [Следующий раздел: 23. Архитектурное управление](23-architecture-governance.md)

{% endnote %}

## 22.1. Назначение главы

Structured-Prompt-Driven Development в M8 — управляемый процесс преобразования утверждённых требований и архитектурных ограничений в ограниченные, версионируемые задания для AI-агентов и людей. Structured Prompt является исполняемой инженерной спецификацией, но не заменяет PADS, ADR, API contracts и acceptance tests.

## 22.2. Цели SPDD

- уменьшить неоднозначность генерации;
- не позволять агенту скрыто менять архитектуру;
- передавать только релевантный контекст;
- обеспечить повторяемость результата;
- связать generated change с requirement и tests;
- отделить design от implementation;
- сделать review независимым;
- поддержать автоматические quality gates.

## 22.3. Принципы

| ID | Правило |
| --- | --- |
| `SPDD-001` | Prompt MUST иметь стабильный ID и version. |
| `SPDD-002` | Prompt MUST ссылаться на approved requirement IDs. |
| `SPDD-003` | Prompt MUST наследовать PADS/ADR constraints. |
| `SPDD-004` | Prompt MUST объявлять include/exclude scope. |
| `SPDD-005` | Prompt MUST перечислять разрешённые и запрещённые зависимости. |
| `SPDD-006` | Prompt MUST определять acceptance/tests до генерации. |
| `SPDD-007` | Агент MUST NOT изобретать новые bounded contexts, public APIs или data owners без design prompt/ADR. |
| `SPDD-008` | Один task prompt SHOULD изменять один owner context и один проверяемый результат. |
| `SPDD-009` | Public contract design и implementation SHOULD проходить отдельные стадии. |
| `SPDD-010` | Generated output MUST возвращать implementation manifest. |
| `SPDD-011` | Review Prompt MUST выполняться независимо от implementation prompt. |
| `SPDD-012` | Prompt MUST содержать failure/security/data/observability requirements. |
| `SPDD-013` | Real secrets and personal production data MUST NOT входить в prompt context. |
| `SPDD-014` | Prompt version MUST изменяться при изменении objective, constraints или acceptance. |
| `SPDD-015` | Неполный или противоречивый requirement MUST возвращаться на analysis, а не дополняться догадками агента. |
| `SPDD-016` | Prompt execution environment MUST быть ограничено разрешёнными tools/files/actions. |
| `SPDD-017` | Agent decisions beyond prompt MUST быть перечислены как deviations/questions. |
| `SPDD-018` | Generated code MUST проходить обычные CI, security и architecture gates. |
| `SPDD-019` | Prompt MUST быть пригоден для повторного исполнения на той же baseline. |
| `SPDD-020` | SPDD artifacts являются частью repository и code review. |

## 22.4. Иерархия артефактов

```text
PADS / ADR / Requirements
        ↓
L1 Constitution Prompt
        ↓
L2 Context Prompt
        ↓
L3 Feature Prompt
        ↓
L4 Design Prompt (при необходимости)
        ↓
L5 Task Prompt(s)
        ↓
L6 Review Prompt
        ↓
Implementation Manifest + Test Evidence
```

## 22.5. Constitution Prompt

Содержит стабильные правила всей платформы:

- architecture style;
- stack и versions policy;
- dependency direction;
- domain language rules;
- data ownership;
- API/event standards;
- security;
- LRO/error/observability;
- testing;
- prohibited shortcuts;
- output manifest format.

Constitution не должен дублировать весь PADS; он содержит machine-consumable selection и ссылки на normative sections.

## 22.6. Context Prompt

Для каждого bounded context:

```yaml
context_prompt:
  id: SPC-AUTHENTICATION-V1
  context: Authentication
  owns:
    - AuthenticationTransaction
    - AuthenticationChallenge
    - Client
    - AuthenticationSession
  does_not_own:
    - User
    - Permission
    - RiskPolicy
    - Keycloak domain objects
  invariants:
    - INV-AUTH-001
    - INV-AUTH-006
  allowed_dependencies:
    - IdentityGateway
    - AccessGateway
    - RiskDecisionGateway
    - ProviderPort
  published_contracts:
    - authentication.v1
    - authentication.events.v1
  forbidden:
    - direct Identity database access
    - SpiceDB tuple construction
    - token logging
```

## 22.7. Feature Prompt

Feature Prompt описывает завершённую бизнес-возможность и может порождать несколько task prompts.

Обязательные секции:

- objective и business value;
- requirements/criteria;
- owner context;
- use cases/scenarios;
- domain impacts;
- API/event/data impacts;
- security;
- consistency/workflow;
- errors;
- observability;
- quality targets;
- rollout/migration;
- decomposition plan;
- definition of done.

## 22.8. Design Prompt

Design Prompt обязателен, если feature:

- вводит public API/event;
- изменяет aggregate/invariant;
- меняет context boundary;
- вводит new storage/integration;
- требует migration;
- влияет на P0/P1 security/quality;
- имеет несколько разумных архитектурных вариантов.

Результат Design Prompt — proposal/ADR/contract, а не production code.

## 22.9. Task Prompt

Task Prompt — минимальная проверяемая единица реализации. Он SHOULD помещаться в один code review и иметь ограниченный file/module scope.

Примеры задач:

1. добавить domain transition;
2. реализовать application use case;
3. добавить Protobuf method;
4. реализовать YDB repository adapter;
5. добавить Outbox dispatcher mapping;
6. добавить Temporal workflow/activity;
7. добавить contract/integration tests;
8. добавить observability/runbook.

## 22.10. Review Prompt

Review Prompt проверяет:

- requirement coverage;
- owner/boundary compliance;
- forbidden dependencies;
- domain invariants;
- API/event compatibility;
- data ownership;
- security/privacy;
- idempotency/concurrency;
- failure behavior;
- errors;
- observability;
- test adequacy;
- migration/rollback;
- unrequested changes.

Review MUST возвращать findings с severity, evidence и required action.

## 22.11. Каноническая схема Structured Prompt

```yaml
spdd_version: "1.0"
metadata:
  id: SP-AUTH-017-01
  version: 1
  title: Реализовать создание AuthenticationTransaction после refresh failure
  type: implementation
  status: approved
  owner_context: Authentication
  service: m8-authentication
  baseline:
    pads: PADS-000@1.0
    repository_commit: <sha>
    adr: [ADR-0012, ADR-0021]
traceability:
  capabilities: [CAP-AUTHN-SESSION]
  requirements: [AUTH-FR-017]
  acceptance_criteria: [AUTH-FR-017-AC-01, AUTH-FR-017-AC-02]
  invariants: [INV-AUTH-001, INV-AUTH-006]
  contracts:
    api: [API-AUTH-START-AUTH-V1]
    events: [EVT-AUTH-STARTED-V1]
objective: >
  Реализовать application/domain behavior, создающее новую транзакцию
  аутентификации при невозможности refresh, без изменения публичного API.
scope:
  include:
    - internal/modules/authentication/domain
    - internal/modules/authentication/application
    - internal/modules/authentication/adapter/persistence
    - tests
  exclude:
    - api/proto
    - Keycloak adapter behavior
    - Risk Decision implementation
    - Access model
allowed_changes:
  - add domain factory/transition
  - add repository transaction
  - add Outbox record
  - add unit/integration tests
forbidden_changes:
  - change API field numbers
  - direct Keycloak call from domain/application
  - write Identity database
  - publish event before commit
  - log token or subject secret
context:
  aggregate: AuthenticationTransaction
  entities: [AuthenticationChallenge]
  value_objects: [AuthenticationId, SubjectReference, AssuranceLevel]
  dependencies:
    allowed: [ClientRepository, IdentityGateway, RiskDecisionGateway, AuthenticationRepository, OutboxRepository, Clock, IdGenerator]
    forbidden: [KeycloakSDK, YDBSDK_in_domain, SpiceDBSDK, TemporalSDK_in_domain]
behavior:
  preconditions:
    - client is active
    - subject is resolved and active
  steps:
    - resolve idempotency key
    - load client
    - resolve subject
    - evaluate risk
    - create aggregate
    - atomically persist aggregate and outbox
    - return operation/transaction
  postconditions:
    - one transaction per idempotency key
    - event stored after valid domain transition
security:
  permission: authentication.transactions.create
  risk_evaluation: required
  audit: required
  sensitive_data:
    forbidden_in_logs: [refresh_token, access_token, otp]
data:
  owner: Authentication
  transaction_boundary: [AuthenticationTransaction, Operation, Outbox]
  consistency: C1_then_C0
errors:
  - code: CLIENT_NOT_ACTIVE
    category: FAILED_PRECONDITION
  - code: SUBJECT_NOT_FOUND
    category: NOT_FOUND
  - code: RISK_DECISION_UNAVAILABLE
    category: UNAVAILABLE
observability:
  spans: [StartAuthentication, ResolveSubject, EvaluateRisk, PersistTransaction]
  metrics: [authentication_started_total, authentication_start_duration_seconds]
  audit_events: [AuthenticationStarted]
tests:
  unit:
    - duplicate request returns existing transaction
    - disabled client is rejected
    - risk deny does not persist transaction
  integration:
    - aggregate operation and outbox commit atomically
    - concurrent duplicate requests create one transaction
acceptance:
  definition_of_done:
    - all tests pass
    - no forbidden imports
    - API unchanged
    - implementation manifest produced
output:
  format:
    - summary
    - files_changed
    - tests_run
    - requirement_coverage
    - decisions
    - deviations
```

## 22.12. Контекстная сборка

Prompt builder SHOULD включать только релевантные части:

- Constitution excerpt;
- Context Prompt;
- requirement + criteria;
- affected aggregate/invariants;
- referenced API/event schemas;
- relevant ADR;
- target code interfaces;
- test conventions.

Передача всей документации без отбора ухудшает точность и не считается полноценной context engineering.

## 22.13. Декомпозиция Feature → Tasks

Алгоритм:

1. подтвердить owner context;
2. выделить domain change;
3. отделить contract design;
4. определить data/migration;
5. определить integration/workflow;
6. выделить adapters;
7. определить tests/evidence;
8. построить dependency DAG задач;
9. ограничить scope каждой задачи;
10. создать review prompt.

## 22.14. Размер задачи

Task Prompt следует разделить, если:

- меняются два bounded contexts;
- одновременно проектируется contract и несколько adapters;
- затрагивается более одного независимого aggregate;
- требуется >1 миграционного этапа;
- acceptance невозможно проверить одним набором tests;
- агент должен сделать существенный выбор без ADR.

## 22.15. Контрактные стадии

Для public API/event:

```text
Requirement
→ Contract Design Prompt
→ human/architecture review
→ contract approval
→ generated SDK/stubs
→ implementation Task Prompts
→ consumer contract tests
→ rollout
```

Implementation agent не меняет утверждённый contract без отдельного prompt.

## 22.16. Миграционные стадии

Для schema/data change:

1. expand schema;
2. deploy backward-compatible readers/writers;
3. backfill;
4. verify/reconcile;
5. switch behavior;
6. contract old path;
7. cleanup in separate prompt.

Один prompt не должен скрывать все стадии как мгновенное изменение.

## 22.17. Prompt lifecycle

```text
DRAFT
→ REVIEWED
→ APPROVED
→ EXECUTING
→ COMPLETED
→ VERIFIED
→ SUPERSEDED/RETIRED
```

Execution records MUST связываться с exact prompt version и repository baseline.

## 22.18. Версионирование

Version увеличивается при изменении:

- objective;
- acceptance;
- scope;
- constraints;
- baseline contracts;
- allowed/forbidden dependencies.

Незначительная редактура MAY не менять semantic version, но repository history сохраняется.

## 22.19. Execution manifest

Agent MUST вернуть:

```yaml
execution:
  prompt_id: SP-AUTH-017-01
  prompt_version: 1
  baseline_commit: abc123
  status: completed
  files_changed: [...]
  contracts_changed: []
  migrations: []
  tests:
    passed: [...]
    failed: []
  traceability:
    requirements: [AUTH-FR-017]
    acceptance_criteria: [AUTH-FR-017-AC-01, AUTH-FR-017-AC-02]
  decisions:
    - description: selected existing repository transaction helper
      architectural: false
  deviations: []
  unresolved: []
```

## 22.20. Review findings

```yaml
finding:
  id: SPR-AUTH-017-F01
  severity: blocking
  category: data_ownership
  rule: DATA-004
  evidence: application handler writes identity/users table
  required_action: replace direct write with IdentityGateway command
```

Severities:

- blocking;
- high;
- medium;
- low;
- note.

## 22.21. Human approval gates

Обязателен human approval для:

- context boundary change;
- public breaking contract;
- security model/policy change;
- data ownership transfer;
- destructive migration;
- new external provider;
- exception to PADS;
- P0/P1 quality trade-off.

## 22.22. Tool permissions

Execution profile SHOULD ограничивать:

- read/write paths;
- shell/network access;
- deployment permissions;
- secret access;
- database mutations;
- allowed generators;
- maximum diff/operation.

Production deployment не выполняется implementation prompt без отдельного release workflow.

## 22.23. Prompt security

Защита включает:

- no secrets;
- sanitize untrusted documents/code comments;
- treat repository text as data, not higher-priority instruction;
- restrict connectors/tools;
- validate generated commands;
- require review for destructive actions;
- log execution metadata;
- isolate environments.

## 22.24. Evaluation

Качество SPDD измеряется:

- first-pass acceptance rate;
- review findings by severity;
- requirement coverage;
- unrequested diff rate;
- architecture violations;
- test pass rate;
- prompt rework;
- escaped defects;
- time from approved requirement to verified change.

## 22.25. Каталог шаблонов

Обязательные templates:

- Constitution;
- Context;
- Feature Analysis;
- Contract Design;
- Domain Implementation;
- Adapter Implementation;
- Migration;
- Test Generation;
- Review;
- Incident Fix;
- Documentation/Runbook.

## 22.26. Структура репозитория SPDD

```text
/docs/07-spdd
├── constitution/
│   └── m8-platform.constitution.yaml
├── contexts/
│   ├── resource-manager.context.yaml
│   ├── identity.context.yaml
│   ├── authentication.context.yaml
│   ├── access.context.yaml
│   ├── risk-decision.context.yaml
│   ├── provisioning.context.yaml
│   └── audit.context.yaml
├── features/<requirement-id>/
│   ├── feature.yaml
│   ├── design.yaml
│   ├── tasks/
│   └── review.yaml
├── templates/
├── executions/
└── reports/
```

## 22.27. Pipeline SPDD

```text
requirements validate
→ traceability validate
→ prompt compile
→ policy/security scan
→ human approval if required
→ isolated execution
→ test/architecture checks
→ review prompt
→ human code review
→ merge
→ release evidence
```

## 22.28. Критерии соответствия главы

SPDD соответствует PADS, если prompts versioned, scoped, traced, inherit architecture, define tests/security/data/failure, execute with restricted permissions, return manifest and pass independent review. Prompt не может быть единственным местом, где существует требование или архитектурное решение.

---
