# M8 Audit Console — demo UI

The existing **Audit** navigation is now functional:

- `/audit/events` — event journal, search, category/outcome/severity/scope/actor type/date UTC filters, pagination and linked correlations.
- `/audit/events/:eventId` — immutable event detail, actor identity, request/trace/correlation, before/after JSON with redaction and links to Long-running Operations / Gateway / Quota Manager.
- `/audit/exports` — scoped export request with CSV or JSONL, snapshot of selected event IDs, explicit Queued → Ready transition, local browser download and export history.

Example event sources cover Resource Manager, Quota Manager, Gateway, Identity, Security and system reconciliation. All event information is synthetic, including actor names, example IP addresses and timestamps.

## Demo semantics

- **Append-only events:** UI does not edit or delete event records. Creating and preparing an export appends a corresponding event.
- **No integrity attestation:** No cryptographic hash, digital signature or immutable storage guarantee is implemented. The UI explicitly states that integrity is NOT verified.
- **Safe redaction:** Recursive redaction masks keys resembling password, token, credential, private key, API key and secret in the event JSON and export. CSV also protects spreadsheet clients from leading formula-like values.
- **Snapshot exports:** The IDs of events selected when the request was created are recorded and remain unchanged, even when later export audit events appear.
- **Manual staging:** Export requests are `Queued`; a separate Prepare action transitions them to `Ready`. No background tasks are simulated as actual work.
- **Browser-only output:** Download creates a local Blob; there is no upload, network call, storage or export approval.
- **Session memory:** All changes reset on browser reload.

## Running

```bash
cd ui/console
pnpm install --no-frozen-lockfile
pnpm lint
node --experimental-strip-types --test src/modules/audit/demo.test.mjs
pnpm build
pnpm dev
```

Open `http://localhost:5173/audit/events`.

## Architecture for production integration

Before production use, add:
- Server-authoritative append-only ingest, actor/service identity, action schema registry, tenant scope and correlation identifiers.
- M8 Access permissions for event read and export, plus sensitive-field policies and audit of download actions.
- YDB-backed hot-indexed audit event queries and an immutable/retained archive with signed export manifests and verifiable hashes.
- Durable export workflow via Temporal (request, authorization/approval, snapshot, export, signed manifest, TTL cleanup) and object storage with time-bound download URLs.
- Stream reconciliation, gap detection, sequence numbers and data retention/erasure controls.

The demo is deliberately **not** evidence-grade or compliance-certified. Read-only event search and local files are implementation examples, not guarantees of audit completeness.
