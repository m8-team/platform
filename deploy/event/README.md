# Event v1.2 deployment and recovery

The Event Layer owns two YDB row tables. Release engineering applies
`schema.yql` against the target database using a schema administrator.
Run the migration before starting publishers or reconciliation. Reapplying
the migration preserves rows. Verify schema compatibility and grants
independently because `IF NOT EXISTS` leaves an existing schema intact.

```sh
ydb -e "$YDB_ENDPOINT" -d "$YDB_DATABASE" yql -f deploy/event/schema.yql
```

Use dedicated, short-lived YDB credentials with TLS. Verify effective and
inherited permissions through negative and positive staging tests.

| Identity | Permissions |
|---|---|
| Migration administrator | Create and describe both tables, manage ACLs, audit restores |
| Reconciler | Database connect plus SELECT and UPDATE ROW on `replica_owners` |
| Publisher | Database connect, SELECT on owners, describe and SELECT/UPDATE ROW on guards, module-owned state access and write access to configured topics |
| Consumer | Its own consumer and projection permissions |

Run one reconciler configuration for each StatefulSet. All publishers using
the same physical topics and slot identity must use one YDB database, the
same owners and guards, and the same route/alias policy. Use database-root
`replica_owners` and `replica_publication_guards` in deployment. Deploying
with different table names can split ownership. The constructor checks path
syntax, distinct tables, unique topic paths and a valid binding. Deployment
review must enforce the shared namespace and prevent extra writers. Mirroring
the same Kubernetes UIDs into another cluster that writes to the same
topics requires a separate identity migration or isolated topics.

## Provision and roll out

1. Create the example StatefulSet at zero replicas. Resolve the real
   StatefulSet UID from its API object and supply it to both manifests
   through the trusted deployment pipeline. Pod UID, Pod name and namespace
   come from the Downward API. The publisher has no Kubernetes API token. Verify DNS and network reachability
   to every YDB endpoint advertised by Discovery, and to the Kubernetes
   API from the reconciler.
2. Configure `slots` as the range from ordinal zero through the largest
   ordinal that can ever hold ownership for this incarnation. Keep this
   maximum in deployment records after scale-down. Expand Role resource
   names and the configured range before scaling beyond it. Each configured
   slot is reconciled independently in a bounded transaction.
3. Build the reconciliation image from the repository. Put the approved
   YDB root bundle into the build context as `ca-certificates.crt`. The
   image exposes this bundle via `SSL_CERT_FILE`. Kubernetes supplies its
   service account CA separately.

   ```sh
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$CONTEXT/event-reconciler" ./cmd/event-reconciler
   docker build -f deploy/event/Dockerfile -t "$RECONCILER_IMAGE" "$CONTEXT"
   ```

4. Apply `reconciler.yaml` with rendered names, UID, endpoint, image and
   credentials. Supply the writer token at `YDB_TOKEN_FILE` through the
   deployment's credential provider. The process rereads projected tokens.
   Kubernetes authentication uses the service account token, CA and named
   GET permissions. Roll the deployment after changing its ConfigMap.
5. Start publisher replicas. Before accepting commands, bind to the
   process's own UID and run `Publisher.CheckReadiness`. Register the
   check with the platform health composition on `/readyz`. Retry startup
   binding within the startup budget while the reconciler assigns ownership.
   Stop admitting requests on SIGTERM, allow bounded work to drain and
   close the YDB driver. Choose grace and readiness budgets consistent with
   caller deadlines and the service SLO. The publisher manifest is a
   composition example for a future business module.
6. Verify the release configuration against a disposable environment.
   The smoke fixture exercises the packaged reconciler, dedicated YDB grants,
   TLS, and StatefulSet recreation.

   ```sh
   bash internal/platform/event/testdata/reconciler-smoke.sh
   ```

7. For recreation, list each predecessor UID and its maximum previously
   authorized slot count in the new reconciler ConfigMap. The new
   reconciler processes predecessors before assigning new slots and
   rejects a predecessor still reported active. Start new command traffic
   only after readiness. The Kubernetes and YDB operations are separate,
   so the release pipeline must retain the predecessor list across failures.

Pod replacement in a StatefulSet reuses the producer prefix and advances
the generation. A deleting or absent Pod clears the active binding.
Scale-up can fill a vacant slot with a new UID and another generation.
A deleted, deleting or replaced StatefulSet leaves a terminal tombstone.
Deleting Kubernetes objects or losing API access alone cannot revoke
YDB access. If observations fail, reconciliation preserves the last
committed binding and reports failure. The maximum stale-owner interval
therefore depends on Kubernetes observation availability and the workload
fencing policy. A security incident requires revocation of the affected YDB
credentials as well as workload isolation. A compromised process with
direct writes to business state or topics can bypass the library.
A Pod UID and a Producer ID identify instances and sources and confer no
cryptographic authentication. Restrict workload creation and mutation
permissions in publisher namespaces.

