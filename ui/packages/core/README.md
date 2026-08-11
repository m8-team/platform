# @m8/core

Framework-free contracts, module definitions, path helpers and registry
validation for M8 declarative UI.

Page trees remain native `json-render` `Spec` values through
`@json-render/next`'s `NextRouteSpec`. This package does not define a second UI
DSL and has no React, Gravity UI or TanStack Query dependency.

## Public API

- `defineModule`, `defineModules`, `ModuleRegistry`
- `ModuleDefinition`, `RouteSpec`, access and navigation contracts
- M8 infrastructure binding expressions (`$state`, `$param`, `$context`)
- `normalizePath`, `joinRoute`
