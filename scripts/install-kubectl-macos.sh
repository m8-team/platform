#!/usr/bin/env bash

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
KUBECTL_VERSION="${KUBECTL_VERSION:-}"
TMP_KUBECTL_DIR=""

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
  if [[ -n "${TMP_KUBECTL_DIR:-}" ]]; then
    rm -rf "$TMP_KUBECTL_DIR"
  fi
}

trap cleanup EXIT

show_help() {
  cat <<'EOF'
Install kubectl for M8 Platform local Kubernetes development on macOS.

Environment variables:
  KUBECTL_VERSION  kubectl version to install. Defaults to Kubernetes stable.txt.
  INSTALL_DIR      directory where kubectl will be installed. Defaults to /usr/local/bin.

Examples:
  scripts/install-kubectl-macos.sh
  KUBECTL_VERSION=v1.31.0 scripts/install-kubectl-macos.sh
  INSTALL_DIR="$HOME/.local/bin" scripts/install-kubectl-macos.sh

Notes:
  - macOS only.
  - Homebrew and jq are not used.
EOF
}

usage_error() {
  err "$1"
  log "Run scripts/install-kubectl-macos.sh --help for usage."
  exit 1
}

check_macos() {
  local os
  os="$(uname -s)"

  if [[ "$os" != "Darwin" ]]; then
    err "this installer supports macOS only; detected '$os'"
    exit 1
  fi
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

resolve_version() {
  if [[ -n "$KUBECTL_VERSION" ]]; then
    printf '%s' "$KUBECTL_VERSION"
    return
  fi

  curl -fsSL https://dl.k8s.io/release/stable.txt
}

show_existing_kubectl() {
  local existing_kubectl

  if existing_kubectl="$(command -v kubectl 2>/dev/null)"; then
    log "Existing kubectl detected at: $existing_kubectl"
    if "$existing_kubectl" version --client=true >/dev/null 2>&1; then
      log "Current kubectl client version:"
      "$existing_kubectl" version --client=true
    else
      log "warning: existing kubectl binary could not report its client version"
    fi
  fi
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

install_kubectl() {
  local arch
  local checksum_url
  local resolved_kubectl
  local target
  local tmp_checksum
  local tmp_kubectl
  local url
  local version

  arch="$(detect_arch)"
  version="$(resolve_version)"
  url="https://dl.k8s.io/release/${version}/bin/darwin/${arch}/kubectl"
  checksum_url="${url}.sha256"
  TMP_KUBECTL_DIR="$(mktemp -d)"
  tmp_kubectl="${TMP_KUBECTL_DIR}/kubectl"
  tmp_checksum="${TMP_KUBECTL_DIR}/kubectl.sha256"
  target="${INSTALL_DIR}/kubectl"

  log "Downloading kubectl ${version} for macOS ${arch}."
  curl -fsSL --retry 3 --retry-delay 2 -o "$tmp_kubectl" "$url"
  curl -fsSL --retry 3 --retry-delay 2 -o "$tmp_checksum" "$checksum_url"

  (
    cd "$TMP_KUBECTL_DIR"
    printf '%s  kubectl\n' "$(cat "$tmp_checksum")" | shasum -a 256 --check >/dev/null
  )
  chmod 0755 "$tmp_kubectl"

  ensure_install_dir

  if [[ -w "$INSTALL_DIR" ]]; then
    mv "$tmp_kubectl" "$target"
  else
    log "Install directory '$INSTALL_DIR' is not writable. sudo is required to install kubectl."
    need_cmd sudo
    sudo mv "$tmp_kubectl" "$target"
  fi

  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      log "warning: '$INSTALL_DIR' is not in PATH for this shell; add it to PATH to run kubectl directly later"
      ;;
  esac

  export PATH="${INSTALL_DIR}:$PATH"
  if ! resolved_kubectl="$(command -v kubectl)"; then
    err "kubectl was installed to '$target', but command -v kubectl failed"
    exit 1
  fi

  log "Installed kubectl at: $target"
  log "kubectl command resolved to: $resolved_kubectl"
  log "kubectl client version:"
  kubectl version --client=true
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
  need_cmd shasum

  show_existing_kubectl
  install_kubectl

  log "kubectl is ready for M8 local Kubernetes development."
}

main "$@"
