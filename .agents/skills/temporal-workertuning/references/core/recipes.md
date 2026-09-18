# Worker Tuning Recipes

End-to-end tuning playbooks that chain documented facts into actionable workflows. Each step cites its source document.

---

## Recipe 1: Initial Worker Setup Checklist (Pre-Production)

Before deploying Workers to production, walk through each item below.

### Step 1: Actively tune Worker options instead of relying on defaults

Default Worker settings are designed for ease in development and testing, not optimal for production environments. Your Workflow complexity, Activity duration, payload sizes, and infrastructure constraints all influence optimal Worker configuration.
<!-- Sources: docs/best-practices/worker.mdx:222-228 -->

### Step 2: Choose a slot supplier strategy

Decide between fixed-size and resource-based slot suppliers:

- **Workflow Tasks** make minimal demands on CPU and normally do not consume much memory. They are well-served by fixed-sized slot suppliers.
- When very low Task completion latency and maximum throughput is important, avoid resource-based auto-tuning slot suppliers.
- Resource-based slot suppliers are useful when resource usage patterns are not yet understood, or for protection from out-of-memory with unpredictable per-task consumption. Outside those scenarios, prefer fixed-size.
- A resource-based tuner can never perform as well as a fixed-size tuner with appropriately chosen configuration.

Do NOT mix `maxConcurrentXXXTask` style options with Worker tuners. Using both styles will cause an error at Worker initialization time.
<!-- Sources: docs/develop/worker-performance.mdx:84-86, 392-406 -->

### Step 3: Configure poller behavior

Use Poller Autoscaling (`PollerBehaviorAutoscaling`) for the majority of use cases. Manually setting the number of pollers too high or too low will result in decreased performance.

**Requirement**: `PollerBehaviorAutoscaling` is only enabled in Temporal Server v1.28.0 and later.
<!-- Sources: docs/develop/worker-performance.mdx:108-110, 235-241 -->

### Step 4: Tune the sticky Workflow cache

Configure the Workflow cache size based on your SDK:

- **Go**: `SetStickyWorkflowCacheSize` (default: 10,000)
- **Java**: `WorkerFactoryOptions#workflowCacheSize` (default: 600) and `maxWorkflowThreadCount` (default: 600). Note: `maxWorkflowThreadCount` is Java SDK only.
- **TypeScript**: Dynamic (e.g., 2,000 for 4 GiB RAM)
- **Python**: `max_cached_workflows` (default: 1,000)
- **.NET**: Default 10,000
<!-- Sources: docs/develop/worker-performance.mdx:349-363; docs/develop/worker-tuning-reference.mdx:96-104 -->

### Step 5: Verify Java SDK invariants (Java only)

If using the Java SDK, confirm these invariants after configuration:

1. `workflowCacheSize` should be less than or equal to `maxWorkflowThreadCount`. Each Workflow has at least one Workflow thread.
2. `maxConcurrentWorkflowTaskExecutionSize` should be less than or equal to `maxWorkflowThreadCount`. It is recommended that `maxWorkflowThreadCount` be at least 2x of `maxConcurrentWorkflowTaskExecutionSize`.
3. `maxConcurrentWorkflowTaskPollers` should be significantly less than `maxConcurrentWorkflowTaskExecutionSize`. And `maxConcurrentActivityTaskPollers` should be significantly less than `maxConcurrentActivityExecutionSize`. The number of pollers should always be lower than the number of executors.
<!-- Sources: docs/develop/worker-performance.mdx:373-379 -->

### Step 6: Separate Task Queues logically

Use separate Task Queues for distinct workloads. This isolation allows you to control rate limiting, prioritize certain workloads, and prevent one workload from starving another. For each Task Queue, ensure you configure at least two Workers to poll.
<!-- Sources: docs/best-practices/worker.mdx:124-128 -->

### Step 7: Set up minimal monitoring alerts

Configure these alerts before going to production:

