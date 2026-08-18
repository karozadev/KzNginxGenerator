#!/bin/sh
# install.sh — one-liner installer for KzNginxGenerator (kznginx).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/karozadev/KzNginxGenerator/main/install.sh | sh
#
# Detects OS/arch, downloads the latest stable release binary from
# GitHub Releases, and installs it to /usr/local/bin/kznginx.
set -eu

REPO="karozadev/KzNginxGenerator"
BIN_NAME="kznginx"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

log() { printf '==> %s\n' "$1"; }
err() { printf 'error: %s\n' "$1" >&2; exit 1; }

detect_os() {
  os=$(uname -s)
  case "$os" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *) err "unsupported OS: $os (only Linux and macOS are supported)" ;;
  esac
}

detect_arch() {
  arch=$(uname -m)
  case "$arch" in
    x86_64 | amd64) echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *) err "unsupported architecture: $arch" ;;
  esac
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd uname

  os=$(detect_os)
  arch=$(detect_arch)

  log "Fetching latest release information for ${REPO}..."
  latest_tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

  [ -n "$latest_tag" ] || err "could not determine the latest release tag"

  version="${latest_tag#v}"
  archive="${BIN_NAME}_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${latest_tag}/${archive}"

  log "Downloading ${BIN_NAME} ${latest_tag} for ${os}/${arch}..."
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  curl -fsSL "$url" -o "${tmpdir}/${archive}" \
    || err "download failed: ${url}"

  log "Extracting archive..."
  tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"

  [ -f "${tmpdir}/${BIN_NAME}" ] || err "binary '${BIN_NAME}' not found in downloaded archive"

  chmod +x "${tmpdir}/${BIN_NAME}"

  log "Installing to ${INSTALL_DIR}/${BIN_NAME}..."
  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmpdir}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
  else
    need_cmd sudo
    sudo mv "${tmpdir}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
  fi

  log "KzNginxGenerator ${latest_tag} installed successfully!"
  log "Run '${BIN_NAME} version' or '${BIN_NAME} ui' to get started."
}

main "$@"
