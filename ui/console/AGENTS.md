# M8 Console — Scoped Agent Instructions

Applies to `ui/console/**`. Also follow the root `AGENTS.md`.

Read `docs/engineering/frontend-ux.md`, `docs/engineering/architecture-security.md`, and `docs/engineering/quality-testing.md` as relevant.

- Use the existing React, TypeScript, Gravity UI, TanStack Router/Query conventions; avoid new UI kits without approval.
- Compose page logic from small typed components and feature-local modules. Keep API transport, server state, and presentation concerns separate.
- Never make UI access control the only enforcement layer. Show or hide actions for usability, but enforce access in backend services.
- Explicitly handle loading/empty/error/forbidden/stale states and tenant/workspace switches; clear tenant-scoped caches appropriately.
- Keep demo/mock data separate from production paths and label it. No hard-coded credentials or customer data.
- Add or update localized text; preserve keyboard navigation, semantics, accessibility, and responsive behavior.
- Verify with `pnpm lint` and `pnpm build` from `ui/console/`; add behavior tests when test infrastructure is available.