| Alert | Threshold |
|---|---|
| `workflow_task_schedule_to_start_latency` (p99) | Alert at >200ms |
| `activity_schedule_to_start_latency` (p99) | Alert at >200ms |
| Sync Match Rate (p99) | Alert at <95% |
| Poll Success Rate (p99) | Alert at <90% |
| `worker_task_slots_available` (p99) | Alert at 0 |
| `sticky_cache_total_forced_eviction` | Alert at a predetermined high number |
| `approximate_backlog_count` | Alert when growing over time |
<!-- Sources: docs/cloud/worker-health.mdx:50-91 -->

### Step 8: Use Worker Versioning for safe deployments

Use Worker Versioning to deploy new Workflow code without breaking running Executions. It maps each Workflow Execution to a specific Worker Deployment Version identified by a build ID.
<!-- Sources: docs/best-practices/worker.mdx:166-176 -->

### Step 9: Manage Event History growth

Do not exceed a few thousand Events in a single Workflow Execution. Use Continue-As-New to continue under a new Workflow Execution with a new Event History. A Workflow Execution may be terminated if any single payload exceeds 2 MB or if the entire Event History exceeds 50 MB.
<!-- Sources: docs/best-practices/worker.mdx:190-208 -->

### Step 10: Run benchmarks

Test your configuration under realistic load to confirm limits and settings are appropriate for your environment.
<!-- Sources: docs/best-practices/worker.mdx:42 -->

---

## Recipe 2: Diagnosing High Schedule-to-Start Latency

Use this decision tree when `workflow_task_schedule_to_start_latency` or `activity_schedule_to_start_latency` P95 exceeds one second.

### Step 1: Check Worker CPU and memory usage

If Worker hosts are fully utilized (near full CPU usage, high load average), go to Step 2a.
If Worker hosts are underutilized, go to Step 2b.
<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:41-46; docs/best-practices/worker.mdx:257-268 -->

### Step 2a: Workers are saturated (high CPU/memory)

High Schedule-to-Start latency with high CPU/memory means Workers are saturated.

- **Action**: Scale up Workers or add more Worker instances.
- **Check for blocked Activities**: If `worker_task_slots_available{worker_type="ActivityWorker"}` is at zero, Activities may be blocked or "zombie" Activities may be consuming slots. Zombie Activities occur when an Activity times out (hits its `StartToClose` or `HeartbeatTimeout`) and has stopped Heartbeating but continues to run, occupying slots as more retries occur.
- **Action for blocked Activities**: Add client-side timeout to downstream API clients. Review Task code to ensure Tasks complete within a reasonable time.
<!-- Sources: docs/best-practices/worker.mdx:260-263; docs/troubleshooting/performance-bottlenecks.mdx:159-173 -->

### Step 2b: Workers are underutilized (low CPU/memory)

High Schedule-to-Start latency with low CPU/memory means Workers are not using their capacity. Proceed through the following sub-checks in order.

#### Sub-check A: Verify executor slots are not depleted

Check `worker_task_slots_available` for the corresponding Worker type. If the metric frequently shows a depleted number of available slots while hosts are underutilized, increase `maxConcurrentWorkflowTaskExecutionSize` or `maxConcurrentActivityExecutionSize`.

Alternatively, consider using a resource-based slot supplier.
<!-- Sources: docs/develop/worker-performance.mdx:895-904 -->

#### Sub-check B: Verify pollers are sufficient

If Worker hosts are underutilized AND `worker_task_slots_available` shows a significant percentage of slots are available AND `schedule_to_start` is abnormally long, increase poller count by adjusting `maxConcurrentWorkflowTaskPollers` or `maxConcurrentActivityTaskPollers` depending on which `schedule_to_start` metric is elevated.

Or use Poller Autoscaling (`PollerBehaviorAutoscaling`) to handle this automatically.
<!-- Sources: docs/develop/worker-performance.mdx:907-917 -->

#### Sub-check C: Verify rate limits are not throttling

If after adjusting pollers and executors you still observe elevated `schedule_to_start`, underutilized Worker hosts, and high `worker_task_slots_available`, check:

