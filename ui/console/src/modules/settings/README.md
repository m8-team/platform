# M8 Console Settings — interactive demo

The existing **Settings** sidebar group is now implemented as five pages under `ui/console`:

| Route | Functionality |
|---|---|
| `/settings/project` | Project identity, slug, namespace, environment and default region, basic validation and versioned local changes |
| `/settings/modules` | Sample Module Registry / Capability Map, categories, tier, dependencies, capabilities, enable/disable with dependency guards |
| `/settings/integrations` | Connector registry for Kubernetes, Temporal, Keycloak, Mimir, Loki, Tempo and Agent Router; metadata-only creation/edit and simulated test |
| `/settings/webhooks` | Event topic subscriptions, HTTPS endpoints, signing secret references, delivery history and simulated retry |
| `/settings/api-tokens` | Token metadata, owner/scopes/expiry, create **Draft** token request and reversible demo-only revocation |

The module is integrated with the existing Gravity UI shell, TanStack Router, English/Russian navigation and sidebar.

## Safety boundaries

This is a **browser-memory demo only**. No requests to M8 Resource Manager, YDB, Kubernetes, Temporal, provider APIs or external endpoints are sent. Changes reset on browser reload. All sample identity and endpoint values are synthetic.

- **Project:** Modifying identifiers, region or namespace does not perform a live resource migration.
- **Module Registry:** A curated set of 14 illustrative modules, **not** the complete canonical registry. Enabled/disabled switches are validated locally and do not deploy code, modify licenses or run migrations.
- **Integrations:** Secret references are vault paths only; **never enter actual access tokens, passwords or credentials**. The Test action does not establish a network connection or claim the provider is healthy.
- **Webhooks:** New webhooks start Disabled; enabling requires a signing secret reference. Retry only adds a synthetic delivery record. Client-side URL validation blocks common local/private hosts, but **real SSRF prevention must be server-enforced** with DNS/IP validation, redirect controls, egress policy and authorization.
- **API Tokens:** New requests create **Draft metadata only**. No usable API token is ever issued, generated, exposed or stored. Existing Active/Revoked records are synthetic; Revoke only updates the demo state.

## Local development

```bash
cd ui/console
pnpm install --no-frozen-lockfile
pnpm lint
node --experimental-strip-types --test src/modules/settings/demo.test.mjs
pnpm build
pnpm dev
```

Open `http://localhost:5173/settings/project`.

## Source files

- `demo.ts` — Settings store, example data, input validation, dependencies and local transitions.
- `SettingsShell.tsx` — shared cards, tabs, modals and form elements.
- `SettingsProjectModulesIntegrations.tsx` — Project, Modules, Integrations.
- `SettingsWebhooksTokens.tsx` — Webhooks and API Tokens.
- `settings.css` — responsive styling using Gravity UI variables.
- `demo.test.mjs` — project, module dependencies, endpoint safety, webhook retries and token metadata tests.

## Production design

Use Resource Manager as the source of truth for project metadata and a versioned Module Registry / Capability Map for module ownership, capabilities and licensing. Every Settings mutation needs a scoped M8 Access check, immutable Audit event, idempotency key, backend validation and optimistic `version` guard.

Module installation requires a durable Temporal operation, dependency and compatibility validation, rollout and reconciliation. Integrations must use provider adapters and secret references resolved server-side, never browser-supplied credentials. Webhook delivery should use a verified outbox/dispatcher with HMAC signatures, SSRF protection, retry budgets and DLQ. Token issuance requires Identity + Access, restricted scopes and TTL, secure token hashing and one-time display, with revocation in an authoritative credential store.
