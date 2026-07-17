---
title: "Runbook: Operation Stuck"
---

# Runbook: Operation Stuck

[Executable Baseline](../../index.md) | [Runbooks](../index.md)

## Действия

Найти workflow_id/run_id, проверить worker queues, retries и compensation. Состояние Operation менять только владельцу workflow.

## Evidence

Зафиксировать incident ID, временной диапазон, trace/correlation IDs,
принятые решения, команды восстановления и итоговую проверку SLO.
