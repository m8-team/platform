# ADR: Resource Manager resource-oriented application slices

## Status

Accepted

## Context

Organizations and workspaces share hierarchy invariants but previously placed
all application behavior in one usecase package. Module composition implicitly
selected in-memory storage through Fx. HTTP Serve errors were discarded, and
List cloned every matching aggregate and formatted UUIDs inside sort comparisons.

## Decision

Keep Resource Manager as one bounded context with resource-oriented application
packages and one file per use case. Commands and queries live alongside their
use cases; explicit module ports cover infrastructure and hierarchy coordination.
Keep rich domain aggregates independent of transport and observability. Use a
plain module constructor with dependencies supplied by the host; Fx remains a
process-level composition mechanism, not a business module dependency.

Instrument application boundaries once for all transports. Inject the observer
and tracer provider; expose local metrics on the management listener. Supervise
server tasks and join them during bounded server shutdown. No background task
may silently discard a listener error.

Retain synchronous completed-operation responses for the existing local adapter.
Do not pretend to implement durable workflows, idempotency or audit without an
Operations backend, persistent store and transactional outbox. Those remain
explicit production composition requirements.

Create contracts use input-only messages. Domain and proto enforce bounded
labels. Undelete supports client version preconditions. Compatibility constraints
do not prevent these contract changes. All Go modules target Go 1.27 and select
Go 1.27.1; devcontainer toolchain selection agrees with the repository.

## Consequences

The same module can be mounted in a monolith or a dedicated API without changing
domain behavior. Constructor dependencies are reviewable, use cases evolve by
resource, and hierarchy tests verify cross-resource consistency. The local
adapter avoids whole-result cloning, uses allocation-free UUID comparisons and
allows independent organizations to mutate concurrently.

The service host still uses Fx. Production persistence, transactional outbox,
idempotency, authorization integration and rate limiting are separate work.
Trace export requires an OTLP endpoint. Metrics are JSON counters, not a full
monitoring backend. Clients must regenerate create request types and restart old
ID-filter pagination sessions. Stored labels must satisfy the new bounds.
