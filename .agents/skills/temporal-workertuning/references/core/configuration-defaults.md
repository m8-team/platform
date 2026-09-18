# Worker Configuration Defaults

SDK default values for Worker configuration options, organized by resource type.

## Compute settings

Compute settings control how many Tasks a Worker can execute concurrently.

### Configuration options

| Setting | Description |
|---------|-------------|
| `MaxConcurrentWorkflowTaskExecutionSize` | Maximum concurrent Workflow Tasks |
| `MaxConcurrentActivityTaskExecutionSize` | Maximum concurrent Activity Tasks |
| `MaxConcurrentLocalActivityTaskExecutionSize` | Maximum concurrent Local Activities |
| `MaxWorkflowThreadCount` / `workflowThreadPoolSize` | Thread pool for Workflow execution |

### Defaults by SDK

| SDK | MaxConcurrentWorkflowTaskExecutionSize | MaxConcurrentActivityTaskExecutionSize | MaxConcurrentLocalActivityTaskExecutionSize | MaxWorkflowThreadCount |
|-----|----------------------------------------|----------------------------------------|---------------------------------------------|------------------------|
| **Go** | 1,000 | 1,000 | 1,000 | - |
| **Java** | 200 | 200 | 200 | 600 |
| **TypeScript** | 40 | 100 | 100 | 1 (reuseV8Context) |
| **Python** | 100 | 100 | 100 | - |
| **.NET** | 100 | 100 | 100 | - |

`reuseV8Context` is TypeScript SDK only. `MaxWorkflowThreadCount` applies to Java SDK only.

Instead of fixed slot counts, resource-based slot suppliers automatically adjust available Task slots based on CPU and memory utilization.

> Worker tuners supersede the existing `maxConcurrentXXXTask` style Worker options.
> Using both styles will cause an error at Worker initialization time.

<!-- Sources: docs/develop/worker-tuning-reference.mdx:56–81, docs/develop/worker-performance.mdx:220–229 -->

## Memory settings

Memory settings control the Workflow cache size and thread pool allocation.

### Configuration options

| Setting | Description |
|---------|-------------|
| `MaxCachedWorkflows` / `StickyWorkflowCacheSize` | Number of Workflows to keep in cache |
| `MaxWorkflowThreadCount` | Thread pool size (Java SDK only) |
| `reuseV8Context` | Reuse V8 context for Workflows (TypeScript SDK only) |

For Go, use `SetStickyWorkflowCacheSize`. For Python, use the `max_cached_workflows` Worker option.

### Defaults by SDK

| SDK | MaxCachedWorkflows / StickyWorkflowCacheSize |
|-----|----------------------------------------------|
| **Go** | 10,000 |
| **Java** | 600 |
| **TypeScript** | Dynamic (e.g., 2000 for 4 GiB RAM) |
| **Python** | 1,000 |
| **.NET** | 10,000 |

### Cache options (Java SDK)

A Workflow Cache is created and shared between all Workers on a single host. These options are defined on `WorkerFactoryOptions`:

- `WorkerFactoryOptions#workflowCacheSize` defines the maximum number of cached Workflow Executions. Each cached Workflow contains at least one Workflow thread and its resources (memory, etc.).
- `maxWorkflowThreadCount` defines the maximum number of Workflow threads that may exist concurrently at any time.

These cache options limit the resource consumption of the in-memory Workflow cache. Workflow cache options are shared between all Workers because the Workflow cache is tightly integrated with the resource consumption of the entire host, including memory and total thread count, which should be limited per host/JVM.

<!-- Sources: docs/develop/worker-tuning-reference.mdx:83–105, docs/develop/worker-performance.mdx:349–363 -->

## IO settings

IO settings control the number of pollers and rate limits for Task Queue interactions.

### Configuration options

| Setting | Description |
|---------|-------------|
| `MaxConcurrentWorkflowTaskPollers` | Number of concurrent Workflow pollers |
| `MaxConcurrentActivityTaskPollers` | Number of concurrent Activity pollers |
| `Namespace APS` | Actions per second limit for Namespace |
| `TaskQueueActivitiesPerSecond` | Activity rate limit per Task Queue |