- If server-side rate limiting per Task Queue is set by `maxTaskQueueActivitiesPerSecond`, remove the limit or adjust the value up.
- If Worker-side rate limiting per Worker is set by `maxWorkerActivitiesPerSecond`, remove the limit.
<!-- Sources: docs/develop/worker-performance.mdx:919-924 -->

#### Sub-check D: Check for network issues

If accompanied by high `temporal_long_request_latency` or `temporal_long_request_failure`, Workers are struggling to reach the Temporal Service. Check the network connection between the Client and Server. If `ResourceExhausted` status codes appear, review rate limits.
<!-- Sources: docs/best-practices/worker.mdx:264-268; docs/troubleshooting/performance-bottlenecks.mdx:183-193 -->

### Step 3: Check the Sync Match Rate

Calculate Sync Match Rate:

```
temporal_cloud_v1_poll_success_sync_count / temporal_cloud_v1_poll_success_count
```

The Sync Match Rate should be at least >95%, but preferably >99%.

- If Sync Match Rate is low AND Schedule-to-Start is high, verify Worker setup is optimized: check CPU usage against task slots, check memory usage against `sticky_cache_size`, increase concurrent pollers if resources allow, and increase the number of available Workers.
- If Sync Match Rate is low, you may need to contact Temporal Cloud support.
<!-- Sources: docs/cloud/worker-health.mdx:155-246 -->

### Step 4: Check for Activity-specific causes

If `activity_schedule_to_start_latency` specifically is high, also check whether `TaskQueueActivitiesPerSecond` is set too low, which can limit the rate at which Activities are started.
<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:56 -->

---

## Recipe 3: Right-Sizing Workers (Compute / Memory / IO Tuning Sequence)

Tune Workers along the three resource dimensions: compute, memory, and IO. Follow these steps in order.

### Phase 1: Compute Tuning

#### Step 1: Review SDK defaults for concurrent execution

Know your SDK's defaults before tuning:

| SDK | MaxConcurrentWorkflowTaskExecutionSize | MaxConcurrentActivityTaskExecutionSize | MaxConcurrentLocalActivityTaskExecutionSize |
|-----|----------------------------------------|----------------------------------------|---------------------------------------------|
| Go | 1,000 | 1,000 | 1,000 |
| Java | 200 | 200 | 200 |
| TypeScript | 40 | 100 | 100 |
| Python | 100 | 100 | 100 |
| .NET | 100 | 100 | 100 |
<!-- Sources: docs/develop/worker-tuning-reference.mdx:70-77 -->

#### Step 2: Determine slot strategy

- Fixed-size slot suppliers work well when you can estimate per-task resource consumption and calculate an upper bound on concurrency. They also handle variable or very high per-task resource needs better than resource-based suppliers when paired with manual tuning.
- Resource-based slot suppliers are useful when resource usage patterns are not yet understood, or for protection from out-of-memory with unpredictable per-task consumption. They cannot match the performance of a properly configured fixed-size supplier.

Do NOT use both `maxConcurrentXXXTask` style options and Worker tuners. Using both causes an error at Worker initialization time.
<!-- Sources: docs/develop/worker-performance.mdx:52-58, 84-86, 392-420 -->

#### Step 3: Monitor compute metrics

| Configuration option | Metric to watch |
|---|---|
| `MaxConcurrentWorkflowTaskExecutionSize` | `worker_task_slots_available {worker_type=WorkflowWorker}` |
| `MaxConcurrentActivityTaskExecutionSize` | `worker_task_slots_available {worker_type=ActivityWorker}` |
| `MaxWorkflowThreadCount` | `workflow_active_thread_count` (Java SDK only) |
| CPU-intensive logic | `workflow_task_execution_latency` |

Also monitor machine CPU consumption (e.g., `container_cpu_usage_seconds_total` in Kubernetes).

