#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-m8-local}"
KUBE_CONTEXT="${KUBE_CONTEXT:-kind-${KIND_CLUSTER_NAME}}"
CLICKSTACK_HELM_REPO="${CLICKSTACK_HELM_REPO:-https://clickhouse.github.io/ClickStack-helm-charts}"
CLICKSTACK_RELEASE="${CLICKSTACK_RELEASE:-m8-clickstack}"
CLICKSTACK_NAMESPACE="${CLICKSTACK_NAMESPACE:-m8-observability}"
CLICKSTACK_OPERATORS_RELEASE="${CLICKSTACK_OPERATORS_RELEASE:-clickstack-operators}"
CLICKSTACK_OPERATORS_NAMESPACE="${CLICKSTACK_OPERATORS_NAMESPACE:-clickstack-system}"
CLICKSTACK_LEGACY_OPERATORS_NAMESPACE="${CLICKSTACK_LEGACY_OPERATORS_NAMESPACE:-clickstack-system}"
CLICKSTACK_VALUES="${CLICKSTACK_VALUES:-${ROOT_DIR}/deploy/local/clickstack-values.yaml}"
CLICKSTACK_TIMEOUT="${CLICKSTACK_TIMEOUT:-15m}"
CLICKSTACK_UI_PORT="${CLICKSTACK_UI_PORT:-8080}"
CLICKSTACK_OTLP_GRPC_PORT="${CLICKSTACK_OTLP_GRPC_PORT:-4317}"
CLICKSTACK_OTLP_HTTP_PORT="${CLICKSTACK_OTLP_HTTP_PORT:-4318}"

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
Manage ClickStack in the M8 local kind environment.

Usage:
  scripts/dev-clickstack.sh install
  scripts/dev-clickstack.sh uninstall
  scripts/dev-clickstack.sh reset
  scripts/dev-clickstack.sh status
  scripts/dev-clickstack.sh port-forward

Environment variables:
  KUBE_CONTEXT                   Kubernetes context. Defaults to kind-${KIND_CLUSTER_NAME}.
  KIND_CLUSTER_NAME              kind cluster name used to derive KUBE_CONTEXT. Defaults to m8-local.
  CLICKSTACK_RELEASE             ClickStack Helm release. Defaults to m8-clickstack.
  CLICKSTACK_NAMESPACE           ClickStack namespace. Defaults to m8-observability.
  CLICKSTACK_OPERATORS_RELEASE   ClickStack operators Helm release. Defaults to clickstack-operators.
  CLICKSTACK_OPERATORS_NAMESPACE ClickStack operators namespace. Defaults to clickstack-system.
  CLICKSTACK_LEGACY_OPERATORS_NAMESPACE
                                  Old operators namespace to clean up during uninstall/reset.
                                  Defaults to clickstack-system.
  CLICKSTACK_VALUES              Helm values file. Defaults to deploy/local/clickstack-values.yaml.
  CLICKSTACK_TIMEOUT             Helm/kubectl wait timeout. Defaults to 15m.
  CLICKSTACK_UI_PORT             Local UI port for port-forward. Defaults to 8080.
  CLICKSTACK_OTLP_GRPC_PORT      Local OTLP gRPC port for port-forward. Defaults to 4317.
  CLICKSTACK_OTLP_HTTP_PORT      Local OTLP HTTP port for port-forward. Defaults to 4318.
EOF
}

usage_error() {
  err "$1"
  log "Run scripts/dev-clickstack.sh --help for usage."
  exit 1
}

kubectl_cmd() {
  kubectl --context "$KUBE_CONTEXT" "$@"
}

helm_cmd() {
  helm --kube-context "$KUBE_CONTEXT" "$@"
}

ensure_cluster() {
  if ! kubectl config get-contexts -o name | grep -Fxq "$KUBE_CONTEXT"; then
    err "Kubernetes context '$KUBE_CONTEXT' was not found"
    err "create the local kind cluster first: make dev:up"
    exit 1
  fi

  if ! kubectl_cmd cluster-info >/dev/null 2>&1; then
    err "Kubernetes context '$KUBE_CONTEXT' is not reachable"
    exit 1
  fi
}

