# M8 Event v1

The flat `event` package shares one Protobuf envelope and publishes batches using
native YDB Query transactions. Business modules own business state, typed
payload contracts, authorization, command receipts and consumer state. Existing
local contracts and outboxes stay in service until each module's cutover.

## Construct and publish

Use `New(Attributes, proto.Message)` to validate and pack a typed payload. Set
the event ID, aggregate revision, tenant, timestamp and correlation fields
outside retryable callbacks, or derive them deterministically from the same
command and authoritative state. The constructor copies metadata and packs the
payload into `Any`. It requires an explicitly supplied occurrence time.

The event owner pairs its event type with its payload via a typed constructor.
This keeps the business type stable across payload descriptor renames.

```go
func newResourceCreated(attrs event.Attributes, payload *resourcev1.ResourceCreated) (*eventv1.Event, error) {
    attrs.Type = "m8.resource.created.v1"
    return event.New(attrs, payload)
}
```

The common package checks the envelope and the structural constraints of the
typed payload. `Publish` resolves and validates the current `Any` again
because Protobuf messages are mutable. The payload descriptor must be linked
into the binary. Keep events immutable during publication. The owning module
enforces business semantics and the event type/payload correspondence.

Create `NewPublisher(driver, DefaultConfig(topics))` at startup using a trusted
map of alias to absolute topic path. The Publisher copies the map and is safe
to share across independent transactions. Use one batch per topic per
transaction. The defaults are 256 KiB per event, 100 events per call and 4 MiB
of serialized envelopes per call. An empty batch with a configured alias is a
no-op. Callers enforce required event emission and their total transaction
budget. Topic provisioning and compatibility verification belong to deployment.

```go
err := owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
    if err := writeBusinessState(ctx, tx); err != nil {
        return err
    }
    return publisher.Publish(ctx, tx, "resource", created)
})
```

The existing transaction owner supplies the same native `query.TxActor` to
state writes and event publication. `NewTransactor(driver, timeout)` is an
optional owner where the application does not already have one. It uses
serializable read-write, an explicit caller deadline and SDK retries with
`WithIdempotent(false)`. Return every publication error from the callback.
Atomicity requires both writes to use the same transaction and all publication
errors to reach its owner. An ignored preflight error can allow state to commit
without its required events. Rollback belongs to the transaction owner.
`Publish` stages messages, and the SDK flushes before commit and closes writer
state on transaction completion. Treat writer initialization, write and commit
errors as distinct diagnostic phases. Avoid external calls and independent
publication retries in the callback.

## Routing and ordering

The routing key encodes the byte length and bytes of tenant ID, aggregate type
and aggregate ID in that order. SDK `BoundPartitionChooser` hashes those bytes
and routes within the server's partition bounds. Tenant and event UUIDs are
canonical lowercase, nonzero UUIDs. Empty tenant is reserved for non-tenant
facts. Opaque aggregate IDs retain their caller-defined representation.
Configure server-side partition bounds for keyed topics. On local YDB 26.3.1.25,
a multi-partition topic created with default disabled auto-partitioning failed
with `no bounds configured`. The test fixture uses `ScaleUp` with fixed and
equal minimum/maximum partitions. Verify bounds in the actual target
environment before enabling writes.

Aggregate revision describes a business state version, event ID identifies an
individual fact, partition and offset identify a physical position, and
producer identity and sequence number belong to the SDK/YDB write stream.
Multiple facts may share one revision. In live tests with unchanged partition
bounds, one aggregate stayed on one partition, sequential commits had
increasing offsets, and a same-transaction batch was read in input order for
one producer. Strict causal order across concurrent commits, independent
producers or partition changes is outside this contract. Consumers deduplicate
by event ID, track revisions, handle stale events and gaps, and reconcile
against authoritative state. Handle distinct events at the same revision
according to the event owner's semantics.

The SDK generates a new producer prefix for every MultiWriter, with one producer per touched partition. YDB documents a recommended
maximum of 100,000 active producer IDs per partition across 14 days and
retirement of idle bindings after 14 days. The pinned local YDB exposed
different settings for its partition source records: a count cap of 6,000,000
and a TTL of 1,382,400 seconds (16 days). Treat the documented binding window,
the observed partition settings, and the workload's acceptance budget as
separate inputs. Use the actual deployment settings when estimating capacity.

