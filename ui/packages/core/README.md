# @m8/core

Module definitions, enablement, dependency ordering, contribution ownership and
route validation.

Page trees remain native `json-render` `Spec` values through
`@json-render/next`'s `NextRouteSpec`. This package does not define a second UI
DSL and has no React, Gravity UI, TanStack Query, query-execution or
operation-execution dependency.

`ModuleDefinition<TQuery, TOperation>` accepts generic contributions constrained
only by `ModuleContribution { id }`. Concrete contracts remain owned by
`@m8/query` and `@m8/operation`.

`ModuleRegistry` validates required dependencies, includes optional dependency
edges only when present, rejects self-dependencies and cycles, and returns a
deterministic topological order. It also detects route/query/operation
collisions and retains lightweight owner mappings.

## Public API

- `defineModule`, `defineModules`, `ModuleRegistry`, `ModuleContribution`
- `selectEnabledModules`, `ModuleEnablementOptions`
- `ModuleDefinition`, `ModuleRouteSpec`, `QueryBinding`
- `RuntimeContext`, route access and navigation contracts
- `normalizePath` and dynamic route canonicalization

Expression, action, state and component semantics remain owned by json-render.
`ModuleRegistry` indexes contribution IDs and owners only; it does not know
query keys, schemas, execution functions, operation modes or HTTP/LRO details.
