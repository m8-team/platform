# @m8/core

Framework-free contracts for unordered M8 business modules and contribution
ownership.

## Module model

A module is a plain immutable bundle of routes, queries and operations. It has
no startup hook, lifecycle, priority or registration-order metadata.

```ts
import {defineModule} from '@m8/core';

export const catalogModule = defineModule({
  id: 'catalog',
  title: 'Catalog',

  routes: {
    '/catalog': catalogRoute,
  },

  queries: [productsQuery],
  operations: [updateProductOperation],
});
```

`ModuleDefinition<TQuery, TOperation>` accepts generic contributions constrained
only by `ModuleContribution { id }`. The concrete contracts remain owned by
`@m8/query` and `@m8/operation`. Page trees remain native json-render `Spec`
values through `@json-render/next`'s `NextRouteSpec`.

## Registration

```ts
export const modules = defineModules([
  identityModule,
  resourceManagerModule,
  catalogModule,
]);
```

That list is the only registration source for every contribution in a business
module. Adding a module does not require separate route, query, operation or
navigation calls.

Registration order has no semantic meaning. `ModuleRegistry` uses the following
composition pipeline:

```text
collect all definitions and contributions
  → validate globally
  → build ID-based registries and ownership
```

Modules, routes, queries and operations are exposed in deterministic technical
order by ID or normalized path. Collisions are always errors; neither the first
nor the last declaration wins.

Routes may reference queries from any registered module by global query ID.
Those references are checked only after all contributions have been collected,
so the target module may appear anywhere in the input list.

## Public API

- `defineModule`, `defineModules`, `ModuleRegistry`, `ModuleContribution`
- `ModuleDefinition`, `ModuleRouteSpec`, `QueryBinding`
- route, query and operation ownership lookups
- `RuntimeContext`, route access and navigation contracts
- route normalization, validation and dynamic-route canonicalization

Expression, action, state and component semantics remain owned by json-render.
The registry does not know query keys, schemas, execution functions, operation
modes or HTTP/LRO details.
