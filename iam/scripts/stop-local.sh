#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/local-common.sh"

load_local_env
ensure_local_state_dirs

stop_managed_service() {
  local name="$1"
  local pid_file
  local pid

  pid_file="$(pid_file_for "${name}")"
  pid="$(read_pid_file "${pid_file}")"

  if [[ -n "${pid}" ]]; then
    if is_pid_running "${pid}"; then
      echo "stopping ${name}: ${pid}"
      stop_pid "${pid}" "${name}"
    fi
    rm -f "${pid_file}"
  fi
}

stop_listener_group() {
  local label="$1"
  shift
  local pids

  pids="$(
    {
      for port in "$@"; do
        listener_pids "${port}"
      done
    } | sort -u
  )"

  if [[ -n "${pids}" ]]; then
    echo "stopping ${label}: ${pids}"
    kill ${pids} 2>/dev/null || true
  fi
}

stop_managed_service "ui"
stop_managed_service "worker"
stop_managed_service "iamd"

stop_listener_group \
  "local listeners" \
  "$(extract_port "${IAM_GRPC_ADDRESS:-}")" \
  "$(extract_port "${IAM_HTTP_ADDRESS:-}")" \
  "${IAM_UI_PORT:-}"

echo "local processes stopped"
