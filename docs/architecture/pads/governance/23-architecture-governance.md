---
title: "PADS: архитектурное управление"
description: "ADR, review gates, exceptions, fitness functions и governance-процессы."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 23. Архитектурное управление {#pads-architecture-governance}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 22. SPDD: проведение требований до Structured Prompt](22-spdd.md) | [Следующий раздел: 24. Глоссарий](24-glossary.md)

{% endnote %}

## 23.1. Назначение главы

Архитектурное управление обеспечивает развитие M8 без рассинхронизации PADS, требований, контрактов и реализации. Управление строится на явных ролях, ADR, автоматических проверках, architecture reviews, exception process и контроле архитектурного долга.

## 23.2. Роли

| Роль | Ответственность |
| --- | --- |
| Platform Architect | PADS, context map, cross-cutting principles, final architecture arbitration |
| Context Owner | domain model, requirements, contracts и quality своего context |
| Service Owner | implementation, SLO, runbooks, lifecycle сервиса |
| Contract Owner | API/event/common schema compatibility |
| Security Owner | threat model, security requirements, exceptions |
| Data Owner | classification, retention, quality, deletion |
| SRE/Operations Owner | SLO, capacity, incident/readiness |
| Requirement Owner | correctness и acceptance requirement |
| Reviewer | независимая проверка design/change |

Одна персона MAY совмещать роли, но responsibilities остаются раздельными.

## 23.3. Нормативные правила

| ID | Правило |
| --- | --- |
| `GOV-001` | PADS является нормативной baseline и version-controlled artifact. |
| `GOV-002` | Значимое архитектурное решение MUST иметь ADR. |
| `GOV-003` | Изменение context boundary MUST обновлять PADS/context map. |
| `GOV-004` | Public contract change MUST пройти Contract Owner review. |
| `GOV-005` | Security-sensitive change MUST пройти Security Owner review. |
| `GOV-006` | Исключение из MUST rule требует approved exception/ADR со сроком. |
| `GOV-007` | Architecture checks SHOULD быть автоматизированы в CI. |
| `GOV-008` | Architecture debt MUST иметь owner, impact и target date. |
| `GOV-009` | Release MUST иметь traceability/evidence manifest. |
| `GOV-010` | Incident corrective action SHOULD обновлять requirements/tests/PADS при необходимости. |
| `GOV-011` | Shared Kernel change MUST иметь review всех affected owners. |
| `GOV-012` | Vendor adoption MUST иметь ACL, exit/continuity analysis и ADR. |
| `GOV-013` | Deprecated contract MUST иметь measured migration plan. |
| `GOV-014` | PADS review MUST проводиться регулярно и при значимых изменениях. |
| `GOV-015` | AI-generated changes подчиняются тем же governance gates. |

## 23.4. ADR

ADR содержит:

```yaml
adr:
  id: ADR-0021
  title: Использование Common Operation и Temporal для длительных процессов
  status: accepted
  date: 2026-07-10
  owners: [PlatformArchitecture, ServiceOwners]
  context: ...
  decision: ...
  alternatives: ...
  consequences:
    positive: [...]
    negative: [...]
  constraints: [...]
  affected:
    pads_sections: [16, 22]
    contexts: [ResourceManager, Provisioning]
    requirements: [OPS-FR-001]
  review_date: 2027-01-10
```

## 23.5. ADR lifecycle

```text
PROPOSED
→ UNDER_REVIEW
→ ACCEPTED
→ SUPERSEDED / DEPRECATED
→ RETIRED
```

Rejected ADR сохраняется для истории с причиной.

## 23.6. Триггеры ADR

ADR обязателен для:

- новой технологии platform-level;
- изменения owner/context boundary;
- нового communication pattern;
- изменения consistency class;
- public API/event major version;
- data ownership transfer;
- security/authentication model;
- multi-region strategy;
- исключения из PADS;
- решения с высокой стоимостью отката.

## 23.7. Architecture review

Design review package SHOULD включать:

- problem/requirements;
- context/capability owner;
- options/trade-offs;
- domain/data model;
- API/events;
- security/threats;
- consistency/failure;
- quality/capacity;
- migration/rollout;
- observability/runbooks;
- traceability;
- draft ADR.

## 23.8. Review gates

| Gate | Проверяет |
| --- | --- |
| Domain gate | language, aggregate, invariant, owner |
| Contract gate | API/event compatibility and consumers |
| Data gate | ownership, classification, migration, retention |
| Security gate | threat model, Access/Risk/Auth, secrets |
| Reliability gate | SLO, failure, DR, capacity |
| SPDD gate | prompt scope, constraints, tests |
| Release gate | evidence, migration, rollback, dashboards |

## 23.9. Исключения

Exception record MUST содержать:

- violated rule;
- business/technical reason;
- risk;
- compensating controls;
- owner;
- scope;
- expiry;
- remediation plan;
- approval;
- monitoring.

Бессрочные исключения запрещены; permanent change оформляется изменением PADS/ADR.

## 23.10. Архитектурный долг

Категории:

- boundary violation;
- shared database/coupling;
- contract debt;
- data migration debt;
- security debt;
- observability debt;
- reliability debt;
- test/traceability debt;
- deprecated dependency.

Debt item MUST иметь severity, interest/impact, owner, target milestone и verification.

## 23.11. Автоматические проверки

CI SHOULD проверять:

- domain does not import infrastructure;
- no cross-service DB clients in wrong module;
- allowed dependency graph;
- buf lint/breaking;
- event envelope/schema;
- requirement IDs exist;
- prompt completeness;
- migration naming/traceability;
- secret scanning;
- security/license/SBOM;
- metric cardinality policy;
- test coverage for acceptance criteria.

## 23.12. Architecture fitness functions

Примеры:

- запрещённые Go imports;
- dependency graph cycles;
- owner package constraints;
- public proto field reuse;
- all mutations have request ID/audit policy;
- all events have required envelope;
- all LRO methods return Operation;
- all public methods have permission annotation;
- all projections declare source/freshness;
- all tasks reference requirements.

## 23.13. Contract governance

Contract Owner отвечает за:

- schema review;
- naming/versioning;
- consumer registry;
- compatibility tests;
- deprecation;
- documentation;
- SDK generation;
- migration communication.

## 23.14. Data governance

Data Owner отвечает за:

- authoritative model;
- classification;
- retention;
- quality;
- lineage;
- deletion;
- access/export;
- backup/restore verification.

## 23.15. Security governance

Security Owner определяет:

- threat modeling standard;
- secure defaults;
- vulnerability SLA;
- exception acceptance;
- incident severity;
- key/secret policy;
- penetration testing scope;
- break-glass control.

## 23.16. Release governance

Release manifest MUST содержать:

```yaml
release:
  id: REL-2026.08.1
  pads_version: PADS-000@1.0
  requirements: [...]
  prompts: [...]
  contracts: [...]
  migrations: [...]
  tests: [...]
  security_evidence: [...]
  quality_evidence: [...]
  deployment_plan: ...
  rollback_plan: ...
  known_exceptions: [...]
```

## 23.17. Change management PADS

PADS versioning:

- patch — редакционные уточнения без изменения нормы;
- minor — новые совместимые нормы/главы;
- major — изменение фундаментальных границ/принципов.

Каждое изменение имеет changelog и affected artifacts.

## 23.18. Review cadence

- per significant design;
- per public contract change;
- quarterly architecture health review;
- semiannual context map review;
- annual PADS baseline review;
- after severity-1 incident;
- before major scale/region expansion.

## 23.19. Метрики управления

- orphan requirements/contracts;
- architecture violations;
- exception count/age;
- ADR lead time;
- deprecated consumer count;
- architecture debt trend;
- SPDD review findings;
- acceptance coverage;
- restore/load/security test freshness;
- incident recurrence.

## 23.20. Критерии соответствия главы

Governance соответствует PADS, если владельцы и gates явны, значимые решения имеют ADR, исключения ограничены сроком, fitness functions автоматизированы, debt управляется, releases имеют evidence и изменения PADS/requirements/contracts остаются синхронизированы.

---
