# Platform SDK (initial v0 API)

Small platform primitives for independently deployed Go services. No service
locator, global logger/provider, generated business framework or runtime DI.
Native clients remain available. Go 1.27.1 is the pinned toolchain.

Read [Architecture](ARCHITECTURE.md) first: Mermaid lifecycles, transaction and
delivery semantics, failure scenarios, boundaries, dependency decisions and rollout.
See the [reference service](../service-template/README.md) for runnable instructions.

```text
platform-sdk/
  app/          runtime + secure stdlib HTTP lifecycle
  lifecycle/    explicit function components
  health/       live, ready, startup
  logging/      slog + sensitive-key redaction
  metadata/     typed request ID and principal context values
  telemetry/    OTel providers, OTLP traces, Prometheus metrics
  connect/      unary policy chain + safe error mapping
  fault/        transport-neutral error codes and field violations
  retry/        explicitly classified bounded retries
  kafka/        native franz-go access, ack/commit, propagation
  transaction/  explicit sql.Tx helper
  idempotency/  same-transaction request deduplication
  outbox/       PostgreSQL insert + leased/fenced dispatch
  inbox/        same-transaction message deduplication
  temporal/     optional native worker lifecycle adapter
  testkit/      event recorder and logger
service-template/
  cmd/api/      composition root and signals
  internal/product/{domain,app,adapter}/
  internal/platform/  service config and IAM adapter
  api/proto/    source contracts
  internal/gen/ generated protobuf + Connect
  migrations/  product + platform schemas
  deploy/      local fixture + Kubernetes reference
```

## Public API navigation

These links point to compilable implementations, not pseudocode. Inspect an API
with `go doc ./app`, `go doc ./kafka`, etc.

| Concern | Public entry points |
|---|---|
| Runtime | [`app.Run(ctx, Config, ...lifecycle.Component)`](app/run.go), [`app.HTTP`](app/http.go) |
| Lifecycle | [`Component{Name, Start, Run, Stop, Force}`](lifecycle/component.go) |
| Probes | [`health.New`, `Registry.Register`, `Started`, `Drain`, `Broken`, `Handler`](health/health.go) |
| Logging | [`logging.New(writer, Info, Config)`, `WithContext(ctx, logger)`](logging/logging.go) |
| Telemetry | [`telemetry.New`, `Providers.Shutdown`, native Tracer/Meter/Propagator](telemetry/telemetry.go) |
| Errors | [`fault.Error`, `Code`, `Violation`, `New`](fault/error.go) |
| Retry | [`retry.Do(ctx, Policy, fn)`](retry/retry.go) |
| Kafka producer | [`NewProducer`, `Producer.Client`, `Send`, `Publish`](kafka/kafka.go) |
| Kafka consumer | [`NewConsumer` (native client), `Consumer.Run`, `HandlerFunc`, optional `DeadLetter`](kafka/kafka.go) |
| Idempotency | [`Execute(ctx, tx, Request, fn) (payload, replayed, error)`](idempotency/sql.go) |
| Outbox | [`Insert`, `Publisher`, `New`, `Worker.Run`, `Step`](outbox/outbox.go) |
| Inbox | [`Execute(ctx, tx, consumer, messageID, fn) (duplicate, error)`](inbox/sql.go) |
| Transactions | [`transaction.Run(ctx, db, func(*sql.Tx) error)`](transaction/sql.go) |
| Connect | [`New`, `Config`, `Authenticator`, `Authorizer`, `Limiter`, `ToError`, `HTTPStatus`](connect/interceptor.go) |

No universal idempotency `Store` is introduced yet: the only supported backend
must share the caller's `sql.Tx`. A detached lease-based store would promise a
different failure model. Introduce a consumer-owned interface when a second real
storage implementation needs it, not as a speculative abstraction.

## Scope and remaining production gates

Implemented: lifecycle rollback/drain, health, logging, tracing, RPC/producer/
consumer/outbox metrics, unary auth/validation/recovery, bounded retry, PostgreSQL
idempotency/outbox/inbox, native Kafka access and sequential consumer runtime.

Deliberately not implemented: YDB/Redis stores, partition worker lanes, retry-topic
scheduler, generic streaming policies, automatic retention, lag/backlog sampling,
distributed rate limiter, complete Temporal instrumentation, OTLP metrics push,
in-memory idempotency pretending to have SQL crash semantics. Kafka/SQL integration
tests are separate opt-in jobs, not evidence supplied by unit tests.

Before adoption: run real-infrastructure and failure-injection tests, configure
broker ACLs/IAM/network policy, retention and alerts, vulnerability checks and
load tests. Keep API v0 until these operational contracts stabilize.