The soak test compares confirmed commits with source records reported by the
partition tablet. For a fixed partition, the observed growth was approximately
one retained record per commit. Estimate the source-ID budget per partition
from the rate of writer attempts touching that partition, including retries,
multiplied by the effective retention interval. Use the hottest partition and
headroom for burst traffic. At 40 new IDs per second, the documented 100,000-ID
planning limit corresponds to about 42 minutes. Test actual eviction and
resource usage at production scale before approving traffic above the validated
budget. See [YDB Topic documentation](https://ydb.tech/docs/en/concepts/datamodel/topic).

## Command receipts and transactional consumers

A retryable command should read a durable receipt scoped to the command and
trusted tenant, validate the original request hash, and return the stored
result when present. Record the receipt, state and required events in the same
transaction. Reconcile ambiguous outcomes via an authoritative receipt lookup
before resubmission. Retry with the same command and logical event identities
and reject reuse of the command ID with a different request hash. Choose
receipt retention to cover the retry horizon. The live fixture simulates
transport failure before the commit RPC and discards a successful server
commit response at the gRPC client boundary.

YDB-only consumers use `PopMessagesBatchTx` and the same `TxActor` for
consumer state and any resulting publication. Commit advances the offset
with those writes. After rollback or restart, the same consumer name can
resume uncommitted offsets. Consumer deduplication by logical Event ID is
a separate application concern. External systems need post-commit delivery,
independent idempotency and reconciliation. This package provides no bridge.

## Operations and security

`Publisher.Stats()` returns cumulative process-local counters for calls,
errors, staged batches, staged messages and serialized staged bytes. Export
these via the service's telemetry. Retries add staging attempts, and staging
can survive in counters after rollback. Collect transaction duration, retries,
commit outcomes and confirmed event counts at the owning transaction boundary
using existing SDK traces and service instrumentation. Use aggregate counters
or configured topic aliases for labels. Avoid event, tenant, aggregate and
correlation identifiers in telemetry labels. There is no telemetry endpoint
or exporter in this package.

Only trusted startup configuration selects physical topics. Use scoped YDB
credentials and TLS, object-level authorization and trusted tenant/source
context at the owning service. Routing provides locality. Per-call limits,
canonical IDs, metadata bounds, duplicate-ID checks and controlled `Any`
resolution reduce accidental abuse. Keep secrets and unnecessary personal
data out of event payloads and metadata. Validation errors omit protobuf
values. Treat wrapped SDK errors as operational diagnostics and sanitize them
at public boundaries.

## Verification and migration gates

Known test environment is disposable local YDB `26.3.1.25`, Go `1.27.4` and
YDB Go SDK `v3.157.1`. The SDK labels transactional and keyed writer paths
experimental. Repeat compatibility and fault tests when upgrading either
server or SDK, and under production-equivalent load and configuration.

```sh
just proto
.bin/buf lint
.bin/buf breaking --against '.git#ref=origin/main'
go test -count=1 ./internal/platform/event/...
go test -race -count=1 ./internal/platform/event/...
go vet ./internal/platform/event/...
go test -run '^$' -bench . -benchmem ./internal/platform/event/...
```

For native tests and benchmarks, set `M8_EVENT_TEST_YDB_DSN` and
`M8_EVENT_TEST_YDB_PREFIX` to an explicitly disposable loopback database.
Use equal published and internal gRPC ports for Discovery. Set
`M8_EVENT_TEST_YDB_REQUIRED=1` to fail on absent configuration, as CI does.
Run `go test -count=1 -v -run '^TestIntegration' ./internal/platform/event/...`
or `go test -run '^$' -bench '^BenchmarkNativeYDB$' -benchmem -benchtime=40x ./internal/platform/event/...`.
The Event Layer workflow runs contract and fast checks, the disposable YDB
suite including a load-induced partition split and hard restart, and a
separate multi-node failover job. The split test checks the inactive parent,
both new child partitions, retained offset ranges and the complete set of
committed IDs. Benchmark latencies include the whole transaction through
commit.

Run the disruption harness from the repository root against Docker. It creates
and destroys its own containers and targets the node reported as the owner
of the test topic tablet in failover mode. The nine-node fixture needs enough
memory to start a local replicated YDB cluster. Failover uses independent
in-memory disks per node. Hard restart uses persistent file disks.

```sh
bash internal/platform/event/testdata/ydb-disruption.sh restart
bash internal/platform/event/testdata/ydb-disruption.sh failover
```

For a separate metadata soak, start the same pinned local image with a
loopback monitor port exposed and keep its gRPC and monitoring settings on
that disposable instance. Run the test with the disposable DSN and prefix,
and set the monitor origin and duration as below. The test drives a bounded
rate and reads only the fixture's partition tablet.

```sh
M8_EVENT_TEST_SOAK_DURATION=8m M8_EVENT_TEST_MONITOR_URL=http://127.0.0.1:38765   go test -count=1 -timeout=10m -v -run '^TestSoakProducerMetadata$' ./internal/platform/event
```


Future migrations must move each command to a native transaction with state
and receipts, provision compatible topics, reconcile and drain its legacy
outbox, and then switch the authoritative publisher. Coordinate cutover if an
existing stream used SHA-256/hex routing, because the new key can repartition
the same aggregate. The repository has no migrated producer using this common
layer. Before production enablement, review external users, topic permissions,
retention, infrastructure recovery budgets, producer metadata capacity,
consumer reconciliation and rollback to a single authoritative path. The
disposable test cluster shares one host. A target multi-zone or multi-host
deployment has its own failure and recovery envelope.

See [ADR-0002](../../../docs/engineering/adr/ADR-0002-native-ydb-transactional-events.md)
and the [pinned Go SDK](https://github.com/ydb-platform/ydb-go-sdk/tree/v3.157.1).
