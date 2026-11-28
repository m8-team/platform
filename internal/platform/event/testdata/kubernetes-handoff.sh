#!/usr/bin/env bash
# Creates a separate kind cluster and YDB for a disposable handoff test.
set -euo pipefail
umask 077
scratch=$(mktemp -d)
cluster=m8-event-handoff
fixture=m8-event-handoff-ydb
export KUBECONFIG="$scratch/kubeconfig"
cleanup() {
  if [[ ${created_ydb:-0} == 1 ]]; then docker rm -f "$fixture" >/dev/null; fi
  if [[ ${created_kind:-0} == 1 ]]; then kind delete cluster --name "$cluster" >/dev/null; fi
  rm -rf "$scratch"
}
trap cleanup EXIT
if kind get clusters | grep -Fxq "$cluster" || docker inspect "$fixture" >/dev/null 2>&1; then
  echo "The disposable cluster or YDB name is already in use" >&2
  exit 1
fi
cat >"$scratch/kind.yaml" <<'YAML'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
YAML
created_kind=1
kind create cluster --name "$cluster" --image kindest/node:v1.34.0 --config "$scratch/kind.yaml" --wait 120s
created_ydb=1
docker run -d --network kind --hostname localhost --name "$fixture" \
  -p 127.0.0.1:32536:32536 -e GRPC_PORT=32536 ydbplatform/local-ydb:26.3.1.25 >/dev/null
healthy=0
for _ in {1..45}; do
  if docker exec "$fixture" /health_check >/dev/null 2>&1; then healthy=1; break; fi
  sleep 1
done
[[ "$healthy" == 1 ]]
ip=$(docker inspect -f '{{(index .NetworkSettings.Networks "kind").IPAddress}}' "$fixture")
arch=$(docker exec "$cluster-worker" uname -m)
case "$arch" in aarch64) arch=arm64 ;; x86_64) arch=amd64 ;; *) exit 1 ;; esac
mkdir -p "$scratch/image"
CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go test -c -o "$scratch/image/event.test" ./internal/platform/event
printf '%s\n' 'FROM scratch' 'COPY event.test /event.test' 'ENTRYPOINT ["/event.test","-test.v","-test.run=^TestKubernetesReplicaWorker$"]' >"$scratch/image/Dockerfile"
docker build -t m8-event-handoff-worker:fixture "$scratch/image"
kind load docker-image --name "$cluster" m8-event-handoff-worker:fixture
kubectl create namespace event-handoff
kubectl -n event-handoff apply --dry-run=server -f deploy/event/reconciler.yaml -f deploy/event/publisher.example.yaml >/dev/null
kubectl -n event-handoff apply -f - <<YAML
apiVersion: v1
kind: ServiceAccount
metadata:
  name: handoff-reader
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: handoff-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  resourceNames: ["publisher-0"]
  verbs: ["get"]
- apiGroups: ["apps"]
  resources: ["statefulsets"]
  resourceNames: ["publisher"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: handoff-reader
subjects:
- kind: ServiceAccount
  name: handoff-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: handoff-reader
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: publisher
spec:
  serviceName: publisher
  replicas: 0
  selector:
    matchLabels:
      app: publisher
  template:
    metadata:
      labels:
        app: publisher
    spec:
      automountServiceAccountToken: false
      hostNetwork: true
      containers:
      - name: publisher
        image: m8-event-handoff-worker:fixture
        imagePullPolicy: Never
        readinessProbe:
          tcpSocket:
            port: 18091
            host: 127.0.0.1
        volumeMounts:
        - name: publisher-ydb
          mountPath: /var/run/publisher-ydb
          readOnly: true
        env:
        - name: M8_EVENT_WORKER
          value: "1"
        - name: M8_EVENT_WORKER_YDB_IP
          value: "$ip"
        - name: POD_UID
          valueFrom:
            fieldRef:
              fieldPath: metadata.uid
      volumes:
      - name: publisher-ydb
        secret:
          secretName: publisher-ydb
YAML
kubectl -n event-handoff create token handoff-reader >"$scratch/reader-token"
kubectl config view --flatten --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d >"$scratch/ca.pem"
kubectl config view -o jsonpath='{.clusters[0].cluster.server}' >"$scratch/server"
M8_EVENT_KUBERNETES_FIXTURE="$scratch" M8_EVENT_TEST_YDB_REQUIRED=1 \
  M8_EVENT_TEST_YDB_DSN=grpc://127.0.0.1:32536/local M8_EVENT_TEST_YDB_PREFIX=/local \
  go test -count=1 -timeout=5m -v -run '^TestIntegrationKubernetesReplicaHandoff$' ./internal/platform/event
