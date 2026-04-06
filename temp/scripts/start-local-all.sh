#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/local-common.sh"

load_local_env
ensure_local_state_dirs

cleanup_on_error() {
  echo "local startup failed, stopping managed processes" >&2
  "${ROOT_DIR}/scripts/stop-local.sh" >/dev/null 2>&1 || true
}

trap cleanup_on_error ERR

start_managed_process() {
  local name="$1"
  local workdir="$2"
  local command="$3"
  local pid_file
  local log_file
  local pid

  pid_file="$(pid_file_for "${name}")"
  log_file="$(log_file_for "${name}")"

  : >"${log_file}"
  SERVICE_WORKDIR="${workdir}" \
  SERVICE_COMMAND="${command}" \
  SERVICE_LOG_FILE="${log_file}" \
  SERVICE_PID_FILE="${pid_file}" \
  python3 - <<'PY'
import os
import subprocess

workdir = os.environ["SERVICE_WORKDIR"]
command = os.environ["SERVICE_COMMAND"]
log_file = os.environ["SERVICE_LOG_FILE"]
pid_file = os.environ["SERVICE_PID_FILE"]

with open(log_file, "ab", buffering=0) as log_handle, open(os.devnull, "rb") as devnull:
    process = subprocess.Popen(
        ["bash", "-lc", command],
        cwd=workdir,
        stdin=devnull,
        stdout=log_handle,
        stderr=subprocess.STDOUT,
        start_new_session=True,
    )

with open(pid_file, "w", encoding="utf-8") as pid_handle:
    pid_handle.write(str(process.pid))
PY

  pid="$(read_pid_file "${pid_file}")"
  if [[ -z "${pid}" ]]; then
    echo "failed to capture pid for ${name}" >&2
    return 1
  fi

  wait_for_pid "${pid}" "${name}"
}

ensure_ui_dependencies() {
  if [[ ! -d "${ROOT_DIR}/ui/node_modules" ]]; then
    echo "installing UI dependencies"
    (cd "${ROOT_DIR}/ui" && npm ci)
  fi
}

echo "stopping previously managed local processes"
"${ROOT_DIR}/scripts/stop-local.sh" >/dev/null 2>&1 || true

echo "checking local ports"
"${ROOT_DIR}/scripts/check-local-ports.sh"

echo "starting local dependency stack"
(cd "${ROOT_DIR}" && docker compose -f deploy/local/docker-compose.yaml up -d)
"${ROOT_DIR}/scripts/wait-local-env.sh"

echo "applying migrations and seed data"
(
  cd "${ROOT_DIR}"
  set -a
  . ./deploy/local/iamd.env
  set +a
  go run ./cmd/migrator
  go run ./cmd/seeder
  go run ./cmd/schema-sync
)

ensure_ui_dependencies

echo "starting iamd"
start_managed_process \
  "iamd" \
  "${ROOT_DIR}" \
  "set -a; . ./deploy/local/iamd.env; set +a; exec go run ./cmd/iamd"
wait_for_port "$(extract_port "${IAM_GRPC_ADDRESS:-}")" "iamd gRPC"
wait_for_port "$(extract_port "${IAM_HTTP_ADDRESS:-}")" "iamd HTTP"

echo "starting worker"
start_managed_process \
  "worker" \
  "${ROOT_DIR}" \
  "set -a; . ./deploy/local/iamd.env; set +a; exec go run ./cmd/worker"

echo "starting UI"
start_managed_process \
  "ui" \
  "${ROOT_DIR}/ui" \
  "export VITE_IAM_API_BASE_URL='${VITE_IAM_API_BASE_URL}'; exec npm run dev -- --host '${IAM_UI_HOST}' --port '${IAM_UI_PORT}'"
wait_for_port "${IAM_UI_PORT}" "UI"

echo
echo "local environment is ready"
echo "gRPC: grpc://127.0.0.1$(printf '%s' "${IAM_GRPC_ADDRESS}")"
echo "REST: http://127.0.0.1$(printf '%s' "${IAM_HTTP_ADDRESS}")"
echo "UI:   http://${IAM_UI_HOST}:${IAM_UI_PORT}"
echo "logs: ${LOG_DIR}"
echo "status: make local-status"
echo "stop:   make local-down"
