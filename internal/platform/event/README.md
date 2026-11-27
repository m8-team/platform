# M8 Event v1

Shared infrastructure, not a bounded context or an event bus. Business modules
retain ownership of their state, event types, typed payload contracts and command
receipts. The common envelope is m8.platform.event.v1.Event. Existing event
contracts and outboxes remain unchanged until their explicit migration.

## API and transaction lifecycle

Use NewPublisher with trusted Config; use NewTransactor only where no existing
native Query SDK transaction owner is available. Pass the same query.TxActor to
business writes and Publisher.Publish. Return publication errors immediately.
Prepare command identity, event IDs and time outside retryable callbacks.
Publish one batch per logical topic per transaction. An empty batch is a no-op
for an authorized topic; callers must enforce whether events are required.

The pinned YDB Go SDK v3.157.1 exposes StartTransactionalWriter, TxWriter.Write
and DoTx. Writers flush on the SDK before-commit callback and close on transaction
completion. The SDK's transactional writer and keyed multiwriter are experimental.
No external calls or asynchronous publication inside the transaction. Only
confirmed DoTx success means committed; do not record successful publication
metrics on Publish return. No independent retries or global writer pools.

The deployed server version, transactional topic support and limits have not
been verified. No legacy fallback is provided. Default application limits:
5 seconds, 256 KiB per event, 100 events and 4 MiB per batch. Provision topics
outside transactions with deployment tooling, initially 4 partitions and 7 days
retention subject to server compatibility, load and consumer downtime.

## Envelope and security

Use a UUID event ID stable across logical retries, versioned event type,
trusted source, aggregate identity and positive revision, occurrence time and
typed protobuf Any payload. Tenant ID is a UUID when present; application
boundaries must require it for tenant-owned facts. Payload types must be linked
into the producer binary so Any.UnmarshalNew resolves them. Both envelope and
decoded payload use Protovalidate. No arbitrary request-driven topic paths.
Use scoped YDB credentials/TLS. No secrets in metadata/payload; producer owns
data minimization. Validation diagnostics do not include payload contents.
No metrics with event/aggregate/tenant IDs. Reuse SDK tracing and service
instrumentation; no OpenTelemetry exporter or metrics framework is introduced.

## Idempotency, ordering and consumers

DoTx uses serializable read-write mode and SDK retries without marking unknown
operations idempotent. A module must read/insert its own durable command receipt
(scope, command ID, request hash, result) in the same transaction. On ambiguous
commit, resolve the receipt before resubmitting. A conflicting request hash is
rejected; choose retention based on retry window. Event IDs survive retries.
Do not retry Write separately.

The routing key hashes length-prefixed tenant/aggregate type/aggregate ID.
SDK BoundPartitionChooser selects partitions, including split-aware routing.
Each attempt uses a new producer prefix; sequence numbers are SDK-assigned,
not aggregate revisions. This is not a logical deduplication mechanism. Producer
churn and partition routing must be benchmarked before production. No strict
cross-transaction or cross-replica aggregate ordering guarantee. Consumers must
deduplicate by event ID, track revisions and fail closed/reconcile on gaps for
security-critical projections. Multiple facts can share a revision.

YDB-only consumers may use PopMessagesBatchTx and the same TxActor to update
state and publish, with offsets committed by YDB. Logical duplicate events may
still require a consumer receipt. Kafka, Temporal and SpiceDB are separate
post-commit consumers, at-least-once across systems. Confirm Kafka ACK before
offset commit, use stable Temporal workflow IDs, and persist SpiceDB checkpoints
and ZedTokens before ACK. Preserve consistency barriers. No bridges added here.

## Migration and verification

This change introduces an independent foundation only. Resource Manager and
Access currently use database/sql and module-local outboxes. Never open a second
transaction for event publication. Migrate each selected command to native
Query SDK, reuse its command receipts, map existing event payload to the common
envelope, provision topic and consumers, run real atomicity/fault/concurrency
tests, drain old outbox with publishers fenced, then switch single ownership.
Do not dual-publish to active consumers, delete pending outbox rows, or enable
the new route before compatibility and cutover checks. Roll back by stopping
writes, reconciling offsets/receipts and restoring a single old publisher only
after accounting for committed native events.

Generate using the repository Buf workflow (just proto); then run buf lint,
buf breaking, go test ./internal/platform/event, go test -race, go vet and
benchmarks. Generated files must not be hand-edited. This draft has not been
built: the execution environment has Go 1.23.2, no Buf and no network access to
GitHub/module proxies. An opt-in live test is provided: M8_EVENT_TEST_YDB_DSN and
M8_EVENT_TEST_YDB_PREFIX must target a disposable database. Run
 go test ./internal/platform/event -run TestIntegrationAtomicPublication -v
for atomic publication, rollback, unavailable topic and transactional consumer
rollback/offset checks. Live execution, production configuration wiring, metrics,
profiling, and module integration are outstanding release gates. Do not merge
or deploy before generated artifacts and verification are complete.

References: https://github.com/ydb-platform/ydb-go-sdk/tree/v3.157.1/examples/topic
and https://ydb.tech/docs/en/concepts/datamodel/topic.
