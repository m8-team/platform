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

collect_pids() {
  local port="$1"

  if [[ -z "${port}" ]]; then
    return 0
  fi

  lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true
}

pids="$(
  {
    collect_pids "$(extract_port "${IAM_GRPC_ADDRESS:-}")"
    collect_pids "$(extract_port "${IAM_HTTP_ADDRESS:-}")"
  } | sort -u
)"

if [[ -z "${pids}" ]]; then
  echo "no local iamd processes found"
  exit 0
fi

echo "stopping local iamd: ${pids}"
kill ${pids}
