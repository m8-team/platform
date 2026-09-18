# Metrics and Monitoring for Worker Performance Tuning

Metrics used to identify Worker bottlenecks, evaluate Task Queue health, and guide tuning decisions.

## Metric naming convention

SDK metrics are documented in Temporal references without a prefix for readability.
At runtime, they use the `temporal_` prefix.
For example, `worker_task_slots_available` in the docs corresponds to `temporal_worker_task_slots_available` at query time.

Cloud metrics use their full name including prefix (e.g., `temporal_cloud_v1_poll_success_count`).

<!-- Sources: docs/develop/worker-performance.mdx:22-24, docs/cloud/worker-health.mdx:104-106 -->

## Compute-related metrics

| Worker configuration option | SDK metric |
|-----------------------------|------------|
| `MaxConcurrentWorkflowTaskExecutionSize` | `worker_task_slots_available {worker_type = WorkflowWorker}` |
| `MaxConcurrentActivityTaskExecutionSize` | `worker_task_slots_available {worker_type = ActivityWorker}` |
| `MaxWorkflowThreadCount` | `workflow_active_thread_count` (Java only) |
| CPU-intensive logic | `workflow_task_execution_latency` |

Also monitor your machine's CPU consumption (for example, `container_cpu_usage_seconds_total` in Kubernetes).

<!-- Sources: docs/develop/worker-tuning-reference.mdx:140-149 -->

### Slot availability metrics

The `worker_task_slots_available` and `worker_task_slots_used` gauges report the number of available executor "slots" that are currently available and unoccupied for a Worker type.
Tag these with `worker_type=WorkflowWorker` for Workflow Task Workers or `worker_type=ActivityWorker` for Activity Workers.

**Important**: Unlike `worker_task_slots_used`, `worker_task_slots_available` can only be used with fixed-size slot suppliers and cannot be used with resource-based slot suppliers.

The `temporal_worker_task_slots_available` metric should always be >0. If it frequently reaches 0, you are likely experiencing a Task backlog.

<!-- Sources: docs/develop/worker-performance.mdx:193-202, docs/cloud/worker-health.mdx:308-319 -->

### Latency metrics

Two schedule-to-start latency timers:

- `workflow_task_schedule_to_start_latency` -- for Workflow Tasks
- `activity_schedule_to_start_latency` -- for Activities

A Schedule-To-Start latency is the time from when a Task is scheduled (placed in a Queue) to when a Worker starts (picks up from the Task Queue) that Task.
These metrics help ensure that Tasks are being processed from the queue in a timely manner.

This latency should be very low, close to zero. Any higher value indicates a bottleneck.

<!-- Sources: docs/develop/worker-performance.mdx:204-209, docs/cloud/worker-health.mdx:110-113, docs/cloud/worker-health.mdx:150-153 -->

## Memory-related metrics

| Worker configuration option | SDK metric |
|-----------------------------|------------|
| `StickyWorkflowCacheSize` | `sticky_cache_total_forced_eviction`, `sticky_cache_size`, `sticky_cache_hit`, `sticky_cache_miss` |

Also monitor your machine's memory consumption (for example, `container_memory_usage_bytes` in Kubernetes).

### Cache metrics

The `sticky_cache_size` and `workflow_active_thread_count` metrics report the size of the Workflow cache and the number of cached Workflow threads.

**Target**: `sticky_cache_size` should report less than or equal to your `WorkflowCacheSize` value.
Also, `sticky_cache_total_forced_eviction` should not be reporting high numbers (relative).

**Note**: `sticky_cache_total_forced_eviction` is available in the Go SDK and the Java SDK only.

**Action**: If you see a high eviction count, verify there are no other inefficiencies in your Worker configuration or resource provisioning (backlog).
If you see the cache size metric exceed the `WorkflowCacheSize`, increase this value if your Worker resources can accommodate it or provision more Workers.

