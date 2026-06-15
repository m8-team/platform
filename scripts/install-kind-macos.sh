#!/usr/bin/env bash

set -euo pipefail

DEFAULT_KIND_VERSION="v0.32.0"
KIND_VERSION="${KIND_VERSION:-$DEFAULT_KIND_VERSION}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TMP_KIND_DIR=""

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

cleanup() {
  if [[ -n "${TMP_KIND_DIR:-}" ]]; then
    rm -rf "$TMP_KIND_DIR"
  fi
}

trap cleanup EXIT

show_help() {
  cat <<'EOF'
Install kind for M8 Platform local Kubernetes development on macOS.

This script downloads the official kind macOS binary directly from:
  https://kind.sigs.k8s.io

Environment variables:
  KIND_VERSION  kind version to install. Defaults to v0.32.0.
  INSTALL_DIR   directory where the kind binary will be installed. Defaults to /usr/local/bin.

Examples:
  scripts/install-kind-macos.sh

  KIND_VERSION=v0.32.0 scripts/install-kind-macos.sh

  INSTALL_DIR="$HOME/.local/bin" scripts/install-kind-macos.sh

Notes:
  - macOS only.
  - Docker, Colima, and kubectl are not installed automatically.
  - If INSTALL_DIR is not writable, sudo will be used for installation.
EOF
}

usage_error() {
  err "$1"
  log "Run scripts/install-kind-macos.sh --help for usage."
  exit 1
}

detect_arch() {
  local machine
  machine="$(uname -m)"

  case "$machine" in
    arm64)
      printf 'arm64'
      ;;
    x86_64 | amd64)
      printf 'amd64'
      ;;
    *)
      err "unsupported macOS architecture '$machine'; supported architectures are arm64 and amd64/x86_64"
      exit 1
      ;;
  esac
}

check_macos() {
  local os
  os="$(uname -s)"

  if [[ "$os" != "Darwin" ]]; then
    err "this installer supports macOS only; detected '$os'"
    exit 1
  fi
}

show_existing_kind() {
  local existing_kind

  if existing_kind="$(command -v kind 2>/dev/null)"; then
    log "Existing kind detected at: $existing_kind"
    if "$existing_kind" version >/dev/null 2>&1; then
      log "Current kind version: $("$existing_kind" version)"
    else
      log "warning: existing kind binary could not report its version"
    fi
    log "Reinstalling kind ${KIND_VERSION}."
    return
  fi

  if [[ -x "${INSTALL_DIR}/kind" ]]; then
    log "Existing kind detected at: ${INSTALL_DIR}/kind"
    if "${INSTALL_DIR}/kind" version >/dev/null 2>&1; then
      log "Current kind version: $("${INSTALL_DIR}/kind" version)"
    else
      log "warning: existing kind binary could not report its version"
    fi
    log "Reinstalling kind ${KIND_VERSION}."
    return
  fi

  log "kind is not installed; installing kind ${KIND_VERSION}."
}

ensure_install_dir() {
  if [[ -d "$INSTALL_DIR" ]]; then
    return
  fi

  if mkdir -p "$INSTALL_DIR" 2>/dev/null; then
    return
  fi

  log "Install directory '$INSTALL_DIR' must be created with sudo."
  need_cmd sudo
  sudo mkdir -p "$INSTALL_DIR"
}

install_kind() {
  local arch
  local installed_version
  local resolved_kind
  local url
  local tmp_kind
  local target

  arch="$(detect_arch)"
  url="https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-darwin-${arch}"
  TMP_KIND_DIR="$(mktemp -d)"
  tmp_kind="${TMP_KIND_DIR}/kind"
  target="${INSTALL_DIR}/kind"

  log "Downloading kind ${KIND_VERSION} for macOS ${arch}."
  curl -fsSL --retry 3 --retry-delay 2 -o "$tmp_kind" "$url"
  chmod 0755 "$tmp_kind"

  ensure_install_dir

  if [[ -w "$INSTALL_DIR" ]]; then
    mv "$tmp_kind" "$target"
  else
    log "Install directory '$INSTALL_DIR' is not writable. sudo is required to install kind."
    need_cmd sudo
    sudo mv "$tmp_kind" "$target"
  fi

  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      log "warning: '$INSTALL_DIR' is not in PATH for this shell; add it to PATH to run kind directly later"
      ;;
  esac

  export PATH="${INSTALL_DIR}:$PATH"
  if ! resolved_kind="$(command -v kind)"; then
    err "kind was installed to '$target', but command -v kind failed"
    exit 1
  fi
  if ! installed_version="$(kind version 2>&1)"; then
    err "kind was installed to '$target', but 'kind version' failed: $installed_version"
    exit 1
  fi

  log "Installed kind at: $target"
  log "kind command resolved to: $resolved_kind"
  log "Installed kind version: $installed_version"
}

check_docker_runtime() {
  if command -v docker >/dev/null 2>&1; then
    if docker info >/dev/null 2>&1; then
      log "Docker runtime is available."
      if command -v colima >/dev/null 2>&1; then
        log "Colima is installed. If you use Colima for Docker runtime, start it with:"
        log "  colima start"
      fi
      return 0
    fi

    err "docker is installed, but the Docker daemon is not running or is not reachable"
    err "start Docker Desktop or another Docker-compatible runtime, then rerun this script"
    if command -v colima >/dev/null 2>&1; then
      log "Colima is installed. You can start Docker runtime with:"
      log "  colima start"
    fi
    return 1
  fi

  err "docker is not installed; install Docker Desktop or Colima before creating kind clusters"
  if command -v colima >/dev/null 2>&1; then
    log "Colima is installed. You can start Docker runtime with:"
    log "  colima start"
  else
    log "Install Docker Desktop or Colima, then rerun this script."
  fi
  return 1
}

check_kubectl() {
  local kubectl_version

  if ! command -v kubectl >/dev/null 2>&1; then
    log "warning: kubectl is not installed; it is needed to work with Kubernetes clusters"
    return
  fi

  if kubectl_version="$(kubectl version --client=true 2>/dev/null)"; then
    log "kubectl client version:"
    log "$kubectl_version"
  else
    log "warning: kubectl is installed, but its client version could not be read"
  fi
}

main() {
  if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    show_help
    exit 0
  fi

  if [[ "$#" -gt 0 ]]; then
    usage_error "unexpected argument: $1"
  fi

  check_macos
  need_cmd curl

  show_existing_kind
  install_kind
  check_kubectl
  check_docker_runtime

  log "kind is ready for M8 local Kubernetes development."
}

main "$@"
