# @m8/runtime

Thin React/Next/json-render integration. It composes the module registry,
TanStack Query and safe operation execution with native json-render providers.
Business modules must not depend on this package.

Gravity UI theming, the application shell and the application component
registry belong to Console. Runtime adds only the private `__RouteRuntime`
integration renderer required to mount route query and access adapters.

`createRuntime` consumes typed query and operation contributions directly from
`ModuleRegistry`; there is no runtime duck typing. `buildNextAppSpec` is the
composition boundary from M8 route contributions to native `NextAppSpec`.
Production route matching remains in `@json-render/next`.

TanStack Query is the server-state source of truth. Route query results are
projected one-way under `/__runtime/queryResults/{binding}` for `$state` reads.
Composition rejects UI bindings/actions that attempt to write the reserved
runtime namespace.

## Public API

- `createRuntime`, `Runtime`, `RuntimeProvider`, `useRuntime`
- `buildNextAppSpec`
- `runtimeNextLoaders` for canonical json-render route-param projection
- `createRuntimeState`, `withRuntimeState`
- `createActionHandlers` (only the operation bridge)
