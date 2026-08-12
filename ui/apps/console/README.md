# M8 Console

Next.js application shell for M8 Platform. Business UI modules are independent
pnpm workspace packages and are composed by the Console at startup.

## Development

Run commands from this directory:

```bash
pnpm install
pnpm dev
```

Full verification:

```bash
pnpm check
```

## Module architecture

```text
              @m8/core
             ▲        ▲
            /          \
     @m8/query      @m8/operation
            \          /
             \        /
              @m8/runtime
                   │
                   ▼
                Console
```

`@m8/resource-manager` depends on `@m8/core`, `@m8/query` and
`@m8/operation`, never on `@m8/runtime`. Its route `page` values are native
json-render `Spec` trees. `src/platform/modules/registry.ts` is the single
module registration source. `src/platform/specs/app.ts` validates platform and
module routes together before building the native `NextAppSpec`; route
collisions cannot be resolved by declaration order.

Gravity UI providers and theming live here, outside generic `@m8/runtime`.
Application navigation consumes `runtime.navigation`; the home `RouteTree` is
only a developer route directory. Production matching, metadata, static params
and loaders use `createNextApp` from `@json-render/next/server`.

To add an already scaffolded business package, export one `ModuleDefinition`
from its package root and add it to `defineModules`:

```ts
export const catalogModule = defineModule({
  id: 'catalog',
  title: 'Catalog',
  routes: {'/catalog': catalogRoute},
  queries: [productsQuery],
  operations: [updateProductOperation],
});

export const moduleRegistry = defineModules([
  resourceManagerModule,
  catalogModule,
]);
```

No additional contribution registration is required. The array can be
reordered without changing runtime semantics.
