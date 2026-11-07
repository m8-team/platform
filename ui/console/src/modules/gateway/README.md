# M8 Gateway Console — interactive demo

The module is integrated into the existing M8 Console shell, routing and English/Russian navigation.

## Pages

| Route | Demo functionality |
|---|---|
| `/gateway/services` | API upstream registry, health and p95 indicators, create/edit service, related routes |
| `/gateway/routes` | HTTPRoute and conceptual AIGatewayRoute inventory, create/edit routes, conflict validation and request matching simulator |
| `/gateway/consumers` | Client registry, auth method, create client metadata, suspend/reactivate without credential issuance |
| `/gateway/policies` | Gateway/Service/Route scoped JWT, API Key, mTLS, CORS, WAF metadata, enable/disable and edit |
| `/gateway/rate-limits` | Gateway/Service/Route/Consumer traffic budgets, window usage, burst, monitor/enforce, enable/disable |
| `/gateway/yaml` | Read-only Gateway API HTTPRoute example and explicitly *non-deployable* conceptual M8 descriptors for AI routes and API services |

Gateway sidebar entries are actual links under the existing AsideHeader.

## Data and safety

The demo store (`demo.ts`) contains services, routes, clients, policies, rate limits and sample activity. Mutations are in memory only and reset on reload. No YDB, Temporal, Kubernetes, Envoy Gateway or Agent Router calls are issued. Policy and rate-limit entries are **not** actually enforced. The simulator performs a simplified host/path/method match, returns a synthetic decision and describes applicable *metadata*; it does **not** validate JWT/API keys/mTLS or proxy traffic.

Consumer create does not generate or display secrets or private key material. Provider observed states and traffic metrics are illustrative.

For HTTPRoute, the YAML page provides a Gateway API example but requires review of the Gateway name/namespace, listeners, backend references and policy attachments before deployment. For Agent Router AIGatewayRoute, the correct CRD must come from the deployed provider's **Resource Type Registry**, so this UI deliberately generates a *conceptual*, not deployable, descriptor. Do not `kubectl apply` it.

## Local development

```bash
cd ui/console
pnpm install --no-frozen-lockfile
pnpm lint
node --experimental-strip-types --test src/modules/gateway/demo.test.mjs
pnpm build
pnpm dev
```

Open `http://localhost:5173/gateway/services`.

## Source layout

- `demo.ts` — types, sample objects, in-memory CRUD and validations.
- `GatewayShell.tsx` — navigation, layout, reusable fields and drawers.
- `GatewayInventory.tsx` — API Services, Routes, Consumers and simulator.
- `GatewayGovernance.tsx` — Policies, Rate Limits and YAML viewer.
- `manifest.ts` — sample Gateway API HTTPRoute / conceptual M8 resource descriptors.
- `gateway.css` — responsive module styles using Gravity UI tokens.
- `demo.test.mjs` — validations, route matching, client status and manifest tests.

## Production follow-up

Use actual M8 resource ownership and scoped permissions. Configure provider adapters, read models with desired/observed state, backend admission, outbox and Temporal LRO for changes, K8s reconciliation, status conditions, and policy-specific schemas. Never expose API key material through list/detail APIs.

The Console may have unrelated build blockers arising from version upgrades in existing components; treat those separately from this demo module.
