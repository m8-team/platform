#!/usr/bin/env bash
# Deploy the actual reconciler to an isolated kind cluster with a disposable TLS YDB.
set -euo pipefail
umask 077
scratch=$(mktemp -d)
cluster=m8-event-v12-reconciler
db_container=m8-event-v12-secure-ydb
namespace=m8-events
image=m8-event-v12-reconciler:smoke
created_cluster=0
created_db=0
export KUBECONFIG="$scratch/kubeconfig"
cleanup() {
  result=$?
  if (( result != 0 && created_cluster == 1 )); then
    kubectl -n "$namespace" get pods >&2 || true
    kubectl -n "$namespace" logs deployment/event-reconciler --tail=12 >&2 || true
    kubectl -n "$namespace" logs deployment/event-reconciler --previous --tail=12 >&2 || true
  fi
  if (( created_db == 1 )); then docker rm -f "$db_container" >/dev/null; fi
  if (( created_cluster == 1 )); then kind delete cluster --name "$cluster" >/dev/null; fi
  rm -rf "$scratch"
  exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT TERM
if kind get clusters | grep -Fxq "$cluster" || docker inspect "$db_container" >/dev/null 2>&1; then
  echo "disposable resource already exists" >&2
  exit 1
fi
kind create cluster --name "$cluster" --image kindest/node:v1.34.0 --wait 90s
created_cluster=1
mkdir -p "$scratch/certs" "$scratch/context"
openssl req -new -newkey rsa:2048 -nodes -x509 -days 2 -subj "/CN=$db_container" \
  -addext "subjectAltName=DNS:$db_container" \
  -keyout "$scratch/certs/key.pem" -out "$scratch/certs/cert.pem" > /dev/null 2>&1
cp "$scratch/certs/cert.pem" "$scratch/certs/ca.pem"
chmod 0644 "$scratch/certs/ca.pem" "$scratch/certs/cert.pem"
docker run -d --platform linux/amd64 --network kind --hostname "$db_container" --name "$db_container" \
  --mount "type=bind,source=$scratch/certs,target=/ydb_certs,readonly" \
  ydbplatform/local-ydb:26.3.1.25 >/dev/null
created_db=1
healthy=0
for _ in {1..45}; do
  if docker exec "$db_container" /health_check >/dev/null 2>&1; then healthy=1; break; fi
  sleep 1
done
[[ "$healthy" == 1 ]]
ip=$(docker inspect -f '{{(index .NetworkSettings.Networks "kind").IPAddress}}' "$db_container")
insecure_sql() {
  docker exec -i "$db_container" /ydb --endpoint grpc://localhost:2136 --database /local --no-discovery sql -f - "$@"
}
insecure_sql < deploy/event/schema.yql
pwd="TestAa1$(openssl rand -hex 18)"
printf "CREATE USER eventsreconciler PASSWORD '%s';\nGRANT CONNECT ON \`/local\` TO eventsreconciler;\nGRANT SELECT, UPDATE ROW ON \`/local/replica_owners\` TO eventsreconciler;\n" "$pwd" |
  insecure_sql >"$scratch/user.log" 2>&1
printf '%s' "$pwd" | docker exec -i "$db_container" sh -c 'umask 077; cat > /tmp/ownership-password'
unset pwd
docker exec "$db_container" /ydb -e "grpcs://$db_container:2135" -d /local --ca-file /ydb_certs/ca.pem \
  --user eventsreconciler --password-file /tmp/ownership-password auth get-token --force >"$scratch/token" 2>"$scratch/token.err"
docker exec "$db_container" rm -f /tmp/ownership-password
[[ -s "$scratch/token" ]]

arch=$(docker exec "$cluster-control-plane" uname -m)
case "$arch" in aarch64) arch=arm64 ;; x86_64) arch=amd64 ;; *) exit 1 ;; esac
CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -o "$scratch/context/event-reconciler" ./cmd/event-reconciler
cp "$scratch/certs/ca.pem" "$scratch/context/ca-certificates.crt"
docker build -f deploy/event/Dockerfile -t "$image" "$scratch/context"
kind load docker-image --name "$cluster" "$image"
kubectl create namespace "$namespace"
kubectl -n "$namespace" create secret generic event-ownership-writer --from-file=token="$scratch/token"
rm -f "$scratch/token"
pause_tag=$(docker exec "$cluster-control-plane" crictl images | awk '/registry.k8s.io\/pause / {print $2; exit}')
[[ -n "$pause_tag" ]]
cat > "$scratch/publisher.yaml" <<YAML
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: publisher
spec:
  serviceName: publisher
  replicas: 1
  selector:
    matchLabels:
      app: publisher
  template:
    metadata:
      labels:
        app: publisher
    spec:
      automountServiceAccountToken: false
      containers:
        - name: publisher
          image: registry.k8s.io/pause:$pause_tag
          imagePullPolicy: Never
