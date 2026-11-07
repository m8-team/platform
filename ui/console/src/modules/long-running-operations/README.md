# M8 Long-running Operations — Console demo

This feature is a **frontend-only prototype** of the M8 control-plane operations registry.

Routes:

- `/platform-operations/long-running` — operations inventory with search, status/service/scope filters, summary metrics and links.
- `/platform-operations/long-running/:operationId` — detailed operation status, execution steps, logs, error/response, workflow IDs and quota reservation correlation.

The sidebar's existing **Platform Operations → Long-running Operations** item now points to the inventory. Quota Manager reservation details link to matching operation details.

## Demo scenarios

- Provision VM, deploy AI routes, create/suspend workspace, update quota policy, delete VM, scale cluster, reconcile ClickHouse database, rotate gateway certificate and create bucket.
- Advance queued/running/waiting operations: update demo progress, steps and logs; operations can eventually succeed.
- Request cancellation: transition to `CancelRequested`, then to `Cancelled` only after another simulated workflow step. Cancellation is not instantaneous.
- Retry a failed retryable operation: create a **new** operation ID with `parentOperationId`; the original failed operation is immutable.
- Reset demo data from the inventory list.
- Filter logs by severity, inspect structured operation JSON and open related Quota Manager reservations.

All mutations stay **in browser memory** until the page reloads. There is **no real backend, auth, RBAC, Temporal workflow execution, YDB, Kubernetes, or provider integration**. In production, use M8 Resource Manager's authoritative LRO API, permission checks, durable reconciliation, idempotent retries and audited cancellation.

## Local validation

```bash
cd ui/console
pnpm install --frozen-lockfile
pnpm lint
pnpm build
pnpm dev
```

Open `http://localhost:5173/platform-operations/long-running`.

## Source

- `demo.ts`: typed operations and deterministic local lifecycle simulation.
- `LongRunningOperations.tsx`: list, detail, progress, timeline, errors and actions.
- `long-running-operations.css`: responsive layout using existing Gravity UI tokens.
- `src/App.tsx`, `src/router.tsx`: navigation and routing.
