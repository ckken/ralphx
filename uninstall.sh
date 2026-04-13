#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="${PREFIX:-${HOME}/.local/bin}"
DATA_DIR="${RALPHX_DATA_DIR:-${HOME}/.local/share/ralphx}"
CONFIG_DIR="${RALPHX_CONFIG_DIR:-${HOME}/.config/ralphx}"

info() {
  printf '[uninstall] %s\n' "$*"
}

remove_if_exists() {
  local path="$1"
  if [[ -e "$path" || -L "$path" ]]; then
    rm -rf "$path"
    info "Removed $path"
  else
    info "Nothing to remove: $path"
  fi
}

main() {
  remove_if_exists "$BIN_DIR/ralphx"
  remove_if_exists "$BIN_DIR/ralphx-doctor"
  remove_if_exists "$CONFIG_DIR/current.env"
  remove_if_exists "$DATA_DIR/releases"
}

main "$@"
