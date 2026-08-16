#!/usr/bin/env bash

set -euo pipefail

namespace=${1:-flink-1c}
deployment_name=${2:-${namespace}}

if ! command -v kubectl >/dev/null 2>&1; then
  echo "required command not found: kubectl" >&2
  exit 1
fi

kubectl -n "${namespace}" get flinkdeployment "${deployment_name}"

jobmanager_status=$(
  kubectl -n "${namespace}" get flinkdeployment "${deployment_name}" \
    -o jsonpath='{.status.jobManagerDeploymentStatus}'
)

if [[ ${jobmanager_status} != READY ]]; then
  echo "JobManager is not ready: ${jobmanager_status:-status unavailable}" >&2
  kubectl -n "${namespace}" describe flinkdeployment "${deployment_name}" >&2
  exit 1
fi

kubectl -n "${namespace}" get service "${deployment_name}-rest"
kubectl -n "${namespace}" get pods -o wide

echo "${deployment_name} is ready in namespace ${namespace}"