In the Java SDK, these poller options use different names: `workflowPollThreadCount` and `activityPollThreadCount`.

### Defaults by SDK

| SDK | MaxConcurrentWorkflowTaskPollers | MaxConcurrentActivityTaskPollers | Namespace APS | TaskQueueActivitiesPerSecond |
|-----|----------------------------------|----------------------------------|---------------|------------------------------|
| **Go** | 2 | 2 | 400 | Unlimited |
| **Java** | 5 | 5 | - | - |
| **TypeScript** | 10 | 10 | - | - |
| **Python** | 5 | 5 | - | - |
| **.NET** | 5 | 5 | - | - |

### Poller autoscaling (recommended)

Enable poller autoscaling by setting `*_task_poller_behavior` options to `PollerBehaviorAutoscaling`. This results in more efficient poller usage, better throughput, and schedule-to-start latency improvements. Names vary slightly depending on the SDK.

`PollerBehaviorAutoscaling` requires Temporal Server v1.28.0 or later.

<!-- Sources: docs/develop/worker-tuning-reference.mdx:107–132, docs/develop/worker-performance.mdx:231–248 -->

## "Large value" drawbacks

Specifying excessively large values without monitoring with SDK and system metrics leads to constant resource contention/stealing. This decreases total throughput and increases latency jitter of the system.

<!-- Sources: docs/develop/worker-performance.mdx:365–369 -->

## Invariants (Java SDK only)

These properties should always be true for a Java Worker's configuration. Perform this sanity check after adjustments to Worker settings.

1. `workflowCacheSize` should be less than or equal to `maxWorkflowThreadCount`. Each Workflow has at least one Workflow thread.
2. `maxConcurrentWorkflowTaskExecutionSize` should be less than or equal to `maxWorkflowThreadCount`. Having more Worker slots than the Workflow cache size leads to resource contention/stealing between executors and unpredictable delays. It is recommended that `maxWorkflowThreadCount` be at least 2x of `maxConcurrentWorkflowTaskExecutionSize`.
3. `maxConcurrentWorkflowTaskPollers` should be significantly less than `maxConcurrentWorkflowTaskExecutionSize`. And `maxConcurrentActivityTaskPollers` should be significantly less than `maxConcurrentActivityExecutionSize`. The number of pollers should always be lower than the number of executors.

<!-- Sources: docs/develop/worker-performance.mdx:371–379 -->

## Choosing slot supplier types

Temporal offers three types of slot suppliers: fixed assignment, resource-based, and custom.

### When to use fixed-size suppliers

- Workflow Tasks make minimal demands on CPU and normally do not consume much memory. They are well-served by fixed-sized slot suppliers.
- When very low Task completion latency and maximum throughput is important, avoid resource-based auto-tuning slot suppliers.
- Scenarios with tasks that have variable, or very high, per-task resource needs should rely on fixed-size suppliers and manual tuning.
- The resource-based tuner can never perform as well as a fixed-size tuner with appropriately chosen configuration, but can offer reasonable performance without the need for profiling.

### When to use resource-based suppliers

- **New workloads where resource usage is not yet understood**: Resource-based suppliers provide reasonable performance without upfront profiling. Once you understand the workload's resource patterns, switching to fixed-size with appropriate numbers will yield better performance.
- **Fluctuating workloads with low per-Task consumption**: Works well when each Task consumes few resources but may run for a relatively long time (e.g., HTTP calls or other blocking I/Os that spend most of their time waiting on external events).
- **Protection from out-of-memory and over-subscription with unpredictable per-task consumption**: Avoids crashes without setting an overly-conservative fixed limit. Keep in mind that auto-tuning can never do a perfect job and may sometimes exceed requested system limits for CPU and memory.

### Custom slot suppliers

For the highest level of control over slot allocation, consider custom slot suppliers. This allows you to tailor the logic of how slots are allocated based on your system requirements. Custom suppliers provide flexibility to optimize for specific use cases that fixed assignment and resource-based suppliers may not fully address.

<!-- Sources: docs/develop/worker-performance.mdx:388–426 -->