Note: `worker_task_slots_available` can only be used with fixed-size slot suppliers and cannot be used with resource-based slot suppliers. Use `worker_task_slots_used` for resource-based suppliers.
<!-- Sources: docs/develop/worker-tuning-reference.mdx:140-149; docs/develop/worker-performance.mdx:196-203 -->

#### Step 4: Adjust executor slots

Increase `maxConcurrentWorkflowTaskExecutionSize` or `maxConcurrentActivityExecutionSize` if:

1. Worker hosts are underutilized (no CPU bottlenecks).
2. `worker_task_slots_available` from the corresponding Worker type frequently shows depleted slots.

Monitor Worker CPU and memory usage while increasing values.
<!-- Sources: docs/develop/worker-performance.mdx:895-904; docs/troubleshooting/performance-bottlenecks.mdx:156-157 -->

### Phase 2: Memory Tuning

#### Step 5: Review cache defaults

| SDK | MaxCachedWorkflows / StickyWorkflowCacheSize |
|-----|----------------------------------------------|
| Go | 10,000 |
| Java | 600 |
| TypeScript | Dynamic (e.g., 2,000 for 4 GiB RAM) |
| Python | 1,000 |
| .NET | 10,000 |
<!-- Sources: docs/develop/worker-tuning-reference.mdx:96-104 -->

#### Step 6: Monitor cache metrics

Watch these metrics to assess cache health:

| Configuration option | Metric to watch |
|---|---|
| `StickyWorkflowCacheSize` | `sticky_cache_total_forced_eviction`, `sticky_cache_size`, `sticky_cache_hit`, `sticky_cache_miss` |

The `sticky_cache_size` should report less than or equal to your `WorkflowCacheSize` value. Also, `sticky_cache_total_forced_eviction` should not be reporting high numbers.
<!-- Sources: docs/develop/worker-tuning-reference.mdx:153-157; docs/cloud/worker-health.mdx:349-351 -->

#### Step 7: Adjust cache if needed

- If `sticky_cache_size` hits `workflowCacheSize` (or `workflow_active_thread_count` hits `maxWorkflowThreadCount` in Java), Workflow Executions will start to be evicted. An evicted Workflow needs to be replayed when it gets any action that may advance it.
- If cache limits are hit and Worker hosts have enough free RAM, increase `workflowCacheSize` (and `maxWorkflowThreadCount` in Java) to decrease replay latency and cost.
- If Workers are memory-bound, consider reducing the cache size.
<!-- Sources: docs/develop/worker-performance.mdx:690-696 -->

### Phase 3: IO Tuning

#### Step 8: Review poller defaults

| SDK | MaxConcurrentWorkflowTaskPollers | MaxConcurrentActivityTaskPollers |
|-----|----------------------------------|----------------------------------|
| Go | 2 | 2 |
| Java | 5 | 5 |
| TypeScript | 10 | 10 |
| Python | 5 | 5 |
| .NET | 5 | 5 |
<!-- Sources: docs/develop/worker-tuning-reference.mdx:122-129 -->

#### Step 9: Enable Poller Autoscaling or adjust manually

Use Poller Autoscaling for the majority of use cases. It automatically selects an appropriate number of pollers based on need, resulting in more efficient poller usage, better throughput, and schedule-to-start latency improvements.

Requires Temporal Server v1.28.0 and later.

For manual adjustment: increase pollers only if Worker hosts are underutilized, `worker_task_slots_available` shows significant available slots, and `schedule_to_start` is abnormally long.
<!-- Sources: docs/develop/worker-performance.mdx:235-240, 907-917 -->

#### Step 10: Monitor IO metrics

| Configuration option | Metric to watch |
|---|---|
| `MaxConcurrentWorkflowTaskPollers` | `num_pollers {poller_type=workflow_task}` |
| `MaxConcurrentActivityTaskPollers` | `num_pollers {poller_type=activity_task}` |
| Network latency | `request_latency {namespace, operation}` |
<!-- Sources: docs/develop/worker-tuning-reference.mdx:160-165 -->

#### Step 11: Check for rate limiting

