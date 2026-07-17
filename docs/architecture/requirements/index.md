---
title: "Requirements Catalog — M8 Platform"
description: "Канонический каталог требований M8 Platform."
keywords:
  - "M8 Platform"
  - "requirements"
---

# M8 Platform Requirements Catalog {#requirements-catalog}
_M8-REQ-000 · Версия 0.1 · Базовый каталог требований · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор документа | `M8-REQ-000` |
| Версия | `0.1` |
| Статус | Базовая аналитическая редакция (`ANALYZED`) |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0` |
| Область | Все ограниченные контексты M8 Platform и сквозные требования |
| Формат | Markdown; пригоден для последующего преобразования в YAML/JSON |

{% note warning "Статус каталога" %}

требования сформированы на основе PADS v1.0 и готовы к продуктовой приоритизации, уточнению контрактов и проведению через SPDD. Статус `ANALYZED` не означает, что все требования включены в один релиз.

{% endnote %}

## Связанные источники

- [PADS](../pads/index.md) — нормативная архитектурная основа.
- [Engineering artifacts](../../engineering-artifacts/index.md) — контракты, traceability, SPDD, ADR и pilot package.
- [Traceability Registry](../../engineering-artifacts/traceability/traceability-registry.md) — машинная трассировка требований к контрактам и проверкам.

## Разделы

| Раздел | Содержание | Главы |
| --- | --- | --- |
| [Обзор и модель](overview/index.md) | Управление артефактом, модель требования и сводное распределение. | [0](overview/00-artifact-control.md), [1](overview/01-requirement-model.md), [2](overview/02-distribution-map.md) |
| [Требования по контекстам](contexts/index.md) | Детальные требования по владельцам и ограниченным контекстам. | [3](contexts/03-platform-cross-cutting.md), [4](contexts/04-resource-manager.md), [5](contexts/05-identity.md), [6](contexts/06-authentication.md), [7](contexts/07-access.md), [8](contexts/08-risk-decision.md), [9](contexts/09-provisioning.md), [10](contexts/10-audit.md), [11](contexts/11-common-operation.md) |
| [Управление и SPDD](governance/index.md) | Сквозные сценарии, контрактная подготовка, приоритеты, SPDD и машинный реестр. | [12](governance/12-architecture-governance-spdd.md), [13](governance/13-cross-cutting-scenarios.md), [14](governance/14-contract-preparation-map.md), [15](governance/15-implementation-priorities.md), [16](governance/16-spdd-backlog.md), [17](governance/17-machine-registry.md) |
| [Приложения](appendices/index.md) | Шаблон требования и проверка целостности каталога. | [A](appendices/appendix-a-requirement-template.md), [B](appendices/appendix-b-integrity-check.md) |

## Полное оглавление

| Глава | Раздел | Назначение |
| --- | --- | --- |
| [0. Управление артефактом](overview/00-artifact-control.md) | Обзор и модель | Назначение, нормативные правила, состояния и статистика редакции. |
| [1. Модель требования](overview/01-requirement-model.md) | Обзор и модель | Обязательные поля требования, Definition of Ready и Definition of Done. |
| [2. Сводная карта распределения](overview/02-distribution-map.md) | Обзор и модель | Распределение требований по владельцам, типам и сервисам. |
| [3. Платформенные и сквозные требования](contexts/03-platform-cross-cutting.md) | Требования по контекстам | Платформенные функциональные, архитектурные, security и NFR требования. |
| [4. Resource Manager](contexts/04-resource-manager.md) | Требования по контекстам | Требования Resource Manager. |
| [5. Identity](contexts/05-identity.md) | Требования по контекстам | Требования Identity. |
| [6. Authentication](contexts/06-authentication.md) | Требования по контекстам | Требования Authentication. |
| [7. Access](contexts/07-access.md) | Требования по контекстам | Требования Access. |
| [8. Risk Decision](contexts/08-risk-decision.md) | Требования по контекстам | Требования Risk Decision. |
| [9. Provisioning](contexts/09-provisioning.md) | Требования по контекстам | Требования Provisioning. |
| [10. Audit](contexts/10-audit.md) | Требования по контекстам | Требования Audit. |
| [11. Common Operation](contexts/11-common-operation.md) | Требования по контекстам | Требования Common Operation. |
| [12. Архитектурное управление и SPDD](governance/12-architecture-governance-spdd.md) | Управление и SPDD | Требования к архитектурному управлению и SPDD. |
| [13. Сквозные сценарии и декомпозиция требований](governance/13-cross-cutting-scenarios.md) | Управление и SPDD | Сквозные сценарии и правила декомпозиции требований. |
| [14. Карта подготовки контрактов](governance/14-contract-preparation-map.md) | Управление и SPDD | Карта подготовки API, событий, данных и операций. |
| [15. Приоритеты реализации](governance/15-implementation-priorities.md) | Управление и SPDD | Приоритетные волны реализации. |
| [16. SPDD backlog](governance/16-spdd-backlog.md) | Управление и SPDD | Backlog Structured Prompt Driven Development. |
| [17. Требования к машинному реестру](governance/17-machine-registry.md) | Управление и SPDD | Требования к машинному представлению каталога. |
| [Приложение A. Шаблон нового требования](appendices/appendix-a-requirement-template.md) | Приложения | Шаблон нового требования. |
| [Приложение B. Проверка целостности редакции](appendices/appendix-b-integrity-check.md) | Приложения | Проверка целостности редакции каталога. |