<!-- Sources: docs/develop/worker-tuning-reference.mdx:152-157, docs/develop/worker-performance.mdx:211-213, docs/cloud/worker-health.mdx:77-87, docs/cloud/worker-health.mdx:344-358 -->

## IO-related metrics

| Worker configuration option | SDK metric |
|-----------------------------|------------|
| `MaxConcurrentWorkflowTaskPollers` | `num_pollers {poller_type = workflow_task}` |
| `MaxConcurrentActivityTaskPollers` | `num_pollers {poller_type = activity_task}` |
| Network latency | `request_latency {namespace, operation}` |

<!-- Sources: docs/develop/worker-tuning-reference.mdx:159-165 -->

## Task Queue metrics

### Cloud and SDK metrics for Task Queues

| Metric | Description |
|--------|-------------|
| `poll_success_sync_count` (Cloud) | Sync match rate (Tasks immediately assigned to Workers) |
| `approximate_backlog_count` (Cloud) | Approximate number of Tasks in a Task Queue |

<!-- Sources: docs/develop/worker-tuning-reference.mdx:167-173 -->

### DescribeTaskQueue statistics

Task Queue statistics are available via the `DescribeTaskQueueEnhanced` method in the Go SDK, with the Temporal CLI `task-queue describe` command, and using `DescribeTaskQueue` through RPC.

The Temporal Service reports information separately for each Task Queue type (not aggregated).

| Statistic | Description |
|-----------|-------------|
| `ApproximateBacklogCount` | Approximate count of Tasks currently backlogged in this Task Queue. May include expired Tasks as well as active Tasks, but will eventually converge to the correct count over time. |
| `ApproximateBacklogAge` | Approximate age of the oldest Task in the backlog, based on the creation time of the Task at the head of the queue. |
| `TasksAddRate` | Approximate Tasks-per-second added to the Task Queue. Averaged over the most recent 30-second interval. Includes sync-matched Tasks. |
| `TasksDispatchRate` | Approximate Tasks-per-second dispatched from the Task Queue. Averaged over the most recent 30-second interval. Includes sync-matched Tasks. |
| `BacklogIncreaseRate` | Net Tasks per second added to the backlog, averaged over the most recent 30 seconds. Calculated as `TasksAddRate - TasksDispatchRate`. Positive = backlog growing; negative = backlog shrinking. |

You can rely on `ApproximateBacklogCount` and `ApproximateBacklogAge` when making scaling decisions.

<!-- Sources: docs/develop/worker-performance.mdx:706-720, docs/develop/worker-performance.mdx:722-771 -->

### Known accuracy limitations for backlog counts

- **Overcount from invalid or expired Tasks**: Tasks belonging to cancelled, terminated, completed, or timed out Workflows and Activities stay in the count until they reach the head of the queue and are processed and discarded.
- **Reset to zero on idle Task Queue unload**: If a Task Queue sees no activity for approximately 5 minutes (no Worker polls, no new Tasks added, and no other Task Queue calls), the Temporal Service unloads it from memory. `ApproximateBacklogCount` reports zero until the Task Queue is reloaded.
- **Sticky queue exclusion**: Sticky queues are not included in these values. Because Sticky queue Tasks only remain valid for a few seconds, this inaccuracy diminishes as the backlog grows.

The actual Task delivery count may be significantly higher than `TasksAddRate`/`TasksDispatchRate` report:

- Eager dispatch: Activities can be requested by an SDK using one Workflow Task completion response. Tasks using Eager dispatch do not pass through Task Queues.
- Tasks passed to Sticky Task Queues are not included in the returned values for `TasksAddRate` and `TasksDispatchRate`.

While individual add and dispatch rates may be inaccurate due to Eager and Sticky Task Queues, the `BacklogIncreaseRate` reliably reflects the rate at which the backlog is shrinking or growing for backlogs older than a few seconds.

<!-- Sources: docs/develop/worker-performance.mdx:732-771 -->

