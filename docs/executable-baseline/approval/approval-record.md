---
title: "Протокол утверждения архитектурной базовой линии"
---

# Протокол утверждения архитектурной базовой линии

[Executable Baseline](../index.md) | [Approval](index.md)

| Поле | Значение |
| --- | --- |
| Решение | `APPROVED` |
| Дата | 10 июля 2026 года |
| Область | M8 Platform |
| Утверждены | PADS, Requirements Catalog, API/Event/Error/Data catalogs, ADR, SPDD Constitution |
| Вступление в силу | Немедленно |
| Изменение | Только через ADR и обновление Traceability Registry |

## Утверждённые решения

1. Границы контекстов Resource Manager, Identity, Authentication, Access, Risk Decision,
   Provisioning и Audit принимаются как нормативные.
2. YDB принимается как основное транзакционное хранилище control plane.
3. Kafka является основной межсервисной шиной событий; YDB Topics допускается для
   локальных YDB-centric потоков после проверки совместимости и эксплуатационной модели.
4. Temporal используется для длительных межконтекстных процессов и компенсаций.
5. Все длительные операции экспонируются через Common Operation,
   совместимый с `google.longrunning.Operation`.
6. Keycloak используется как внешний IAM/OIDC runtime; CIBA является основным
   безредиректным сценарием аутентификации.
7. SpiceDB используется как движок relationship-based authorization.
8. Публичные контракты задаются Protobuf и обслуживаются ConnectRPC/gRPC.
9. Межсервисная доставка фактов использует Transactional Outbox/Inbox.
10. Multi-region начинается с active/passive control plane и эволюционирует
    к cell-based модели только после доказанной необходимости.

## Утверждение требований

Все 214 требований имеют архитектурный статус `APPROVED`.
Реализация разбита на `MVP-1`, `MVP-2`, `MVP-3` и `GA-1`.
