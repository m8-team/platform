---
title: "План реализации"
---

# План реализации

[Executable Baseline](../index.md) | [Roadmap Baseline](index.md)

## Этап 0 — Baseline activation

Импортировать комплект в репозиторий, назначить CODEOWNERS, включить CI,
опубликовать Protobuf и event schemas, создать dev environment.

## Этап 1 — Platform foundation

Common Operation, request context, error model, idempotency, Outbox/Inbox,
Audit client, Access client, OpenTelemetry и YDB migration runner.

## Этап 2 — Первый вертикальный срез

Реализовать `AUTH-FR-017` end-to-end: API → Risk/Identity → aggregate →
Operation/Outbox → Kafka → Audit → acceptance evidence.

## Этап 3 — MVP-1

Resource hierarchy, Identity basics, Access model, CIBA/step-up, Audit search,
UI Operations и Access Explorer.

## Этап 4 — MVP-2

Lifecycle, merge/anonymization, access reviews, advanced providers, audit export.

## Этап 5 — MVP-3

Risk policies, manual review, Provisioning, drivers, reconciliation и drift.

## Этап 6 — GA

Multi-region readiness, SDK, load/chaos/DR evidence, SLO and operational acceptance.