YAML
export M8_SMOKE_IMAGE="$image" M8_SMOKE_YDB_HOST="$db_container" M8_SMOKE_YDB_IP="$ip"
render() {
  M8_SMOKE_SET_UID="$1" M8_SMOKE_PREVIOUS_UID="${2:-}" ruby -ryaml -rjson -e '
    docs = YAML.load_stream(File.read("deploy/event/reconciler.yaml"))
    config = docs.find { |item| item["kind"] == "ConfigMap" }
    data = JSON.parse(config["data"]["replicas.json"])
    data["uid"] = ENV.fetch("M8_SMOKE_SET_UID")
    data["predecessors"] = [{ "uid" => ENV["M8_SMOKE_PREVIOUS_UID"], "slots" => 2 }] unless ENV["M8_SMOKE_PREVIOUS_UID"].empty?
    config["data"]["replicas.json"] = JSON.generate(data)
    deploy = docs.find { |item| item["kind"] == "Deployment" }
    spec = deploy["spec"]["template"]["spec"]
    spec["hostAliases"] = [{ "ip" => ENV.fetch("M8_SMOKE_YDB_IP"), "hostnames" => [ENV.fetch("M8_SMOKE_YDB_HOST")] }]
    container = spec["containers"].first
    container["image"] = ENV.fetch("M8_SMOKE_IMAGE")
    container["imagePullPolicy"] = "Never"
    container["env"].find { |item| item["name"] == "YDB_DSN" }["value"] = "grpcs://#{ENV.fetch("M8_SMOKE_YDB_HOST")}:2135/local"
    puts YAML.dump_stream(*docs)
  ' | kubectl -n "$namespace" apply -f -
}
set_uid() {
  kubectl -n "$namespace" get statefulset publisher -o jsonpath='{.metadata.uid}'
}
pod_uid() {
  kubectl -n "$namespace" wait --for=create pod/publisher-0 --timeout=20s >&2
  kubectl -n "$namespace" get pod publisher-0 -o jsonpath='{.metadata.uid}'
}
expect_owner() {
  local controller=$1 uid=$2 generation=$3 found
  for _ in {1..35}; do
    found=$(printf "SELECT CAST(COUNT(*) AS Uint64) AS matches FROM replica_owners WHERE replica = '%s/%s/publisher-0' AND pod_uid = '%s' AND generation = %su;\n" \
      "$namespace" "$controller" "$uid" "$generation" | insecure_sql --format json-unicode)
    if printf '%s\n' "$found" | grep -Eq '"matches"[[:space:]]*:[[:space:]]*"?1"?'; then
      echo "ownership observed for incarnation $controller at generation $generation"
      return 0
    fi
    sleep 1
  done
  echo "expected owner transition was not observed" >&2
  return 1
}
kubectl -n "$namespace" apply -f "$scratch/publisher.yaml"
first_set=$(set_uid)
first_pod=$(pod_uid)
render "$first_set"
kubectl -n "$namespace" rollout status deployment/event-reconciler --timeout=35s
expect_owner "$first_set" "$first_pod" 1
kubectl -n "$namespace" delete statefulset publisher --wait=true --timeout=25s
kubectl -n "$namespace" wait --for=delete pod/publisher-0 --timeout=25s
expect_owner "$first_set" retired 2
kubectl -n "$namespace" apply -f "$scratch/publisher.yaml"
second_set=$(set_uid)
second_pod=$(pod_uid)
[[ "$first_set" != "$second_set" && "$first_pod" != "$second_pod" ]]
render "$second_set" "$first_set"
kubectl -n "$namespace" rollout restart deployment/event-reconciler
kubectl -n "$namespace" rollout status deployment/event-reconciler --timeout=35s
expect_owner "$second_set" "$second_pod" 1
expect_owner "$first_set" retired 2
echo "Production reconciler smoke with Kubernetes RBAC, TLS, dedicated YDB token and recreation passed"
