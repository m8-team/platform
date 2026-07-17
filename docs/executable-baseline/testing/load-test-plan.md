---
title: "План нагрузочного тестирования"
---

# План нагрузочного тестирования

[Executable Baseline](../index.md) | [Testing Baseline](index.md)

Профили: permission check, StartAuthentication, audit ingestion, event dispatch,
Resource Manager reads, Operation polling. Для каждого профиля измеряются p50/p95/p99,
error rate, queue lag, CPU, memory, YDB RU/latency и saturation.

Обязательные тесты: steady state, spike x5, soak 8h, dependency slowdown,
hot partition и replay backlog.
