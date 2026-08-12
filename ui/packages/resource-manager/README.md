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
workspace package in Console and add `resourceManagerModule` to
`defineModules([...])`. It deliberately does not import `@m8/runtime`.

`resourceManagerModule` is the only registration source for its routes,
queries, operations and navigation metadata. Console must not
repeat those contribution lists.

Query results contain resource data only. Detail links are constructed by the
Console table renderer through route configuration, and delete-operation input
is assembled from json-render state bindings at the UI boundary.
