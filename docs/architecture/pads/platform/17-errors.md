---
title: "PADS: модель ошибок"
description: "Слои ошибок, категории, коды, retryability и error catalog."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 17. Модель ошибок {#pads-errors}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: Long-Running Operations (LRO)](16-operations.md) | [Следующий раздел: 18. Наблюдаемость](18-observability.md)

{% endnote %}

## 17.1. Назначение главы

Модель ошибок обеспечивает единообразное различение transport failure, validation problem, business rejection, security denial, concurrency conflict и infrastructure failure. Ошибка должна быть одновременно полезной клиенту, оператору и автоматизированному обработчику, не раскрывая чувствительные детали.

## 17.2. Слои ошибки

| Слой | Назначение |
| --- | --- |
| Transport status | широкая категория gRPC/HTTP |
| Service error code | стабильная предметная причина |
| Details | типизированные параметры и ссылки |
| User message code | локализуемое объяснение |
| Operator diagnostics | безопасные внутренние детали по correlation ID |
| Retry metadata | возможность и задержка повторения |

## 17.3. Принципы

| ID | Правило |
| --- | --- |
| `ERR-001` | Клиент MUST принимать решение по code/category, а не парсить текст. |
| `ERR-002` | Error code MUST принадлежать owner service. |
| `ERR-003` | Один code MUST иметь стабильный смысл внутри major API version. |
| `ERR-004` | Domain error MUST быть transport-independent до adapter mapping. |
| `ERR-005` | Internal stack trace MUST NOT возвращаться клиенту. |
| `ERR-006` | Error response MUST NOT раскрывать existence/security details, если это создаёт enumeration risk. |
| `ERR-007` | Retryability MUST быть явной. |
| `ERR-008` | Optimistic conflict MUST отличаться от generic internal error. |
| `ERR-009` | Validation MUST указывать field violations. |
| `ERR-010` | Long-running failure MUST сохраняться в Operation. |
| `ERR-011` | Provider errors MUST переводиться ACL в M8 taxonomy. |
| `ERR-012` | Unexpected error MUST иметь correlation ID и safe public message. |
| `ERR-013` | Error code catalog MUST быть versioned и documented. |
| `ERR-014` | Logs MUST не дублировать sensitive payload только из-за ошибки. |
| `ERR-015` | Structured Prompt MUST перечислять expected errors и mapping. |

## 17.4. Канонические категории

| Категория | gRPC | HTTP | Пример |
| --- | --- | --- | --- |
| Invalid input | `INVALID_ARGUMENT` | 400 | неверный filter, field violation |
| Unauthenticated | `UNAUTHENTICATED` | 401 | token отсутствует/недействителен |
| Permission denied | `PERMISSION_DENIED` | 403 | нет права выполнить действие |
| Not found | `NOT_FOUND` | 404 | ресурс отсутствует или скрыт policy |
| Conflict | `ALREADY_EXISTS`/`ABORTED` | 409 | duplicate, revision conflict |
| Failed precondition | `FAILED_PRECONDITION` | 400/409 | недопустимое состояние |
| Resource exhausted | `RESOURCE_EXHAUSTED` | 429 | quota/rate limit |
| Cancelled | `CANCELLED` | 499/409 | операция отменена |
| Deadline | `DEADLINE_EXCEEDED` | 504 | budget исчерпан |
| Unavailable | `UNAVAILABLE` | 503 | transient dependency outage |
| Internal | `INTERNAL` | 500 | unexpected invariant/infrastructure failure |
| Data loss | `DATA_LOSS` | 500 | обнаружено повреждение данных |

## 17.5. Формат

```yaml
error:
  category: FAILED_PRECONDITION
  code: AUTH_CHALLENGE_NOT_PENDING
  message_code: authentication.challenge.not_pending
  message: Challenge cannot be completed in its current state.
  retryable: false
  correlation_id: corr_123
  details:
    resource: authenticationChallenges/ch_123
    current_state: EXPIRED
    allowed_states: [PENDING]
```

`message` MAY быть fallback и не является стабильным контрактом.

## 17.6. Именование кодов

Формат:

```text
<DOMAIN>_<CONDITION>
```

Примеры:

- `PROJECT_NOT_ACTIVE`;
- `USER_ALREADY_DISABLED`;
- `AUTH_CHALLENGE_EXPIRED`;
- `ACCESS_RELATIONSHIP_CONFLICT`;
- `RISK_POLICY_NOT_PUBLISHED`;
- `PROVISIONING_DRIVER_UNAVAILABLE`;
- `AUDIT_EXPORT_TOO_LARGE`.

Коды не должны содержать vendor name, если ошибка имеет предметный смысл.

## 17.7. Validation details

Validation response SHOULD содержать список:

```yaml
violations:
  - field: project.display_name
    rule: string.min_len
    description_code: common.validation.too_short
  - field: update_mask.paths[0]
    rule: mutable_field
    description_code: common.validation.output_only
```

Необходимо различать синтаксическую и предметную validation.

## 17.8. Precondition failures

Precondition detail SHOULD включать:

- resource;
- violated condition type;
- current state/revision;
- required state;
- remediation code;
- conflicting operation reference.

## 17.9. Concurrency conflicts

Ошибки:

- `ETAG_MISMATCH`;
- `REVISION_CONFLICT`;
- `IDEMPOTENCY_CONFLICT`;
- `OPERATION_ALREADY_RUNNING`.

Client SHOULD перечитать ресурс и повторно сформировать намерение; blind retry запрещён.

## 17.10. Security errors

Для anti-enumeration несколько внутренних причин MAY отображаться одинаково публично. Внутренний Audit/diagnostic code сохраняется отдельно.

Пример: login response не должен различать несуществующего и заблокированного пользователя, если policy запрещает раскрытие.

## 17.11. Retryability

| Ошибка | Retry |
| --- | --- |
| Invalid argument | нет |
| Permission denied | нет, пока не изменены права |
| Failed precondition | после изменения состояния |
| Revision conflict | после re-read/re-evaluate |
| Resource exhausted | после `retry_after` |
| Unavailable | exponential backoff + jitter |
| Deadline exceeded | только если request идемпотентен |
| Internal | обычно нет автоматического retry без classification |

## 17.12. Provider errors

ACL переводит provider codes в M8:

```text
Keycloak session_not_active → AUTH_SESSION_NOT_ACTIVE
SpiceDB invalid_revision → ACCESS_CONSISTENCY_TOKEN_INVALID
Cloud quota exceeded → PROVISIONING_PROVIDER_QUOTA_EXCEEDED
Kubernetes forbidden → PROVISIONING_PROVIDER_PERMISSION_DENIED
```

Raw provider body MAY сохраняться только в protected diagnostics.

## 17.13. Ошибки событий

Consumer error classification:

- transient processing;
- schema unsupported;
- validation invalid;
- stale revision;
- missing dependency;
- permanent business rejection;
- poison event;
- side-effect uncertain.

Каждая категория имеет retry/DLQ/reconciliation policy.

## 17.14. Ошибки Operation

Operation failure MUST хранить original stable code и failed stage. Повторное получение Operation возвращает тот же terminal error. Временные retry attempts не должны преждевременно становиться terminal failure.

## 17.15. Локализация

Клиентская локализация выполняется по `message_code` и parameters. Серверные английские/русские тексты не считаются стабильным API. Security-sensitive parameters проходят allowlist.

## 17.16. Логирование

Unexpected errors логируются один раз на ответственном уровне с:

- code/category;
- correlation/trace;
- operation/resource reference;
- safe context;
- stack trace в protected log;
- owner component.

Повторное логирование на каждом слое SHOULD избегаться.

## 17.17. Error Catalog

```yaml
error_code:
  code: PROJECT_NOT_ACTIVE
  owner: m8-resource-manager
  category: FAILED_PRECONDITION
  retryable: conditional
  public: true
  message_code: resource_manager.project.not_active
  requirement_ids: [RM-FR-020]
  methods:
    - ServicesService.RegisterService
```

## 17.18. Тестирование

- mapping domain → transport;
- error code stability;
- field violation tests;
- anti-enumeration tests;
- redaction tests;
- retry classification;
- provider ACL mapping;
- Operation terminal error persistence;
- unknown error fallback;
- localization parameter safety.

## 17.19. SPDD-требования

Prompt MUST перечислять expected error codes, triggering conditions, transport mapping, retryability, details, audit/log behavior and tests. ИИ-агент не может вводить новый public code без реестра.

## 17.20. Критерии соответствия главы

Ошибка соответствует PADS, если имеет стабильный code, корректную category, безопасный message, explicit retry semantics, typed details, correlation и documented mapping, а domain layer не зависит от transport status.

---
