#!/usr/bin/env bash

set -euo pipefail

runtime_image=${1:-flink-runtime:validation}
sql_root=${2:-deploy/flink/installations/flink-1c/sql}

if ! command -v docker >/dev/null 2>&1; then
  echo "required command not found: docker" >&2
  exit 1
fi

for sql_group in sources views sinks jobs; do
  if [[ ! -d ${sql_root}/${sql_group} ]]; then
    echo "SQL directory not found: ${sql_root}/${sql_group}" >&2
    exit 1
  fi
done

sql_client_log=$(mktemp)
# Invoked indirectly by the EXIT trap.
# shellcheck disable=SC2317
cleanup() {
  rm -f -- "${sql_client_log}"
}
trap cleanup EXIT

shopt -s nullglob
definition_files=(
  "${sql_root}"/sources/*.sql
  "${sql_root}"/views/*.sql
  "${sql_root}"/sinks/*.sql
)
job_files=("${sql_root}"/jobs/*.sql)
shopt -u nullglob

if [[ ${#definition_files[@]} -eq 0 || ${#job_files[@]} -eq 0 ]]; then
  echo "expected table/view definitions and at least one SQL job" >&2
  exit 1
fi

if ! {
  awk '{print}' "${definition_files[@]}"
  for job_file in "${job_files[@]}"; do
    printf 'EXPLAIN PLAN FOR\n'
    awk '{print}' "${job_file}"
  done
  printf 'QUIT;\n'
} | docker run --rm --interactive "${runtime_image}" \
  /opt/flink/bin/sql-client.sh embedded >"${sql_client_log}" 2>&1; then
  cat "${sql_client_log}" >&2
  echo "Flink SQL Client failed" >&2
  exit 1
fi

if grep -Fq '[ERROR]' "${sql_client_log}"; then
  cat "${sql_client_log}" >&2
  echo "Flink SQL syntax or connector validation failed" >&2
  exit 1
fi

echo "Flink SQL syntax and plans validated"
