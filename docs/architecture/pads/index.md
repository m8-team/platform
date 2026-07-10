---
title: "PADS — M8 Platform Architecture & Domain Specification"
description: "Нормативная архитектурная и доменная спецификация M8 Platform."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
  - "domain specification"
---

# PADS: M8 Platform Architecture & Domain Specification {#pads}
_PADS-000 · Версия 1.0 · Базовая архитектура и предметная модель · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор документа | PADS-000 |
| Версия | 1.0 |
| Статус | Базовая нормативная спецификация |
| Владелец | Sergey Gorbachev |
| Платформа | M8 Platform |
| Область | Resource Manager, Identity, Authentication, Access, Risk Decision, Provisioning, Audit, Common Operation |
| Архитектурный стиль | Domain-Driven Design, Clean Architecture, API First, Event-Driven, Control Plane |
| Базовый стек | Go, Protobuf, ConnectRPC, buf.validate / Protovalidate, YDB, YDB Topics, Redis, Temporal, SpiceDB, Keycloak, OpenTelemetry |

{% note warning "Нормативный статус" %}

Настоящий документ является основным источником истины для границ платформы, предметного языка, владения данными, распределения требований и SPDD. Любое отклонение должно быть оформлено ADR или ограниченным по сроку архитектурным исключением.

{% endnote %}

## Как пользоваться

- Начинайте с [назначения и области действия](overview/01-scope.md), если нужно понять границы PADS.
- Используйте [карту контекстов](domain/08-context-map.md) и [спецификации ограниченных контекстов](domain/09-bounded-contexts.md) при проектировании модулей и сервисов.
- Проверяйте изменения API, событий и данных через [правила API](platform/12-api-design.md), [правила событий](platform/13-events.md) и [владение данными](platform/11-data-ownership.md).
- Для требований и Codex/SPDD-процесса используйте [распределение требований](governance/20-requirements.md), [модель трассировки](governance/21-traceability.md) и [SPDD](governance/22-spdd.md).

## Разделы

| Раздел | Содержание | Главы |
| --- | --- | --- |
| [Обзор и основания](overview/index.md) | Нормативный статус, область действия, видение, цели и архитектурные принципы. | [0](overview/00-document-control.md), [1](overview/01-scope.md), [2](overview/02-vision.md), [3](overview/03-design-goals.md), [4](overview/04-architecture-principles.md) |
| [Предметная модель и контексты](domain/index.md) | Язык, бизнес-возможности, доменная модель, карта контекстов и Shared Kernel. | [5](domain/05-ubiquitous-language.md), [6](domain/06-capability-map.md), [7](domain/07-domain-model.md), [8](domain/08-context-map.md), [9](domain/09-bounded-contexts.md), [10](domain/10-shared-kernel.md) |
| [Платформенная архитектура](platform/index.md) | Данные, API, события, интеграции, безопасность, операции, ошибки, наблюдаемость и качество. | [11](platform/11-data-ownership.md), [12](platform/12-api-design.md), [13](platform/13-events.md), [14](platform/14-integration-consistency.md), [15](platform/15-security.md), [16](platform/16-operations.md), [17](platform/17-errors.md), [18](platform/18-observability.md), [19](platform/19-quality-attributes.md) |
| [Требования, SPDD и управление](governance/index.md) | Распределение требований, трассировка, SPDD, архитектурное управление и глоссарий. | [20](governance/20-requirements.md), [21](governance/21-traceability.md), [22](governance/22-spdd.md), [23](governance/23-architecture-governance.md), [24](governance/24-glossary.md) |
| [Приложения](appendices/index.md) | Структура репозитория, минимальная готовность и план следующих артефактов. | [A](appendices/appendix-a-repository-structure.md), [B](appendices/appendix-b-definition-of-done.md), [C](appendices/appendix-c-next-artifacts.md) |

## Полное оглавление

