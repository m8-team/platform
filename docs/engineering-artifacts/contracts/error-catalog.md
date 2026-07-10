---
title: "M8 Platform Error Catalog"
description: "Каталог канонических ошибок."
keywords:
  - "M8 Platform"
  - "contracts"
---

# M8 Platform Error Catalog {#error-catalog}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Contracts](index.md) | [PADS: модель ошибок](../../architecture/pads/platform/17-errors.md) | `error-catalog.yaml`

{% endnote %}

_M8-ERR-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-ERR-000` |
| Версия | `0.1` |
| Статус | Базовая проектная редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Канонические platform и service errors |

# 1. Правила

- Код ошибки стабилен и пригоден для машинной обработки.
- Пользовательское сообщение не раскрывает внутренние идентификаторы, provider errors или секреты.
- Retryable задаётся каталогом, но клиент также обязан учитывать deadline и idempotency.
- gRPC/Connect mapping не заменяет domain code.

# 2. Реестр

| ID | Code | gRPC | HTTP | Retry | Описание |
| --- | --- | --- | --- | --- | --- |
| `ERR-COMMON-001` | `INVALID_ARGUMENT` | `INVALID_ARGUMENT` | 400 | нет | Один или несколько параметров не прошли проверку. |
| `ERR-COMMON-002` | `UNAUTHENTICATED` | `UNAUTHENTICATED` | 401 | нет | Не подтверждена идентичность вызывающей стороны. |
| `ERR-COMMON-003` | `PERMISSION_DENIED` | `PERMISSION_DENIED` | 403 | нет | Недостаточно полномочий для операции. |
| `ERR-COMMON-004` | `NOT_FOUND` | `NOT_FOUND` | 404 | нет | Запрошенный ресурс не найден или не раскрывается вызывающей стороне. |
| `ERR-COMMON-005` | `ALREADY_EXISTS` | `ALREADY_EXISTS` | 409 | нет | Ресурс с конфликтующим уникальным ключом уже существует. |
| `ERR-COMMON-006` | `FAILED_PRECONDITION` | `FAILED_PRECONDITION` | 412 | нет | Состояние ресурса не допускает операцию. |
| `ERR-COMMON-007` | `REVISION_MISMATCH` | `ABORTED` | 409 | да | Переданная revision/ETag устарела. |
| `ERR-COMMON-008` | `IDEMPOTENCY_CONFLICT` | `ALREADY_EXISTS` | 409 | нет | Ключ идемпотентности ранее использован с другим телом. |
| `ERR-COMMON-009` | `RATE_LIMITED` | `RESOURCE_EXHAUSTED` | 429 | да | Превышен лимит запросов или попыток. |
| `ERR-COMMON-010` | `DEADLINE_EXCEEDED` | `DEADLINE_EXCEEDED` | 504 | да | Операция не завершена в установленный срок. |
| `ERR-COMMON-011` | `DEPENDENCY_UNAVAILABLE` | `UNAVAILABLE` | 503 | да | Обязательная зависимость временно недоступна. |
| `ERR-COMMON-012` | `INTERNAL_ERROR` | `INTERNAL` | 500 | да | Непредвиденная внутренняя ошибка без раскрытия деталей. |
| `ERR-COMMON-013` | `CANCELLED` | `CANCELLED` | 499 | нет | Операция отменена. |
| `ERR-COMMON-014` | `SCOPE_REQUIRED` | `INVALID_ARGUMENT` | 400 | нет | Не указана обязательная ресурсная область. |
| `ERR-COMMON-015` | `POLICY_DENIED` | `PERMISSION_DENIED` | 403 | нет | Политика безопасности запретила действие. |
| `ERR-COMMON-016` | `STEP_UP_REQUIRED` | `FAILED_PRECONDITION` | 412 | нет | Для продолжения требуется более высокий AAL. |
| `ERR-COMMON-017` | `RISK_REVIEW_REQUIRED` | `FAILED_PRECONDITION` | 412 | нет | Требуется ручная проверка риска. |
| `ERR-COMMON-018` | `OPERATION_IN_PROGRESS` | `FAILED_PRECONDITION` | 412 | да | Эквивалентная операция уже выполняется. |
| `ERR-COMMON-019` | `DATA_CLASSIFICATION_VIOLATION` | `FAILED_PRECONDITION` | 412 | нет | Данные нарушают правила классификации или минимизации. |
| `ERR-COMMON-020` | `CONTRACT_VERSION_UNSUPPORTED` | `UNIMPLEMENTED` | 501 | нет | Версия публичного контракта не поддерживается. |
| `ERR-RM-001` | `RESOURCE_HAS_DEPENDENCIES` | `FAILED_PRECONDITION` | 412 | нет | Ресурс нельзя удалить до обработки зависимостей. |
| `ERR-RM-002` | `RESOURCE_PATH_CONFLICT` | `ALREADY_EXISTS` | 409 | нет | Целевой путь или имя ресурса конфликтует. |
| `ERR-RM-003` | `MOVE_NOT_ALLOWED` | `FAILED_PRECONDITION` | 412 | нет | Перемещение нарушает иерархические инварианты. |
| `ERR-ID-001` | `EXTERNAL_IDENTITY_CONFLICT` | `ALREADY_EXISTS` | 409 | нет | Внешняя идентичность уже связана с другим Subject. |
| `ERR-ID-002` | `LAST_IDENTITY_UNLINK_FORBIDDEN` | `FAILED_PRECONDITION` | 412 | нет | Нельзя отвязать последнюю пригодную идентичность. |
| `ERR-ID-003` | `USER_MERGE_CONFLICT` | `FAILED_PRECONDITION` | 412 | нет | Пользователи не могут быть безопасно объединены. |
| `ERR-AUTH-001` | `CLIENT_NOT_FOUND` | `NOT_FOUND` | 404 | нет | Клиент не найден. |
| `ERR-AUTH-002` | `CLIENT_DISABLED` | `FAILED_PRECONDITION` | 412 | нет | Клиент отключён. |
| `ERR-AUTH-003` | `FLOW_NOT_ALLOWED` | `PERMISSION_DENIED` | 403 | нет | Клиенту не разрешён запрошенный flow. |
| `ERR-AUTH-004` | `SUBJECT_NOT_FOUND` | `NOT_FOUND` | 404 | нет | Subject не разрешён. |
| `ERR-AUTH-005` | `CHALLENGE_EXPIRED` | `FAILED_PRECONDITION` | 412 | нет | Challenge истёк. |
| `ERR-AUTH-006` | `CHALLENGE_ATTEMPTS_EXCEEDED` | `RESOURCE_EXHAUSTED` | 429 | нет | Исчерпан лимит попыток challenge. |
| `ERR-AUTH-007` | `CALLBACK_INVALID` | `UNAUTHENTICATED` | 401 | нет | Callback не прошёл проверку подписи, state или nonce. |
| `ERR-AUTH-008` | `HANDOFF_ALREADY_REDEEMED` | `FAILED_PRECONDITION` | 412 | нет | Handoff уже использован. |
| `ERR-AUTH-009` | `PROVIDER_UNAVAILABLE` | `UNAVAILABLE` | 503 | да | Поставщик аутентификации временно недоступен. |
| `ERR-AUTH-010` | `RISK_DECISION_UNAVAILABLE` | `UNAVAILABLE` | 503 | да | Невозможно получить обязательное решение риска. |
| `ERR-ACC-001` | `AUTHORIZATION_MODEL_INVALID` | `INVALID_ARGUMENT` | 400 | нет | Модель полномочий не прошла проверку. |
| `ERR-ACC-002` | `RELATIONSHIP_CONFLICT` | `ABORTED` | 409 | да | Изменение отношения конфликтует с текущей revision. |
| `ERR-ACC-003` | `ACCESS_BACKEND_UNAVAILABLE` | `UNAVAILABLE` | 503 | да | Движок доступа недоступен; применяется заданный fail mode. |
| `ERR-RISK-001` | `RISK_POLICY_INVALID` | `INVALID_ARGUMENT` | 400 | нет | Политика риска не прошла валидацию. |
| `ERR-RISK-002` | `RISK_DECISION_EXPIRED` | `FAILED_PRECONDITION` | 412 | нет | Решение риска больше нельзя использовать. |
| `ERR-RISK-003` | `INSUFFICIENT_RISK_SIGNALS` | `FAILED_PRECONDITION` | 412 | да | Недостаточно обязательных сигналов для решения. |
| `ERR-PROV-001` | `RESOURCE_DEFINITION_NOT_FOUND` | `NOT_FOUND` | 404 | нет | Определение управляемого ресурса не найдено. |
| `ERR-PROV-002` | `DRIVER_UNAVAILABLE` | `UNAVAILABLE` | 503 | да | Драйвер временно недоступен. |
| `ERR-PROV-003` | `PLACEMENT_UNAVAILABLE` | `FAILED_PRECONDITION` | 412 | да | Подходящее размещение не найдено. |
| `ERR-PROV-004` | `RECONCILIATION_FAILED` | `INTERNAL` | 500 | да | Reconciliation завершён ошибкой. |
| `ERR-PROV-005` | `MANUAL_REMEDIATION_REQUIRED` | `FAILED_PRECONDITION` | 412 | нет | Автоматическое восстановление невозможно. |
| `ERR-AUD-001` | `AUDIT_PROVENANCE_INVALID` | `UNAUTHENTICATED` | 401 | нет | Происхождение AuditEvent не подтверждено. |
| `ERR-AUD-002` | `AUDIT_EVENT_REJECTED` | `INVALID_ARGUMENT` | 400 | нет | AuditEvent нарушает схему или правила минимизации. |
| `ERR-AUD-003` | `AUDIT_INTEGRITY_VIOLATION` | `DATA_LOSS` | 500 | нет | Проверка целостности выявила несоответствие. |
| `ERR-OPS-001` | `OPERATION_NOT_CANCELLABLE` | `FAILED_PRECONDITION` | 412 | нет | Операция не поддерживает отмену в текущем состоянии. |
| `ERR-OPS-002` | `OPERATION_RESULT_TYPE_MISMATCH` | `INTERNAL` | 500 | нет | Тип результата не соответствует контракту операции. |

# 3. Error details

Разрешены типизированные details: `FieldViolation`, `ResourceInfo`, `RetryInfo`, `PreconditionFailure`, `ErrorInfo`, `Help`. Внешние provider payload и stack trace клиенту не возвращаются.
