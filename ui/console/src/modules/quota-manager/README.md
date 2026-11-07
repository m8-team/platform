# M8 Quota Manager — Console demo

The Quota Manager is mounted in the existing M8 Console shell (Gravity UI 7, TanStack Router) at:

- `/quota-manager/overview`
- `/quota-manager/policies`
- `/quota-manager/reservations`
- `/quota-manager/requests`
- `/quota-manager/reconciliation`

The Platform Operations sidebar links to all five pages. The selector switches between the example Acme organization and its AI Studio project. Labels adapt to the Console's English/Russian language setting.

## Demo interactions

1. **Overview:** filter by resource category, inspect used/reserved/available and open quota details.
2. **Quota Policies:** search/filter policies, edit a numeric limit, switch between Dedicated and Shared, preview the delta and validate parent-budget constraints. Changes are visible on Overview.
3. **Reservations:** filter by status, inspect operation timelines, simulate a retry on a `Reconciling` reservation. An uncertain reservation is never force-released.
4. **Quota Requests:** submit a request with justification, then approve/reject it. An approved request can be applied to update the demo policy after another quota validation.
5. **Reconciliation:** filter drift findings, compare desired and observed values, run a simulated scan or mark a finding reviewed. No provider resource changes are made.

All demo mutations are **in-memory only** and reset on reload. There are **no calls to YDB, Kubernetes, Temporal, Quota Manager API, or provider API**, and no real RBAC is enforced. Before production use, implement server-authoritative effective quotas, idempotent reservations, admission, approvals, permissions, reconciliation and audited repairs.

## Local checks

From `ui/console`:

```bash
pnpm install --frozen-lockfile
pnpm lint
pnpm build
pnpm dev
```

Then open `http://localhost:5173/quota-manager/overview`.

## Implementation

- `demo.ts`: isolated in-memory demo store and validation rules.
- `QuotaManager.tsx`: five route components, details drawers, forms and tables.
- `quota-manager.css`: responsive styles using Gravity UI variables.
- `src/router.tsx`, `src/App.tsx`, `src/i18n.ts`: route and sidebar integration.

The Console currently defaults to a fake organization/project context; Quota Manager's scope selector is explicitly a **separate demo data selector** rather than a binding to live Resource Manager IDs.
