---
title: "M8 Platform Docs"
description: "Единая точка входа в архитектуру, требования, инженерные артефакты, исполнимую baseline-документацию и API-справочники M8 Platform."
keywords:
  - "M8 Platform"
  - "documentation"
  - "architecture"
  - "requirements"
---

# M8 Platform Docs {#m8-platform-docs}

Единая точка входа в документацию M8 Platform: архитектурную спецификацию, каталог требований, инженерные артефакты, исполнимую базовую линию и справочники API.

{% note info "Как читать" %}

Начинайте с PADS, если нужно понять архитектурные правила и границы платформы. Переходите к Requirements Catalog для требований, к Engineering Artifacts для контрактов и трассировки, к Executable Baseline для утверждённой реализации и эксплуатационных материалов.

{% endnote %}

## Быстрый вход

| Раздел | Назначение |
| --- | --- |
| [PADS](architecture/pads/index.md) | Нормативная архитектурная и доменная спецификация M8 Platform. |
| [Requirements Catalog](architecture/requirements/index.md) | Канонический каталог требований, классы требований и распределение по контекстам. |
| [Engineering Artifacts](engineering-artifacts/index.md) | Контракты, ADR, traceability, SPDD, pilot package и governance CI. |
| [Executable Architecture Baseline](executable-baseline/index.md) | Утверждённая baseline: contracts, DDL, Go scaffold, CI/CD, operations, security и testing. |

## Основные маршруты

| Задача | Куда идти |
| --- | --- |
| Разобраться в границах платформы | [Назначение и область действия](architecture/pads/overview/01-scope.md) |
| Проверить bounded contexts | [Карта контекстов](architecture/pads/domain/08-context-map.md) и [спецификации контекстов](architecture/pads/domain/09-bounded-contexts.md) |
| Проектировать API | [Правила проектирования API](architecture/pads/platform/12-api-design.md) и [API Contract Catalog](engineering-artifacts/contracts/api-contract-catalog.md) |
| Проектировать события | [Правила проектирования событий](architecture/pads/platform/13-events.md) и [Event Contract Catalog](engineering-artifacts/contracts/event-contract-catalog.md) |
| Работать с длительными операциями | [Long-Running Operations (LRO)](architecture/pads/platform/16-operations.md) |
| Найти класс требования | [Классы требований](architecture/pads/governance/20-requirements.md#requirement-classes) |
| Проверить трассировку | [Traceability Registry](engineering-artifacts/traceability/traceability-registry.md) |
| Найти эксплуатационные материалы | [Operations](executable-baseline/operations/index.md) и [Runbooks](executable-baseline/operations/runbooks/index.md) |

## Сервисы и API

| Сервис | Документация |
| --- | --- |
| Access | [Обзор](services/access/index.md) · [REST API](services/access/api-reference/rest/index.md) |
| IAM | [Обзор](services/iam/index.md) · [Authentication](services/iam/api-reference/authentication.md) · [REST API](services/iam/api-reference/rest/index.md) |
| Resource Manager | [Обзор](services/resource-manager/index.md) · [Authentication](services/resource-manager/api-reference/authentication.md) · [REST API](services/resource-manager/api-reference/rest/index.md) |

## Управляющие артефакты

| Артефакт | Ссылка |
| --- | --- |
| ADR baseline | [Engineering ADR](engineering-artifacts/adr/index.md) |
| SPDD constitution | [SPDD Constitution](engineering-artifacts/spdd/constitution.md) |
| Governance CI | [Fitness functions и validator](engineering-artifacts/governance-ci/index.md) |
| Validation report | [Executable Baseline Validation Report](executable-baseline/validation-report.md) |
