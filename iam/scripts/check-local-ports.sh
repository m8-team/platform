#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/local-common.sh"

load_local_env

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
check_port "${IAM_UI_PORT:-}"
