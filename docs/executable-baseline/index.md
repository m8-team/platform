---
title: Executable Architecture Baseline
---

# M8 Platform Executable Architecture Baseline

**Идентификатор:** `M8-EAB-001`  
**Версия:** `1.0`  
**Статус:** `APPROVED`  
**Дата:** 10 июля 2026 года

{% note info %}

Этот раздел переводит [PADS](../architecture/pads/index.md), [Requirements Catalog](../architecture/requirements/index.md) и [Engineering Artifacts](../engineering-artifacts/index.md) в исполнимую базовую линию: утвержденные контракты, схемы событий, DDL, Go-каркас, SPDD-пакеты, CI/CD, тестовые планы, эксплуатацию и security-модели.

{% endnote %}

## Навигация

| Раздел | Содержание |
| --- | --- |
| [Approval](approval/index.md) | протокол утверждения и реестр approved artifacts |
| [Product](product/index.md) | MVP scope, release waves, approved requirements |
| [Contracts](contracts/index.md) | Protobuf, ConnectRPC, AsyncAPI, event schemas, errors |
| [Data](data/index.md) | YDB DDL, ownership, sizing, migration and backup policy |
| [Repository](repository/index.md) | executable Go repository skeleton and pilot vertical slice |
| [Deployment](deployment/index.md) | Kubernetes baseline, topology, secrets and DR |
| [CI](ci/index.md) | GitLab CI, Makefile and architecture validation scripts |
| [Testing](testing/index.md) | test strategy, acceptance, load, chaos and DR plans |
| [SPDD](spdd/index.md) | execution model and generated feature/review/task prompts |
| [Operations](operations/index.md) | SLO, alerts, dashboard spec and runbooks |
| [UI](ui/index.md) | Gravity UI specification |
| [SDK and CLI](sdk/index.md) | SDK/CLI contract and client examples |
| [Security](security/index.md) | security verification plan and threat models |
| [Roadmap](roadmap/index.md) | implementation roadmap, Definition of Ready and Done |
| [Reference](reference/index.md) | canonical links replacing historical embedded reference copies |

## Baseline Counts

| Объект | Количество |
| --- | ---: |
| Approved requirements | 214 |
| MVP-1 requirements | 83 |
| API contracts | 156 |
| Event contracts | 116 |
| Event JSON schemas | 117 |
| Canonical errors | 52 |
| Data entities | 38 |
| Protobuf files | 15 |
| YDB DDL files | 11 |
| Kubernetes manifests | 9 |
| Go scaffold files | 18 |
| SPDD feature prompts | 214 |
| SPDD review prompts | 214 |
| SPDD task prompts | 342 |

## Raw Control Files

| Артефакт | Путь |
| --- | --- |
| Baseline manifest | `manifest.yaml` |
| Validation report | [validation-report.md](validation-report.md) |