- Server-side: `maxTaskQueueActivitiesPerSecond`
- Worker-side: `maxWorkerActivitiesPerSecond`

Remove or increase these limits if they are causing elevated `schedule_to_start` with underutilized Workers.
<!-- Sources: docs/develop/worker-performance.mdx:919-924 -->

---

## Recipe 4: Cache Tuning Workflow

Use this when you suspect cache configuration is impacting Worker performance.

### Step 1: Establish a baseline

Monitor these metrics to understand current cache behavior:

- `sticky_cache_size` -- number of Workflow Executions currently cached
- `sticky_cache_hit` -- total cache hits (Worker found Workflow in cache)
- `sticky_cache_miss` -- total cache misses (Worker must fetch Event History and replay)
- `sticky_cache_total_forced_eviction` -- total forced evictions from cache
<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:259-305; docs/cloud/worker-health.mdx:306-311 -->

### Step 2: Determine if cache is too small

**Indicators that cache is too small:**

- `sticky_cache_total_forced_eviction` is increasing at a high rate. A forced eviction means a Workflow Execution was removed from the cache before it completed because the cache was full and needed to make room.
- `sticky_cache_size` has reached your configured `WorkflowCacheSize` value.
- High `workflow_task_replay_latency` (exceeds a few milliseconds), indicating frequent replays.
- High ratio of `sticky_cache_miss` to `sticky_cache_hit`.

**Action**: If your Workers have enough free RAM and are not close to reasonable thread limits, increase `workflowCacheSize` (and `maxWorkflowThreadCount` in Java) to decrease the overall latency and cost of replays.
<!-- Sources: docs/develop/worker-performance.mdx:690-696; docs/troubleshooting/performance-bottlenecks.mdx:294-305, 107-110 -->

### Step 3: Determine if cache is too large

**Indicators that cache is too large:**

- High Worker memory usage correlating with high `sticky_cache_size`. If you observe high memory usage and high `sticky_cache_size`, you can be reasonably sure the cache is contributing to memory pressure.
- Workers are memory-bound but `sticky_cache_total_forced_eviction` is low (cache is not being forced to evict, it is just consuming too much memory).

**Action**: Reduce the cache size. Experiment with different cache sizes in a staging environment to find the optimal setting.
<!-- Sources: docs/best-practices/worker.mdx:277-287 -->

### Step 4: Java-specific cache checks

In the Java SDK only:

- When `sticky_cache_size` hits `workflowCacheSize` OR `workflow_active_thread_count` hits `maxWorkflowThreadCount`, evictions begin.
- Ensure `workflowCacheSize` is less than or equal to `maxWorkflowThreadCount`.
- Ensure `maxConcurrentWorkflowTaskExecutionSize` is less than or equal to `maxWorkflowThreadCount`. It is recommended that `maxWorkflowThreadCount` be at least 2x of `maxConcurrentWorkflowTaskExecutionSize`.

In CoreSDK-based SDKs (like TypeScript), `sticky_cache_size` should be monitored and adjusted on a per Worker and Task Queue basis.
<!-- Sources: docs/develop/worker-performance.mdx:690-704, 373-378 -->

### Step 5: Investigate high replay latency

If `workflow_task_replay_latency` is high (exceeds a few milliseconds), check for:

- Large Event Histories (consider using Continue-As-New)
- Slow Data Converters (especially those that perform encryption or interact with external services)
- Large payloads in Activities or Signals
- Complex Workflow logic or computationally intensive operations
- Worker resource constraints (high CPU or memory pressure)
<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:107-118 -->

---

## Recipe 5: Scaling Decision Framework

Use this to decide when to add Workers, remove Workers, or adjust configuration.

### Decision: When to add more Workers

**Signals that you need more Workers:**

