# Product service reference

## Responsibility

Owns tenant-scoped products and their creation events. Does not own identities,
tenants, Kafka infrastructure or cross-service workflows. This is an independent
reference module, not an additional M8 business module or a migration of existing
M8 services.

## APIs and events

`product.v1.ProductService` exposes `GetProduct` and `CreateProduct` using Connect,
gRPC and gRPC-Web unary protocols. GetProduct declares `NO_SIDE_EFFECTS` and supports
the Connect HTTP GET protocol; no separate unprotected REST handler is registered.
Create requires an idempotency key, validated on transport and application boundaries.
Both operations check permissions again in the application layer. Tenant comes
from verified identity, never from the request body. Queries always include tenant.

Publishes `company.product.created.v1` to `product.events` through a transactional
outbox, keyed by product ID. Event ID is stable, headers include tenant/schema and
trace context. Consumes no events: consumer/inbox adapters are SDK opt-ins.

## Run locally

Requires Go 1.27.1, buf and an active Docker daemon. From this directory:

```sh
make local
export LOCAL_INSECURE=true
export DEV_AUTH_TOKEN=local-only-0123456789abcdef0123456789abcdef
export DATABASE_URL='postgres://product:local-only-password@localhost:5432/product?sslmode=disable'
export KAFKA_BROKERS=localhost:9092
go run ./cmd/api
```

Local fixture credentials above are public examples, never production secrets.
Compose binds host ports to loopback and creates a disposable database. Migrations
run on initial database creation only. Production migrations are a separate gated
deployment job, never application startup hooks.

```sh
curl -s http://localhost:8080/product.v1.ProductService/CreateProduct \
  -H "Authorization: Bearer $DEV_AUTH_TOKEN" \
  -H 'Content-Type: application/json' -H 'Connect-Protocol-Version: 1' \
  -d '{"name":"Coffee","idempotencyKey":"create-coffee-0001"}'

# Repeat the same command: the same product ID is returned.
# Change name but retain key: aborted/conflict.
curl -sG http://localhost:8080/product.v1.ProductService/GetProduct \
  -H "Authorization: Bearer $DEV_AUTH_TOKEN" \
  --data-urlencode 'connect=v1' --data-urlencode 'encoding=json' \
  --data-urlencode 'message={"id":"REPLACE_WITH_PRODUCT_ID"}'
curl -s http://localhost:8081/health/ready
curl -s http://localhost:8081/metrics
```

For generated Go clients, pass `connect.WithHTTPGet()` to enable safe GET calls.
Do not log the authorization header or put access tokens in URLs.

## Verify

```sh
make proto test race vet
export TEST_DATABASE_URL="$DATABASE_URL"
export TEST_KAFKA_BROKERS="$KAFKA_BROKERS"
make integration
go test ./internal/product/domain -fuzz=FuzzProduct -fuzztime=10s
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Integration tests refuse to infer a production database: explicit TEST variables
are required. They create unique test records and retain them for inspection in
the disposable fixture. Use a fresh fixture for reproducible runs. Root `go test
./...` does NOT traverse these nested modules; the Makefile checks both explicitly.

## Production configuration

Required: `DATABASE_URL` with `sslmode=verify-full`, `KAFKA_BROKERS`,
`TLS_CERT_FILE`, `TLS_KEY_FILE`, `AUTH_INTROSPECTION_URL` (HTTPS), `AUTH_CLIENT_ID`,
`AUTH_CLIENT_SECRET`. Optional: `HTTP_ADDRESS`, `ADMIN_ADDRESS`,
`OTEL_EXPORTER_OTLP_ENDPOINT` (OTLP gRPC host:port), `SHUTDOWN_TIMEOUT` (default 30s).
Public HTTP and Kafka use TLS outside explicit local mode. Configure Kafka SASL or
client certificates through native franz-go options for your broker ACL model;
the reference does not invent broker credentials. PostgreSQL account permissions
must be limited to this service's schemas; migrations use a separate role.

Production identity uses RFC 7662 introspection with company `tenant_id` and `roles`
extensions. Required roles: `product.read`, `product.create`. Introspection has a
3s timeout, rejects redirects and fails closed. Adapt it to the company's real IAM
contract before deploying. No token cache or implicit retry is enabled.

Management port has no public Service route. Restrict it with cluster-specific
NetworkPolicy and scrape authorization/network controls. No pprof endpoint is
publicly registered. The supplied probes give startup up to 60s; dependency
failures affect readiness only. There is no blind preStop sleep: drain starts on
SIGTERM. Add a measured routing-propagation delay only if your ingress needs it,
within the 45s Pod budget. Streaming requires a dedicated policy/server, not
disabling unary limits globally.

Metrics are Prometheus-compatible and can be scraped into Mimir. Traces use OTLP
when an endpoint is configured; configure the collector separately. Production
rollout must add backlog/dead-row alerts, broker lag monitoring, retention jobs,
rate-limit policy, retry/DLQ operational procedures, schema/event compatibility
checks, vulnerability scanning and load/failure-injection tests. This is a tested
reference foundation, not a claim of production certification.

See [SDK architecture](../platform-sdk/ARCHITECTURE.md) for lifecycle diagrams,
failure semantics, SDK APIs, dependency rationale and versioning.
