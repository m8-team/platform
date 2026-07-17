---
title: "Политика миграций и восстановления"
---

# Политика миграций и восстановления

[Executable Baseline](../index.md) | [Data Baseline](index.md)

1. Миграции только вперёд; destructive change выполняется через expand/migrate/contract.
2. Каждая миграция имеет requirement/ADR linkage и проверку на копии production schema.
3. Rollback приложения не должен требовать rollback схемы.
4. Backup проверяется restore-тестом не реже одного раза в квартал.
5. RPO control plane: 15 минут; RTO: 4 часа для MVP, 1 час для GA.
6. Outbox и Inbox входят в backup, но допускают безопасный replay.
7. Audit использует отдельную retention и integrity policy.
