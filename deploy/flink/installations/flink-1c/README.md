# Flink installation: flink-1c

## Responsibility

Owns the isolated Flink runtime and 1C CDC SQL pipelines in namespace `flink-1c`.

## Owns

- `FlinkDeployment/flink-1c`
- `Service/flink-1c-rest`, created by the operator
- `Deployment` and `Service` named `flink-1c-sql-gateway`
- namespace-scoped ServiceAccount, Role, and RoleBinding
- Kafka brokers, topics, principal, JAAS Secret, and SQL connector settings
- object-storage credentials and the `flink-1c` state prefix
- resource sizing, slots, environment overlays, and business SQL

## Does not own

- the shared runtime image build
- the shared Flink Kubernetes Operator
- any other Flink installation

## State isolation

```text
s3://bucket-flink-production-k8s-zenden-cloud/flink-1c/ha
s3://bucket-flink-production-k8s-zenden-cloud/flink-1c/checkpoints
s3://bucket-flink-production-k8s-zenden-cloud/flink-1c/savepoints
```

Never reuse these prefixes for another installation.

## SQL layout

SQL is submitted in this order using the same SQL Gateway session:

1. `sql/sources/*.sql`
2. `sql/views/*.sql`
3. `sql/sinks/*.sql`
4. `sql/jobs/*.sql`

`sql/kustomization.yaml` packages these non-secret files into a hash-named
ConfigMap mounted at `/opt/flink/installation-sql`. Argo CD owns that artifact and
rolls this installation's Gateway when SQL changes; job submission is still an
explicit REST action.

The price-types source uses an upsert changelog keyed by `_IDRRef`. Value fields
exclude the primary key during decoding, so create/update/delete identity comes
from the Kafka key even when a delete record has a null value.

The normalized view preserves opaque 1C reference and numeric encodings as
strings. It decodes the explicitly documented `AA==` and `AQ==` boolean values.
Changing any remaining opaque field to a numeric type requires a separately
validated 1C binary-decoding contract.
