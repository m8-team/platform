---
title: "PADS: модель трассировки"
description: "Граф трассировки, coverage rules, evidence и automation."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 21. Модель трассировки {#pads-traceability}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 20. Распределение требований](20-requirements.md) | [Следующий раздел: 22. SPDD: проведение требований до Structured Prompt](22-spdd.md)

{% endnote %}

## 21.1. Назначение главы

Трассировка доказывает, что бизнес-возможность реализована корректными контрактами, кодом и тестами, а каждое изменение кода имеет обоснование. M8 применяет двунаправленную трассировку от цели до production evidence.

## 21.2. Граф трассировки

```text
Platform Vision
→ Design Goal
→ Business Capability
→ Requirement
→ Acceptance Criterion
→ Domain Model / Invariant
→ ADR
→ API / Event / Data Contract
→ Structured Prompt
→ Code Component
→ Test
→ Deployment / Release
→ Runtime Evidence / SLO
```

Связи являются типизированными, а не только текстовыми ссылками.

## 21.3. Типы узлов

| Тип | Пример ID |
| --- | --- |
| Goal | `DG-004` |
| Principle | `AP-021` |
| Capability | `CAP-AUTHN-SESSION` |
| Requirement | `AUTH-FR-017` |
| Criterion | `AUTH-FR-017-AC-01` |
| Invariant | `INV-AUTH-006` |
| ADR | `ADR-0012` |
| API | `API-AUTH-START-AUTH-V1` |
| Event | `EVT-AUTH-STARTED-V1` |
| Data contract | `DATA-AUTH-TRANSACTION-V1` |
| Prompt | `SP-AUTH-017-01` |
| Code | package/file/symbol or generated component ID |
| Test | `UT-*`, `CT-*`, `IT-*`, `AT-*` |
| Release | `REL-2026.08.1` |
| Evidence | dashboard/test report/audit/query ID |

## 21.4. Типы связей

| Связь | Значение |
| --- | --- |
| `realizes` | артефакт реализует capability/requirement |
| `verifies` | test/evidence проверяет criterion/NFR |
| `constrains` | principle/ADR ограничивает design |
| `owns` | context/service владеет requirement/contract |
| `publishes` | producer публикует event/API |
| `consumes` | consumer зависит от contract |
| `depends_on` | requirement/use case требует другой capability |
| `supersedes` | новый artifact заменяет старый |
| `generated_from` | code/prompt produced from specification |
| `affects` | change impact without direct realization |

## 21.5. Traceability record

```yaml
traceability:
  id: TRACE-AUTH-FR-017
  requirement: AUTH-FR-017
  capability: CAP-AUTHN-SESSION
  goals: [DG-004, DG-018]
  principles: [AP-004, AP-021, AP-052]
  domain:
    aggregate: DM-AUTH-TRANSACTION
    invariants: [INV-AUTH-001, INV-AUTH-006]
  decisions: [ADR-0012, ADR-0021]
  contracts:
    api: [API-AUTH-START-AUTH-V1]
    events: [EVT-AUTH-STARTED-V1]
    errors: [AUTH_CLIENT_NOT_ACTIVE, RISK_DECISION_UNAVAILABLE]
  prompts:
    feature: SP-AUTH-017
    tasks: [SP-AUTH-017-01, SP-AUTH-017-02]
    review: SPR-AUTH-017
  code:
    - internal/modules/authentication/application/start_authentication.go
    - internal/modules/authentication/domain/transaction.go
  tests:
    unit: [UT-AUTH-017-01]
    contract: [CT-AUTH-017-01]
    integration: [IT-AUTH-017-01]
    acceptance: [AT-AUTH-017-01, AT-AUTH-017-02]
  releases: [REL-2026.08.1]
  evidence:
    - slo://authentication/start
    - audit-query://AUTH-FR-017
```

## 21.6. Двунаправленность

Должны поддерживаться запросы:

- от requirement к code/tests/release;
- от code change к requirement/ADR/prompt;
- от failed test к affected requirements;
- от API field к consumers and requirements;
- от runtime incident к capability/design decisions;
- от deprecated requirement к artifacts to remove.

## 21.7. Coverage rules

Минимальные обязательные связи:

| Артефакт | Обязательные ссылки |
| --- | --- |
| Requirement | capability, owner, criteria |
| Public API method | requirement, permission, error catalog, tests |
| Event | requirement, producer, consumers, schema tests |
| Structured Prompt | requirement, contracts, tests, constraints |
| Code PR | requirement/prompt/ADR if applicable |
| Acceptance test | criterion |
| Release | requirements and evidence |
| ADR | affected goals/principles/contexts |

## 21.8. Coverage metrics

- requirements with acceptance criteria;
- requirements with owner;
- requirements with API/event contract;
- criteria verified by tests;
- public methods without requirement;
- events without consumers/catalog;
- code changes without requirement/prompt;
- released requirements without runtime evidence;
- deprecated artifacts still consumed.

Target для обязательных production artifacts — 100% по критичным связям.

## 21.9. Изменения и impact analysis

При изменении узла система SHOULD вычислять transitive impact.

Примеры:

- изменение `Project` resource name → API, Access tuples, events, projections, tests, migrations;
- изменение AAL policy → Authentication, Risk, clients, audit, acceptance;
- удаление event field → all consumers and replay tools;
- изменение Operation state → SDK/UI/workflows/tests.

## 21.10. Baseline и versioning

Traceability graph versioned вместе с PADS/repository. Release фиксирует immutable baseline: какие версии requirements, contracts, prompts и code вошли.

## 21.11. Evidence

Evidence MAY включать:

- CI test report;
- contract compatibility report;
- security scan;
- load test;
- deployment record;
- SLO dashboard snapshot/reference;
- audit query;
- migration/reconciliation report;
- architecture review.

Evidence MUST иметь timestamp, environment, version и owner.

## 21.12. Автоматизация

Рекомендуется хранить metadata в front matter/YAML и проверять CI:

- ID existence;
- no orphan artifacts;
- acceptance coverage;
- contract owner;
- prompt completeness;
- code annotations/PR references;
- release manifest completeness.

## 21.13. Трассировка generated code

Generated files SHOULD содержать source reference и не редактироваться вручную. Generator version является частью evidence.

## 21.14. Трассировка миграций

Каждая migration MUST ссылаться на requirement/ADR, иметь compatibility phase, verification query и rollback/roll-forward procedure.

## 21.15. Трассировка incidents

Post-incident review SHOULD связывать:

- affected capability;
- violated quality scenario;
- requirements/contracts;
- design decision;
- missing/failed control;
- corrective requirements/prompts/tests.

## 21.16. Отчёты

Минимальные отчёты:

1. Requirement Coverage;
2. Acceptance Coverage;
3. API/Event Orphans;
4. Architecture Constraint Coverage;
5. Security Control Coverage;
6. Release Traceability Manifest;
7. Deprecated Artifact Usage;
8. Quality Attribute Evidence.

## 21.17. SPDD и трассировка

Prompt ID является обязательным узлом. Ответ агента SHOULD возвращать machine-readable manifest:

```yaml
implementation_manifest:
  prompt_id: SP-AUTH-017-01
  requirement_ids: [AUTH-FR-017]
  files_changed: [...]
  contracts_changed: []
  tests_added: [UT-AUTH-017-01, IT-AUTH-017-01]
  decisions_made: []
  deviations: []
```

## 21.18. Критерии соответствия главы

Трассировка соответствует PADS, если она двунаправленна, versioned, автоматически проверяема, покрывает requirements/contracts/prompts/code/tests/releases/evidence и позволяет выполнить impact analysis без ручного поиска по всему проекту.

---