ensure_values_file() {
  if [[ ! -f "$CLICKSTACK_VALUES" ]]; then
    err "ClickStack values file was not found: $CLICKSTACK_VALUES"
    exit 1
  fi
}

add_helm_repo() {
  helm repo add clickstack "$CLICKSTACK_HELM_REPO" --force-update >/dev/null
  helm repo update clickstack >/dev/null
}

install_clickstack() {
  need_cmd helm
  need_cmd kubectl
  ensure_cluster
  ensure_values_file
  add_helm_repo

  log "Installing ClickStack operators into namespace '${CLICKSTACK_OPERATORS_NAMESPACE}'."
  helm_cmd upgrade --install "$CLICKSTACK_OPERATORS_RELEASE" clickstack/clickstack-operators \
    --namespace "$CLICKSTACK_OPERATORS_NAMESPACE" \
    --create-namespace \
    --set "mongodb-operator.operator.watchNamespace=${CLICKSTACK_NAMESPACE}" \
    --set "clickhouse-operator.watchNamespaces[0]=${CLICKSTACK_NAMESPACE}" \
    --wait \
    --timeout "$CLICKSTACK_TIMEOUT"

  log "Installing ClickStack into namespace '${CLICKSTACK_NAMESPACE}'."
  helm_cmd upgrade --install "$CLICKSTACK_RELEASE" clickstack/clickstack \
    --namespace "$CLICKSTACK_NAMESPACE" \
    --create-namespace \
    --values "$CLICKSTACK_VALUES" \
    --wait \
    --timeout "$CLICKSTACK_TIMEOUT"

  restart_collector
  wait_clickstack
  print_endpoints
}

release_exists() {
  local release="$1"
  local namespace="$2"

  helm_cmd status "$release" --namespace "$namespace" >/dev/null 2>&1
}

uninstall_release() {
  local release="$1"
  local namespace="$2"

  if release_exists "$release" "$namespace"; then
    log "Uninstalling Helm release '${release}' from namespace '${namespace}'."
    helm_cmd uninstall "$release" --namespace "$namespace" --wait --timeout "$CLICKSTACK_TIMEOUT"
  else
    log "Helm release '${release}' is not installed in namespace '${namespace}'."
  fi
}

uninstall_clickstack() {
  need_cmd helm
  need_cmd kubectl
  ensure_cluster

  uninstall_release "$CLICKSTACK_RELEASE" "$CLICKSTACK_NAMESPACE"
  uninstall_release "$CLICKSTACK_OPERATORS_RELEASE" "$CLICKSTACK_OPERATORS_NAMESPACE"
  if [[ "$CLICKSTACK_LEGACY_OPERATORS_NAMESPACE" != "$CLICKSTACK_OPERATORS_NAMESPACE" ]]; then
    uninstall_release "$CLICKSTACK_OPERATORS_RELEASE" "$CLICKSTACK_LEGACY_OPERATORS_NAMESPACE"
  fi
}

wait_clickstack() {
  log "Waiting for ClickStack app and collector workloads."
  kubectl_cmd rollout status "deployment/${CLICKSTACK_RELEASE}-app" \
    --namespace "$CLICKSTACK_NAMESPACE" \
    --timeout "$CLICKSTACK_TIMEOUT"
  kubectl_cmd rollout status "daemonset/${CLICKSTACK_RELEASE}-otel-collector-agent" \
    --namespace "$CLICKSTACK_NAMESPACE" \
    --timeout "$CLICKSTACK_TIMEOUT"
}

restart_collector() {
  local daemonset="${CLICKSTACK_RELEASE}-otel-collector-agent"

  if kubectl_cmd get "daemonset/${daemonset}" --namespace "$CLICKSTACK_NAMESPACE" >/dev/null 2>&1; then
    log "Restarting ClickStack collector to apply local collector config."
    kubectl_cmd rollout restart "daemonset/${daemonset}" --namespace "$CLICKSTACK_NAMESPACE" >/dev/null
  fi
}

