# Platform Health

`internal/platform/health` is the reusable M8 Platform health module. It is technical platform foundation code and is not owned by any business module.

Domain modules may expose `HealthChecks() []health.Check`. They should not know about HTTP, gRPC, Kubernetes probes, or adapter packages.

## Probe Kinds

- `/livez` uses `CheckKindLiveness`. Use it for process-local checks only. Do not register external dependencies as liveness checks by default.
- `/readyz` uses `CheckKindReadiness`. Use it for checks that decide whether the service can receive traffic.
- `/startupz` uses `CheckKindStartup`. Use it for startup gating checks.
- `/healthz` uses `CheckKindReadiness`. Use it as a diagnostic endpoint for readiness checks.

## Aggregation

- Empty results are `HEALTHY`.
- A required dependency with `UNHEALTHY` or `UNKNOWN` makes the snapshot `UNHEALTHY`.
- An optional dependency with `UNHEALTHY` or `UNKNOWN` makes the snapshot `DEGRADED` unless a required check failed.
- Any `DEGRADED` check makes the snapshot `DEGRADED` unless a required check failed.

## Register Dependency Checks

```go
func (m *ResourceManager) HealthChecks() []health.Check {
    return []health.Check{
        {
            Spec: health.CheckSpec{
                Name: "resource-manager.storage",
                Target: health.Target{
                    Kind:   health.TargetKindDependency,
                    Name:   "postgres",
                    Module: "resource-manager",
                },
                Kinds:       []health.Kind{health.CheckKindReadiness},
                Criticality: health.CriticalityRequired,
            },
            Checker: checks.NewPingChecker("postgres", m.db.PingContext),
        },
    }
}
```

```go
registry := health.NewRegistry()

err := health.Register(registry,
    health.Check{
        Spec: health.CheckSpec{
            Name: "postgres",
            Target: health.Target{
                Kind:   health.TargetKindDependency,
                Name:   "postgres",
                Module: "resource-manager",
            },
            Kinds:       []health.Kind{health.CheckKindReadiness},
            Criticality: health.CriticalityRequired,
        },
        Checker: checks.NewPingChecker("postgres", postgres.PingContext),
    },
    health.Check{
        Spec: health.CheckSpec{
            Name: "kafka",
            Target: health.Target{
                Kind:   health.TargetKindDependency,
                Name:   "kafka",
                Module: "resource-manager",
            },
            Kinds:       []health.Kind{health.CheckKindReadiness},
            Criticality: health.CriticalityOptional,
        },
        Checker: checks.NewPingChecker("kafka", kafka.Ping),
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
