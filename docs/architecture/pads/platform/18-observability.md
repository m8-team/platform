---
title: "PADS: наблюдаемость"
description: "Tracing, metrics, logs, SLI/SLO, dashboards, alerts и runbooks."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 18. Наблюдаемость {#pads-observability}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 17. Модель ошибок](17-errors.md) | [Следующий раздел: 19. Атрибуты качества](19-quality-attributes.md)

{% endnote %}

## 18.1. Назначение главы

Наблюдаемость обеспечивает возможность понять состояние, производительность, безопасность и корректность M8 Platform по внешним сигналам. Она включает traces, metrics, logs, events, audit и профили, но не смешивает их назначения.

## 18.2. Принципы

| ID | Правило |
| --- | --- |
| `OBS-001` | Observability MUST проектироваться вместе с feature, а не после инцидента. |
| `OBS-002` | Каждый входящий запрос MUST иметь trace/correlation/request identity. |
| `OBS-003` | Асинхронное событие MUST сохранять correlation и causation. |
| `OBS-004` | Metrics MUST иметь bounded cardinality. |
| `OBS-005` | Logs MUST быть structured и redacted. |
| `OBS-006` | Audit MUST NOT заменяться application logs. |
| `OBS-007` | SLI MUST измерять пользовательский результат, а не только uptime процесса. |
| `OBS-008` | Каждая критичная dependency MUST иметь latency/error/saturation telemetry. |
| `OBS-009` | Long-running process MUST показывать stage, age и stuck condition. |
| `OBS-010` | Alert MUST иметь owner, severity и runbook. |
| `OBS-011` | Telemetry MUST не содержать секреты и необоснованные персональные данные. |
| `OBS-012` | Sampling MUST сохранять ошибки и security-critical traces согласно policy. |
| `OBS-013` | Dashboards MUST строиться по владельцам capabilities и SLO. |
| `OBS-014` | Telemetry schema SHOULD следовать OpenTelemetry semantic conventions. |
| `OBS-015` | Structured Prompt MUST определить telemetry и acceptance signals. |

## 18.3. Корреляционные идентификаторы

| ID | Назначение |
| --- | --- |
| `request_id` | идемпотентность/идентификация API request |
| `trace_id` | distributed trace |
| `span_id` | локальная операция trace |
| `correlation_id` | сквозной бизнес-процесс |
| `causation_id` | непосредственная причина |
| `operation_id` | длительная операция |
| `event_id` | событие |
| `decision_id` | Access/Risk decision |
| `audit_event_id` | audit record |

Они не взаимозаменяемы, но SHOULD быть связаны.

## 18.4. Tracing

Trace SHOULD включать spans:

- API ingress;
- AuthGuard;
- Access check;
- Risk decision;
- application use case;
- repository transaction;
- Outbox enqueue;
- downstream call;
- Temporal activity;
- external provider operation.

Sensitive attributes MUST redacted. Resource IDs MAY использоваться только согласно classification и cardinality policy.

## 18.5. Асинхронная трассировка

Producer injects trace context в event metadata. Consumer создаёт linked span, а не обязательно child span, если обработка отложена. Correlation сохраняется независимо от trace retention.

## 18.6. Metrics model

Базовые классы:

- request rate;
- success/error rate;
- latency histogram;
- saturation/concurrency;
- queue/backlog;
- business outcome;
- data freshness;
- security decisions;
- workflow/operation duration;
- dependency health.

## 18.7. Cardinality

Запрещённые labels:

- user ID;
- raw project ID при миллионах значений;
- event ID;
- trace ID;
- error message;
- arbitrary URL;
- full resource name.

Разрешённые labels SHOULD быть bounded enums: service, method, status, error code, operation type, stage, region.

## 18.8. Logs

Структурный log SHOULD содержать:

```yaml
severity: ERROR
service: m8-provisioning
component: reconciler
message_code: provisioning.activity.failed
trace_id: ...
correlation_id: ...
operation_id: operations/op_123
resource_type: ManagedResource
error_code: PROVISIONING_DRIVER_UNAVAILABLE
retryable: true
```

Свободный текст является дополнением, а не основной структурой.

## 18.9. Audit и logs

| Audit | Logs |
| --- | --- |
| доказательство действия | диагностика работы системы |
| immutable/controlled retention | operational retention |
| actor/action/target/outcome | component/context/error |
| compliance access | developer/SRE access |
| стабильная schema | может меняться быстрее |

Критичное действие MUST иметь AuditEvent даже при наличии подробного log.

## 18.10. SLI и SLO

SLI classes:

- availability;
- latency;
- correctness;
- freshness;
- durability;
- completion time;
- security control effectiveness.

Примеры:

```text
Authentication success latency p95
Access check availability
Outbox publish lag p99
Projection freshness p99
Provisioning operation completion within target
Audit event durability
```

## 18.11. Error budget

Для SLO-critical capability MUST быть error budget. Его расход влияет на release velocity, risk acceptance и remediation priority.

## 18.12. Service dashboards

Каждый сервис SHOULD иметь:

1. overview;
2. API RED metrics;
3. dependency health;
4. database/storage;
5. events/outbox/inbox;
6. operations/workflows;
7. business outcomes;
8. security signals;
9. SLO/error budget;
10. deployment version.

## 18.13. Контекстные метрики

### Resource Manager

- resource create/update/delete rate;
- hierarchy conflicts;
- stuck deletion operations;
- projection publish lag.

### Identity

- active/disabled users;
- identity linking failures;
- duplicate issuer+subject conflicts;
- privacy deletion progress.

### Authentication

- flow start/completion;
- challenge success/failure;
- step-up rate;
- refresh failure → re-auth rate;
- provider latency;
- suspicious attempts.

### Access

- check latency/availability;
- allow/deny distribution;
- consistency token errors;
- relationship write lag;
- SpiceDB dependency health.

### Risk Decision

- decisions by outcome/reason;
- policy version distribution;
- evaluation latency;
- signal availability;
- manual review backlog.

### Provisioning

- operation duration;
- reconciliation rate;
- drift count;
- external provider errors;
- retry/compensation;
- resources by condition.

### Audit

- ingestion durability;
- rejected events;
- integrity verification;
- query/export duration;
- retention/deletion jobs.

## 18.14. Alerting

Alert SHOULD быть symptom-based и actionable. Каждый alert имеет:

- name;
- condition;
- severity;
- owner;
- impact;
- runbook;
- dashboard;
- deduplication key;
- escalation;
- silence policy.

Alert на каждую единичную ошибку без агрегации запрещён.

## 18.15. Stuck detection

Operation/workflow считается stuck по типизированным thresholds:

- долго в QUEUED;
- отсутствует progress;
- повторяется один stage;
- heartbeat отсутствует;
- outbox не публикуется;
- reconciliation не достигает convergence.

Stuck detector MUST не создавать duplicate remediation.

## 18.16. Sampling

Sampling policy MUST учитывать:

- baseline head sampling;
- tail sampling errors/slow requests;
- 100% security-critical decision traces при допустимой privacy;
- tenant fairness;
- cost limits;
- incident override.

## 18.17. Telemetry retention

Retention различается:

- high-volume traces — короткий срок;
- metrics — агрегированный долгий срок;
- logs — operational/compliance class;
- audit — policy/legal retention;
- profiles — строго ограниченный доступ.

## 18.18. Data quality observability

Критичные pipelines MUST измерять:

- event lag;
- projection completeness;
- revision gaps;
- duplicates;
- reconciliation mismatch;
- schema rejection;
- stale data age.

## 18.19. Deployment observability

Telemetry MUST включать:

- build/version;
- environment;
- region/zone;
- deployment ID;
- configuration revision;
- feature flag/policy version where relevant.

Это позволяет связать regressions с изменением.

## 18.20. Runbooks

Runbook MUST содержать:

- symptom;
- user impact;
- diagnostics queries;
- safe remediation;
- rollback;
- escalation;
- evidence preservation;
- post-check.

## 18.21. Тестирование наблюдаемости

- trace propagation tests;
- log redaction tests;
- metric cardinality checks;
- alert simulation;
- dashboard query validation;
- audit completeness;
- telemetry under dependency failure;
- sampling policy tests;
- stuck operation tests.

## 18.22. SPDD-требования

Feature prompt MUST указать:

```yaml
observability:
  spans: [StartAuthentication, EvaluateRisk, CreateTransaction]
  metrics:
    - authentication_started_total
    - authentication_start_duration_seconds
  logs:
    - code: authentication.start.failed
      sensitive_fields: [subject_hint]
  audit_events:
    - AuthenticationStarted
  alerts:
    - authentication_provider_error_rate_high
  slo_impact: interactive_authentication
```

## 18.23. Критерии соответствия главы

Feature соответствует PADS, если его request/event/workflow можно проследить, пользовательский результат измерим, logs structured/redacted, metrics bounded, SLO и alerts имеют owner/runbook, а audit отделён от diagnostics.

---
