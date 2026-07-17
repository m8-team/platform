# Optional Flink Applications

Apply these Argo CD Applications after bootstrap when Flink is required:

```bash
kubectl apply -f gitops/optional/flink/application.yaml
```

The manifest creates:

- `m8-flink-operator`: installs Apache Flink Kubernetes Operator from the Apache Helm chart.
- `m8-flink`: reconciles `gitops/components/system/flink` into namespace `m8-data` as a `FlinkDeployment`.

The runtime application uses `SkipDryRunOnMissingResource=true` because the `FlinkDeployment` CRD is provided by the operator application.
