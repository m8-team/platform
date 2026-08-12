# @m8/runtime

Thin React/Next/json-render integration. It composes an already validated module
registry with TanStack Query and safe operation execution. Business modules do
not import this package.

Gravity UI theming, the application shell and the application component
registry belong to Console. Runtime adds only the private `__RouteRuntime`
renderer needed to mount route query and access adapters.

`createRuntime` receives query and operation definitions directly from
`ModuleRegistry`; there is no runtime duck typing or module startup sequence.
`buildNextAppSpec` is the collision-safe composition boundary from platform
routes and M8 route contributions to a native `NextAppSpec`. It validates all
routes before constructing a deterministically keyed result.

TanStack Query is the server-state source of truth. Route query results are
projected one-way under `/__runtime/queryResults/{binding}` for `$state` reads.
Composition rejects UI bindings/actions and loader data that attempt to write
the reserved runtime namespace.

`@json-render/next@0.19.0` permits one named loader per route.
`createRuntimeNextLoaders(customLoaders)` wraps each custom loader and merges its
business data with M8 route params in the same initial state. Consumers with no
custom loaders can use `runtimeNextLoaders` directly.

## Public API

- `createRuntime`, `Runtime`, `RuntimeProvider`, `useRuntime`
- `buildNextAppSpec`
- `createRuntimeNextLoaders`, `runtimeNextLoaders`
- `createRuntimeState`, `withRuntimeState`
- `createActionHandlers` (the operation bridge)
