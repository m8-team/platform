# @m8/query

Typed query definitions, Zod input/output boundaries, a pure registry and
resolver for M8 infrastructure bindings. TanStack Query integration is isolated
under `@m8/query/react`; TanStack remains the owner of remote server state.
Query execution receives the current `RuntimeContext` and `AbortSignal`.
Binding normalization removes null/undefined optional filters while preserving
fields whose Zod schema explicitly accepts null.
