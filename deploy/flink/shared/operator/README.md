# Shared Flink Kubernetes Operator

This wrapper pins the upstream Apache Flink Kubernetes Operator chart to 1.15.0.
There is one release in namespace `flink-operator` for all installations.

`watchNamespaces: []` is intentional: in upstream chart 1.15.0 it means watch all
namespaces. Restricting this list would require updating the operator every time a
new installation namespace is added. The resulting operator ClusterRole is the
minimum scope compatible with that requirement; runtime job RBAC remains local to
each installation.
