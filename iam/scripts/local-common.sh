#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${TMPDIR:-/tmp}/m8-platform-iam-local"
PID_DIR="${STATE_DIR}/pids"
LOG_DIR="${STATE_DIR}/logs"

load_local_env() {
  set -a
  . "${ROOT_DIR}/deploy/local/iamd.env"
  set +a

  export IAM_UI_HOST="${IAM_UI_HOST:-127.0.0.1}"
  export IAM_UI_PORT="${IAM_UI_PORT:-5173}"
  export VITE_IAM_API_BASE_URL="${VITE_IAM_API_BASE_URL:-http://${IAM_UI_HOST}${IAM_HTTP_ADDRESS:-:8082}}"
}

ensure_local_state_dirs() {
  mkdir -p "${PID_DIR}" "${LOG_DIR}"
}

extract_port() {
  local addr="$1"
  printf '%s\n' "${addr##*:}"
}

pid_file_for() {
  local name="$1"
  printf '%s/%s.pid\n' "${PID_DIR}" "${name}"
}

log_file_for() {
  local name="$1"
  printf '%s/%s.log\n' "${LOG_DIR}" "${name}"
}

read_pid_file() {
  local file="$1"

  if [[ -f "${file}" ]]; then
    tr -d '[:space:]' <"${file}"
  fi
}

is_pid_running() {
  local pid="${1:-}"
  [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null
}

listener_pids() {
  local port="$1"

  if [[ -z "${port}" ]]; then
    return 0
  fi

  lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true
}

wait_for_port() {
  local port="$1"
  local label="$2"
  local attempts="${3:-60}"
  local sleep_seconds="${4:-1}"
  local i

  for ((i = 1; i <= attempts; i += 1)); do
    if lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1; then
      return 0
    fi
    sleep "${sleep_seconds}"
  done

  echo "${label} did not start on port ${port}" >&2
  return 1
}

wait_for_pid() {
  local pid="$1"
  local label="$2"
  local attempts="${3:-10}"
  local sleep_seconds="${4:-1}"
  local i

  for ((i = 1; i <= attempts; i += 1)); do
    if is_pid_running "${pid}"; then
      return 0
    fi
    sleep "${sleep_seconds}"
  done

  echo "${label} is not running" >&2
  return 1
}

stop_pid() {
  local pid="$1"
  local label="$2"
  local i

  if ! is_pid_running "${pid}"; then
    return 0
  fi

  kill "${pid}" 2>/dev/null || true
  for ((i = 1; i <= 10; i += 1)); do
    if ! is_pid_running "${pid}"; then
      return 0
    fi
    sleep 1
  done

  echo "force stopping ${label} (${pid})"
  kill -9 "${pid}" 2>/dev/null || true
}
