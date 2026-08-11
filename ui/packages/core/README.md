# @m8/core

Module definitions, route contributions, path helpers and module graph
validation.

Page trees remain native `json-render` `Spec` values through
`@json-render/next`'s `NextRouteSpec`. This package does not define a second UI
DSL and has no React, Gravity UI or TanStack Query dependency.

## Public API

- `defineModule`, `defineModules`, `ModuleRegistry`
- `ModuleDefinition`, `ModuleRouteSpec`, `QueryBinding`
- `RuntimeContext`, route access and navigation contracts
- `normalizePath` and dynamic route canonicalization

Expression, action, state and component validation semantics are imported from
json-render. `ModuleRegistry` does not index queries or operations and does not
inspect UI action trees.
