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
installed/enabled/registered module source. `src/platform/specs/app.ts` merges
the registered native `NextRouteSpec` routes into the platform `NextAppSpec`.

Gravity UI providers and theming live here, outside generic `@m8/runtime`.
Application navigation consumes `runtime.navigation`; the home `RouteTree` is
only a developer route directory. Production matching, metadata, static params
and loaders use `createNextApp` from `@json-render/next/server`.

To add a module:

1. Create `ui/packages/<module-name>/package.json` with explicit exports.
2. Depend on `@m8/core` and optionally `@m8/query` / `@m8/operation`.
3. Export one `ModuleDefinition` from the package root.
4. Add the package to Console with `workspace:*`.
5. Add it to `installedModules` in `src/platform/modules/registry.ts`.
6. Run `pnpm install` and `pnpm check`.
