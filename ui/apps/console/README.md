# M8 Console

Next.js application shell for M8 Platform. Business UI modules are independent pnpm workspace packages and are composed by the Console at startup.

## Development

Run commands from this directory:

```bash
pnpm install
pnpm dev
```

Verification:

```bash
pnpm check
```

`check` type-checks every module package, type-checks and lints the Console, and runs the production build.

## Module architecture

```text
apps/console
  depends on @m8/module-sdk
  depends on @m8/resource-manager

packages/json-render-module-sdk
  owns shared json-render module, query and operation contracts

packages/resource-manager
  owns Resource Manager routes, queries, operations and API adapter
```

Console declares module packages in `package.json` using `workspace:*`. `src/platform/specs/app.ts` is the composition root: it imports package entry points, passes their module definitions to `defineModules`, and merges the generated routes into the application spec.

To add a module:

1. Create `ui/packages/<module-name>/package.json` with a unique package name and public `exports` entry.
2. Depend on `@m8/module-sdk`; never import from `apps/console`.
3. Export one `ModuleDefinition` from the package root.
4. Add the package to Console dependencies using `workspace:*`.
5. Register the exported module in `src/platform/specs/app.ts`.
6. Run `pnpm install` and `pnpm check`.
