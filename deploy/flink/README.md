# Apache Flink GitOps

This directory owns the Kubernetes and Argo CD delivery model for Apache Flink.
It deliberately separates shared infrastructure from independently operated
installations.

```text
deploy/flink/
├── shared/
│   ├── image/                  # one generic Flink 2.2.1 runtime image
│   └── operator/               # one operator chart, version 1.15.0
├── installations/
│   └── flink-1c/               # namespace-scoped runtime and business SQL
├── argocd/
│   ├── operator.application.yaml
│   └── installations/
└── scripts/
```

## Ownership model

The shared image contains Flink, Java, the Kafka SQL connector, the built-in
Presto S3 filesystem plugin, and Yandex CA certificates. It contains no brokers,
topics, usernames, passwords, SQL, manifests, or environment-specific settings.

The shared operator watches all namespaces. The upstream chart's exact value is
`watchNamespaces`; `[]` means all namespaces. The operator's cluster-scoped RBAC
is therefore required. Job identities and permissions are disabled in the chart
and owned by each installation through namespace-scoped RBAC.

An installation owns its `FlinkDeployment`, SQL Gateway, Kafka configuration,
secret references, sizing, state paths, SQL, environment overlays, and Argo CD
Application. Adding an installation never adds another operator.

Installation SQL is rendered as a non-secret, hash-named ConfigMap and mounted
read-only at `/opt/flink/installation-sql` in that installation's Gateway. Argo CD
therefore tracks SQL changes and rolls only the affected Gateway. Submitting or
replacing a streaming job remains an explicit SQL Gateway REST operation.

## Runtime image lifecycle

Build the generic image locally:

```bash
make flink:image-build FLINK_RUNTIME_TAG=2.2.1-local
```

GitLab CI publishes immutable revision tags. It reports the registry digest but
does not edit Git or deploy. Promote a release by changing only `newName` and
`digest` in:

```text
installations/<name>/base/kustomization.yaml
```

Both the Flink cluster and its SQL Gateway use the symbolic `flink-runtime`
image, so each installation has exactly one repository-and-digest pin. The
existing generic registry repository is retained for an evolutionary rollout;
it contains no installation name. The initial `flink-1c` digest deliberately
preserves the previous revision until the first reduced runtime image is
published and explicitly promoted.

## Secrets

Examples are excluded from Kustomize. Never commit their populated copies.
For a manual bootstrap, create the namespace and apply the two local files before
Argo CD starts the workloads:

```bash
kubectl create namespace flink-1c --dry-run=client -o yaml | kubectl apply -f -

cp deploy/flink/installations/flink-1c/base/kafka-secret.example.yaml \
  deploy/flink/installations/flink-1c/base/kafka-secret.yaml
cp deploy/flink/installations/flink-1c/base/object-storage-secret.example.yaml \
  deploy/flink/installations/flink-1c/base/object-storage-secret.yaml

# Fill placeholders locally, then:
kubectl apply -f deploy/flink/installations/flink-1c/base/kafka-secret.yaml
kubectl apply -f deploy/flink/installations/flink-1c/base/object-storage-secret.yaml
```

For production, manage Secrets with the platform's external secret controller or
encrypted Git workflow while keeping the Secret names and keys defined by the
examples.

## Argo CD bootstrap

The repository must be pushed before these Applications are created. Bootstrap
in dependency order:

```bash
kubectl apply -f deploy/flink/argocd/project.yaml
kubectl apply -f deploy/flink/argocd/operator.application.yaml

kubectl -n argocd wait application/flink-operator \
  --for=jsonpath='{.status.health.status}'=Healthy --timeout=10m

kubectl apply -f deploy/flink/argocd/installations/flink-1c.application.yaml
```

The Applications use automated prune and self-heal. The operator Application is
wave `-10`; installation Applications start at wave `-5`. Inside an installation,
RBAC is wave `-1`, the cluster is wave `0`, and SQL Gateway is wave `1`.

## Validate and operate `flink-1c`

```bash
make flink:validate

deploy/flink/scripts/check-flink.sh flink-1c
deploy/flink/scripts/check-kafka.sh flink-1c
deploy/flink/scripts/check-sql-gateway.sh flink-1c
```

For interactive REST access:

```bash
kubectl -n flink-1c port-forward svc/flink-1c-sql-gateway 8083:8083
```

In another terminal, submit all SQL files in deterministic dependency order:

```bash
deploy/flink/scripts/submit-sql.sh \
  http://127.0.0.1:8083 \
  deploy/flink/installations/flink-1c/sql
```

The first pipeline consumes the compacted 1C price-types topic by Kafka key and
writes normalized changelog rows to `flink-1c.price-types.normalized.v1`. Provision
that output topic and grant the installation Kafka principal write permission
before submitting the job.

## Add another independent installation

```bash
deploy/flink/scripts/create-installation.sh flink-pricing
```

The command copies only infrastructure skeleton files, creates empty SQL
directories, assigns the new namespace/name/state prefix, creates an Argo CD
Application, and registers it in the Argo Kustomization. It does not copy 1C SQL.

Before committing, define the new installation's image digest, sizing, slots,
Kafka brokers/principal/topics, object-storage prefix, SQL, and secret delivery.
Run `make flink:validate`, commit, and let Argo CD reconcile it independently.

## CI/CD boundary

`.gitlab-ci.yml` validates Helm/Kustomize/Kubernetes manifests, reviews the SQL
contract, checks shell scripts, builds the image, and publishes it from protected
branches. It contains no production cluster mutation command. Argo CD is the only
production reconciler.
