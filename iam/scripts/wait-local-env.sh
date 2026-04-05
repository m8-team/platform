#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/deploy/local/docker-compose.yaml"

wait_http() {
  local name="$1"
  local url="$2"
  local attempts="${3:-60}"
  local delay="${4:-2}"

  for ((i=1; i<=attempts; i++)); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      echo "${name} is ready: ${url}"
      return 0
    fi
    sleep "${delay}"
  done

  echo "${name} did not become ready: ${url}" >&2
  return 1
}

wait_cmd() {
  local name="$1"
  shift

  for ((i=1; i<=60; i++)); do
    if "$@" >/dev/null 2>&1; then
      echo "${name} is ready"
      return 0
    fi
    sleep 2
  done

  echo "${name} did not become ready" >&2
  return 1
}

wait_http "YDB UI" "http://127.0.0.1:8765"
wait_cmd "Redis" docker compose -f "${COMPOSE_FILE}" exec -T redis redis-cli ping
wait_http "Keycloak realm" "http://127.0.0.1:8081/realms/m8/.well-known/openid-configuration"
wait_http "SpiceDB metrics" "http://127.0.0.1:9090/metrics"
wait_cmd "Temporal cluster" docker compose -f "${COMPOSE_FILE}" exec -T temporal temporal operator cluster health --address 127.0.0.1:7233

cat <<'EOF'

Local IAM dependency stack is ready.

Endpoints:
- YDB UI: http://127.0.0.1:8765
- Keycloak: http://127.0.0.1:8081
- SpiceDB gRPC: 127.0.0.1:50051
- Temporal UI: http://127.0.0.1:8233
- Redis: 127.0.0.1:6379

Credentials:
- Keycloak admin: admin / admin
- Keycloak test user in realm m8: test-admin / admin
- SpiceDB preshared key: dev-spicedb-key

Next:
- make run-local
- make worker-local
EOF
