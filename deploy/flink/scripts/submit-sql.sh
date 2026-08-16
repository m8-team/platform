#!/usr/bin/env bash

set -euo pipefail

gateway_url=${1:-http://127.0.0.1:8083}
sql_root=${2:-deploy/flink/installations/flink-1c/sql}
gateway_url=${gateway_url%/}

for required_command in curl jq; do
  if ! command -v "${required_command}" >/dev/null 2>&1; then
    echo "required command not found: ${required_command}" >&2
    exit 1
  fi
done

if [[ ! -d ${sql_root} ]]; then
  echo "SQL directory not found: ${sql_root}" >&2
  exit 1
fi

shopt -s nullglob
sql_files=(
  "${sql_root}"/sources/*.sql
  "${sql_root}"/views/*.sql
  "${sql_root}"/sinks/*.sql
  "${sql_root}"/jobs/*.sql
)
shopt -u nullglob

if [[ ${#sql_files[@]} -eq 0 ]]; then
  echo "no SQL files found under ${sql_root}" >&2
  exit 1
fi

session_name="gitops-$(basename -- "$(dirname -- "${sql_root}")")-$(date +%s)"
session_payload=$(jq -n --arg session_name "${session_name}" '{sessionName: $session_name}')
session_response=$(
  curl --fail-with-body --silent --show-error \
    -H 'Content-Type: application/json' \
    -X POST \
    -d "${session_payload}" \
    "${gateway_url}/v1/sessions"
)
session_handle=$(jq -er '.sessionHandle' <<<"${session_response}")

close_session() {
  curl --silent --show-error -X DELETE \
    "${gateway_url}/v1/sessions/${session_handle}" >/dev/null || true
}
trap close_session EXIT

wait_for_operation() {
  local operation_handle=$1
  local status_response status

  for _ in $(seq 1 180); do
    status_response=$(
      curl --fail-with-body --silent --show-error \
        "${gateway_url}/v1/sessions/${session_handle}/operations/${operation_handle}/status"
    )
    status=$(jq -er '.status' <<<"${status_response}")

    case ${status} in
      FINISHED)
        return 0
        ;;
      ERROR | CANCELED | CLOSED | TIMEOUT)
        curl --silent --show-error \
          "${gateway_url}/v1/sessions/${session_handle}/operations/${operation_handle}/result/0?rowFormat=JSON" \
          >&2 || true
        echo "operation ${operation_handle} failed with status ${status}" >&2
        return 1
        ;;
    esac

    sleep 1
  done

  echo "operation ${operation_handle} timed out while waiting for completion" >&2
  return 1
}

for sql_file in "${sql_files[@]}"; do
  echo "submitting ${sql_file}"
  statement_payload=$(jq -Rs '{statement: .}' "${sql_file}")
  statement_response=$(
    curl --fail-with-body --silent --show-error \
      -H 'Content-Type: application/json' \
      -X POST \
      -d "${statement_payload}" \
      "${gateway_url}/v1/sessions/${session_handle}/statements"
  )
  operation_handle=$(jq -er '.operationHandle' <<<"${statement_response}")
  wait_for_operation "${operation_handle}"
  curl --fail-with-body --silent --show-error -X DELETE \
    "${gateway_url}/v1/sessions/${session_handle}/operations/${operation_handle}/close" \
    >/dev/null
done

echo "submitted ${#sql_files[@]} SQL files through session ${session_handle}"
