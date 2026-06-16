# Local ClickStack

M8 local Kubernetes can run ClickStack for logs, traces and metrics during local development.

## Start

```bash
make dev:up
make clickstack:install
make clickstack:port-forward
```

Open the UI:

```text
http://localhost:8080
```

## OTLP Endpoints

Inside the kind cluster:

```text
OTLP gRPC: m8-clickstack-otel-collector.m8-observability.svc.cluster.local:4317
OTLP HTTP: http://m8-clickstack-otel-collector.m8-observability.svc.cluster.local:4318
```

From the host, after `make clickstack:port-forward`:

```text
OTLP gRPC: localhost:4317
OTLP HTTP: http://localhost:4318
```

Local API key:

```text
m8-local-clickstack-api-key
```

The local collector runs as a DaemonSet and collects Kubernetes pod logs, host metrics,
kubelet metrics, cluster metrics and OTLP telemetry sent by M8 services.

The ClickStack collector uses supervisor/OpAMP mode. Local Kubernetes receivers are
added through `CUSTOM_OTELCOL_CONFIG_FILE` instead of replacing the base collector
configuration. The collector image is pinned to the ClickStack app version to keep
OpAMP config and the embedded ClickHouse exporter compatible. `make clickstack:install`
restarts the collector DaemonSet after Helm upgrade so mounted custom collector config
changes are applied reliably.

ClickStack operators are installed into `clickstack-system` by default and are configured
to watch the ClickStack namespace. This matters because the bundled MongoDB operator
only reconciles watched namespaces; without that override, the ClickStack app waits
forever for `m8-clickstack-mongodb-svc`.

## Manage

```bash
make clickstack:status
make clickstack:uninstall
make clickstack:reset
```

Useful overrides:

```bash
make CLICKSTACK_UI_PORT=18080 clickstack:port-forward
make CLICKSTACK_NAMESPACE=m8-observability-dev clickstack:install
make KUBE_CONTEXT=kind-m8-local clickstack:status
```
