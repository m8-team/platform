# Optional Flink Runtime

This component defines the optional Apache Flink runtime for platform features that need stream or batch processing.

Flink is managed by Apache Flink Kubernetes Operator. The operator is installed by `gitops/optional/flink/application.yaml`, then this component creates a session cluster through `FlinkDeployment`.

It is intentionally optional and is not referenced by the default root `ApplicationSet`.

Enable it through:

```bash
kubectl apply -f gitops/optional/flink/application.yaml
```

The default manifest creates a `flink.apache.org/v1beta1` `FlinkDeployment` session cluster in `m8-data`.

Production notes:

- Mirror `ghcr.io/apache/flink-kubernetes-operator:1.15.0` and `flink:2.2.0` into the private registry and pin both by digest before production use.
- Replace local filesystem checkpoint/savepoint paths with object storage.
- Enable the operator admission webhook only after wiring non-default certificate and keystore secret management.
- Operator `1.15.0` uses `jobManager.resource` and `taskManager.resource`; do not use the newer `resources` fields until the installed CRD supports them.
- Keep Flink application jobs in separate GitOps paths as `FlinkSessionJob` resources; this component owns only the runtime session cluster.
