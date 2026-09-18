# Troubleshooting Worker Performance Bottlenecks

Diagnose and resolve Worker performance bottlenecks using Temporal SDK metrics, with a problem-cause-action structure for each symptom.

## Schedule-to-Start Latency Spikes

### Workflow Task Schedule-to-Start Latency

**Metric:** `temporal_workflow_task_schedule_to_start_latency`

**Threshold:** P95 higher than one second.

This metric represents the time between when a Workflow Task is scheduled (enqueued) and when it is picked up by a Worker for processing.

**Potential causes:**

- Insufficient Worker capacity: not enough Workers or Workers are overloaded, causing Tasks to wait longer in the queue.
- Worker configuration issues: too few pollers or Task slots.
- High Workflow lock latency: many updates to a single execution can cause Workflow lock latency, which affects schedule-to-start latency. Reduce the rate of Signals.
- Network latency: Workers in a different region from the Temporal cluster, or large payload size.

**Diagnostic steps:**

1. Check Worker CPU and memory usage.
2. Review Worker configuration (number of pollers, Task slots, etc.).
3. Look for any spikes in Workflow or Activity starts that might be overwhelming the system.
4. Ensure Workers are in the same region as the Temporal cluster if possible.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:30-47 -->

### Activity Task Schedule-to-Start Latency

**Metric:** `temporal_activity_schedule_to_start_latency`

**Threshold:** P95 higher than one second.

This metric represents the time between when an Activity Task is scheduled (enqueued) and when it is picked up by a Worker for processing.

**Potential causes:**

- Insufficient Worker capacity: not enough Workers or Workers are overloaded.
- Worker configuration issues: too few pollers or Task slots.
- Task Queue configuration: setting `TaskQueueActivitiesPerSecond` too low can limit the rate at which Activities are started.
- Network latency: Workers in a different region from the Temporal cluster, or large payload size.

**Diagnostic steps:**

1. Check Worker CPU and memory usage.
2. Review Worker configuration (number of pollers, Task slots, etc.).
3. Look for any spikes in Workflow or Activity starts that might be overwhelming the system.
4. Ensure Workers are in the same region as the Temporal cluster if possible.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:49-65 -->

### Interpreting Schedule-to-Start With Other Metrics

No single metric tells the full story. Correlate schedule-to-start latency with resource metrics:

- **High schedule-to-start latency AND high CPU/memory:** Workers are saturated. Scale up your Workers or add more Workers. Workers may also be blocked on Activities.
- **High schedule-to-start latency AND low CPU/memory:** Workers are underutilized. Increase the number of pollers, executor slots, or both. If accompanied by high `temporal_long_request_latency` or `temporal_long_request_failure`, Workers are struggling to reach the Temporal Service.
- **Low schedule-to-start latency AND low CPU/memory:** Depending on your workload, this could be normal. If you are consistently seeing low memory and CPU usage, you may be over-provisioning your Workers and can consider scaling down.

<!-- Sources: docs/best-practices/worker.mdx:256-274 -->

## Workflow Task Execution Latency

**Metric:** `temporal_workflow_task_execution_latency`

This metric represents the time taken by a Worker to execute a Workflow Task. The SDK raises a "Deadlock detected during Workflow run" error or TMPRL1101 when a Workflow Task takes more than one or two seconds to complete.

**Potential causes:**

- CPU-intensive work in Workflow Task code.
- Slow local Activities (Workflow Task execution time includes Local Activity execution time).
- Slow Workflow replay (execution time includes Workflow Replay time; see replay latency below).
- Worker resource constraints: high CPU usage on Worker pods.
- Infinite loops or blocking calls in Workflow code.
- Slow data conversion: custom Data Converter taking too long to encode/decode payloads (e.g., talking to a remote encryption service).

**Diagnostic steps:**

1. Monitor Worker CPU and memory utilization.
2. Ensure Workers have adequate resources and are properly scaled for your workload.
3. Consider running Workflow code in a profiler using a replayer to see where CPU cycles are spent.
4. Review Workflow code for potential optimizations or to remove blocking operations.
5. Disable deadlock detection for Data Converter: does not reduce Task execution latency but removes the "Deadlock detected" or TMPRL1101 error. In Go, wrap with `workflow.DataConverterWithoutDeadlockDetection`. In Java, surround Data Converter code with `WorkflowUnsafe.deadlockDetectorOff`.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:84-104 -->

## Replay Latency

**Metric:** `workflow_task_replay_latency`

High if it exceeds a few milliseconds. Workflow Task replay reconstructs the Workflow's state by re-executing the Workflow code from the beginning using the recorded Event History.

**Potential causes:**

