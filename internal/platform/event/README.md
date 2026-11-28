# M8 Event v1.2

Stable Producer Identity is required for every publication. See the
[deployment and recovery runbook](../../../deploy/event/README.md).

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
typed payload. The scoped publish function resolves and validates the current `Any` again
because Protobuf messages are mutable. The payload descriptor must be linked
into the binary. Keep events immutable during publication. The owning module
enforces business semantics and the event type/payload correspondence.

Create one publisher per process and replica slot with trusted deployment
identity and explicit operation timeout. Register its readiness method with
`health/checks.NewPingCheck` in the service composition root.

```go
owner, err := event.NewTransactor(driver, operationTimeout)
if err != nil {
    return err
}
claim, err := slot.Bind(ctx, owner, podUID)
if err != nil {
    return err
}
publisher, err := event.NewPublisher(owner, cfg, claim, guardTable)
if err != nil {
    return err
}
if err := publisher.CheckReadiness(ctx); err != nil {
    return err
}
return publisher.WithinTransaction(ctx, func(
    ctx context.Context, tx query.TxActor, publish func(string, ...*eventv1.Event) error,
) error {
    // Resolve an existing command receipt before repeating business writes.
    if err := writeBusinessState(ctx, tx); err != nil {
        return err
    }
    if err := publish("resource", created); err != nil {
        return err
    }
    return writeCommandReceipt(ctx, tx)
})
```

The Repository uses the supplied actor for business queries. The SDK owns
the Serializable RW transaction, retries with `WithIdempotent(false)`,
flushes writers before commit, and cleans them up on completion. The lane
covers admission through SDK completion. Ownership is read and the guard
advanced on every attempt. Each physical topic gets one nonempty call per
attempt. Any failed publication poisons the attempt even if the caller
ignores its error. The callback and actor are scoped to the attempt. Return
promptly on cancellation and avoid network calls or independent retries
inside the callback. The owning module decides which integration events
must exist. A successful callback with no publication does not enforce a
business event obligation.

The actor supports transactional business queries. The SDK transactional
topic reader requires the SDK's concrete transaction lifecycle capability.
A transactional consume-and-forward operation through this protected API
would require a separately reviewed integration. The historical fixture
continues to cover native forwarding using direct SDK calls.

