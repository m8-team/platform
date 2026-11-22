# Protobuf Contracts — Scoped Agent Instructions

Applies to `proto/**`. Also follow the root `AGENTS.md`.

Read `docs/engineering/backend-api.md` and `docs/engineering/quality-testing.md` before editing.

- Contract-first; preserve numbered fields, field meanings, JSON/HTTP shapes, resource names, and published event semantics.
- Use package `m8.platform.<module>.vN` and versioned paths following established module conventions.
- Add `google.api.field_behavior` and `buf.validate` constraints as appropriate, but do not substitute schema annotations for server-side authorization, optimistic concurrency, or runtime validation.
- For edits, define update masks, immutable/output-only semantics, revision handling, and idempotency; for lists, define filtering, ordering, pagination and authorization semantics.
- Use `reserved` for permanently removed field numbers/names when appropriate. Do not renumber fields or reuse retired numbers.
- Run `buf lint` and `buf breaking` against the intended baseline; breaking exceptions require explicit authorization and a recorded rollout plan.
- Generated Go clients/servers and OpenAPI are outputs of `buf.gen.yaml`; never patch them manually.
