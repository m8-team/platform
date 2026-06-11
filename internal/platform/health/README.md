# Platform Health

`internal/platform/health` is the reusable M8 Platform health module. It is technical platform foundation code and is not owned by any business module.

Domain modules may expose `HealthChecks() []health.Config`. They should not know about HTTP, gRPC, Kubernetes probes, or adapter packages.

## Probe Kinds

- `/livez` uses `KindLiveness`. Use it for process-local checks only. Do not register external dependencies as liveness checks by default.
- `/readyz` uses `KindReadiness`. Use it for checks that decide whether the service can receive traffic.
- `/startupz` uses `KindStartup`. Use it for startup gating checks.
- `/healthz` uses `KindReadiness`. Use it as a diagnostic endpoint for readiness checks.

## Aggregation

- Empty results are `HEALTHY`.
- A required dependency with `UNHEALTHY` or `UNKNOWN` makes the snapshot `UNHEALTHY`.
- An optional dependency with `UNHEALTHY` or `UNKNOWN` makes the snapshot `DEGRADED` unless a required check failed.
- Any `DEGRADED` check makes the snapshot `DEGRADED` unless a required check failed.

## Register Dependency Checks

```go
func (m *ResourceManager) HealthChecks() []health.Config {
    return []health.Config{
        {
            Spec: health.Spec{
                Name: "resource-manager.storage",
                Target: health.Target{
                    Kind:   health.TargetKindDependency,
                    Name:   "postgres",
                    Module: "resource-manager",
                },
                Kinds:       []health.Kind{health.KindReadiness},
                Criticality: health.CriticalityRequired,
            },
            Check: checks.NewPingCheck("postgres", m.db.PingContext),
        },
    }
}
```

```go
registry := health.NewRegistry()

err := health.Register(registry,
    health.Config{
        Spec: health.Spec{
            Name: "postgres",
            Target: health.Target{
                Kind:   health.TargetKindDependency,
                Name:   "postgres",
                Module: "resource-manager",
            },
            Kinds:       []health.Kind{health.KindReadiness},
            Criticality: health.CriticalityRequired,
        },
        Check: checks.NewPingCheck("postgres", postgres.PingContext),
    },
    health.Config{
        Spec: health.Spec{
            Name: "kafka",
            Target: health.Target{
                Kind:   health.TargetKindDependency,
                Name:   "kafka",
                Module: "resource-manager",
            },
            Kinds:       []health.Kind{health.KindReadiness},
            Criticality: health.CriticalityOptional,
        },
        Check: checks.NewPingCheck("kafka", kafka.Ping),
    },
)
if err != nil {
    return err
}
```

## HTTP Adapter

```go
registry := health.NewRegistry()
mux := http.NewServeMux()

healthhttp.NewHandler(registry).RegisterRoutes(mux)
```

The HTTP adapter returns JSON snapshots. `HEALTHY` and `DEGRADED` return HTTP 200. `UNHEALTHY` and `UNKNOWN` return HTTP 503.

## gRPC Adapter

```go
adapter := healthgrpc.NewAdapter(
    registry,
    healthgrpc.WithPeriod(5*time.Second),
    healthgrpc.WithServiceNames("m8.resource-manager.v1.ProjectService"),
)

grpc_health_v1.RegisterHealthServer(grpcServer, adapter.Server())
go adapter.Start(ctx)
```

The adapter evaluates readiness. `HEALTHY` and `DEGRADED` map to `SERVING`; `UNHEALTHY` maps to `NOT_SERVING`; `UNKNOWN` maps to `UNKNOWN`.

## Fx

```go
var ResourceManagerModule = fx.Module(
    "resource-manager",
    fx.Provide(NewResourceManager),
    fx.Invoke(func(registry health.Registry, m *ResourceManager) error {
        return health.Register(registry, m.HealthChecks()...)
    }),
)

app := fx.New(
    health.FxModule,
    ResourceManagerModule,
)
```
