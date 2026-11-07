# M8 Observability Console — demo UI

M8 Console now implements all five existing **Observability** sidebar entries:

| Route | Functionality |
|---|---|
| `/observability/metrics` | Service-level RPS, error rate, latency p95 and CPU trends; time-range and service selectors; cross-service health |
| `/observability/logs` | Structured log explorer, service/level/text/correlation filters, deep-linked log details |
| `/observability/logs/:logId` | Structured log attributes, Trace ID and Long-running Operation links |
| `/observability/traces` | Trace inventory, service/status/text filters, slow trace filter and service dependencies |
| `/observability/traces/:traceId` | Span waterfall, timing/parent relations, attributes, related logs and LRO links |
| `/observability/alerts` | Alert incidents, statuses, acknowledgement/resolution/muting, alert rules and create/toggle rule |
| `/observability/alerts/:incidentId` | Incident details, condition and timeline, trace/LRO correlations and demo actions |
| `/observability/slo` | SLO/SLI cards, availability/latency objectives, 7d/30d windows, remaining error budgets |

## Data and security boundaries

The demo store includes six M8 services (API Gateway, Resource Manager, Quota Manager, Identity, AI Gateway, Temporal Worker), 24-point metric sequences, structured logs, distributed trace span trees, alert rules and incidents, and SLOs. All values are **synthetic**.

- The interface is **in-memory only**; state resets on reload.
- No PromQL or LogQL/TraceQL query is executed, and no calls are made to Mimir, Loki, Tempo, Prometheus, OpenTelemetry, Alertmanager or the backend.
- No credentials, private logs or secrets are accessed.
- New alert rules are metadata only; toggling a rule, acknowledging/muting/resolving an incident is never propagated to real providers.
- SLOs use illustrative good/total event counts. The UI calculates an error budget and marks breaches; synthetic trends are not evidence of actual reliability.
- Logs and traces have cross-links to demo Long-running Operations where IDs align.

## Development

```bash
cd ui/console
pnpm install --no-frozen-lockfile
pnpm lint
node --experimental-strip-types --test src/modules/observability/demo.test.mjs
pnpm build
pnpm dev
```

Open `http://localhost:5173/observability/metrics`.

## Implementation

- `demo.ts`: synthetic metrics/logs/traces, alert state transitions and error-budget functions.
- `ObservabilityShell.tsx`: shared navigation, visualization and UI primitives.
- `ObservabilityMetricsLogs.tsx`: Metrics dashboard and Log Explorer.
- `ObservabilityTraces.tsx`: trace list, span-waterfall details, related logs.
- `ObservabilityAlertsSlo.tsx`: Alert incident/runbook-style details, rules and SLOs.
- `observability.css`: responsive layout using Gravity UI tokens.
- `demo.test.mjs`: metric shape, search, trace integrity, incident lifecycle and SLO budget tests.

## Production integration plan

1. Authoritative M8 Resource Manager scope/ownership model for service and resource filtering; use M8 Access permissions on every query and mutation.
2. Tenant-aware read APIs proxying PromQL to **Mimir/Prometheus**, LogQL to **Loki**, and trace queries to **Tempo**; strict query limits, retention and label cardinality rules.
3. OpenTelemetry ingestion with consistent `tenant_id`, `org_id`, `project_id`, `service_id`, `resource_id`, `operation_id`, `trace_id`; do not trust user-controlled labels to authorize access.
4. Verified alert evaluation using **Mimir Ruler/Prometheus + Alertmanager**, persisted acknowledgment/escalation state and audit events.
5. Approved SLI definitions and rolling-window burn rates with reliable metric semantics, sampled/unsampled caveats, freshness and incident timelines.
6. Secure redaction, retention, query audit and privacy controls for log payloads.

No production observability claims, external network calls, provider actions or alert delivery are implied by this PR.
