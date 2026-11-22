# Frontend & UX Standard

**Status:** PROPOSED · **Scope:** `ui/console/**`. Current stack includes React, TypeScript, Gravity UI, TanStack Router, TanStack Query, Vite and pnpm.

## A. Architecture and implementation

### FE-001 — Current stack first (MUST)

Use React + TypeScript and existing Gravity UI components/tokens, routing and fetching infrastructure. Do not add another component library, state manager, router, charting stack or runtime dependency without clear justification and approval. Prefer typed, functional components and feature-local composition.

### FE-002 — Feature ownership (MUST)

Feature code should live under `src/modules/<feature>/`; cross-feature platform services under `src/platform/`; shared presentational components under `src/components/` only if truly reusable. Avoid new monolithic page files and cross-feature imports into private implementation details. Keep API contracts, view models and rendering concerns distinct.

### FE-003 — Type safety and source of truth (MUST)

Do not use `any` or unsafe casts to silence errors. Use generated/typed API clients where available, and map protocol responses to view models explicitly when necessary. Server data belongs in server-state management (`TanStack Query`); local UI state is kept local. Avoid duplicating server state in global stores. Validate untrusted form input and server responses at appropriate boundaries.

### FE-004 — Tenant and resource context (MUST)

Always distinguish tenant, organization, workspace, project and resource identity. Scope fetch/query keys by the trusted selected tenant/scope and relevant authorization context. On context change, invalidate/remove stale scoped data and prevent a previous tenant's response from rendering in the new context. A client-side scope selector does not grant authorization.

### FE-005 — API integration and errors (MUST)

Do not build new API behavior on mock objects while presenting it as production-backed. Use central clients/request wrappers for auth, trace IDs, error normalization and retry policy. Retry only safe/idempotent requests with bounded backoff. Mutation UI should support optimistic updates only with safe rollback and revision/conflict handling. Handle `401`, `403`, `404`, `409`/conflict, network errors and expired sessions distinctly where appropriate.

### FE-006 — Forms and editing (MUST)

Forms MUST match the API's writable fields, immutability rules and validation messages. Use explicit field masks/partial updates where the contract requires; never send output-only data blindly. Make destructive actions confirmable, accessible and responsive to idempotent retries. Surface expected revision conflicts as recoverable UI states rather than silent overwrites.

## B. UX and accessibility

### UX-001 — Every screen state (MUST)

Provide meaningful loading, initial empty, no-results, error, forbidden and success states. State must reflect pending asynchronous operations. Do not show fake success before a mutation is confirmed. Preserve unsaved form changes on recoverable errors.

### UX-002 — Accessibility and keyboard behavior (MUST)

Use semantic controls, labels, focus order, appropriate ARIA only when needed, visible focus, error association, keyboard-operable actions and sufficient contrast. Reuse Gravity UI accessibility conventions. Every icon-only button requires an accessible name. Do not rely on color alone to represent state.

### UX-003 — Internationalization (MUST)

New user-facing text should use the existing i18n system and translation keys; avoid scattered untranslated literals. Ensure dates, numbers and currencies use locale-aware formatting; never use binary floating-point operations for displaying finance-derived values unless conversion is explicitly safe and presentation-only.

### UX-004 — Responsive and consistent design (SHOULD)

Use Gravity UI spacing, typography, layout and feedback patterns, with predictable navigation/breadcrumbs. Verify critical tables/forms at common desktop and narrower viewports. Keep very large tables paginated or virtualized, and show explicit loading/refresh state.

### UX-005 — Performance and resilience (SHOULD)

Avoid expensive computation in render, unnecessary re-fetching or mounting high-cost charts offscreen. Debounce user input appropriately, cache safely per tenant, and cancel obsolete requests where supported. Lazy-load large modules only when it improves navigation without hiding errors.

## C. Demo/fixture discipline

### FE-007 — Demo data separation (MUST)

Current console includes some static illustrative data. New/changed pages MUST clearly distinguish fixture/demo mode from live API data. Demo fixtures belong in explicit fixture files or a documented dev adapter, never mixed with real backend responses in production behavior. Never create users, resources or permission states that appear genuinely live without a source.

### FE-008 — Authorization representation (MUST)

The UI MAY hide/disable unavailable actions and show explanatory forbidden states, but service-level authorization remains authoritative. Never display unauthorized tenant data because it was present in a client cache; do not leak data through breadcrumbs, counts, dropdown options or client logs.

## D. Verification

Run `pnpm lint` and `pnpm build` in `ui/console`. For behavior changes, add component/integration tests using the project's test framework when available; if no framework exists, document the gap and at least provide a manual scenario test. Validate loading/error/forbidden, tenant switch, keyboard access and appropriate responsive behavior. Do not claim that current package scripts include a test runner unless actually added.

**UI review prompts:** Is the scope explicit? Are server and local state separated? Does a revoked permission or tenant switch clear displayed data? Does an API error preserve unsaved input? Are mocks clearly labeled? Is every major screen state accessible?
