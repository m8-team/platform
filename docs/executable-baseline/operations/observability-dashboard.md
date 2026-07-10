---
title: "Наблюдаемость и dashboards"
---

# Наблюдаемость и dashboards

[Executable Baseline](../index.md) | [Operations Baseline](index.md)

Каждый сервис публикует RED-метрики, dependency latency, YDB transaction latency,
Outbox/Inbox lag, operation state, audit delivery и business counters.
Все логи содержат request, correlation, causation, actor, project, operation и trace IDs.

Обязательные dashboards:
1. Platform Executive SLO;
2. Authentication funnel;
3. Access decision latency/cache;
4. Outbox/Inbox delivery;
5. YDB saturation and hot partitions;
6. Temporal workflow health;
7. Audit integrity and export;
8. Provisioning reconciliation/drift.