### Querying Task Queue info with Temporal CLI

```
temporal task-queue describe \
    --task-queue YourTaskQueueName \
    [additional options]
```

This command retrieves poller information, backlog statistics, and task reachability for Task types (available in Temporal Server v1.25.0, Temporal CLI 1.1 and later).

<!-- Sources: docs/develop/worker-performance.mdx:795-806 -->

### Querying Task Queue info with the Go SDK

Retrieve Task Queue data using the Go SDK by calling `DescribeTaskQueueEnhanced`.
Specify the Task Queue name and set `ReportStats` to `true`:

```go
for _, taskQueueName := range taskQueueNames {
        resp, err := s.client.DescribeTaskQueueEnhanced(ctx, client.DescribeTaskQueueEnhancedOptions{
            TaskQueue:   taskQueueName,
            ReportStats: true,
        })
        if err != nil {
            log.Printf("Error describing task queue %s: %v", taskQueueName, err)
        }

        // Get the backlog count from the enhanced response
        backlogCount += getBacklogCount(resp)
    }
```

<!-- Sources: docs/develop/worker-performance.mdx:816-833 -->

### Evaluating Worker availability

Each Temporal Server records the last time of each poll request, displayed in `temporal task-queue describe` output.

- A `LastAccessTime` value exceeding one minute may indicate that the Worker fleet is at capacity or that Workers have shut down or been removed.
- Values under 5 minutes typically suggest the Worker fleet is at capacity (all Workflow and Activity slots are full).
- Values over 5 minutes since the last poll request usually suggest that Workers have shut down or been removed. Workers are removed if 5 minutes have passed since the last poll request.

<!-- Sources: docs/develop/worker-performance.mdx:836-848 -->

### Managing Worker fleet with Task Queue data

- `ApproximateBacklogAge` shows how long Tasks have been waiting to be dispatched. If this time grows too long, more Workers can boost Workflow efficiency.
- Calculate the demand per Worker by dividing the number of backlogged Tasks (`ApproximateBacklogCount`) by the number of Workers.
- Determine if your task processing rate is within an acceptable range using the per-Worker demand, the backlog consumption rate (`TasksDispatchRate`), and the dispatch latency (`ApproximateBacklogAge`).

A large backlog of Tasks with too few Workers will slow down Workflow Execution completions and decrease processing efficiency. Adding more Workers speeds up completion rates and improves throughput. An empty backlog indicates low Worker utilization, allowing you to reduce your fleet and associated costs.

<!-- Sources: docs/develop/worker-performance.mdx:849-863 -->

## Failure metrics

| Metric | Description |
|--------|-------------|
| `long_request_failure` | Failures for long-running operations (polling, history retrieval) |
| `request_failure` | Failures for standard operations (Task completion responses) |

Common failure codes:

- `RESOURCE_EXHAUSTED` -- Rate limits exceeded
- `DEADLINE_EXCEEDED` -- Operation timeout
- `NOT_FOUND` -- Resource not found

<!-- Sources: docs/develop/worker-tuning-reference.mdx:183-193 -->

## Worker health monitoring patterns

### Minimal observations (alert thresholds)

Configure these alerts first for baseline application health intelligence:

1. **Schedule To Start latency** (SDK metrics for both Workflow and Activity Executions):
   - Alert at >200ms for your p99 value
   - Plot >100ms for your p95 value

2. **Sync Match Rate** (Grafana panel):
   - Alert at <95% for your p99 value
   - Plot <99% for your p95 value

3. **Poll Success Rate** (Grafana panel):
   - Alert at <90% for your p99 value
   - Plot <95% for your p95 value

4. **`temporal_worker_task_slots_available`** (SDK metric):
   - Alert at 0 for your p99 value

5. **`temporal_sticky_cache_size`** (SDK metric):
   - Plot at {value} > {WorkflowCacheSize.Value}

