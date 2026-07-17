---
title: "Runbook: Outbox Lag"
---

# Runbook: Outbox Lag

[Executable Baseline](../../index.md) | [Runbooks](../index.md)

## Действия

Проверить dispatcher health, Kafka availability, pending age, retries и poison events. Не удалять события вручную.

## Evidence

Зафиксировать incident ID, временной диапазон, trace/correlation IDs,
принятые решения, команды восстановления и итоговую проверку SLO.
