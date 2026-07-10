---
title: "Runbook: Risk Unavailable"
---

# Runbook: Risk Unavailable

[Executable Baseline](../../index.md) | [Runbooks](../index.md)

## Действия

Authentication применяет fail-closed для privileged flows и явно возвращает dependency unavailable. Проверить Risk Decision, cache и circuit breaker.

## Evidence

Зафиксировать incident ID, временной диапазон, trace/correlation IDs,
принятые решения, команды восстановления и итоговую проверку SLO.