6. **`temporal_sticky_cache_total_forced_eviction`** (SDK metric, Go and Java only):
   - Alert at >{predetermined_high_number}

7. **`temporal_cloud_v1_approximate_backlog_count`** (Cloud metric):
   - Alert when the value is growing over time for a given Task Queue

<!-- Sources: docs/cloud/worker-health.mdx:48-92 -->

### Detect Task backlog

**Symptoms**: Tasks are waiting to find Workers to run on, causing a delay in Workflow execution. Detected by watching Schedule To Start latency, sync match rate, and approximate backlog count.

**Metrics to monitor**:

- SDK metric: `workflow_task_schedule_to_start_latency`
- SDK metric: `activity_schedule_to_start_latency`
- Cloud metric: `temporal_cloud_v1_poll_success_count`
- Cloud metric: `temporal_cloud_v1_poll_success_sync_count`
- Cloud metric: `temporal_cloud_v1_approximate_backlog_count`

If your Schedule To Start latency alert triggers or is high, check the Sync Match Rate to decide if you need to adjust your Worker or fleet.

#### Prometheus query samples: Schedule To Start latency

**Workflow Task Latency, 99th percentile**

```promql
histogram_quantile(0.99, sum(rate(temporal_workflow_task_schedule_to_start_latency_seconds_bucket[5m])) by (le, namespace, task_queue))
```

**Workflow Task Latency, average**

```promql
sum(increase(temporal_workflow_task_schedule_to_start_latency_seconds_sum[5m])) by (namespace, task_queue)
/
sum(increase(temporal_workflow_task_schedule_to_start_latency_seconds_count[5m])) by (namespace, task_queue)
```

**Activity Task Latency, 99th percentile**

```promql
histogram_quantile(0.99, sum(rate(temporal_activity_schedule_to_start_latency_seconds_bucket[5m])) by (le, namespace, task_queue))
```

**Activity Task Latency, average**

```promql
sum(increase(temporal_activity_schedule_to_start_latency_seconds_sum[5m])) by (namespace, task_queue)
/
sum(increase(temporal_activity_schedule_to_start_latency_seconds_count[5m])) by (namespace, task_queue)
```

<!-- Sources: docs/cloud/worker-health.mdx:93-153 -->

### Sync Match Rate {#sync-match-rate}

The sync match rate measures the rate of Tasks delivered to Workers without having to be persisted (Workers are up and available to pick them up) to the rate of all delivered Tasks.

A sync match is when a Task is immediately matched to a Worker via the Sticky Queue.
An async match is when a Task cannot be matched to the Sticky Queue for a Worker (no Worker has cached the Workflow, or the Task times out during processing). The Task then returns to the general Task Queue.

**Calculate Sync Match Rate**:

```
temporal_cloud_v1_poll_success_sync_count / temporal_cloud_v1_poll_success_count = N
```

#### Prometheus query sample: Sync Match Rate

```promql
sum by(temporal_namespace) (
    temporal_cloud_v1_poll_success_sync_count{temporal_namespace=~"$namespace"}
)
/
sum by(temporal_namespace) (
    temporal_cloud_v1_poll_success_count{temporal_namespace=~"$namespace"}
)
```

**Target**: The Sync Match Rate should be at least >95%, but preferably >99%.

<!-- Sources: docs/cloud/worker-health.mdx:155-185 -->

### Handling Task backlog issues

#### High Schedule To Start latency and high sync match rate

Three typical causes:

- There are not enough Workers to perform work
- Each Worker is either under-resourced or misconfigured to handle enough work
- There is congestion caused by the environment (e.g., network) hosting the Worker(s) and Temporal Cloud

Actions:

- Increase the number of available Workers
- Verify that your Worker hosts are appropriately resourced
- Increase the Worker configuration value for concurrent pollers for Workers/task executions (if your Worker resources can accommodate the increased load)

#### High Schedule To Start latency and low sync match rate