1. `ApproximateBacklogAge` (from `temporal task-queue describe`) shows Tasks have been waiting too long. If this time grows too long, more Workers can boost efficiency.
2. `BacklogIncreaseRate` is positive, meaning the backlog is growing by approximately N Tasks per second. As this rate increases, you may need to add more Workers until demand and capacity are balanced.
3. High `schedule_to_start` latency AND high Worker CPU/memory utilization. Workers are saturated -- scale up vertically (more CPU/memory) or horizontally (more Worker instances).
<!-- Sources: docs/develop/worker-performance.mdx:856-867; docs/best-practices/worker.mdx:260-262 -->

### Decision: When to remove Workers (scale down)

**Signals that you have too many Workers:**

1. Low Poll Success Rate AND low `schedule_to_start` latency AND low Worker host resource utilization at the same time. You might have too many Workers; consider sizing down.
2. An empty backlog indicates low Worker utilization, allowing you to reduce your fleet and associated costs.
3. Low Schedule-to-Start latency with consistently low CPU and memory usage. You may be over-provisioning Workers.

Poll Success Rate formula:

```
(poll_success + poll_success_sync) / (poll_success + poll_success_sync + poll_timeouts)
```

Poll Success Rate should be >90% in most cases of systems with a steady load. For high volume and low latency, try to target >95%.
<!-- Sources: docs/develop/worker-performance.mdx:849-893; docs/cloud/worker-health.mdx:278-298; docs/best-practices/worker.mdx:269-271 -->

### Decision: When to adjust Worker configuration (not fleet size)

**Signals that individual Worker configuration needs tuning:**

1. High `schedule_to_start` latency AND low Worker host CPU/memory (Workers are underutilized). Increase executor slots or poller counts rather than adding more Workers.
2. `worker_task_slots_available` frequently at zero while hosts are underutilized. Increase `maxConcurrentWorkflowTaskExecutionSize` or `maxConcurrentActivityExecutionSize`.
3. `worker_task_slots_available` shows a significant percentage of slots available AND `schedule_to_start` is abnormally long. Increase poller count.
<!-- Sources: docs/develop/worker-performance.mdx:895-917; docs/cloud/worker-health.mdx:312-342 -->

### Scale-down safety check

Before shutting down a Worker, verify it does not have too many active Tasks. This is especially relevant for Workers handling long-running, expensive Activities.

- If `worker_task_slots_available` is at or near zero, the Worker is running active Tasks. Shutting it down could trigger expensive retries or timeouts for long-running Activities.
- Use Graceful Shutdowns to allow the Worker to complete its current Tasks before shutting down. All SDKs provide a way to configure Graceful Shutdowns (e.g., Go SDK has the `WorkerStopTimeout` option).
<!-- Sources: docs/best-practices/worker.mdx:289-299 -->

### Using Task Queue data for scaling decisions

Query Task Queue information using the CLI:

```
temporal task-queue describe \
    --task-queue YourTaskQueueName \
    [additional options]
```

Key data points:

- `ApproximateBacklogCount` -- approximate count of Tasks currently backlogged.
- `ApproximateBacklogAge` -- approximate age of the oldest Task in the backlog.
- `TasksAddRate` and `TasksDispatchRate` -- approximate Tasks per second added to or dispatched from the queue, averaged over the most recent 30 seconds.
- `BacklogIncreaseRate` -- net Tasks per second added to the backlog (`TasksAddRate - TasksDispatchRate`). Positive means growing; negative means shrinking.
- Calculate per-Worker demand: divide `ApproximateBacklogCount` by the number of Workers.

A `LastAccessTime` value exceeding one minute may indicate Workers are at capacity or have shut down. Values over 5 minutes usually suggest Workers have shut down or been removed.
<!-- Sources: docs/develop/worker-performance.mdx:795-863 -->

### Metrics to correlate for scaling decisions

No single metric tells the full story. Monitor all of these on your Worker dashboard:

- Worker CPU and memory utilization
- `workflow_task_schedule_to_start_latency` and `activity_task_schedule_to_start_latency`
- `worker_task_slots_available`
- `temporal_long_request_failure`, `temporal_request_failure`, `temporal_long_request_latency`, and `temporal_request_latency`
<!-- Sources: docs/best-practices/worker.mdx:244-255 -->