- Large Event Histories: Workflows with long histories take more time to replay.
- Data Converter performance: slow Data Converters, especially those that perform encryption or interact with external services.
- Large payloads: Activities or Signals with large payloads slow down replay, especially if the Data Converter needs to process them.
- Complex Workflow logic: computationally intensive operations such as scheduling many concurrent child Workflows or Activities.
- Frequent cache evictions: Workers often evict Workflow Executions from their cache (due to memory constraints or frequent restarts), leading to more replays and higher latency.
- Worker resource constraints: high CPU utilization or memory pressure on Worker nodes.

**Diagnostic steps:**

1. Monitor the `temporal_workflow_task_replay_latency` metric.
2. Analyze Workflow History size: check the number of events in Workflow histories and consider using Continue-As-New for long-running Workflows.
3. Optimize Data Converters if using custom ones, especially for encryption or complex serialization.
4. Review payload sizes: large Activity or Signal payloads slow down replay.
5. Profile Workflow code to identify CPU-intensive parts.
6. Manage Worker cache: frequent cache evictions lead to more replays. Tune the Worker's cache size and eviction policies.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:105-127 -->

## Activity Execution Latency

**Metric:** `temporal_activity_execution_latency`

This metric measures the time from when a Worker starts processing an Activity Task until it reports to the service that the Task is complete or failed.

**Potential causes:**

- Activity implementation: the most common cause. Time-consuming operations or slow external API calls.
- External dependencies: Activity constrained by an external resource or service that all Activities access.
- Worker resource constraints: under-resourced Worker nodes or high CPU utilization.
- Network latency: high latency between Workers and external services or the Temporal service.

**Diagnostic steps:**

1. Monitor the `activity_execution_latency` metric, which can be filtered by Activity type and Activity Task Queue.
2. Optimize Activity implementation to reduce latency, especially with external services or database interactions.
3. Check Worker CPU and memory utilization for adequate resources.
4. Examine Worker configuration, particularly `(Max)ConcurrentActivityExecutionSize` and `(Max)WorkerActivitiesPerSecond`, to ensure they are not limiting Activity execution.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:128-144 -->

## Slot Depletion

### Workflow Worker Slot Depletion

**Metric:** `temporal_worker_task_slots_available{worker_type="WorkflowWorker"}`

This metric indicates the number of available slots for executing Workflow Tasks on a Worker. It may go to zero for several reasons:

- High Workflow Task load: more Tasks than the Worker can handle concurrently (incoming rate exceeds completion rate).
- Worker configuration: the number of available slots is determined by `MaxConcurrentWorkflowTaskExecutionSize`. If set too low, the Worker may not have enough slots.
- High `temporal_workflow_task_execution_latency` and `workflow_task_replay_latency`.

**Actions:**

1. Monitor Worker CPU and memory usage while increasing `(Max)ConcurrentWorkflowTaskExecutionSize` to add more execution slots.
2. Scale up Workers both vertically (increasing CPU and memory) and horizontally (increasing Worker instances).

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:145-158 -->

### Activity Worker Slot Depletion

**Metric:** `temporal_worker_task_slots_available{worker_type="ActivityWorker"}`

This metric indicates the number of available slots for executing Activity Tasks on a Worker. It may go to zero for several reasons:

- Blocked Activities and Zombie Activities: the most common cause. Zombie Activities occur when an Activity times out (hits its `StartToClose` or `HeartbeatTimeout` timeout) and has stopped Heartbeating but continues to run, occupying slots as more retries occur. This can happen if:
  - The Activity code is blocking on a downstream service call or an infinite loop.
  - There is a mismatch between the Activity's `StartToClose` timeout and any client-side timeouts for external calls.
- Resource utilization: high CPU or memory usage on Workers can cause Activities to block and not release slots.

**Actions:**

1. Monitor Worker CPU and memory usage while increasing `(Max)ConcurrentActivityExecutionSize` to add more execution slots.
2. Add client-side timeout to your downstream API client.
3. Review Task code to ensure Tasks complete within a reasonable time measured by `temporal_activity_execution_latency`.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:159-174 -->

## Network Request Failures

### High `temporal_long_request_failure`

This metric counts failed RPC long poll requests for `PollWorkflowTaskQueue`, `PollActivityTaskQueue`, and `GetWorkflowExecutionHistory` (when polling new events).

**Potential causes:**

- Network issues: problems with the connection between Client and Server, including firewalls and proxies.
- Rate limiting: request rate exceeds configured limits, causing rejections. Often indicated by a `ResourceExhausted` status code.
- Server errors: Temporal Server experiencing issues responding to long poll requests.

**Diagnostic steps:**

1. Check the operation and the status or code tag of the `temporal_long_request_failure` metric to see the type of errors.
2. If you receive a `ResourceExhausted` status code, review rate limits or contact Temporal Support for Temporal Cloud.
3. Check the network connection between Client and Server.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:176-194 -->

