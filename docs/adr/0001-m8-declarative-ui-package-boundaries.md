# ADR-0001: M8 declarative UI package boundaries

## Status

Accepted

## Context

M8 Console previously split module contracts between a `module-sdk` package
and application-local registries. Query and operation definitions shared that
package even though their execution concerns differ. React, Next.js and
Gravity UI integration require framework dependencies that business modules
must not inherit.

## Decision

Use four workspace packages:

- `@m8/core` owns framework-free module contracts, validation, path helpers and
  `NextAppSpec` composition.
- `@m8/query` owns typed query definitions, Zod validation, binding resolution
  and an optional TanStack Query React adapter.
- `@m8/operation` owns registered mutation definitions and the authorization,
  confirmation, audit and invalidation pipeline.
- `@m8/runtime` owns React, Next.js, json-render, TanStack Query and Gravity UI
  composition.

Native json-render `Spec` remains the component-tree language. M8 adds only
infrastructure metadata around native `NextRouteSpec`; that metadata is removed
before routes are passed to `@json-render/next`.

## Consequences

Business modules have small, explicit dependencies and cannot accidentally
couple to React runtime infrastructure. Queries and operations are independently
testable and validated at runtime. The Console composition root must explicitly
provide registries and adapters. Backend authorization remains authoritative;
client-side module filtering is presentation composition only.