Verify that you have not set a value for `ScheduleToStartTimeout` in your Activity Options, as this may skew your observations.

It may be acceptable for your use case to have low sync match rate (e.g., known workloads or intentional throttling).

**Successful async polls**:

```
temporal_cloud_v1_poll_success_count - temporal_cloud_v1_poll_success_sync_count = N
```

```promql
sum by(temporal_namespace, task_type) (
    temporal_cloud_v1_poll_success_count{temporal_namespace=~"$namespace"}
)
-
sum by(temporal_namespace, task_type) (
    temporal_cloud_v1_poll_success_sync_count{temporal_namespace=~"$namespace"}
)
```

Monitor approximate backlog count to observe Task Queue depth directly:

```promql
temporal_cloud_v1_approximate_backlog_count{temporal_namespace=~"$namespace", temporal_task_queue=~"$task_queue"}
```

Actions:

- Check the system CPU usage against `task_slots` and adjust `maxConcurrentWorkflowTaskExecutionSize` and `maxConcurrentActivityExecutionSize` settings as necessary.
- Check the system memory usage against `sticky_cache_size` and adjust sticky cache size as necessary.
- Increase the Worker config for concurrent pollers for Workflow or Activity `task_slots`, if your Worker resources can accommodate the increased load.
- Increase the number of available Workers.

<!-- Sources: docs/cloud/worker-health.mdx:187-252 -->

### Detect greedy Worker resources {#detect-greedy-workers}

You can have too many Workers. If the Poll Success Rate shows low numbers, you might have too many resources polling Temporal Cloud.

**Metrics to monitor**:

- Cloud metric: `temporal_cloud_v1_poll_success_count`
- Cloud metric: `temporal_cloud_v1_poll_success_sync_count`
- Cloud metric: `temporal_cloud_v1_poll_timeout_count`
- SDK metric: `workflow_task_schedule_to_start_latency`
- SDK metric: `activity_schedule_to_start_latency`

**Calculate Poll Success Rate**:

```
(temporal_cloud_v1_poll_success_count)
/
(temporal_cloud_v1_poll_success_count + temporal_cloud_v1_poll_timeout_count)
```

**Target**: Poll Success Rate should be >90% in most cases of systems with a steady load. For high volume and low latency, try to target >95%.

**Interpretation**: If you see all of the following at the same time, you might have too many Workers:

- Low poll success rate
- Low Schedule To Start latency
- Low Worker host resource utilization

**Actions**:

- Reduce the number of Workers polling the impacted Task Queue, OR
- Reduce the concurrent pollers per Worker, OR
- Both of the above

<!-- Sources: docs/cloud/worker-health.mdx:254-299 -->

### Detect misconfigured Workers {#detect-misconfigured-workers}

Worker configuration can negatively affect Task processing efficiency.

**Metrics to monitor**:

- SDK metric: `worker_task_slots_available`
- SDK metric: `sticky_cache_size`
- SDK metric: `sticky_cache_total_forced_eviction`

The `maxConcurrentWorkflowTaskExecutionSize` and `maxConcurrentActivityExecutionSize` define the number of total available slots for the Worker. If set too low, the Worker will not be able to keep up processing Tasks.

**Target**: The `temporal_worker_task_slots_available` metric should always be >0.

#### Prometheus query samples: slot availability

**Over Time**

```promql
avg_over_time(temporal_worker_task_slots_available{namespace="$namespace",worker_type="WorkflowWorker"}[10m])
```

**Current Time**

```promql
temporal_worker_task_slots_available{namespace="default", worker_type="WorkflowWorker", task_queue="$task_queue_name"}
```

**Action**: Increase the `maxConcurrentWorkflowTaskExecutionSize` and `maxConcurrentActivityExecutionSize` values and keep an eye on your Worker resource metrics (CPU utilization, etc) to make sure you haven't created a new issue.

<!-- Sources: docs/cloud/worker-health.mdx:300-343 -->
