---
title: "PADS: атрибуты качества"
description: "Сценарии качества, quality gates и каталог атрибутов качества."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 19. Атрибуты качества {#pads-quality-attributes}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 18. Наблюдаемость](18-observability.md) | [Следующий раздел: 20. Распределение требований](../governance/20-requirements.md)

{% endnote %}

## 19.1. Назначение главы

Атрибуты качества задают измеримые свойства архитектуры. Функциональная корректность без доступности, безопасности, восстанавливаемости и сопровождаемости не считается достаточной.

## 19.2. Формат сценария качества

Каждый значимый NFR SHOULD оформляться:

```yaml
quality_scenario:
  id: QA-AVAIL-001
  source: external_client
  stimulus: access_check_request
  environment: normal_and_single_zone_failure
  artifact: m8-access
  response: return_valid_decision_or_controlled_error
  measure:
    availability: 99.99%
    latency_p95_ms: 50
```

Значения являются baseline и уточняются SLO/нагрузочным профилем.

## 19.3. Приоритеты

| Приоритет | Атрибут |
| --- | --- |
| P0 | безопасность, целостность, изоляция, audit durability |
| P1 | доступность decision/authentication control plane, восстановление |
| P2 | производительность, масштабируемость, operability |
| P3 | удобство разработки, portability, cost efficiency |

При конфликте P0 имеет преимущество, если ADR явно не доказывает иной риск-баланс.

## 19.4. Доступность

Начальные целевые классы:

| Класс | Пример | Availability target |
| --- | --- | --- |
| Critical decision | Access Check, token/session validation | 99.99% |
| Interactive control | Start Authentication, Risk Evaluate | 99.95% |
| Control-plane mutation | Create Project, Bind Role | 99.9% |
| Long-running orchestration | Provisioning completion | 99.9% accepted requests; completion SLO отдельно |
| Reporting/export | Audit export, analytics | 99.5% |

Точные SLO утверждаются отдельно и MAY отличаться по плану/региону.

## 19.5. Производительность

Начальные latency objectives:

| Операция | p50 | p95 | p99 |
| --- | --- | --- | --- |
| Access Check | 15 ms | 50 ms | 100 ms |
| Risk Evaluate interactive | 50 ms | 200 ms | 500 ms |
| Get resource | 30 ms | 150 ms | 400 ms |
| Control-plane mutation acceptance | 100 ms | 500 ms | 1 s |
| Start Authentication | 100 ms | 500 ms | 1 s без ожидания challenge completion |

Targets исключают клиентскую сеть и MAY уточняться нагрузочными тестами.

## 19.6. Пропускная способность

Capacity plan MUST задавать:

- requests/sec per method;
- events/sec;
- concurrent operations;
- active authentication transactions;
- relationship cardinality;
- audit ingestion rate;
- data volume/retention;
- peak factor;
- growth forecast.

Сервис MUST иметь load test до предполагаемого peak + safety margin.

## 19.7. Масштабируемость

Архитектура SHOULD масштабироваться горизонтально при соблюдении:

- stateless API adapters;
- partitioned storage;
- bounded in-memory state;
- distributed-safe idempotency;
- no process-local locks for global invariants;
- partition-aware event processing;
- workload isolation.

## 19.8. Надёжность и долговечность

Критичные committed state и AuditEvent MUST выдерживать отказ процесса/узла без потери подтверждённой записи. Outbox обеспечивает eventual publish после local commit.

## 19.9. Восстанавливаемость

Начальные классы:

| Данные | RPO | RTO |
| --- | --- | --- |
| Resource/Identity/Access state | ≤ 5 минут или storage-native stronger | ≤ 60 минут |
| Authentication active state | ≤ 1 минута; short-lived transactions MAY restart | ≤ 30 минут |
| Audit | near-zero committed loss | ≤ 4 часа ingestion/query recovery |
| Projections/cache | reconstructable | зависит от rebuild SLO |
| Analytics | ≤ 24 часа | ≤ 24 часа |

Конкретные значения подтверждаются backup/restore tests.

## 19.10. Согласованность

Каждый API/процесс MUST объявлять consistency class главы 14. Нельзя использовать расплывчатое «данные актуальны» без freshness или source semantics.

## 19.11. Безопасность

Security quality включает:

- authentication strength;
- authorization correctness;
- tenant isolation;
- secret protection;
- vulnerability remediation;
- audit completeness;
- abuse resistance;
- incident response.

Метрики и тесты определены главами 15 и 18.

## 19.12. Privacy

Система MUST обеспечивать:

- minimization;
- purpose limitation;
- retention;
- deletion/anonymization;
- access control;
- export traceability;
- data residency where required.

## 19.13. Сопровождаемость

Цели:

- bounded context changes локализованы;
- domain layer testable без infrastructure;
- contracts versioned;
- ADR объясняют значимые решения;
- architecture rules автоматизированы;
- ownership и runbooks явны;
- технический долг измерим.

## 19.14. Тестируемость

Каждый use case MUST быть тестируем с fake/port dependencies. Обязательны unit, contract, integration и acceptance levels согласно риску.

## 19.15. Развёртываемость

Сервисы SHOULD поддерживать:

- backward-compatible rolling deployment;
- health/readiness;
- graceful shutdown;
- schema migration before/after compatibility;
- feature flags/policy rollout;
- rollback;
- canary where risk warrants.

## 19.16. Совместимость

API/events MUST следовать главам 12–13. Producer и consumer SHOULD поддерживать переходный период N/N-1 версии, если deployment независим.

## 19.17. Наблюдаемость

Каждый SLO MUST иметь измеримый SLI, dashboard и alert. Нельзя утверждать quality target, если отсутствует источник измерения.

## 19.18. Удобство использования API

API quality включает:

- consistent naming;
- discoverable resource patterns;
- stable errors;
- SDK generation;
- examples;
- predictable pagination;
- clear idempotency;
- documentation completeness.

## 19.19. Portability и vendor isolation

Domain/application MUST быть изолированы от vendor SDK через ports/adapters. Полная vendor-neutral реализация не является самоцелью; ACL обязателен для сохранения domain language и тестируемости.

## 19.20. Cost efficiency

Каждый сервис SHOULD контролировать:

- storage growth;
- telemetry cardinality;
- event retention;
- idle capacity;
- external API cost;
- workflow history;
- cache efficiency;
- export volume.

Cost optimization не должна нарушать P0/P1 свойства без ADR.

## 19.21. Accessibility и localization

Для UI-facing contracts SHOULD поддерживаться:

- stable message codes;
- localization-ready errors/progress;
- timezone-aware timestamps;
- accessible state semantics;
- отсутствие цвет-only indicators.

## 19.22. Quality Attribute Catalog

| ID | Атрибут | Owner evidence |
| --- | --- | --- |
| `QA-SEC-001` | Cross-project isolation | security tests + Access model |
| `QA-DUR-001` | Audit durability | storage/restore tests |
| `QA-AVAIL-001` | Access check availability | SLO dashboard |
| `QA-PERF-001` | Access p95 latency | load test + production SLI |
| `QA-REC-001` | Resource state recovery | DR exercise |
| `QA-EVOL-001` | API backward compatibility | buf breaking |
| `QA-OBS-001` | end-to-end traceability | trace propagation test |
| `QA-MAINT-001` | domain independence | architecture tests |
| `QA-DATA-001` | projection freshness | lag SLI |
| `QA-OPS-001` | operation completion | operation SLO |

## 19.23. Quality gates

Release MUST блокироваться при:

- breaking contract without approved version;
- failed authorization isolation tests;
- unbounded metric cardinality;
- critical vulnerability beyond policy;
- missing migration rollback;
- failed required load/DR test;
- missing owner/runbook for critical alert;
- missing requirement/acceptance traceability.

## 19.24. SPDD-требования

Feature Prompt MUST перечислять impacted QA IDs, quantitative targets, test method and evidence. Формулировка «быстро», «надёжно», «масштабируемо» без measure недопустима.

## 19.25. Критерии соответствия главы

Quality requirement соответствует PADS, если оформлен измеримым сценарием, имеет owner, SLI/test, приоритет, архитектурную тактику и release evidence.

---
