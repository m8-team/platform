---
title: "M8 Platform ADR Baseline"
description: "ADR baseline."
keywords:
  - "M8 Platform"
  - "ADR"
---

# M8 Platform ADR Baseline {#adr-baseline}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [ADR index](index.md) | [PADS governance](../../architecture/pads/governance/23-architecture-governance.md)

{% endnote %}

_M8-ADR-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-ADR-000` |
| Версия | `0.1` |
| Статус | Базовый набор решений |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | ADR-0001—ADR-0010 |

# 1. Реестр

| ID | Статус | Название | Основание |
| --- | --- | --- | --- |
| `ADR-0001` | Accepted | Границы bounded contexts и сервисная декомпозиция | PADS §8–10 |
| `ADR-0002` | Accepted | YDB как основное транзакционное хранилище control plane | PADS §11, §14 |
| `ADR-0003` | Proposed | YDB Topics и Kafka для интеграционных событий | PADS §13–14 |
| `ADR-0004` | Accepted | Temporal для длительных межсервисных процессов | PADS §14, §16 |
| `ADR-0005` | Accepted | Common Operation на основе google.longrunning | PADS §10, §16 |
| `ADR-0006` | Accepted | Keycloak как поставщик протокольной аутентификации и CIBA | PADS §9.4, §15 |
| `ADR-0007` | Accepted | SpiceDB как движок проверки отношений доступа | PADS §9.5, §15 |
| `ADR-0008` | Accepted | Protobuf и ConnectRPC как основной API transport | PADS §12 |
| `ADR-0009` | Accepted | Transactional Outbox и Inbox | PADS §13–14 |
| `ADR-0010` | Proposed | Базовые предположения multi-region | PADS §11, §19 |

# 2. Правила

Accepted ADR обязателен до его supersede. Proposed ADR блокирует необратимую реализацию затрагиваемой части. Любое изменение PADS через ADR должно содержать impact, migration и дату пересмотра.