### High `temporal_request_failure_total`

This metric counts the number of RPC requests made by the Temporal Client that have failed.

**Potential causes:**

- Network issues: problems with the connection between Client and Server.
- Client errors: misconfiguration or resource exhaustion on the Client.
- Operation errors: operations like `SignalWorkflowExecution` or `TerminateWorkflowExecution` can fail if acting on a closed Workflow Execution that no longer exists (completed and removed from persistence at Namespace retention time).
- Rate limiting: indicated by a `ResourceExhausted` status code.
- Request size limit: Worker tries to return an Activity response larger than the blob size limit (2MB), causing the service to reject it.
- Server errors.

**Diagnostic steps:**

1. Check the status or code tag to see the type of errors.
2. Look at the operation tag to see which operations are failing.
3. Monitor Temporal Server and Client logs for error messages or warnings.
4. Check the network connection between Client and Server.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:195-213 -->

### High `temporal_request_latency`

This metric measures the latency of gRPC requests made by the Temporal Client.

**Potential causes:**

- Network latency: physical distance and network conditions between Client and Server.
- Network transfer time: larger payloads take longer to transfer. For example, large payloads in `RespondWorkflowTaskCompleted` affect latency, especially when Workflows schedule multiple Activities with large inputs.
- Resource exhaustion: running out of CPU or memory on the client or server.
- Client configuration: improper configuration such as setting thread pool sizes too aggressively or memory constraints too low for allocated threads, causing Tasks to overwhelm the client.
- Server load: Temporal Server under heavy load.

**Diagnostic steps:**

1. Monitor `temporal_request_latency` to identify when and where latency spikes occur.
2. Check the network connection between Client and Server.
3. Monitor resource usage on both Client and Server.
4. Review Client configuration for workload optimization.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:214-232 -->

## Caching Issues

### Sticky Cache Size

**Metric:** `temporal_sticky_cache_size`

Represents the number of Workflow Executions currently cached in a Worker's memory. The sticky cache keeps Workflow state in memory, reducing the need to reconstruct from Event History for every Task.

There is a direct relationship between sticky cache size and Worker memory consumption. As the cache size increases, so does memory usage.

Monitor this metric alongside Worker memory usage. A sudden increase in `sticky_cache_size` can correlate with increased memory consumption and potential performance issues.

- If memory consumption is too high, reduce the maximum sticky cache size.
- If you have available memory and want to improve performance, increase it.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:259-278 -->

### Cache Hit and Miss Rates

**Metrics:** `temporal_sticky_cache_hit_total` and `temporal_sticky_cache_miss_total`

- A "hit" means the Worker finds the Workflow in its cache when processing a Workflow Task, allowing immediate processing without fetching the full Event History from the server and Replaying.
- A "miss" means the Worker did not find the Workflow in its cache and must fetch the Event History and Replay.

A high rate of cache hits with a low rate of cache misses indicates Workflows are being scheduled efficiently, with minimal need for fetching Event Histories and Replaying.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:280-293 -->

### Forced Evictions

**Metric:** `temporal_sticky_cache_total_forced_eviction_total`

A forced eviction means a Workflow Execution was removed from the cache before it completed, typically because the cache was full and needed to make room for other Workflow Executions. If the Worker needs to process more Tasks for the evicted Workflow, it must fetch the entire Event History from the Temporal Service and Replay.

A high rate of forced evictions could indicate that your cache size is too small for your workload. You may need to increase the `WorkflowCacheSize` setting if Worker resources can accommodate it.

<!-- Sources: docs/troubleshooting/performance-bottlenecks.mdx:295-306 -->

### Optimizing Worker Cache

Workers keep a cache of Workflow Executions to reduce replay overhead. The `temporal_sticky_cache_size` tracks the size of the cache. If you observe high memory usage for Workers and high `temporal_sticky_cache_size`, the cache is likely contributing to memory pressure.

Having a high `temporal_sticky_cache_size` by itself is not necessarily an issue, but if Workers are memory-bound, consider reducing the cache size to allow more concurrent executions. Experiment with different cache sizes in a staging environment to find the optimal setting for your Workflows.

<!-- Sources: docs/best-practices/worker.mdx:277-287 -->

## Workflow Cache Tuning

When the number of cached Workflow Executions reported by `sticky_cache_size` hits `workflowCacheSize` or the number of threads reported by the `workflow_active_thread_count` metrics gauge hits `maxWorkflowThreadCount`, Workflow Executions will start to be evicted from the cache. An evicted Workflow Execution will need to be replayed when it gets any action that may advance it.

