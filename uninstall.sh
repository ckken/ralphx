#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_BIN_DIR="${HOME}/.local/bin"
BIN_DIR="${PREFIX:-$DEFAULT_BIN_DIR}"
BUILD_DIR="$SCRIPT_DIR/bin"
COMMAND_NAME="ralphx"
DOCTOR_NAME="ralphx-doctor"
TARGET_BINARY="$BUILD_DIR/$COMMAND_NAME"
DOCTOR_BINARY="$BUILD_DIR/$DOCTOR_NAME"
INSTALL_PATH="$BIN_DIR/$COMMAND_NAME"
DOCTOR_INSTALL_PATH="$BIN_DIR/$DOCTOR_NAME"

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

info() {
  printf '[uninstall] %s\n' "$*"
}

remove_wrapper() {
  local install_path="$1"
  local target_binary="$2"

  if [[ -L "$install_path" ]]; then
    rm -f "$install_path"
    info "Removed symlink: $install_path"
    return 0
  fi

  if [[ -f "$install_path" ]]; then
    if grep -qF "$target_binary" "$install_path" 2>/dev/null; then
      rm -f "$install_path"
      info "Removed wrapper: $install_path"
      return 0
    fi
    die "Refusing to remove non-ralphx file: $install_path"
  fi

  info "Nothing to uninstall: $install_path"
}

main() {
  remove_wrapper "$INSTALL_PATH" "$TARGET_BINARY"
  remove_wrapper "$DOCTOR_INSTALL_PATH" "$DOCTOR_BINARY"
}

main "$@"
