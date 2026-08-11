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
json-render `Spec` trees. `src/platform/specs/app.ts` is the composition root
that registers modules and merges their native `NextRouteSpec` routes into the
platform `NextAppSpec`.

To add a module:

1. Create `ui/packages/<module-name>/package.json` with explicit exports.
2. Depend on `@m8/core` and optionally `@m8/query` / `@m8/operation`.
3. Export one `M8ModuleDefinition` from the package root.
4. Add the package to Console with `workspace:*`.
5. Register it in `src/platform/specs/app.ts`.
6. Run `pnpm install` and `pnpm check`.
