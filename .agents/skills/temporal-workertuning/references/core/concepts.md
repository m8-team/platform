# Worker Tuning Core Concepts

Core concepts for configuring Temporal Worker performance through task slots, slot suppliers, worker tuners, and eager task execution.

## Task Slots

A **Worker Task Slot** represents the capacity of a Temporal Worker to execute a single concurrent Task.
When a Worker starts processing a Task, it occupies one slot.
The number of available slots directly affects how many Tasks a Worker can handle simultaneously. <!-- Sources: docs/develop/worker-performance.mdx:33-38 -->

## Slot Types

There are four slot types:

| Slot Type | Purpose |
|---|---|
| Workflow | Executes Workflow Tasks |
| Activity | Executes Activity Tasks |
| Local Activity | Executes Local Activity Tasks |
| Nexus | Executes Nexus Operations |

<!-- Sources: docs/develop/worker-performance.mdx:45 -->

## Slot Suppliers

A **Slot Supplier** defines a strategy to provide slots for a Worker, increasing or decreasing the Worker's slot count.
The supplier determines when it is acceptable to begin a new Task.
Each supplier manages one slot type.
An available slot determines whether or not a Worker is willing to poll for, and execute, a new Task of that type. <!-- Sources: docs/develop/worker-performance.mdx:42-46 -->

### Fixed-Size Slot Suppliers

Hands out slots up to a preset limit.
Useful when you have a concrete idea of how many resources your Tasks will consume and can determine an upper bound on how many should run at once.
For the absolute best performance, review your hardware and environment characteristics, then calculate an appropriate fixed-size limit.
Evaluate the maximum number of slots you can support without oversubscribing or hitting out-of-memory conditions. <!-- Sources: docs/develop/worker-performance.mdx:52-58 -->

### Resource-Based Slot Suppliers

Hands out slots based on real-time CPU and memory usage.
You set target utilization for both CPU and memory, and the Slot Supplier tries to reach those values without exceeding them under load.
A resource-based supplier will account for memory limits imposed in containerized environments.
It dynamically adjusts the number of available slots for different task types with respect to current system resources.

When running in a containerized environment, all SDKs use cgroups for both CPU and memory. CPU is accounted for at the container level.

> **Warning:** You cannot guarantee that the targets for resource-based suppliers will never be exceeded. Resources consumed during a Task cannot be known ahead of time.

<!-- Sources: docs/develop/worker-performance.mdx:60-70, 79-80 -->

### Custom Slot Suppliers

Hands out slots based on custom logic that you define.
Use this approach when you need complete control over when Workers accept and execute Tasks.
Custom suppliers provide flexibility to optimize for specific use cases that fixed-size and resource-based suppliers may not fully address. <!-- Sources: docs/develop/worker-performance.mdx:72-75, 422-424 -->

## Slot Permits

Slot Suppliers issue `SlotPermit`s.
These represent the right to use a slot of a specific type (Workflow, Activity, Local Activity, or Nexus).
You control whether a Worker can perform certain Tasks by issuing or withholding permits. <!-- Sources: docs/develop/worker-performance.mdx:444-446 -->

## Custom Slot Supplier Interface

Custom Slot Suppliers must implement these functions:

| Function | Description |
|---|---|
| `reserveSlot` | Called before polling for new tasks. Your implementation can block and must return a Slot Permit once it decides to accept new work. |
| `tryReserveSlot` | Called for slot reservations in cases like eager activity processing. This must not block. |
| `markSlotUsed` | Called when a slot is about to be used for a task (not while it is held during polling). It provides information about the task. |
| `releaseSlot` | Called when a slot is no longer needed, whether or not it was used. |

SDK-specific interface/class names:

| Language | Type Name |
|---|---|
| Go | `SlotSupplier` |
| Java | `SlotSupplier` |
| Python | `CustomSlotSupplier` |
| TypeScript | `CustomSlotSupplier` |
| .NET | `CustomSlotSupplier` |

<!-- Sources: docs/develop/worker-performance.mdx:436-455 -->

## Worker Tuners

Worker tuning is the process of defining customized slot suppliers for the different task slots of a Worker to fine-tune its performance.
A **Worker Tuner** instance exists per-Worker, providing slot suppliers for different slot types (Activity, Workflow, Nexus, or Local Activity Tasks).
A tuner assigns different suppliers to each slot type.
For example, it might provide a fixed-size slot supplier for Workflows and use a resource-based supplier for Activities.

A **composite tuner** lets you mix different slot supplier strategies for each Task type.
For example, you can use fixed-size slot suppliers for Workflow and Nexus Tasks while using resource-based slot suppliers for Activity and Local Activity Tasks.

