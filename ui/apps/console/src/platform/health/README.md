# Service Health console panel

The footer's **Service Health** panel uses the existing platform health adapter. It fetches
`/livez`, `/readyz`, `/startupz` and `/healthz` on opening and every 15 seconds while
open. The summary badge uses readiness. Backend HTTP 503 is expected for valid unhealthy
snapshots and is not treated as a network error.

In development, Vite proxies `/health/resource-manager/*` to Resource Manager's
separate health listener at `http://127.0.0.1:8081` (default
`M8_HEALTH_HTTP_ADDR=:8081`). Set `M8_RESOURCE_MANAGER_HEALTH_PROXY_TARGET` when
the backend's health listener runs elsewhere. The frontend can alternatively use
`VITE_RESOURCE_MANAGER_HEALTH_URL` for a different same-origin reverse-proxy path.

When deploying the console outside Vite, configure the same reverse-proxy route,
with appropriate access control for internal diagnostics. Do not expose detailed
health results publicly. To add another backend, add its health proxy route and
a corresponding entry to `monitoredServices` in `serviceHealth.ts`.
The standalone development runner `cmd/resource-manager-v1-local` does not serve
these probes; use the full `cmd/resource-manager` binary for this panel.
