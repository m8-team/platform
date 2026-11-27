# ADR-0002 — Native YDB Transactional Events

- Status: Proposed implementation; architecture directed by the user
- Date: 2026-11-27
- Affected: shared event infrastructure; future module-by-module migration
- Requirements: ARCH-001, ARCH-002, ARCH-003, ARCH-006, DATA-002, DATA-003

## Decision

Use one common Protobuf Event envelope and a small shared Go package
internal/platform/event. YDB is the transaction coordinator. Business modules
continue owning state and typed payload semantics; do not adapt the shared
layer to existing local event envelopes. Use one explicit native serializable
query transaction for state, command receipt and event publication.

## Alternatives

Reject polling outbox, CDC outbox, Watermill and external event bus frameworks
for new native paths: they duplicate atomicity or require additional relay
coordination. A legacy outbox may remain necessary for non-YDB transactional
sources or during a safely gated migration.

## Migration, security and verification

See internal/platform/event/README.md. No legacy producers or consumers change
in this foundation PR. No production writes, deployments or destructive
migrations. Verify code generation, SDK/server compatibility, live atomicity,
ambiguous commit behavior, consumer offsets, concurrency and performance before
enabling native publication. Existing module-local events are temporary legacy
contracts, not the design authority for the new shared layer.
