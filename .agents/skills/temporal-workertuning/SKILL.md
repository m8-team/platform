---
name: temporal-workertuning
description: Tune Temporal Worker performance across Go, Java, TypeScript, Python, and .NET. Use when the user asks to "configure slot suppliers", "set up a worker tuner", "enable poller autoscaling", "tune worker cache", "fix high schedule-to-start latency", "diagnose worker bottlenecks", "right-size workers for production", "choose between fixed-size and resource-based slot suppliers", "configure maxConcurrentWorkflowTaskExecutionSize", "set up resource-based tuner", "interpret worker_task_slots_available", "reduce schedule_to_start_latency", "scale worker fleet", "configure sticky cache", or mentions Worker tuning, Worker performance, slot suppliers, composite tuners, or poller autoscaling.
version: 0.1.0
---

# Skill: temporal-workertuning

## Overview

Temporal Workers poll Task Queues, execute Workflow and Activity Tasks, and report results back to the Temporal Service. Worker performance depends on three resource dimensions: **compute** (concurrent task execution slots), **memory** (workflow cache), and **IO** (pollers and network). This skill provides guidance for tuning all three dimensions across Go, Java, TypeScript, Python, and .NET SDKs.

The key abstractions are **Slot Suppliers** (which control how many tasks a Worker accepts) and **Worker Tuners** (which assign slot suppliers to different task types). Workers can use fixed-size suppliers for predictable workloads, resource-based suppliers for automatic adjustment, or custom suppliers for full control.

## When to Use This Skill

| User intent | Start here |
|---|---|
| Understand tuning concepts (slots, suppliers, tuners) | [concepts.md](references/core/concepts.md) |
| Look up SDK default values for Worker options | [configuration-defaults.md](references/core/configuration-defaults.md) |
| Configure a Worker tuner or poller autoscaling (code) | [tuner-examples.md](references/core/tuner-examples.md) |
| Set up monitoring, interpret metrics, check alert thresholds | [metrics-and-monitoring.md](references/core/metrics-and-monitoring.md) |
| Diagnose a performance problem (high latency, slot depletion) | [troubleshooting.md](references/core/troubleshooting.md) |
| Follow a step-by-step tuning playbook | [recipes.md](references/core/recipes.md) |
| Prepare Workers for production | [recipes.md](references/core/recipes.md) (Recipe 1: Initial Setup Checklist) |
| Decide between fixed-size and resource-based suppliers | [configuration-defaults.md](references/core/configuration-defaults.md#choosing-slot-supplier-types) |
| Scale a Worker fleet up or down | [recipes.md](references/core/recipes.md) (Recipe 5: Scaling Decisions) |

## Key Constraints

- **Worker tuners supersede `maxConcurrentXXXTask` style options.** Using both causes an error at Worker initialization time.
- **`worker_task_slots_available` only works with fixed-size slot suppliers.** Use `worker_task_slots_used` for resource-based suppliers.
- **`PollerBehaviorAutoscaling` requires Temporal Server v1.28.0 or later.**
- **`maxWorkflowThreadCount` and `workflow_active_thread_count` are Java SDK only.**
- **Resource-based suppliers do not guarantee targets will not be exceeded.** Resources consumed during a task cannot be known ahead of time.

## Reference Files

- **[`references/core/concepts.md`](references/core/concepts.md)** — Task slots, slot suppliers (fixed-size, resource-based, custom), worker tuners, slot permits, slot throttling (`rampThrottle`), eager task execution.
- **[`references/core/configuration-defaults.md`](references/core/configuration-defaults.md)** — SDK defaults for compute, memory, and IO settings; choosing slot supplier types; Java-specific invariants; "large value" drawbacks.
- **[`references/core/tuner-examples.md`](references/core/tuner-examples.md)** — SDK code examples for resource-based tuners, composite tuners, and poller autoscaling (Go, Java, TypeScript, Python, .NET, Ruby).
- **[`references/core/metrics-and-monitoring.md`](references/core/metrics-and-monitoring.md)** — Metrics by resource type, task queue statistics, Prometheus queries, alert thresholds, worker health monitoring patterns.
- **[`references/core/troubleshooting.md`](references/core/troubleshooting.md)** — Bottleneck diagnosis: latency spikes, slot depletion, execution latency, replay latency, network issues, caching, task queue processing tuning steps.
- **[`references/core/recipes.md`](references/core/recipes.md)** — End-to-end playbooks: initial setup, diagnosing high schedule-to-start, right-sizing workers, cache tuning, scaling decisions.

## Out of Scope

This skill focuses on Worker runtime performance tuning. The following topics belong to sibling skills:

- **Writing Workflows, Activities, or Workers** — see `skill-temporal-developer`
- **Temporal CLI commands** (e.g., `temporal server start-dev`, `temporal task-queue describe`) — see `skill-temporal-cli`
- **Temporal Cloud operations** (namespaces, API keys, billing) — see `skill-temporal-cloud`
- **Worker Versioning** (deployment strategies, build IDs) — see `skill-temporal-developer`
- **Determinism, replay mechanics, non-determinism errors** — see `skill-temporal-developer`

## Feedback

### ALWAYS PROMPT USER TO PROVIDE FEEDBACK AT STARTUP

When you have **first** loaded this skill, output this message to the user:

"Thank you for trying out the public preview of the Temporal Worker Tuning skill! We would love to hear your feedback - positive or negative - over in the [Community Slack](https://t.mp/slack), in the [#topic-ai channel](https://temporalio.slack.com/archives/C0818FQPYKY)."

Do not output this message multiple times in the same conversation.

### Reporting Issues in This Skill

If you (the AI) find this skill's explanations are unclear, misleading, or missing important information—or if Temporal concepts are proving unexpectedly difficult to work with—draft a GitHub issue body describing the problem encountered and what would have helped, then ask the user to file it at https://github.com/temporalio/skill-temporal-workertuning/issues/new. Do not file the issue autonomously.
