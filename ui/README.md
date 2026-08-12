# M8 UI

M8 UI is a thin platform/domain layer over `json-render`. It does not define a
second UI framework.

## Ownership

`json-render` owns:

- UI specs, elements, bindings, visibility and watchers
- UI state and UI actions
- component catalogs, component registries and rendering
- Next.js route matching, layouts, metadata, SSR and server loaders

M8 owns:

- module identity, dependencies, enablement and contribution ownership
- typed server queries and TanStack Query integration
- business operations, authorization, audit and long-running-operation hooks
- the thin adapters between those platform concepts and `json-render`

## Package boundaries

```text
@m8/core
   ↑
   ├── @m8/query
   ├── @m8/operation
   └── @m8/runtime → core + query + operation + json-render

business modules → core + query + operation
Console          → runtime + business modules + Gravity UI
```

`@m8/core` contains only module/platform contracts. Query execution belongs to
`@m8/query`; operation execution belongs to `@m8/operation`; Gravity UI belongs
to Console.

## Module registration

The Console registration source is
`apps/console/src/platform/modules/registry.ts`:

```text
installed modules
  → selectEnabledModules(...)
  → dependency validation and topological sorting
  → ModuleRegistry
  → buildNextAppSpec(...)
```

A business package exports its module definition as its public registration
root. Routes, queries and operations are not registered a second time in the
Console.

## Query lifecycle

```text
QueryDefinition
  → ModuleRegistry ownership
  → QueryRegistry
  → TanStack Query (source of truth)
  → one-way UI projection at /__runtime/queryResults/*
  → json-render $state reads
```

The runtime projection exposes stable loading/data/error fields to UI specs. It
does not implement caching, retries, polling, invalidation or freshness, and UI
actions cannot write the reserved `__runtime` namespace.

## Operation lifecycle

```text
json-render action
  → executeOperation adapter
  → OperationRegistry
  → OperationDefinition.execute
  → authorization / audit / LRO / query invalidation adapters
```

UI specs use stable operation IDs and never know backend endpoints.

## Next.js rendering

Production matching, metadata, static params, loader execution and page data
come from `createNextApp` in `@json-render/next/server`. Version `0.19.0`
returns `getPageData`, `generateMetadata` and `generateStaticParams`; the
catch-all page only handles Next.js `notFound()` and renders the returned data.
Named route params are projected by a registered json-render server loader, so
there is no second production route matcher in M8.

## Validation

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm --filter console build
```
