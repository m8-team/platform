# Worker Tuner and Poller Autoscaling SDK Examples

SDK code examples for configuring Worker tuners (resource-based and composite) and poller autoscaling.

---

## Poller Autoscaling

Poller Autoscaling automatically selects an appropriate number of pollers based on need. It results in more efficient poller usage, better throughput, and schedule-to-start latency improvements. Enable it by setting the `*_task_poller_behavior` options to `PollerBehaviorAutoscaling` (names vary slightly by SDK). Poller Autoscaling will be the default configuration in future versions of Temporal SDKs.

> **Requirement:** `PollerBehaviorAutoscaling` is only enabled in Temporal Server v1.28.0 and later.

<!-- Sources: docs/develop/worker-performance.mdx:231-241 -->

### Go SDK

[Go SDK docs](https://pkg.go.dev/go.temporal.io/sdk/worker#PollerBehaviorAutoscalingOptions)

```go
w := worker.New(c, "my-task-queue", worker.Options{
  WorkflowTaskPollerBehavior: worker.NewPollerBehaviorAutoscaling(worker.PollerBehaviorAutoscalingOptions{}),
  ActivityTaskPollerBehavior: worker.NewPollerBehaviorAutoscaling(worker.PollerBehaviorAutoscalingOptions{}),
  NexusTaskPollerBehavior: worker.NewPollerBehaviorAutoscaling(worker.PollerBehaviorAutoscalingOptions{}),
})
```

<!-- Sources: docs/develop/worker-performance.mdx:255-263 -->

### Java SDK

[Java SDK docs](https://javadoc.io/doc/io.temporal/temporal-sdk/latest/io/temporal/worker/tuning/PollerBehaviorAutoscaling.html)

```java
public class WorkerExample {
    public static void main(String[] args) {
        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();
        WorkflowClient client = WorkflowClient.newInstance(service);
        WorkerFactory factory = WorkerFactory.newInstance(client);
        WorkerOptions workerOptions = WorkerOptions.newBuilder()
            .setWorkflowTaskPollersBehavior(new PollerBehaviorAutoscaling())
            .setActivityTaskPollersBehavior(new PollerBehaviorAutoscaling())
            .setNexusTaskPollersBehavior(new PollerBehaviorAutoscaling())
            .build();

        Worker worker = factory.newWorker("my-task-queue", workerOptions);
    }
}
```

<!-- Sources: docs/develop/worker-performance.mdx:265-282 -->

### Python SDK

[Python SDK docs](https://python.temporal.io/temporalio.worker.PollerBehaviorAutoscaling.html)

```python
worker = Worker(
    client,
    task_queue="my-task-queue",
    workflows=[MyWorkflow],
    activities=[my_activity],
    
    workflow_task_poller_behavior=PollerBehaviorAutoscaling(),
    activity_task_poller_behavior=PollerBehaviorAutoscaling(),
    nexus_task_poller_behavior=PollerBehaviorAutoscaling(),
)
```

<!-- Sources: docs/develop/worker-performance.mdx:284-297 -->

### TypeScript SDK

[TypeScript SDK docs](https://typescript.temporal.io/api/interfaces/proto.temporal.api.sdk.v1.WorkerConfig.IAutoscalingPollerBehavior)

```ts
const worker = await Worker.create({
  connection,
  taskQueue: 'my-task-queue',
  workflowsPath: require.resolve('./workflows'),
  activities,
  
  workflowTaskPollerBehavior: PollerBehavior.autoscaling(),
  activityTaskPollerBehavior: PollerBehavior.autoscaling(),
  nexusTaskPollerBehavior: PollerBehavior.autoscaling(),
});
```

<!-- Sources: docs/develop/worker-performance.mdx:299-312 -->

### .NET SDK

[DotNet SDK docs](https://dotnet.temporal.io/api/Temporalio.Worker.Tuning.PollerBehavior.Autoscaling.html)

```csharp
using var worker = new TemporalWorker(
    client,
    new TemporalWorkerOptions("my-task-queue")
    {
        WorkflowTaskPollerBehavior = new PollerBehavior.Autoscaling(),
        ActivityTaskPollerBehavior = new PollerBehavior.Autoscaling(),
        NexusTaskPollerBehavior = new PollerBehavior.Autoscaling(),
    }
    .AddWorkflow<MyWorkflow>()
    .AddActivity(MyActivities.MyActivity)
);
```

<!-- Sources: docs/develop/worker-performance.mdx:314-328 -->

### Ruby SDK

[Ruby SDK docs](https://ruby.temporal.io/Temporalio/Worker/PollerBehavior/Autoscaling.html)

```ruby
worker = Temporalio::Worker.new(
  client,
  'my-task-queue',
  workflows: [MyWorkflow],
  activities: [MyActivity],
  
  workflow_task_poller_behavior: Temporalio::Worker::PollerBehavior::Autoscaling.new,
  activity_task_poller_behavior: Temporalio::Worker::PollerBehavior::Autoscaling.new,
  nexus_task_poller_behavior: Temporalio::Worker::PollerBehavior::Autoscaling.new,
)
```

<!-- Sources: docs/develop/worker-performance.mdx:330-344 -->

---

## Resource-Based Tuners

Each tuner provides slot suppliers for various Task types. The following examples focus on Activities and Local Activities, since Workflow Tasks normally do not need resource-based tuning.

<!-- Sources: docs/develop/worker-performance.mdx:479-484 -->

### Go SDK

[features/snippets/worker_tuner/worker_tuner.go](https://github.com/temporalio/features/blob/main/features/snippets/worker_tuner/worker_tuner.go)

```go
func resourceBasedTuner() (worker.Options, error) {
	tuner, err := worker.NewResourceBasedTuner(worker.ResourceBasedTunerOptions{
		TargetMem:    0.8,
		TargetCpu:    0.9,
		InfoSupplier: sysinfo.SysInfoProvider(),
	})
	if err != nil {
		return worker.Options{}, err
	}
	return worker.Options{
		Tuner: tuner,
	}, nil
}
```

<!-- Sources: docs/develop/worker-performance.mdx:488-507 -->

### Java SDK

```java
// Just resource based
WorkerOptions.newBuilder()
    .setWorkerTuner(
        ResourceBasedTuner.newBuilder()
            .setControllerOptions(
                ResourceBasedControllerOptions.newBuilder(0.8, 0.9).build())
            .build())
    .build())
```

<!-- Sources: docs/develop/worker-performance.mdx:561-569 -->

### TypeScript SDK

```tsx
// Just resource based
const resourceBasedTunerOptions: ResourceBasedTunerOptions = {
  targetMemoryUsage: 0.8,
  targetCpuUsage: 0.9,
};
const workerOptions = {
  tuner: {
    tunerOptions: resourceBasedTunerOptions,
  },
};
```

<!-- Sources: docs/develop/worker-performance.mdx:592-602 -->

### Python SDK

```python
# Just a resource based tuner, with poller autoscaling
tuner = WorkerTuner.create_resource_based(
    target_memory_usage=0.5,
    target_cpu_usage=0.5,
)
worker = Worker(
    client,
    task_queue="foo",
    tuner=tuner,
    workflow_task_poller_behavior=PollerBehaviorAutoscaling(),
    activity_task_poller_behavior=PollerBehaviorAutoscaling()
)
```

<!-- Sources: docs/develop/worker-performance.mdx:628-640 -->

### .NET C# SDK

```csharp
// Just resource based
var worker = new TemporalWorker(
    Client,
    new TemporalWorkerOptions("my-task-queue")
    {
        Tuner = WorkerTuner.CreateResourceBased(0.8, 0.9),
    });
```

<!-- Sources: docs/develop/worker-performance.mdx:665-672 -->

---

## Composite Tuners

A composite tuner lets you mix different slot supplier strategies for each Task type. For example, you can use fixed-size slot suppliers for Workflow and Nexus Tasks while using resource-based slot suppliers for Activity and Local Activity Tasks.

<!-- Sources: docs/develop/worker-performance.mdx:510-513 -->

### Go SDK

[features/snippets/worker_tuner/worker_tuner.go](https://github.com/temporalio/features/blob/main/features/snippets/worker_tuner/worker_tuner.go)

```go
func compositeTuner() (worker.Options, error) {
	options := worker.DefaultResourceControllerOptions()
	options.MemTargetPercent = 0.8
	options.CpuTargetPercent = 0.9
	options.InfoSupplier = sysinfo.SysInfoProvider()
	controller := worker.NewResourceController(options)
	wfSS, err := worker.NewFixedSizeSlotSupplier(10)
	if err != nil {
		return worker.Options{}, err
	}

	actSS, err := worker.NewResourceBasedSlotSupplier(controller, worker.DefaultActivityResourceBasedSlotSupplierOptions())
	if err != nil {
		return worker.Options{}, err
	}
	laSS, err := worker.NewResourceBasedSlotSupplier(controller, worker.DefaultActivityResourceBasedSlotSupplierOptions())
	if err != nil {
		return worker.Options{}, err
	}
	nexusSS, err := worker.NewFixedSizeSlotSupplier(10)
	if err != nil {
		return worker.Options{}, err
	}

	compositeTuner, err := worker.NewCompositeTuner(worker.CompositeTunerOptions{
		WorkflowSlotSupplier:      wfSS,
		ActivitySlotSupplier:      actSS,
		LocalActivitySlotSupplier: laSS,
		NexusSlotSupplier:         nexusSS,
	})
	if err != nil {
		return worker.Options{}, err
	}
	return worker.Options{
		Tuner: compositeTuner,
	}, nil
}
```

<!-- Sources: docs/develop/worker-performance.mdx:515-557 -->

### Java SDK

```java
// Combining different types
SlotSupplier<WorkflowSlotInfo> workflowTaskSlotSupplier = new FixedSizeSlotSupplier<>(10);
SlotSupplier<ActivitySlotInfo> activityTaskSlotSupplier =
    ResourceBasedSlotSupplier.createForActivity(
        resourceController, ResourceBasedTuner.DEFAULT_ACTIVITY_SLOT_OPTIONS);
SlotSupplier<LocalActivitySlotInfo> localActivitySlotSupplier =
    ResourceBasedSlotSupplier.createForLocalActivity(
        resourceController, ResourceBasedTuner.DEFAULT_ACTIVITY_SLOT_OPTIONS);
SlotSupplier<NexusSlotInfo> nexusSlotSupplier = new FixedSizeSlotSupplier<>(10);

WorkerOptions.newBuilder()
    .setWorkerTuner(
        new CompositeTuner(
            workflowTaskSlotSupplier,
            activityTaskSlotSupplier,
            localActivitySlotSupplier,
            nexusSlotSupplier))
    .build();
```

<!-- Sources: docs/develop/worker-performance.mdx:570-588 -->

### TypeScript SDK

```tsx
// Combining different types
const resourceBasedTunerOptions: ResourceBasedTunerOptions = {
  targetMemoryUsage: 0.8,
  targetCpuUsage: 0.9,
};
const workerOptions = {
  tuner: {
    activityTaskSlotSupplier: {
      type: 'resource-based',
      tunerOptions: resourceBasedTunerOptions,
    },
    workflowTaskSlotSupplier: {
      type: 'fixed-size',
      numSlots: 10,
    },
    localActivityTaskSlotSupplier: {
      type: 'resource-based',
      tunerOptions: resourceBasedTunerOptions,
    },
  },
};
```

<!-- Sources: docs/develop/worker-performance.mdx:603-623 -->

### Python SDK

```python
# Combining different types, with poller autoscaling
resource_based_options = ResourceBasedTunerConfig(0.8, 0.9)
tuner = WorkerTuner.create_composite(
    workflow_supplier=FixedSizeSlotSupplier(10),
    activity_supplier=ResourceBasedSlotSupplier(
        ResourceBasedSlotConfig(),
        resource_based_options,
    ),
    local_activity_supplier=ResourceBasedSlotSupplier(
        ResourceBasedSlotConfig(),
        resource_based_options,
    ),
)
worker = Worker(
    client,
    task_queue="foo",
    tuner=tuner,
    workflow_task_poller_behavior=PollerBehaviorAutoscaling(),
    activity_task_poller_behavior=PollerBehaviorAutoscaling()
)
```

<!-- Sources: docs/develop/worker-performance.mdx:641-660 -->

### .NET C# SDK

```csharp
// Combining different types
var resourceTunerOptions = new ResourceBasedTunerOptions(0.8, 0.9);
var worker = new TemporalWorker(
    Client,
    new TemporalWorkerOptions("my-task-queue")
    {
        Tuner = new WorkerTuner(
             new FixedSizeSlotSupplier(10),
             new ResourceBasedSlotSupplier(
                 new ResourceBasedSlotSupplierOptions(),
                 resourceTunerOptions),
             new ResourceBasedSlotSupplier(
                 new ResourceBasedSlotSupplierOptions(),
                 resourceTunerOptions)),
    });
```

<!-- Sources: docs/develop/worker-performance.mdx:673-688 -->
