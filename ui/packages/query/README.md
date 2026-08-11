# @m8/query

Typed query definitions, Zod input/output boundaries and a pure registry.
TanStack Query integration is isolated
under `@m8/query/react`; TanStack remains the owner of remote server state.
Query execution receives the current `RuntimeContext` and `AbortSignal`.
Binding normalization removes null/undefined optional filters while preserving
fields whose Zod schema explicitly accepts null. Cache keys are centrally
prefixed with query ID and tenant/resource/actor scope through
`runtimeScopeKey`.

## Public API

- `defineQuery`, `QueryDefinition`
- `QueryRegistry`, `QueryRuntime`, `runtimeScopeKey`
- `QueryProvider`, `useRegisteredQuery`, `QueryClient` from `@m8/query/react`
