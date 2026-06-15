#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-m8-local}"
KIND_CONFIG="${KIND_CONFIG:-${ROOT_DIR}/deploy/local/kind-config.yaml}"

log() {
  printf '%s\n' "$*"
}

err() {
  printf 'error: %s\n' "$*" >&2
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "required command '$1' was not found"
    exit 1
  fi
}

show_help() {
  cat <<'EOF'
Manage the M8 local kind cluster.

Usage:
  scripts/dev-kind-cluster.sh up
  scripts/dev-kind-cluster.sh down
  scripts/dev-kind-cluster.sh reset

Environment variables:
  KIND_CLUSTER_NAME  kind cluster name. Defaults to m8-local.
  KIND_CONFIG        kind cluster config path. Defaults to deploy/local/kind-config.yaml.
EOF
}

check_docker_runtime() {
  need_cmd docker

  if docker info >/dev/null 2>&1; then
    return
  fi

  err "docker is installed, but the Docker daemon is not running or is not reachable"
  err "start Docker Desktop or another Docker-compatible runtime before creating the kind cluster"
  if command -v colima >/dev/null 2>&1; then
    log "Colima is installed. You can start Docker runtime with:"
    log "  colima start"
  fi
  exit 1
}

cluster_exists() {
  kind get clusters | grep -Fxq "$KIND_CLUSTER_NAME"
}

create_cluster() {
  local -a args

  need_cmd kind
  need_cmd kubectl
  check_docker_runtime

  if cluster_exists; then
    log "kind cluster '${KIND_CLUSTER_NAME}' already exists."
  else
    args=(create cluster --name "$KIND_CLUSTER_NAME")
    if [[ -f "$KIND_CONFIG" ]]; then
      args+=(--config "$KIND_CONFIG")
      log "Creating kind cluster '${KIND_CLUSTER_NAME}' using config: $KIND_CONFIG"
    else
      log "Creating kind cluster '${KIND_CLUSTER_NAME}' with kind defaults."
    fi
    kind "${args[@]}"
  fi

  kubectl cluster-info --context "kind-${KIND_CLUSTER_NAME}"
  kubectl get nodes --context "kind-${KIND_CLUSTER_NAME}"
}

delete_cluster() {
  need_cmd kind

  if cluster_exists; then
    log "Deleting kind cluster '${KIND_CLUSTER_NAME}'."
    kind delete cluster --name "$KIND_CLUSTER_NAME"
  else
    log "kind cluster '${KIND_CLUSTER_NAME}' does not exist."
  fi
}

main() {
  case "${1:-}" in
    up)
      create_cluster
      ;;
    down)
      delete_cluster
      ;;
    reset)
      delete_cluster
      create_cluster
      ;;
    --help | -h)
      show_help
      ;;
    "")
      show_help
      exit 1
      ;;
    *)
      err "unexpected command: $1"
      show_help
      exit 1
      ;;
  esac
}

main "$@"
