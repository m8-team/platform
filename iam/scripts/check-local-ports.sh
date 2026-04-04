#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

set -a
. "${ROOT_DIR}/deploy/local/iamd.env"
set +a

extract_port() {
  local addr="$1"
  printf '%s\n' "${addr##*:}"
}

check_port() {
  local port="$1"
  local tmp_file

  if [[ -z "${port}" ]]; then
    return 0
  fi

  tmp_file="/tmp/iam-port-check.${port}"
  if lsof -nP -iTCP:"${port}" -sTCP:LISTEN >"${tmp_file}" 2>/dev/null; then
    echo "port ${port} is already in use:"
    cat "${tmp_file}"
    rm -f "${tmp_file}"
    echo "run 'make stop-local' to stop a previous local iamd, or change the port in deploy/local/iamd.env"
    return 1
  fi
}

check_port "$(extract_port "${IAM_GRPC_ADDRESS:-}")"
check_port "$(extract_port "${IAM_HTTP_ADDRESS:-}")"
