---
title: "CI checks"
description: "Governance CI checks."
keywords:
  - "M8 Platform"
  - "governance CI"
---

# CI checks {#ci-checks}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Governance CI](index.md) | [Fitness functions](fitness-functions.md) | `ci-checks.yaml`

{% endnote %}

| ID | Name | Gate | Mechanism |
| --- | --- | --- | --- |
| FF-001 | No cross-context database access | block | go/import + configuration scan |
| FF-002 | Domain dependency direction | block | go list / depguard |
| FF-003 | Protobuf lint and breaking | block | buf lint + buf breaking |
| FF-004 | Event schema compatibility | block | schema registry check |
| FF-005 | Requirement ID validity | block | artifact validator |
| FF-006 | Traceability completeness | block for implementing | artifact validator |
| FF-007 | Outbox atomicity | block | integration test |
| FF-008 | Consumer idempotency | block | integration/fault test |
| FF-009 | Mutation idempotency | block | concurrency test |
| FF-010 | Secret leakage | block | gitleaks + fixture scan |
| FF-011 | Permission annotation | block | proto/custom lint |
| FF-012 | Risk gate annotation | block | catalog lint |
| FF-013 | Audit coverage | block | test/static annotation |
| FF-014 | Telemetry identifiers | warn/block critical | integration test |
| FF-015 | LRO contract | block | proto lint |
| FF-016 | YDB migration ownership | block | migration path lint |
| FF-017 | ADR required for boundary change | block | diff policy |
| FF-018 | PII classification | block | data registry lint |
| FF-019 | SLO metadata | warn/block critical | service catalog lint |
| FF-020 | Prompt scope | block | SPDD schema lint |