> **Warning:** Worker tuners supersede the existing `maxConcurrentXXXTask` style Worker options. Using both styles will cause an error at Worker initialization time.

<!-- Sources: docs/develop/worker-performance.mdx:89-100, 383-386, 512-513 -->

## Choosing a Slot Supplier Type

| Consideration | Recommendation |
|---|---|
| Workflow Tasks (minimal CPU, low memory) | Fixed-size slot suppliers |
| Very low latency and maximum throughput required | Avoid resource-based auto-tuning; use fixed-size |
| Workloads with resource usage patterns you do not fully understand | Resource-based for good-enough performance without profiling |
| Tasks with variable or very high per-task resource needs | Fixed-size suppliers with manual tuning |
| New workload where resource usage is not yet understood | Resource-based to avoid profiling upfront; switch to fixed-size once usage patterns are clear |
| Fluctuating workloads with low per-task consumption (e.g., HTTP calls, blocking I/O) | Resource-based |
| Protection from out-of-memory with unpredictable per-task consumption | Resource-based (may still sometimes exceed requested limits) |
| Complete control over slot allocation | Custom slot suppliers |

The resource-based tuner can never perform as well as a fixed-size tuner with appropriately chosen configuration, but can offer reasonable performance without the need for profiling.

<!-- Sources: docs/develop/worker-performance.mdx:388-424 -->

## Slot Throttling (rampThrottle)

Slot throttling is a mechanism to control the rate at which new slots for concurrent tasks are made available for processing.
This concept is part of the resource-based auto-tuning feature for Workers.
By waiting a brief period between making slots available, the Worker can assess how resource usage has changed since the last task began processing.

This throttle is called `rampThrottle` in the SDK options for resource-based slot suppliers.
It defines the minimum time the Worker will wait between handing out new slots after passing the minimum slots number.

**A higher `rampThrottle` trades off performance for safety.**

For example, if a just-started Worker were to have no throttle and there was a backlog of Tasks, it might immediately accept 100 Tasks at once.
If each Task allocated 1GB of RAM, the Worker would likely run out of memory and crash.
The throttle enforces a wait before handing out new slots (after a minimum number of slots have been occupied) so you can measure newly consumed resources. <!-- Sources: docs/develop/worker-performance.mdx:458-478 -->

## Eager Task Execution

As a latency optimization, Activity and Workflow Tasks may be started eagerly in a local Worker under the right circumstances.

> **Warning:** Eager start does not respect Worker versioning. An eagerly started Workflow may run on any available local Worker even if that Worker is not the Current or Ramping version of its Worker deployment.

### Eager Activity Start

Eager Activity Start may happen automatically if the Worker processing a Workflow Task has also registered the Activity Definition being called.
If it does, it may try to reserve an Activity Slot for the execution of the Activity, and the server may respond to the Workflow Task completion with the Activity Task for the Worker to execute immediately. <!-- Sources: docs/develop/worker-performance.mdx:118-123 -->

### Eager Workflow Start

Eager Workflow Start reduces the latency required to initiate a Workflow Execution.
It is recommended for short-lived Workflows that use Local Activities to interact with external services, especially when these interactions are initiated in the first Workflow Task and the Workflow is deployed near the Temporal Server to minimize network delay.

Eager Workflow Start is available in Public Preview in the Go, Java, Python, and .NET SDKs.
It is enabled for all Temporal Cloud users and self-hosted Temporal Server 1.29.0+.
You must set `request_eager_start` (or similar name) to true when starting each Workflow for Eager Workflow Start to be used.

**Requirements:**

- The Starter and the Worker must share a Client located in the same process.
- `request_eager_start` (or similar name) must be set to true in the Start Workflow call.
- The Worker must have a Workflow Task slot available and the Workflow Definition registered.

When set, the Worker can execute the first task of the Workflow locally without first making a round-trip to the Temporal Server.
This is typically most useful in combination with a Local Activity executing in the first Workflow Task, since other Workflow API calls that require waiting on something will force a round-trip.

**How it works:**

1. The Starter sets `request_eager_start` to true in the Start Workflow Options.
2. The SDK tries to locate a local Worker that is willing to execute the first Workflow Task, and reserves an execution slot for it.
3. If successful, the SDK provides a hint to the server that eager mode is preferred for the new Workflow.
4. The server registers the start of the Workflow in history and assigns the first Workflow Task to the Starter, all in the same DB update.
5. The first task is included in the server response; no matching step is required.
6. The SDK extracts the task from the response and dispatches it to the local Worker.

To recover from errors, Eager Workflow Start falls back to the non-eager path.
For example, when the first Task is returned eagerly but the local Worker fails or times out while processing the task, the server retries this task non-eagerly after WorkflowTaskTimeout. <!-- Sources: docs/develop/worker-performance.mdx:112-177 -->
