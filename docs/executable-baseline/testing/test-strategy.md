---
title: "Стратегия тестирования M8"
---

# Стратегия тестирования M8

[Executable Baseline](../index.md) | [Testing Baseline](index.md)

| Уровень | Что проверяется | Gate |
| --- | --- | --- |
| Unit | агрегаты, value objects, policies | каждый MR |
| Application | use case, идемпотентность, transaction boundary | каждый MR |
| Repository | YDB mapping, optimistic locking, migration | каждый MR |
| Contract | Protobuf, event schema, compatibility | каждый MR |
| Integration | Outbox/Inbox, Kafka, Keycloak, SpiceDB, Temporal | merge/release |
| Acceptance | критерии требований | release |
| Security | authn/authz, secrets, abuse, threat controls | release |
| Load | SLO, saturation, partitioning | milestone |
| Chaos/DR | dependency failure, replay, restore | quarterly |

Ни одно требование не считается реализованным без acceptance evidence.
