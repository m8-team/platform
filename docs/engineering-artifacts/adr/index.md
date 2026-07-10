---
title: "Engineering Artifacts: ADR"
description: "ADR baseline and baseline decisions."
keywords:
  - "M8 Platform"
  - "ADR"
---

# ADR {#engineering-adr}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [PADS: архитектурное управление](../../architecture/pads/governance/23-architecture-governance.md)

{% endnote %}

| Артефакт | Назначение |
| --- | --- |
| [M8 Platform ADR Baseline](baseline.md) | ADR baseline. |
| [ADR-0001: bounded contexts](adr-0001-bounded-contexts.md) | Границы bounded contexts и сервисная декомпозиция. |
| [ADR-0002: YDB control plane](adr-0002-ydb-control-plane.md) | YDB как основное транзакционное хранилище control plane. |
| [ADR-0003: YDB Topics и Kafka](adr-0003-ydb-topics-kafka.md) | YDB Topics и Kafka для интеграционных событий. |
| [ADR-0004: Temporal](adr-0004-temporal.md) | Temporal для длительных межсервисных процессов. |
| [ADR-0005: Common Operation](adr-0005-common-operation.md) | Common Operation на основе google.longrunning. |
| [ADR-0006: Keycloak CIBA](adr-0006-keycloak-ciba.md) | Keycloak как поставщик протокольной аутентификации и CIBA. |
| [ADR-0007: SpiceDB](adr-0007-spicedb.md) | SpiceDB как движок проверки отношений доступа. |
| [ADR-0008: Protobuf и ConnectRPC](adr-0008-protobuf-connectrpc.md) | Protobuf и ConnectRPC как основной API transport. |
| [ADR-0009: Transactional Outbox и Inbox](adr-0009-transactional-outbox-inbox.md) | Transactional Outbox и Inbox. |
| [ADR-0010: multi-region](adr-0010-multi-region.md) | Базовые предположения multi-region. |
