#!/usr/bin/env bash

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
HELM_VERSION="${HELM_VERSION:-}"
TMP_HELM_DIR=""

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
  if [[ -n "${TMP_HELM_DIR:-}" ]]; then
    rm -rf "$TMP_HELM_DIR"
  fi
}

trap cleanup EXIT

show_help() {
  cat <<'EOF'
Install Helm for M8 Platform local Kubernetes development on macOS.

Environment variables:
  HELM_VERSION  Helm version to install. Defaults to the latest Helm GitHub release.
  INSTALL_DIR   directory where helm will be installed. Defaults to /usr/local/bin.

Examples:
  scripts/install-helm-macos.sh
  HELM_VERSION=v3.16.0 scripts/install-helm-macos.sh
  INSTALL_DIR="$HOME/.local/bin" scripts/install-helm-macos.sh

Notes:
  - macOS only.
  - Homebrew and jq are not used.
EOF
}

usage_error() {
  err "$1"
  log "Run scripts/install-helm-macos.sh --help for usage."
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
  local latest_url
  local version

  if [[ -n "$HELM_VERSION" ]]; then
    printf '%s' "$HELM_VERSION"
    return
  fi

  latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/helm/helm/releases/latest)"
  version="${latest_url##*/}"

  if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    err "could not resolve latest Helm version from '$latest_url'"
    exit 1
  fi

  printf '%s' "$version"
}

show_existing_helm() {
  local existing_helm

  if existing_helm="$(command -v helm 2>/dev/null)"; then
    log "Existing helm detected at: $existing_helm"
    if "$existing_helm" version --short >/dev/null 2>&1; then
      log "Current helm version: $("$existing_helm" version --short)"
    else
      log "warning: existing helm binary could not report its version"
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

install_helm() {
  local archive_name
  local archive_url
  local arch
  local checksum_name
  local checksum_url
  local extracted_helm
  local resolved_helm
  local target
  local version

  arch="$(detect_arch)"
  version="$(resolve_version)"
  archive_name="helm-${version}-darwin-${arch}.tar.gz"
  checksum_name="${archive_name}.sha256sum"
  archive_url="https://get.helm.sh/${archive_name}"
  checksum_url="https://get.helm.sh/${checksum_name}"
  TMP_HELM_DIR="$(mktemp -d)"
  target="${INSTALL_DIR}/helm"

  log "Downloading Helm ${version} for macOS ${arch}."
  curl -fsSL --retry 3 --retry-delay 2 -o "${TMP_HELM_DIR}/${archive_name}" "$archive_url"
  curl -fsSL --retry 3 --retry-delay 2 -o "${TMP_HELM_DIR}/${checksum_name}" "$checksum_url"

  (
    cd "$TMP_HELM_DIR"
    shasum -a 256 --check "$checksum_name" >/dev/null
  )

  tar -xzf "${TMP_HELM_DIR}/${archive_name}" -C "$TMP_HELM_DIR"
  extracted_helm="${TMP_HELM_DIR}/darwin-${arch}/helm"
  chmod 0755 "$extracted_helm"

  ensure_install_dir

  if [[ -w "$INSTALL_DIR" ]]; then
    mv "$extracted_helm" "$target"
  else
    log "Install directory '$INSTALL_DIR' is not writable. sudo is required to install helm."
    need_cmd sudo
    sudo mv "$extracted_helm" "$target"
  fi

  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      log "warning: '$INSTALL_DIR' is not in PATH for this shell; add it to PATH to run helm directly later"
      ;;
  esac

  export PATH="${INSTALL_DIR}:$PATH"
  if ! resolved_helm="$(command -v helm)"; then
    err "helm was installed to '$target', but command -v helm failed"
    exit 1
  fi

  log "Installed helm at: $target"
  log "helm command resolved to: $resolved_helm"
  log "Installed helm version: $(helm version --short)"
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
  need_cmd tar

  show_existing_helm
  install_helm

  log "Helm is ready for M8 local Kubernetes development."
}

main "$@"
