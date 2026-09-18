# Temporal Worker Tuning Skill

A skill for tuning [Temporal](https://temporal.io/) Worker performance — configuring slot suppliers, worker tuners, poller autoscaling, cache settings, and diagnosing bottlenecks using SDK metrics.

> [!WARNING]
> This Skill is currently in Public Preview, and will continue to evolve and improve.
> We would love to hear your feedback - positive or negative - over in the [Community Slack](https://t.mp/slack), in the [#topic-ai channel](https://temporalio.slack.com/archives/C0818FQPYKY).

## Installation

### As a Plugin

This skill is packaged as a plugin for major coding agents:

- **Claude Code**: [temporalio/claude-temporal-plugin](https://github.com/temporalio/claude-temporal-plugin)
- **Cursor**: [temporalio/cursor-temporal-plugin](https://github.com/temporalio/cursor-temporal-plugin)
- **OpenAI Codex**: [temporalio/codex-temporal-plugin](https://github.com/temporalio/codex-temporal-plugin)

See each repo's README for installation instructions.

### Standalone Installation

```bash
mkdir -p ~/.claude/skills && git clone https://github.com/temporalio/skill-temporal-workertuning ~/.claude/skills/temporal-workertuning
```

Adjust the installation directory based on your coding agent.

## What this skill covers

- **Tuning concepts** — task slots, slot suppliers (fixed-size, resource-based, custom), worker tuners, composite tuners, slot throttling, eager task execution
- **SDK configuration defaults** — compute, memory, and IO settings across Go, Java, TypeScript, Python, and .NET with per-SDK default values
- **Code examples** — resource-based tuners, composite tuners, and poller autoscaling for all supported SDKs
- **Metrics and monitoring** — SDK metrics by resource type, task queue statistics, Prometheus query examples, alert thresholds, worker health patterns
- **Troubleshooting** — diagnosing schedule-to-start latency spikes, slot depletion, execution/replay latency, network failures, cache evictions
- **Tuning playbooks** — initial production setup checklist, right-sizing workers, cache tuning workflow, scaling decision framework

## Currently supported Temporal SDK languages

- [x] Go
- [x] Java
- [x] TypeScript
- [x] Python
- [x] .NET
- [x] Ruby (poller autoscaling only)

## What this skill does NOT cover

- **Writing Workflows, Activities, or Workers** — use [skill-temporal-developer](https://github.com/temporalio/skill-temporal-developer)
- **Temporal CLI commands** — use [skill-temporal-cli](https://github.com/temporalio/skill-temporal-cli)
- **Worker Versioning** (deployment strategies, build IDs) — use [skill-temporal-developer](https://github.com/temporalio/skill-temporal-developer)
- **Determinism, replay mechanics, non-determinism errors** — use [skill-temporal-developer](https://github.com/temporalio/skill-temporal-developer)