If the cache limits are hit and Worker hosts have enough free RAM and are not close to reasonable thread limits, you may choose to increase `workflowCacheSize` and `maxWorkflowThreadCount` limits to decrease the overall latency and cost of Replays.

If the opposite occurs (hosts are resource-constrained), consider decreasing the limits.

**Note:** `maxWorkflowThreadCount` and `workflow_active_thread_count` are for the Java SDK only. In CoreSDK-based SDKs like TypeScript, this metric works differently and should be monitored and adjusted on a per Worker and Task Queue basis.

<!-- Sources: docs/develop/worker-performance.mdx:690-704 -->

## Task Queue Processing Tuning

The following steps limit delays in Task Queue processing due to insufficient or unbalanced Workers. Review these if you notice high `schedule_to_start` metrics. The steps are in recommended order of execution.

<!-- Sources: docs/develop/worker-performance.mdx:868-873 -->

### Step 1: Hosts and Resources Provisioning

If currently provisioned Worker hosts are fully utilized (near full CPU usage, high load average, etc.), additional Worker hosts must be provisioned to increase the capacity of the Workers pool.

**It is possible to have too many Workers.** Monitor the poll success (`poll_success`/`poll_success_sync`) and poll timeout (`poll_timeouts`) Server metric counters.

Poll Success Rate = (`poll_success` + `poll_success_sync`) / (`poll_success` + `poll_success_sync` + `poll_timeouts`)

Poll Success Rate should be >90% in most cases of systems with a steady load. For high volume and low latency, try to target >95%.

If you see all three of the following at the same time:

1. Low Poll Success Rate, and
2. Low `schedule_to_start_latency`, and
3. Low Worker hosts resource utilization

then you might have too many Workers. Consider sizing down.

<!-- Sources: docs/develop/worker-performance.mdx:875-894 -->

### Step 2: Worker Executor Slots Sizing

The main area to focus on when tuning is the number of Worker Executor Slots. Increase the maximum number of working slots by adjusting `maxConcurrentWorkflowTaskExecutionSize` or `maxConcurrentActivityExecutionSize` if both of the following conditions are met:

1. The Worker hosts are underutilized (no bottlenecks on CPU, load average, etc.).
2. The `worker_task_slots_available` metric from the corresponding Worker type frequently shows a depleted number of available Worker slots.

Alternatively, consider using a resource-based slot supplier.

<!-- Sources: docs/develop/worker-performance.mdx:896-904 -->

### Step 3: Poller Count

Sometimes it can be appropriate to increase the number of Task pollers. This is usually more common when Workers have somewhat high latency communicating with the server. You can use automated poller tuning to handle this automatically.

Consider manual adjustment if:

1. The Worker hosts are underutilized (no bottlenecks on CPU, load average, etc.).
2. `worker_task_slots_available` metric from the corresponding Worker type shows that a significant percentage of Worker slots are available on a regular basis.
3. The `schedule_to_start` metric is abnormally long.

Then consider increasing the number of pollers by adjusting `maxConcurrentWorkflowTaskPollers` or `maxConcurrentActivityTaskPollers`, depending on which type of `schedule_to_start` metric is elevated.

<!-- Sources: docs/develop/worker-performance.mdx:906-918 -->

### Step 4: Rate Limiting

If, after adjusting the poller and executor counts, you still observe an elevated `schedule_to_start`, underutilized Worker hosts, or high `worker_task_slots_available`, check the following:

- If server-side rate limiting per Task Queue is set by `WorkerOptions#maxTaskQueueActivitiesPerSecond`, remove the limit or adjust the value up.
- If Worker-side rate limiting per Worker is set by `WorkerOptions#maxWorkerActivitiesPerSecond`, remove the limit.

<!-- Sources: docs/develop/worker-performance.mdx:919-925 -->

## Safe Scale-Down

Before shutting down a Worker, verify that it does not have too many active Tasks. This is especially relevant if Workers are handling long-running, expensive Activities.

If `worker_task_slots_available` is at or near zero, the Worker is running active Tasks. Shutting it down could trigger expensive retries or timeouts for long-running Activities. Use Graceful Shutdowns to allow the Worker to complete its current Tasks before shutting down. All SDKs provide a way to configure Graceful Shutdowns.

<!-- Sources: docs/best-practices/worker.mdx:289-299 -->

## Recommended Monitoring Dashboard

Keep all of the following metrics on your Worker monitoring dashboard. When you observe anomalies, correlate across multiple metrics to identify root causes:

- Worker CPU and memory utilization
- `workflow_task_schedule_to_start_latency` and `activity_task_schedule_to_start_latency`
- `worker_task_slots_available`
- `temporal_long_request_failure`, `temporal_request_failure`, `temporal_long_request_latency`, and `temporal_request_latency`

<!-- Sources: docs/best-practices/worker.mdx:246-254 -->
