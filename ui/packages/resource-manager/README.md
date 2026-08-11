# M8 Resource Manager UI Module

## Responsibility

Owns the Resource Manager Console contribution: organizations and projects routes, query definitions, operations and its API adapter.

## Owns

- `/resource-manager` routes
- Organization and project query definitions
- Create and delete project operations
- Resource Manager API client boundary

## Does Not Own

- Console application composition
- Shared json-render module contracts and runtime orchestration
- Component renderer implementations

## Integration

The module uses `defineModule` from `@m8/core`, query definitions from
`@m8/query`, and operation definitions from `@m8/operation`. Install the
workspace package in Console and pass `resourceManagerModule` to the module
registry. It deliberately has no dependency on `@m8/runtime`.

`resourceManagerModule` is the only registration source for its routes,
queries, operations, availability and navigation metadata. Console must not
repeat those contribution lists.