| Глава | Раздел | Назначение |
| --- | --- | --- |
| [0. Управление документом](overview/00-document-control.md) | Обзор и основания | Версии, нормативный язык и правила использования PADS. |
| [1. Назначение и область действия](overview/01-scope.md) | Обзор и основания | Область действия, системные границы и критерии соответствия PADS. |
| [2. Видение платформы](overview/02-vision.md) | Обзор и основания | Миссия, целевое состояние и роль M8 Platform. |
| [3. Цели проектирования](overview/03-design-goals.md) | Обзор и основания | Цели предметной архитектуры, модульности, данных, безопасности, эксплуатации и SPDD. |
| [4. Архитектурные принципы](overview/04-architecture-principles.md) | Обзор и основания | Нормативные архитектурные принципы и проверки соответствия. |
| [5. Единый язык предметной области](domain/05-ubiquitous-language.md) | Предметная модель и контексты | Владение понятиями, контекстный язык и правила именования. |
| [6. Карта бизнес-возможностей платформы](domain/06-capability-map.md) | Предметная модель и контексты | Декомпозиция возможностей платформы, зависимости и связь с требованиями. |
| [7. Модель предметной области](domain/07-domain-model.md) | Предметная модель и контексты | Агрегаты, сущности, объекты-значения, события и инварианты. |
| [8. Карта контекстов](domain/08-context-map.md) | Предметная модель и контексты | Каталог ограниченных контекстов и допустимые отношения между ними. |
| [9. Спецификации ограниченных контекстов](domain/09-bounded-contexts.md) | Предметная модель и контексты | Нормативные спецификации контекстов Resource Manager, Identity, Authentication, Access, Risk Decision, Provisioning, Audit и Common Operation. |
| [10. Shared Kernel и общие контракты](domain/10-shared-kernel.md) | Предметная модель и контексты | Границы общего ядра, общие контракты и правила совместимости. |
| [11. Владение данными](platform/11-data-ownership.md) | Платформенная архитектура | Владельцы данных, проекции, репликация, удаление и retention. |
| [12. Правила проектирования API](platform/12-api-design.md) | Платформенная архитектура | API-first правила, совместимость, авторизация, ошибки и валидация. |
| [13. Правила проектирования событий](platform/13-events.md) | Платформенная архитектура | События, outbox, envelope, версионирование и потребители. |
| [14. Модель интеграции и согласованности](platform/14-integration-consistency.md) | Платформенная архитектура | Синхронные и асинхронные интеграции, согласованность и деградация. |
| [15. Архитектура безопасности](platform/15-security.md) | Платформенная архитектура | Trust boundaries, authentication, authorization, audit, secrets, threat modeling и Secure SDLC. |
| [Long-Running Operations (LRO)](platform/16-operations.md) | Платформенная архитектура | Каноническая модель long-running operations (LRO): progress, cancellation, idempotency, Temporal и safeguards. |
| [17. Модель ошибок](platform/17-errors.md) | Платформенная архитектура | Слои ошибок, категории, коды, retryability и error catalog. |
| [18. Наблюдаемость](platform/18-observability.md) | Платформенная архитектура | Tracing, metrics, logs, SLI/SLO, dashboards, alerts и runbooks. |
| [19. Атрибуты качества](platform/19-quality-attributes.md) | Платформенная архитектура | Сценарии качества, quality gates и каталог атрибутов качества. |
| [20. Распределение требований](governance/20-requirements.md) | Требования, SPDD и управление | Классы требований, пространства идентификаторов, DoR, DoD и начальное распределение. |
| [21. Модель трассировки](governance/21-traceability.md) | Требования, SPDD и управление | Граф трассировки, coverage rules, evidence и automation. |
| [22. SPDD: проведение требований до Structured Prompt](governance/22-spdd.md) | Требования, SPDD и управление | Prompt hierarchy, structured prompt schema, lifecycle, security и evaluation. |
| [23. Архитектурное управление](governance/23-architecture-governance.md) | Требования, SPDD и управление | ADR, review gates, exceptions, fitness functions и governance-процессы. |
| [24. Глоссарий](governance/24-glossary.md) | Требования, SPDD и управление | Глоссарий терминов, сокращения и правила ведения. |
| [Приложение A. Начальная структура репозитория](appendices/appendix-a-repository-structure.md) | Приложения | Начальная структура репозитория M8 Platform. |
| [Приложение B. Минимальное определение готовности](appendices/appendix-b-definition-of-done.md) | Приложения | Минимальное определение готовности для реализации. |
| [Приложение C. План последующих артефактов](appendices/appendix-c-next-artifacts.md) | Приложения | Следующие обязательные артефакты и рекомендуемый пилот. |
