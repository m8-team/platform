# M8 Platform UI

The canonical web console is the Vite application in `apps/console`.

## Stack

- Vite
- React
- TypeScript
- Gravity UI UIKit
- Gravity UI Icons
- Gravity UI Table
- TanStack Router
- TanStack Query

## Layout

```text
apps/console/                 main M8 web console
packages/module-sdk/         contracts for independently developed UI modules
```

Resource Manager pages keep transport state in their module. Cursor pagination,
filtering, and ordering are sent to the service through TanStack Query; reusable
tables must explicitly select either client-side or server-side data processing.

## Development

```bash
cd apps/console
pnpm install
pnpm dev
```

## Checks

```bash
pnpm lint
pnpm build
```

The service request console is enabled automatically during Vite development.
For a diagnostic production build, opt in explicitly with
`VITE_ENABLE_REQUEST_CONSOLE=true`; request tokens and sensitive body fields are
redacted. Response previews are limited to JSON payloads up to 32 KiB, and
unknown or sensitive header values are omitted.