Readiness verifies the immutable binding, guard table shape and read access,
topic existence, active partition bounds and optional `MinTopicRetention`.
Write permissions, effective quotas and topic sizing are release checks.
Readiness sends no events and does not mutate fencing state.

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
multiplied by the effective retention interval. Under the count cap, TTL
cleanup of the implicit source records used by this writer also requires the
record's offset to fall below the partition's start offset. Plan for message
retention and the time needed for that offset to advance. Lowering the source
TTL alone can leave records in place. Use the hottest partition and headroom
for burst traffic. At 40 new IDs per second, the documented 100,000-ID planning
limit corresponds to about 42 minutes. Test resource usage and eviction with
the production settings before approving traffic above the validated budget.
See [YDB Topic documentation](https://ydb.tech/docs/en/concepts/datamodel/topic)
and the [pinned Source ID cleanup implementation](https://github.com/ydb-platform/ydb/blob/26.3.1.25/ydb/core/persqueue/pqtablet/partition/sourceid.cpp#L285).

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
can survive in counters after rollback. `Transactor.Stats()` counts SDK
attempts and retries, successful and failed invocations, and cumulative
invocation duration, ownership transitions and explicit commit conflicts. A successful lazy transaction can finish without a
commit RPC. These counters count invocation outcomes. The explicit
YDB `UNDETERMINED` count is limited to the SDK commit trace. Transport errors
such as a lost successful response can also make the outcome uncertain.
Use SDK trace hooks to export per-invocation latency distributions and writer
initialization errors. Count receipt-confirmed operations and committed
business events at the owning boundary. Use aggregate counters
or configured topic aliases for labels. Publisher stats include cumulative lane wait, acquisition errors, in-use and capacity gauges, and rejected ownership checks. Avoid event, tenant, aggregate and
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

## v1.1 Stable Producer Identity verification

With YDB Go SDK `v3.157.1`, `StartTransactionalWriter` accepts
`WithProducerIDPrefix` inside `WithWriteToManyPartitions`. The SDK appends
`-<partition_id>` to the prefix. The partition writer pool reads the init
response's `LastSeqNum` and advances the new MultiWriter's automatic sequence
counter before sending to that partition. A writer-local sequence counter can
create gaps across partitions. The bound chooser handles current partition
bounds and the split path uses a separate producer for each child. The SDK
flushes before commit and stops writers on transaction completion. The Query
SDK decides when to retry a serializable read-write transaction. An
application callback must recreate its messages and writers on each attempt.

The loopback sequence fixture confirmed four durable state mutations and
the exact payloads, contiguous offsets and increasing sequences of eight
events in one topic and four in a second topic. Each attempt uses the same
Producer ID for both physical topics. It covers a rollback after staging,
one injected retriable callback result after staging, opening the same
identity through a second driver and a
discarded successful commit response followed by a state lookup and another
publication with the same identity. The replay decision for a real command
still belongs to its durable receipt. A lost response when the server is
still resolving the commit remains a separate case.

The duplicate-identity fixture blocks the first producer's gRPC
`SendMsg` after it receives a sequence number. A second independent YDB
driver commits a different mutation and event using the same identity and
partition. After releasing the first write, both transactions report
success. Both state keys exist, but the topic contains only the second
event. Fencing did not prevent suppression on YDB `26.3.1.25`.
Stable publication therefore requires both ownership verification and the
transactional guard. The historical direct-SDK tests intentionally retain
the unprotected configuration as a negative control.

### Replica handoff during a partition

Kubernetes can preserve exclusive use of an ordinal by waiting for confirmed
termination. A control-plane partition can make the wait indefinite. Force
deletion allows a replacement while the old Pod can still reach YDB.
The out-of-service taint assumes the node is already shut down. Kubernetes
leader election and YDB Coordination Node semaphores coordinate ownership
among participating clients. Neither makes the native Topic or Query API
check that ownership at commit. Readiness, lease expiry and cancellation
cannot revoke a request the old process already sent.

The live handoff test checks a possible storage-side fence. Every publishing
transaction reads the replica generation in the same serializable YDB
transaction as state and topic writes. A privileged fixture increments the
generation while the old writer is delayed before sending a sequenced write
or before sending commit after flushing. It resumes the old writer after the
replacement commits, and also checks the order in which the stale, flushed
transaction aborts before the replacement opens its writer. YDB returns
`ABORTED`, and the SDK retry sees the new generation and rejects the old
owner. The replacement can reuse the same Producer ID. The fixture checks
both exact committed state keys and both expected event payloads, offsets
and producer sequences.

`ReplicaSlot.Reconcile` implements the handoff. A trusted operator reads the
generation in a serializable YDB transaction, then issues a direct Kubernetes
Pod GET with no `resourceVersion` override. It validates the namespace, Pod
name, non-terminating Pod UID and pinned StatefulSet controller UID from
deployment configuration. It increments the generation only when the UID
changes. If a competing handoff changes the row, the SDK retries the entire
operation and obtains a fresh Pod observation. The operator supplies an
authenticated HTTPS client and a bounded transaction timeout.

Provision the following ownership table through the deployment owner. Give
the reconciler a dedicated Kubernetes credential with `get` on the named Pod
and a separate YDB credential with read and update rights on this table.
Publisher credentials get read-only access to the ownership table. Enforce
YDB authentication and restrict who can create or modify Pods and
StatefulSets in the deployment namespace. Deliver the publisher's own
`metadata.uid` through a Downward API environment variable. Keep that value
fixed at process startup and use `ReplicaSlot.Bind` to obtain an immutable
`PodOwnership` for use across transaction retries.

```sql
CREATE TABLE replica_owners (
    replica Utf8 NOT NULL,
    pod_uid Utf8 NOT NULL,
    generation Uint64 NOT NULL,
    PRIMARY KEY (replica)
);
```

The Kubernetes fixture runs two actual Pods with the same StatefulSet
ordinal on different kind worker nodes. It drops traffic in both directions
between the old worker and the API server, keeps YDB reachable, force-deletes
the old Pod object and waits for the replacement. The test also delays a
successful read of the old Pod across the replacement. It checks the
replacement's stable-ID publications, rejection of the old UID and
authorization to update the ownership row. Run it on a disposable Docker
host with kind and kubectl available:

```sh
bash internal/platform/event/testdata/kubernetes-handoff.sh
```

The fixture runs on one Docker host. See the
[Kubernetes shutdown guidance](https://kubernetes.io/docs/concepts/cluster-administration/node-shutdown/),
[leader election contract](https://pkg.go.dev/k8s.io/client-go/tools/leaderelection),
and [YDB serializable transaction semantics](https://ydb.tech/docs/en/concepts/transactions).

### Stable publication

The producer prefix is `m8-` followed by the SHA-256 hexadecimal digest of
`namespace/StatefulSetUID/PodName`. The SDK adds its partition suffix.
Pod UID and generation are held in the ownership binding. One transaction
reads ownership and advances a separate guard while writing business state
and events. The guard makes overlapping attempts conflict across same-Pod
process restarts. A retry uses fresh writers and initializes their
sequence state through the SDK.

The Kubernetes fixture runs the protected Publisher and restarts its process
without changing Pod UID. It covers a confirmed YDB commit with its response
withheld from the application and a process exit during the commit send.
The restarted process publishes before the fixture checks the interrupted
outcome. The live YDB overlap test separately holds an old commit request
while another publisher with the same UID commits. The old transaction is
aborted and SDK retry publishes after the winning event. Both checks compare
exact committed state, event payloads, Producer ID and offsets.

Deploy the reconciler and both tables before enabling the Publisher.
Production deployments and business-module wiring are managed separately.

### Metadata budget

The local A/B metadata run used one partition, six workers, batches of one,
and a target of 40 transactions/s for one minute per strategy. The topic
end equaled confirmed commits in both cases. The random fixture retained
2,399 Source IDs after 2,399 commits. Six exclusive test producers retained
6 IDs after 2,324 commits and were already at 6 at 20 seconds. These are
tablet counts on the disposable fixture. Budget active IDs per partition
as the sum of replicas times lanes across deployments that write to the
topic. Include inactive split partitions and distinct historical identities
from scale-down, renames and reconfiguration until observed eviction.
The accelerated TTL fixture verified reusing an evicted identity at the
next offset with sequence 1. Count-cap eviction and production retention
remain unverified. Lane counts and p99 targets require representative
workload rates, batch sizes, traffic skew and failover headroom.

The official YDB metrics reference documents `topic.producers_count` for
the topic and `topic.partition.producers_count_max` for its largest
partition count. On the pinned local image the `topics` Prometheus subgroup
had no samples during the single-partition soak. The `tablets` subgroup
exported `PQ_SourceIdCreated`, a cumulative creation counter aggregated
by tablet type, with no topic label. Use that counter to investigate
creation activity. Capacity alerts need a retained-count gauge with topic
labels, verified against the configured and historical identity budget on
the target YDB export. Production metric export and alerts remain a release
gate. Keep the disposable tablet HTML sampler in the test fixture. See the
[YDB metrics reference](https://ydb.tech/docs/en/reference/observability/metrics/#topic-metrics)
and [Kubernetes StatefulSet force deletion](https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/).

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
or `go test -run '^$' -bench '^BenchmarkNativeYDB$' -benchmem -benchtime=1000x ./internal/platform/event/...`.
The Event Layer workflow runs contract and fast checks, the disposable YDB
suite including a load-induced partition split and hard restart, and a
separate multi-node failover job. The pull request workflow compares metadata
for one minute per strategy, including protected publication. A manual workflow dispatch can run 8 or 30
minutes for the protected Publisher using the same disposable server image. The split test induces a split with the baseline pressure workload and
reuses a stable prefix before and after the split. It checks the inactive
parent, both children, partition-specific IDs, retained offsets and all
committed IDs. The optional `M8_EVENT_TEST_STABLE_SPLIT_LOAD=1` experiment
uses the stable prefix for the pressure workload too. That run reached its
80-second deadline before observing a split and remains unverified. Benchmark latencies include the whole transaction through
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
M8_EVENT_TEST_SOAK_DURATION=8m M8_EVENT_TEST_MONITOR_URL=http://127.0.0.1:38765 \
  go test -count=1 -timeout=20m -v -run '^TestSoakProducerMetadata$' ./internal/platform/event
```


The accelerated TTL test uses a separate disposable container. Start the image
with `--enable-feature-flag enable_topic_retention_delete_last_blob` after the
image name, expose the loopback monitor port, and use the same DSN, prefix and
monitor environment as for the soak. Run
`go test -count=1 -timeout=3m -v -run '^TestProducerMetadataTTL$' ./internal/platform/event`.
CI requires the test via `M8_EVENT_TEST_TTL_REQUIRED=1`. The test shows that a
source record with a 10-second TTL disappears once its message has expired.
A young source survives its message's removal, and an expired source survives
while its message still reads at the original offset. With default settings,
the single-blob trial kept its start offset at zero for the 70-second
observation. The test intentionally enables last-blob deletion to advance the
offset on a tiny topic. Production-TTL timing and capacity eviction remain
outside this test.


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
