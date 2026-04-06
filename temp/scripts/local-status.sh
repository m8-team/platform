#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/local-common.sh"

load_local_env
ensure_local_state_dirs

print_service_status() {
  local name="$1"
  local details="$2"
  local pid_file
  local pid
  local log_file

  pid_file="$(pid_file_for "${name}")"
  log_file="$(log_file_for "${name}")"
  pid="$(read_pid_file "${pid_file}")"

  if is_pid_running "${pid}"; then
    echo "${name}: running (pid ${pid}) ${details}"
    echo "  log: ${log_file}"
    return 0
  fi

  echo "${name}: stopped"
  if [[ -f "${log_file}" ]]; then
    echo "  log: ${log_file}"
  fi
}

echo "Managed processes"
print_service_status "iamd" "grpc=127.0.0.1${IAM_GRPC_ADDRESS} http=127.0.0.1${IAM_HTTP_ADDRESS}"
print_service_status "worker" "task-queue=${IAM_TEMPORAL_TASK_QUEUE}"
print_service_status "ui" "url=http://${IAM_UI_HOST}:${IAM_UI_PORT}"
echo
echo "Dependency stack"
(cd "${ROOT_DIR}" && docker compose -f deploy/local/docker-compose.yaml ps)