## Operational response

- **Retire or decommission.** Stop command admission and drain. Keep the
  reconciler and its Role active until ownership rows are vacant after
  scale-down or terminal after StatefulSet deletion. Verify the generation
  on every authorized ordinal before removing the reconciler and revoking
  credentials. A former process with an admitted transaction must conflict
  with the retirement. Durable rows and guard revisions stay in YDB.
- **Recover ownership after a restore.** Fence all old processes and
  credentials first. Stop reconciliation. Restore business tables, receipts,
  topic position and fencing state as a coordinated procedure. Use fresh
  workload identities and a new StatefulSet incarnation if continuity of
  generations or sequence metadata cannot be established. Re-enable
  reconciliation, verify all bindings, readiness and event/receipt
  consistency, then resume traffic. Never restore only owners or guards
  while a prior credential can still write. Rollback stops traffic and
  preserves both fencing tables. Roll back image/config only to a build
  that honors the same fences.
- **Ownership rejection or node failure.** Inspect Kubernetes object and
  reconciler availability, UID and generation for the slot, effective ACLs
  and the previous authorized process. Restore trusted observation and
  verify a handoff from YDB before resuming command traffic.
- **Lane saturation.** Compare wait, acquisition failures, native
  transaction duration and writer initialization latency. Shorten
  callbacks and evaluate topic throttling, conflicts and skew. Scale
  stable slots only after a service SLO and realistic load test.
- **Ambiguous commit.** Read the durable command receipt with the original
  command identity and fingerprint. The module decides how to reconcile
  an absent or pending receipt and retries with the same logical event
  IDs. Inspect transaction conflicts and timeouts. A staged event counter
  is not evidence of a committed business operation.
- **Topic incompatibility.** Inspect active partition coverage, retention
  and quotas. Verify topic layout, permissions and transactional writes
  with equivalent credentials against dedicated test topics. Readiness
  checks metadata without sending synthetic events or mutating the guard.
- **Producer growth or consumer lag.** Compare creation rate against rollout
  and restart history, check retained count from supported YDB metrics,
  and investigate unexpected writer identities. Configure lag alerts in
  the consumer's owning service using the consumer SLO.

## Monitoring and capacity

Use the supported YDB Prometheus exporter and a scoped scrape target.
In the local 26.3.1.25 node exporter at `/counters/prometheus`, the
observed counters include `PQ_SourceIdCreated` and `PQ_WriteError` with
`counters="tablets",type="PersQueue",category="app"`,
`Issues_YQL:KIKIMR_LOCKS_INVALIDATED{counters="kqp"}`, and the
`Transactions_Pure_ServerDurationMs` histogram. `LatencyMs` and `infly`
are available with `counters="grpc",subsystem="serverStats",stream="StreamWrite"`.
These stream series aggregate the node without topic labels. The creation
counter is node level and cannot supply retained-per-partition counts. Confirm the
deployment's aggregation and labels before enabling the example rules.
Configure the [supported YDB monitoring](https://ydb.tech/docs/en/reference/observability/metrics/)
and check all topic, partition and stream series at the actual cluster.
Where retained-count telemetry is missing, inventory stable deployment
slots and historical incarnations, then work with the YDB operator to
validate retained counts through supported monitoring. The test harness
can sample a disposable partition tablet. The production code makes
no calls to HTML monitoring pages.

Budget each hot partition for the active slots across all writers plus
historical incarnations and retired slots until eviction, split parents
and children, external writers, and YDB internal records. The YDB
recommendation of roughly 100,000 sources in 14 days per partition is
planning guidance. Verify effective Source ID lifetime and caps and use
an explicit margin. M8 example warning/critical planning thresholds are
50,000/80,000, subject to cluster evidence. Report actual retained count
separately from source creation.

Publish and transaction snapshots expose staged counts, confirmed
DoTx outcomes, explicit ambiguous commit responses, explicit commit
conflicts, handoffs, retirements, ownership rejection and lane use/wait.
Counters are cumulative per process. A successful no-op DoTx can avoid a
commit RPC. An aborted commit response is an observed commit-stage
conflict, and an explicit UNDETERMINED response is a lower bound on
unknown outcomes. Handoff metrics count direct active UID replacements
and retirement metrics count confirmed transitions to vacancy or a
terminal tombstone. Service telemetry should integrate these snapshots
and SDK timing hooks. This repository has no configured OpenTelemetry
provider or exporter. Metric wiring, histogram SLO alerts and verifying
native retained-count gauges on the production configuration remain
release gates.

Keep one lane. Measure end-to-end service time and shared partition
contention before setting capacity. The event limits are 512 KiB
serialized envelope, 100 events and 8 MiB per call by default. Several
topics can be staged per transaction. Its owner must budget total state
and event bytes, time, and retry cost across all topics.