status_clickstack() {
  need_cmd helm
  need_cmd kubectl
  ensure_cluster

  if release_exists "$CLICKSTACK_OPERATORS_RELEASE" "$CLICKSTACK_OPERATORS_NAMESPACE"; then
    helm_cmd status "$CLICKSTACK_OPERATORS_RELEASE" --namespace "$CLICKSTACK_OPERATORS_NAMESPACE"
  elif [[ "$CLICKSTACK_LEGACY_OPERATORS_NAMESPACE" != "$CLICKSTACK_OPERATORS_NAMESPACE" ]] && release_exists "$CLICKSTACK_OPERATORS_RELEASE" "$CLICKSTACK_LEGACY_OPERATORS_NAMESPACE"; then
    log "ClickStack operators release is installed in legacy namespace '${CLICKSTACK_LEGACY_OPERATORS_NAMESPACE}'."
    helm_cmd status "$CLICKSTACK_OPERATORS_RELEASE" --namespace "$CLICKSTACK_LEGACY_OPERATORS_NAMESPACE"
  else
    log "ClickStack operators release is not installed."
  fi

  if release_exists "$CLICKSTACK_RELEASE" "$CLICKSTACK_NAMESPACE"; then
    helm_cmd status "$CLICKSTACK_RELEASE" --namespace "$CLICKSTACK_NAMESPACE"
  else
    log "ClickStack release is not installed."
  fi

  log "ClickStack namespace resources:"
  kubectl_cmd get pods,svc,pvc --namespace "$CLICKSTACK_NAMESPACE" || true
}

print_endpoints() {
  log "ClickStack UI:"
  log "  make clickstack:port-forward"
  log "  http://localhost:${CLICKSTACK_UI_PORT}"
  log "In-cluster OTLP endpoints:"
  log "  gRPC: ${CLICKSTACK_RELEASE}-otel-collector.${CLICKSTACK_NAMESPACE}.svc.cluster.local:4317"
  log "  HTTP: http://${CLICKSTACK_RELEASE}-otel-collector.${CLICKSTACK_NAMESPACE}.svc.cluster.local:4318"
  log "Local OTLP endpoints after port-forward:"
  log "  gRPC: localhost:${CLICKSTACK_OTLP_GRPC_PORT}"
  log "  HTTP: http://localhost:${CLICKSTACK_OTLP_HTTP_PORT}"
}

port_forward_clickstack() {
  local otlp_pid
  local ui_pid

  need_cmd kubectl
  ensure_cluster

  log "Forwarding ClickStack UI to http://localhost:${CLICKSTACK_UI_PORT}"
  kubectl_cmd port-forward \
    --namespace "$CLICKSTACK_NAMESPACE" \
    "svc/${CLICKSTACK_RELEASE}-app" \
    "${CLICKSTACK_UI_PORT}:3000" &
  ui_pid="$!"

  log "Forwarding OTLP gRPC to localhost:${CLICKSTACK_OTLP_GRPC_PORT} and OTLP HTTP to http://localhost:${CLICKSTACK_OTLP_HTTP_PORT}"
  kubectl_cmd port-forward \
    --namespace "$CLICKSTACK_NAMESPACE" \
    "svc/${CLICKSTACK_RELEASE}-otel-collector" \
    "${CLICKSTACK_OTLP_GRPC_PORT}:4317" \
    "${CLICKSTACK_OTLP_HTTP_PORT}:4318" &
  otlp_pid="$!"

  cleanup_port_forward() {
    kill "$ui_pid" "$otlp_pid" >/dev/null 2>&1 || true
  }

  trap cleanup_port_forward INT TERM EXIT
  wait "$ui_pid" "$otlp_pid"
}

main() {
  case "${1:-}" in
    install)
      install_clickstack
      ;;
    uninstall)
      uninstall_clickstack
      ;;
    reset)
      uninstall_clickstack
      install_clickstack
      ;;
    status)
      status_clickstack
      ;;
    port-forward)
      port_forward_clickstack
      ;;
    --help | -h)
      show_help
      ;;
    "")
      show_help
      exit 1
      ;;
    *)
      usage_error "unexpected command: $1"
      ;;
  esac
}

main "$@"
