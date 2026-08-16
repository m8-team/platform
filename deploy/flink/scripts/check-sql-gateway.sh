#!/usr/bin/env bash

set -euo pipefail

namespace=${1:-flink-1c}
local_port=${2:-8083}
gateway_name=${namespace}-sql-gateway

for required_command in kubectl curl; do
  if ! command -v "${required_command}" >/dev/null 2>&1; then
    echo "required command not found: ${required_command}" >&2
    exit 1
  fi
done

kubectl -n "${namespace}" rollout status "deployment/${gateway_name}" --timeout=2m

port_forward_log=$(mktemp)
kubectl -n "${namespace}" port-forward \
  "service/${gateway_name}" "${local_port}:8083" >"${port_forward_log}" 2>&1 &
port_forward_pid=$!

# Invoked indirectly by the EXIT trap.
# shellcheck disable=SC2317
cleanup() {
  kill "${port_forward_pid}" >/dev/null 2>&1 || true
  wait "${port_forward_pid}" >/dev/null 2>&1 || true
  rm -f -- "${port_forward_log}"
}
trap cleanup EXIT

for _ in $(seq 1 30); do
  if ! kill -0 "${port_forward_pid}" >/dev/null 2>&1; then
    cat "${port_forward_log}" >&2
    exit 1
  fi
  if curl --fail --silent --show-error \
    "http://127.0.0.1:${local_port}/v1/info"; then
    echo
    echo "SQL Gateway is ready at http://127.0.0.1:${local_port}"
    exit 0
  fi
  sleep 1
done

cat "${port_forward_log}" >&2
echo "SQL Gateway did not become reachable" >&2
exit 1
